package idgen

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// GormLeafIDGenerator 基于GORM的美团Leaf分布式ID生成器
type GormLeafIDGenerator struct {
	dao           LeafDAO        // 数据访问对象
	bufferMap     sync.Map       // 业务标识到SegmentBuffer的映射
	isPreloading  sync.Map       // 记录哪些业务标识正在预加载
	initMutexes   sync.Map       // 业务标识初始化互斥锁
	config        *LeafConfig    // 配置信息
	metrics       sync.Map       // 业务标识到指标的映射
	cleanupTicker *time.Ticker   // 清理定时器
	stopChan      chan struct{}  // 停止信号
	wg            sync.WaitGroup // 等待组
	tableCreated  int32          // 表是否已创建的原子标记
}

// LeafConfig Leaf配置
type LeafConfig struct {
	DefaultStep      int32         `json:"default_step"`      // 默认步长
	PreloadThreshold float64       `json:"preload_threshold"` // 预加载阈值
	CleanupInterval  time.Duration `json:"cleanup_interval"`  // 清理间隔
	MaxStepSize      int32         `json:"max_step_size"`     // 最大步长
	MinStepSize      int32         `json:"min_step_size"`     // 最小步长
	StepAdjustRatio  float64       `json:"step_adjust_ratio"` // 步长调整比例
}

// DefaultLeafConfig 默认配置
func DefaultLeafConfig() *LeafConfig {
	return &LeafConfig{
		DefaultStep:      1000,
		PreloadThreshold: 0.9,
		CleanupInterval:  time.Hour,
		MaxStepSize:      100000,
		MinStepSize:      100,
		StepAdjustRatio:  2.0,
	}
}

// NewGormLeafIDGenerator 创建新的GORM Leaf ID生成器
func NewGormLeafIDGenerator(db *gorm.DB, config *LeafConfig) *GormLeafIDGenerator {
	if config == nil {
		config = DefaultLeafConfig()
	}

	dao := NewGormLeafDAO(db)

	generator := &GormLeafIDGenerator{
		dao:      dao,
		config:   config,
		stopChan: make(chan struct{}),
	}

	// 启动清理协程
	generator.startCleanupWorker()

	return generator
}

// NextID 获取下一个ID
func (g *GormLeafIDGenerator) NextID(ctx context.Context, bizTag string) (int64, error) {
	// 获取或创建SegmentBuffer（这会确保metrics对象存在）
	buffer, err := g.getOrCreateBuffer(ctx, bizTag)
	if err != nil {
		g.updateMetrics(bizTag, func(m *LeafMetrics) {
			m.IncTotalRequests()
			m.IncFailedRequests()
		})
		return 0, fmt.Errorf("failed to get buffer for bizTag %s: %w", bizTag, err)
	}

	// 更新指标（在buffer创建后）
	g.updateMetrics(bizTag, func(m *LeafMetrics) {
		m.IncTotalRequests()
	})

	// 从buffer获取ID
	id, err := g.getIDFromBuffer(ctx, buffer)
	if err != nil {
		g.updateMetrics(bizTag, func(m *LeafMetrics) {
			m.IncFailedRequests()
		})
		return 0, err
	}

	g.updateMetrics(bizTag, func(m *LeafMetrics) {
		m.IncSuccessRequests()
	})

	return id, nil
}

// BatchNextID 批量获取ID
func (g *GormLeafIDGenerator) BatchNextID(ctx context.Context, bizTag string, count int) ([]int64, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		id, err := g.NextID(ctx, bizTag)
		if err != nil {
			return nil, fmt.Errorf("failed to get ID at index %d: %w", i, err)
		}
		ids[i] = id
	}

	return ids, nil
}

// getOrCreateBuffer 获取或创建SegmentBuffer
func (g *GormLeafIDGenerator) getOrCreateBuffer(ctx context.Context, bizTag string) (*SegmentBuffer, error) {
	// 先尝试从缓存获取
	if buffer, exists := g.bufferMap.Load(bizTag); exists {
		return buffer.(*SegmentBuffer), nil
	}

	// 确保指标对象存在（在任何操作之前）
	g.metrics.LoadOrStore(bizTag, NewLeafMetrics())

	// 获取或创建该业务标识的互斥锁
	mutexInterface, _ := g.initMutexes.LoadOrStore(bizTag, &sync.Mutex{})
	mutex := mutexInterface.(*sync.Mutex)

	mutex.Lock()
	defer mutex.Unlock()

	// 再次检查缓存（双重检查锁定）
	if buffer, exists := g.bufferMap.Load(bizTag); exists {
		return buffer.(*SegmentBuffer), nil
	}

	// 缓存中不存在，需要初始化
	return g.initBuffer(ctx, bizTag)
}

