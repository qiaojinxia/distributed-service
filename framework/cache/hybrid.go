package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type HybridCache struct {
	l1Cache        Cache
	l2Cache        Cache
	config         HybridConfig
	writeBackQueue chan *writeBackItem
	stats          HybridStats
	mutex          sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

type HybridConfig struct {
	L1Config           Config        `json:"l1_config" yaml:"l1_config"`
	L2Config           Config        `json:"l2_config" yaml:"l2_config"`
	SyncStrategy       SyncStrategy  `json:"sync_strategy" yaml:"sync_strategy"`
	WriteBackEnabled   bool          `json:"write_back_enabled" yaml:"write_back_enabled"`
	WriteBackInterval  time.Duration `json:"write_back_interval" yaml:"write_back_interval"`
	WriteBackBatchSize int           `json:"write_back_batch_size" yaml:"write_back_batch_size"`
	L1TTL              time.Duration `json:"l1_ttl" yaml:"l1_ttl"`
	L2TTL              time.Duration `json:"l2_ttl" yaml:"l2_ttl"`
}


type writeBackItem struct {
	key        string
	value      interface{}
	expiration time.Duration
	timestamp  time.Time
}

type HybridStats struct {
	L1Hits      int64
	L1Misses    int64
	L2Hits      int64
	L2Misses    int64
	L1Sets      int64
	L2Sets      int64
	Writebacks  int64
	Errors      int64
	LastUpdated time.Time
}

func NewHybridCache(config HybridConfig, manager *Manager) (*HybridCache, error) {
	// 创建L1缓存（通常是内存缓存）
	var l1Cache Cache
	var err error
	
	if config.L1Config.Type == TypeMemory {
		l1Builder := &MemoryBuilder{}
		l1Cache, err = l1Builder.Build(config.L1Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create L1 cache: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported L1 cache type: %s", config.L1Config.Type)
	}

	// 创建L2缓存（通常是Redis缓存）
	var l2Cache Cache
	
	if config.L2Config.Type == TypeRedis {
		// 从管理器中获取Redis构建器
		if err := manager.CreateCache(config.L2Config); err != nil {
			return nil, fmt.Errorf("failed to create L2 cache config: %w", err)
		}
		
		l2Cache, err = manager.GetCache(config.L2Config.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get L2 cache: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported L2 cache type: %s", config.L2Config.Type)
	}

	// 设置默认值
	if config.WriteBackInterval == 0 {
		config.WriteBackInterval = time.Minute * 5
	}
	if config.WriteBackBatchSize == 0 {
		config.WriteBackBatchSize = 100
	}
	if config.L1TTL == 0 {
		config.L1TTL = time.Hour
	}
	if config.L2TTL == 0 {
		config.L2TTL = time.Hour * 24
	}

	hc := &HybridCache{
		l1Cache:        l1Cache,
		l2Cache:        l2Cache,
		config:         config,
		writeBackQueue: make(chan *writeBackItem, 1000),
		stats:          HybridStats{LastUpdated: time.Now()},
		stopChan:       make(chan struct{}),
	}

	// 启动写回工作协程
	if config.WriteBackEnabled && config.SyncStrategy == SyncStrategyWriteBack {
		hc.wg.Add(1)
		go hc.writeBackWorker()
	}

	return hc, nil
}

func (h *HybridCache) Get(ctx context.Context, key string) (interface{}, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 先从L1缓存查询
	value, err := h.l1Cache.Get(ctx, key)
	if err == nil {
		h.stats.L1Hits++
		return value, nil
	}

	h.stats.L1Misses++

	// L1未命中，从L2缓存查询
	value, err = h.l2Cache.Get(ctx, key)
	if err != nil {
		h.stats.L2Misses++
		if errors.Is(err, ErrKeyNotFound) {
			return nil, ErrKeyNotFound
		}
		h.stats.Errors++
		return nil, err
	}

	h.stats.L2Hits++

	// 将L2的数据写入L1（缓存升级）
	if setErr := h.l1Cache.Set(ctx, key, value, h.config.L1TTL); setErr != nil {
		// 记录错误但不影响返回结果
		h.stats.Errors++
	}

	return value, nil
}

func (h *HybridCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	switch h.config.SyncStrategy {
	case SyncStrategyWriteThrough:
		return h.writeThrough(ctx, key, value, expiration)
	case SyncStrategyWriteBack:
		return h.writeBack(ctx, key, value, expiration)
	case SyncStrategyWriteAround:
		return h.writeAround(ctx, key, value, expiration)
	default:
		return h.writeThrough(ctx, key, value, expiration)
	}
}

func (h *HybridCache) writeThrough(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 同时写入L1和L2
	l1TTL := h.config.L1TTL
	if expiration > 0 && expiration < l1TTL {
		l1TTL = expiration
	}

	l2TTL := h.config.L2TTL
	if expiration > 0 && expiration < l2TTL {
		l2TTL = expiration
	}

	// 写入L1
	if err := h.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
		h.stats.Errors++
		return fmt.Errorf("failed to set L1 cache: %w", err)
	}
	h.stats.L1Sets++

	// 写入L2
	if err := h.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
		h.stats.Errors++
		return fmt.Errorf("failed to set L2 cache: %w", err)
	}
	h.stats.L2Sets++

	return nil
}

func (h *HybridCache) writeBack(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 先写入L1
	l1TTL := h.config.L1TTL
	if expiration > 0 && expiration < l1TTL {
		l1TTL = expiration
	}

	if err := h.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
		h.stats.Errors++
		return fmt.Errorf("failed to set L1 cache: %w", err)
	}
	h.stats.L1Sets++

	// 加入写回队列
	if h.config.WriteBackEnabled {
		l2TTL := h.config.L2TTL
		if expiration > 0 && expiration < l2TTL {
			l2TTL = expiration
		}

		select {
		case h.writeBackQueue <- &writeBackItem{
			key:        key,
			value:      value,
			expiration: l2TTL,
			timestamp:  time.Now(),
		}:
		default:
			// 队列满了，直接写入L2
			if err := h.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
				h.stats.Errors++
			} else {
				h.stats.L2Sets++
			}
		}
	}

	return nil
}