// initBuffer 初始化SegmentBuffer
func (g *GormLeafIDGenerator) initBuffer(ctx context.Context, bizTag string) (*SegmentBuffer, error) {
	// 确保表已创建
	if err := g.ensureTableCreated(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure table created: %w", err)
	}

	// 获取或创建叶子分配记录
	leafAlloc, err := g.dao.GetLeafAlloc(ctx, bizTag)
	if err != nil {
		if errors.Is(err, ErrBizTagNotFound) {
			// 自动创建新的业务标识
			description := fmt.Sprintf("Auto created for %s", bizTag)
			createErr := g.dao.CreateLeafAlloc(ctx, bizTag, g.config.DefaultStep, description)
			if createErr != nil {
				return nil, fmt.Errorf("failed to create leaf alloc: %w", createErr)
			}

			// 重新获取
			leafAlloc, err = g.dao.GetLeafAlloc(ctx, bizTag)
			if err != nil {
				return nil, fmt.Errorf("failed to get leaf alloc after creation: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get leaf alloc: %w", err)
		}
	}

	// 初始化指标（在加载segment之前）
	g.metrics.Store(bizTag, NewLeafMetrics())

	// 创建SegmentBuffer
	buffer := NewSegmentBuffer(bizTag, leafAlloc.Step)

	// 加载第一个号段
	segment, err := g.loadSegment(ctx, bizTag, leafAlloc.Step)
	if err != nil {
		return nil, fmt.Errorf("failed to load initial segment: %w", err)
	}

	buffer.mutex.Lock()
	buffer.Segments[0] = segment
	buffer.InitOK = true
	buffer.mutex.Unlock()

	// 存储到缓存
	g.bufferMap.Store(bizTag, buffer)

	return buffer, nil
}

// loadSegment 加载号段
func (g *GormLeafIDGenerator) loadSegment(ctx context.Context, bizTag string, step int32) (*LeafSegment, error) {
	// 从数据库获取新的号段
	leafAlloc, err := g.dao.UpdateMaxID(ctx, bizTag, step)
	if err != nil {
		return nil, fmt.Errorf("failed to update maxID ID: %w", err)
	}

	// 计算号段范围
	minID := leafAlloc.MaxID - int64(step) + 1
	maxID := leafAlloc.MaxID

	segment := NewLeafSegment(minID, maxID, step)

	g.updateMetrics(bizTag, func(m *LeafMetrics) {
		m.IncSegmentLoads()
	})

	return segment, nil
}

// getIDFromBuffer 从buffer获取ID
func (g *GormLeafIDGenerator) getIDFromBuffer(ctx context.Context, buffer *SegmentBuffer) (int64, error) {
	if !buffer.InitOK {
		return 0, ErrBufferNotReady
	}

	// 检查是否需要预加载
	if buffer.ShouldPreload(g.config.PreloadThreshold) {
		g.asyncPreloadNext(ctx, buffer)
	}

	// 尝试从当前号段获取ID
	current := buffer.Current()
	if current != nil && current.IsAvailable() {
		return current.GetCurrentID(), nil
	}

	// 当前号段用完，尝试切换到下一个号段
	if buffer.CanSwitchToNext() {
		buffer.SwitchPos()
		g.updateMetrics(buffer.Key, func(m *LeafMetrics) {
			m.IncBufferSwitches()
		})

		// 从新的当前号段获取ID
		current = buffer.Current()
		if current != nil && current.IsAvailable() {
			return current.GetCurrentID(), nil
		}
	}

	// 如果还是无法获取ID，同步加载号段
	return g.syncLoadAndGetID(ctx, buffer)
}

// asyncPreloadNext 异步预加载下一个号段
func (g *GormLeafIDGenerator) asyncPreloadNext(ctx context.Context, buffer *SegmentBuffer) {
	bizTag := buffer.Key

	// 检查是否已在预加载
	if _, isPreloading := g.isPreloading.LoadOrStore(bizTag, true); isPreloading {
		return
	}

	// 启动预加载协程
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer g.isPreloading.Delete(bizTag)

		// 使用新的context避免上下文问题
		newCtx := context.Background()

		// 确保表已创建
		if err := g.ensureTableCreated(newCtx); err != nil {
			log.Printf("Failed to ensure table created for preload: %v", err)
			return
		}

		// 动态调整步长
		newStep := g.adjustStep(buffer)

		// 加载新号段
		segment, err := g.loadSegment(newCtx, bizTag, newStep)
		if err != nil {
			log.Printf("Failed to preload segment for %s: %v", bizTag, err)
			return
		}

		// 设置到下一个位置
		buffer.SetNextSegment(segment)
		buffer.mutex.Lock()
		buffer.Step = newStep
		buffer.mutex.Unlock()
	}()
}

// syncLoadAndGetID 同步加载号段并获取ID
func (g *GormLeafIDGenerator) syncLoadAndGetID(ctx context.Context, buffer *SegmentBuffer) (int64, error) {
	// 动态调整步长
	newStep := g.adjustStep(buffer)

	// 同步加载号段
	segment, err := g.loadSegment(ctx, buffer.Key, newStep)
	if err != nil {
		return 0, fmt.Errorf("failed to sync load segment: %w", err)
	}

	// 更新当前号段
	buffer.mutex.Lock()
	buffer.Segments[buffer.CurrentPos] = segment
	buffer.Step = newStep
	buffer.mutex.Unlock()

	return segment.GetCurrentID(), nil
}

// adjustStep 动态调整步长
func (g *GormLeafIDGenerator) adjustStep(buffer *SegmentBuffer) int32 {
	current := buffer.Current()
	buffer.mutex.RLock()
	currentStep := buffer.Step
	buffer.mutex.RUnlock()

	if current == nil {
		return currentStep
	}

	// 根据消耗速度调整步长
	buffer.mutex.RLock()
	duration := time.Since(buffer.UpdateTime)
	buffer.mutex.RUnlock()

	consumptionRate := float64(current.Max-current.Min+1) / duration.Seconds()

	newStep := currentStep

	switch {
	case consumptionRate > 1000: // 高消耗率，增加步长
		newStep = int32(float64(newStep) * g.config.StepAdjustRatio)
		if newStep > g.config.MaxStepSize {
			newStep = g.config.MaxStepSize
		}
	case consumptionRate < 10: // 低消耗率，减少步长
		newStep = int32(float64(newStep) / g.config.StepAdjustRatio)
		if newStep < g.config.MinStepSize {
			newStep = g.config.MinStepSize
		}
	}

	return newStep
}

// updateMetrics 更新指标
func (g *GormLeafIDGenerator) updateMetrics(bizTag string, updateFunc func(*LeafMetrics)) {
	metricsInterface, _ := g.metrics.LoadOrStore(bizTag, NewLeafMetrics())
	metrics := metricsInterface.(*LeafMetrics)
	updateFunc(metrics)
}

// GetMetrics 获取指标
func (g *GormLeafIDGenerator) GetMetrics(bizTag string) *LeafMetrics {
	if metricsInterface, exists := g.metrics.Load(bizTag); exists {
		original := metricsInterface.(*LeafMetrics)

		// 创建一个副本，使用原子操作读取当前值
		metrics := &LeafMetrics{
			TotalRequests:   atomic.LoadInt64(&original.TotalRequests),
			SuccessRequests: atomic.LoadInt64(&original.SuccessRequests),
			FailedRequests:  atomic.LoadInt64(&original.FailedRequests),
			SegmentLoads:    atomic.LoadInt64(&original.SegmentLoads),
			BufferSwitches:  atomic.LoadInt64(&original.BufferSwitches),
			LastUpdateTime:  original.LastUpdateTime,
		}

		metrics.CalculateQPS()
		return metrics
	}
	return NewLeafMetrics()
}

// GetAllMetrics 获取所有指标
func (g *GormLeafIDGenerator) GetAllMetrics() map[string]*LeafMetrics {
	result := make(map[string]*LeafMetrics)

	g.metrics.Range(func(key, value interface{}) bool {
		bizTag := key.(string)
		metrics := value.(*LeafMetrics)
		metrics.CalculateQPS()
		result[bizTag] = metrics
		return true
	})

	return result
}

// CreateBizTag 创建业务标识
func (g *GormLeafIDGenerator) CreateBizTag(ctx context.Context, bizTag string, step int32, description string) error {
	if step <= 0 {
		step = g.config.DefaultStep
	}

	return g.dao.CreateLeafAlloc(ctx, bizTag, step, description)
}