func (h *HybridCache) writeAround(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 只写入L2，不写L1
	l2TTL := h.config.L2TTL
	if expiration > 0 && expiration < l2TTL {
		l2TTL = expiration
	}

	if err := h.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
		h.stats.Errors++
		return fmt.Errorf("failed to set L2 cache: %w", err)
	}
	h.stats.L2Sets++

	return nil
}

func (h *HybridCache) Delete(ctx context.Context, key string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var lastErr error

	// 从L1删除
	if err := h.l1Cache.Delete(ctx, key); err != nil {
		h.stats.Errors++
		lastErr = err
	}

	// 从L2删除
	if err := h.l2Cache.Delete(ctx, key); err != nil {
		h.stats.Errors++
		lastErr = err
	}

	return lastErr
}

func (h *HybridCache) Exists(ctx context.Context, key string) (bool, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 先检查L1
	if exists, err := h.l1Cache.Exists(ctx, key); err == nil && exists {
		return true, nil
	}

	// 再检查L2
	return h.l2Cache.Exists(ctx, key)
}

func (h *HybridCache) Clear(ctx context.Context) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var lastErr error

	if err := h.l1Cache.Clear(ctx); err != nil {
		h.stats.Errors++
		lastErr = err
	}

	if err := h.l2Cache.Clear(ctx); err != nil {
		h.stats.Errors++
		lastErr = err
	}

	return lastErr
}

func (h *HybridCache) Close() error {
	close(h.stopChan)
	h.wg.Wait()

	var lastErr error

	if err := h.l1Cache.Close(); err != nil {
		lastErr = err
	}

	if err := h.l2Cache.Close(); err != nil {
		lastErr = err
	}

	return lastErr
}

func (h *HybridCache) GetStats() HybridStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.stats
}

func (h *HybridCache) ResetStats() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.stats = HybridStats{LastUpdated: time.Now()}
}