// UpdateStep 更新步长
func (g *GormLeafIDGenerator) UpdateStep(ctx context.Context, bizTag string, newStep int32) error {
	if newStep < g.config.MinStepSize || newStep > g.config.MaxStepSize {
		return fmt.Errorf("step size %d is out of range [%d, %d]",
			newStep, g.config.MinStepSize, g.config.MaxStepSize)
	}

	err := g.dao.UpdateStep(ctx, bizTag, newStep)
	if err != nil {
		return err
	}

	// 更新内存中的buffer
	if bufferInterface, exists := g.bufferMap.Load(bizTag); exists {
		buffer := bufferInterface.(*SegmentBuffer)
		buffer.mutex.Lock()
		buffer.Step = newStep
		buffer.mutex.Unlock()
	}

	return nil
}

// DeleteBizTag 删除业务标识
func (g *GormLeafIDGenerator) DeleteBizTag(ctx context.Context, bizTag string) error {
	err := g.dao.DeleteLeafAlloc(ctx, bizTag)
	if err != nil {
		return err
	}

	// 从内存中清理
	g.bufferMap.Delete(bizTag)
	g.metrics.Delete(bizTag)
	g.isPreloading.Delete(bizTag)

	return nil
}

// GetBufferStatus 获取buffer状态
func (g *GormLeafIDGenerator) GetBufferStatus(bizTag string) map[string]interface{} {
	status := make(map[string]interface{})

	if bufferInterface, exists := g.bufferMap.Load(bizTag); exists {
		buffer := bufferInterface.(*SegmentBuffer)

		current := buffer.Current()
		next := buffer.Next()

		buffer.mutex.RLock()
		status["biz_tag"] = bizTag
		status["current_pos"] = buffer.CurrentPos
		status["next_ready"] = buffer.NextReady
		status["init_ok"] = buffer.InitOK
		status["step"] = buffer.Step
		status["update_time"] = buffer.UpdateTime
		buffer.mutex.RUnlock()

		if current != nil {
			status["current_segment"] = map[string]interface{}{
				"min":         current.Min,
				"max":         current.Max,
				"cursor":      current.Cursor,
				"usage_ratio": current.UsageRatio(),
				"remaining":   current.Remaining(),
			}
		}

		if next != nil {
			status["next_segment"] = map[string]interface{}{
				"min":         next.Min,
				"max":         next.Max,
				"cursor":      next.Cursor,
				"usage_ratio": next.UsageRatio(),
				"remaining":   next.Remaining(),
			}
		}
	}

	return status
}

// startCleanupWorker 启动清理工作协程
func (g *GormLeafIDGenerator) startCleanupWorker() {
	g.cleanupTicker = time.NewTicker(g.config.CleanupInterval)

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		for {
			select {
			case <-g.cleanupTicker.C:
				g.cleanup()
			case <-g.stopChan:
				return
			}
		}
	}()
}

// cleanup 清理不活跃的buffer和指标
func (g *GormLeafIDGenerator) cleanup() {
	cutoffTime := time.Now().Add(-g.config.CleanupInterval * 2)

	// 清理不活跃的buffer
	g.bufferMap.Range(func(key, value interface{}) bool {
		bizTag := key.(string)
		buffer := value.(*SegmentBuffer)

		if buffer.UpdateTime.Before(cutoffTime) {
			g.bufferMap.Delete(bizTag)
			g.metrics.Delete(bizTag)
			g.isPreloading.Delete(bizTag)
		}

		return true
	})
}

// Close 关闭生成器
func (g *GormLeafIDGenerator) Close() error {
	close(g.stopChan)

	if g.cleanupTicker != nil {
		g.cleanupTicker.Stop()
	}

	g.wg.Wait()
	return nil
}

// ensureTableCreated 确保表已创建（线程安全）
func (g *GormLeafIDGenerator) ensureTableCreated(ctx context.Context) error {
	if atomic.LoadInt32(&g.tableCreated) == 1 {
		return nil
	}

	// 直接调用创建表，DAO方法内部会检查表是否存在
	err := g.dao.CreateTable(ctx)
	if err == nil {
		atomic.StoreInt32(&g.tableCreated, 1)
	}
	return err
}

// CreateTable 创建表
func (g *GormLeafIDGenerator) CreateTable(ctx context.Context) error {
	err := g.dao.CreateTable(ctx)
	if err == nil {
		atomic.StoreInt32(&g.tableCreated, 1)
	}
	return err
}