func (h *HybridCache) writeBackWorker() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.config.WriteBackInterval)
	defer ticker.Stop()

	batch := make([]*writeBackItem, 0, h.config.WriteBackBatchSize)

	for {
		select {
		case <-h.stopChan:
			// 处理剩余的写回任务
			h.flushWriteBackBatch(batch)
			return

		case item := <-h.writeBackQueue:
			batch = append(batch, item)
			if len(batch) >= h.config.WriteBackBatchSize {
				h.flushWriteBackBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				h.flushWriteBackBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (h *HybridCache) flushWriteBackBatch(batch []*writeBackItem) {
	ctx := context.Background()

	for _, item := range batch {
		if err := h.l2Cache.Set(ctx, item.key, item.value, item.expiration); err != nil {
			h.mutex.Lock()
			h.stats.Errors++
			h.mutex.Unlock()
		} else {
			h.mutex.Lock()
			h.stats.L2Sets++
			h.stats.Writebacks++
			h.mutex.Unlock()
		}
	}
}

// MGet 支持批量操作的混合缓存
func (h *HybridCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	missedKeys := make([]string, 0)

	// 先从L1获取
	if batchL1, ok := h.l1Cache.(BatchCache); ok {
		l1Result, err := batchL1.MGet(ctx, keys)
		if err == nil {
			for key, value := range l1Result {
				result[key] = value
				h.stats.L1Hits++
			}

			// 找出L1未命中的key
			for _, key := range keys {
				if _, exists := l1Result[key]; !exists {
					missedKeys = append(missedKeys, key)
					h.stats.L1Misses++
				}
			}
		} else {
			missedKeys = keys
			h.stats.L1Misses += int64(len(keys))
		}
	} else {
		// L1不支持批量操作，逐个获取
		for _, key := range keys {
			value, err := h.l1Cache.Get(ctx, key)
			if err == nil {
				result[key] = value
				h.stats.L1Hits++
			} else {
				missedKeys = append(missedKeys, key)
				h.stats.L1Misses++
			}
		}
	}

	// 从L2获取未命中的key
	if len(missedKeys) > 0 {
		if batchL2, ok := h.l2Cache.(BatchCache); ok {
			l2Result, err := batchL2.MGet(ctx, missedKeys)
			if err == nil {
				for key, value := range l2Result {
					result[key] = value
					h.stats.L2Hits++

					// 写回L1
					_ = h.l1Cache.Set(ctx, key, value, h.config.L1TTL)
				}

				// 统计L2未命中
				for _, key := range missedKeys {
					if _, exists := l2Result[key]; !exists {
						h.stats.L2Misses++
					}
				}
			} else {
				h.stats.L2Misses += int64(len(missedKeys))
				h.stats.Errors++
			}
		} else {
			// L2不支持批量操作，逐个获取
			for _, key := range missedKeys {
				value, err := h.l2Cache.Get(ctx, key)
				if err == nil {
					result[key] = value
					h.stats.L2Hits++

					// 写回L1
					_ = h.l1Cache.Set(ctx, key, value, h.config.L1TTL)
				} else {
					h.stats.L2Misses++
					if !errors.Is(err, ErrKeyNotFound) {
						h.stats.Errors++
					}
				}
			}
		}
	}

	return result, nil
}

type HybridBuilder struct {
	manager *Manager
}

func NewHybridBuilder(manager *Manager) *HybridBuilder {
	return &HybridBuilder{manager: manager}
}

func (b *HybridBuilder) Build(config Config) (Cache, error) {
	hybridConfig := HybridConfig{
		SyncStrategy:       SyncStrategyWriteThrough,
		WriteBackEnabled:   true,
		WriteBackInterval:  time.Minute * 5,
		WriteBackBatchSize: 100,
		L1TTL:              time.Hour,
		L2TTL:              time.Hour * 24,
	}

	if settings := config.Settings; settings != nil {
		if data, err := json.Marshal(settings); err == nil {
			_ = json.Unmarshal(data, &hybridConfig)
		}
	}

	return NewHybridCache(hybridConfig, b.manager)
}
