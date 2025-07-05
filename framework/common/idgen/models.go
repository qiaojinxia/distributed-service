package idgen

import (
	"sync"
	"sync/atomic"
	"time"
)

// LeafAlloc 叶子分配表模型
type LeafAlloc struct {
	BizTag      string    `gorm:"primaryKey;column:biz_tag;type:varchar(128);not null;comment:业务标识" json:"biz_tag"`
	MaxID       int64     `gorm:"column:max_id;type:bigint;not null;default:1;comment:当前最大ID" json:"max_id"`
	Step        int32     `gorm:"column:step;type:int;not null;default:1000;comment:步长" json:"step"`
	Description string    `gorm:"column:description;type:varchar(256);comment:描述" json:"description"`
	UpdateTime  time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间" json:"update_time"`
	AutoClean   int8      `gorm:"column:auto_clean;type:tinyint;not null;default:0;comment:自动清理标识" json:"auto_clean"`
}

// TableName 指定表名
func (LeafAlloc) TableName() string {
	return "leaf_alloc"
}

// LeafSegment 内存中的号段结构
type LeafSegment struct {
	Min    int64 `json:"min"`     // 号段最小值
	Max    int64 `json:"max"`     // 号段最大值
	Cursor int64 `json:"cursor"`  // 当前游标位置
	Step   int32 `json:"step"`    // 步长
	InitOK bool  `json:"init_ok"` // 是否初始化成功
}

// NewLeafSegment 创建新的号段
func NewLeafSegment(min, max int64, step int32) *LeafSegment {
	return &LeafSegment{
		Min:    min,
		Max:    max,
		Cursor: min - 1, // 游标从min-1开始，首次使用时会递增到min
		Step:   step,
		InitOK: true,
	}
}

// GetCurrentID 获取当前ID并递增游标（线程安全）
func (s *LeafSegment) GetCurrentID() int64 {
	// 使用原子操作确保线程安全
	return atomic.AddInt64(&s.Cursor, 1)
}

// IsAvailable 检查号段是否可用
func (s *LeafSegment) IsAvailable() bool {
	return s.InitOK && atomic.LoadInt64(&s.Cursor) < s.Max
}

// UsageRatio 计算使用率
func (s *LeafSegment) UsageRatio() float64 {
	if s.Max == s.Min {
		return 1.0
	}
	cursor := atomic.LoadInt64(&s.Cursor)
	used := cursor - s.Min + 1
	total := s.Max - s.Min + 1
	return float64(used) / float64(total)
}

// Remaining 剩余ID数量
func (s *LeafSegment) Remaining() int64 {
	return s.Max - atomic.LoadInt64(&s.Cursor)
}

// IsNearlyExhausted 检查是否即将用完（使用率超过90%）
func (s *LeafSegment) IsNearlyExhausted() bool {
	return s.UsageRatio() >= 0.9
}

// SegmentBuffer 双缓冲区结构
type SegmentBuffer struct {
	Key        string          `json:"key"`         // 业务标识
	Segments   [2]*LeafSegment `json:"segments"`    // 双缓冲区
	CurrentPos int             `json:"current_pos"` // 当前使用的缓冲区位置(0或1)
	NextReady  bool            `json:"next_ready"`  // 下一个缓冲区是否准备好
	InitOK     bool            `json:"init_ok"`     // 是否初始化成功
	Step       int32           `json:"step"`        // 当前步长
	UpdateTime time.Time       `json:"update_time"` // 最后更新时间
	mutex      sync.RWMutex    // 读写锁
}

// NewSegmentBuffer 创建新的号段缓冲区
func NewSegmentBuffer(key string, step int32) *SegmentBuffer {
	return &SegmentBuffer{
		Key:        key,
		Segments:   [2]*LeafSegment{nil, nil},
		CurrentPos: 0,
		NextReady:  false,
		InitOK:     false,
		Step:       step,
		UpdateTime: time.Now(),
	}
}

// Current 获取当前使用的号段
func (sb *SegmentBuffer) Current() *LeafSegment {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.Segments[sb.CurrentPos]
}

// Next 获取下一个号段
func (sb *SegmentBuffer) Next() *LeafSegment {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.Segments[1-sb.CurrentPos]
}

// SwitchPos 切换到下一个缓冲区
func (sb *SegmentBuffer) SwitchPos() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.CurrentPos = 1 - sb.CurrentPos
	sb.NextReady = false
	sb.UpdateTime = time.Now()
}

// SetNextSegment 设置下一个号段
func (sb *SegmentBuffer) SetNextSegment(segment *LeafSegment) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	nextPos := 1 - sb.CurrentPos
	sb.Segments[nextPos] = segment
	sb.NextReady = true
}

// GetID 从当前号段获取ID
func (sb *SegmentBuffer) GetID() (int64, error) {
	current := sb.Current()
	if current == nil || !current.IsAvailable() {
		return 0, ErrSegmentNotAvailable
	}

	return current.GetCurrentID(), nil
}

// ShouldPreload 判断是否应该预加载下一个号段
func (sb *SegmentBuffer) ShouldPreload(threshold float64) bool {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	current := sb.Segments[sb.CurrentPos]
	if current == nil || !current.InitOK {
		return false
	}

	// 当前号段使用率超过阈值且下一个号段未准备好时触发预加载
	return current.UsageRatio() >= threshold && !sb.NextReady
}

// CanSwitchToNext 判断是否可以切换到下一个号段
func (sb *SegmentBuffer) CanSwitchToNext() bool {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	current := sb.Segments[sb.CurrentPos]
	// 当前号段用完且下一个号段已准备好时可以切换
	return (current == nil || !current.IsAvailable()) && sb.NextReady
}

// 预定义错误
var (
	ErrSegmentNotAvailable = NewLeafError("SEGMENT_NOT_AVAILABLE", "segment not available")
	ErrBufferNotReady      = NewLeafError("BUFFER_NOT_READY", "buffer not ready")
	ErrBizTagNotFound      = NewLeafError("BIZ_TAG_NOT_FOUND", "biz tag not found")
)

// LeafError 自定义错误类型
type LeafError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewLeafError(code, message string) *LeafError {
	return &LeafError{
		Code:    code,
		Message: message,
	}
}

func (e *LeafError) Error() string {
	return e.Message
}

// LeafMetrics 监控指标
type LeafMetrics struct {
	TotalRequests   int64        `json:"total_requests"`   // 总请求数
	SuccessRequests int64        `json:"success_requests"` // 成功请求数
	FailedRequests  int64        `json:"failed_requests"`  // 失败请求数
	SegmentLoads    int64        `json:"segment_loads"`    // 号段加载次数
	BufferSwitches  int64        `json:"buffer_switches"`  // 缓冲区切换次数
	AverageQPS      float64      `json:"average_qps"`      // 平均QPS
	LastUpdateTime  time.Time    `json:"last_update_time"` // 最后更新时间
	mutex           sync.RWMutex // 保护指标的读写锁
}

// NewLeafMetrics 创建新的指标对象
func NewLeafMetrics() *LeafMetrics {
	return &LeafMetrics{
		LastUpdateTime: time.Now(),
	}
}

// IncTotalRequests 增加总请求数
func (m *LeafMetrics) IncTotalRequests() {
	atomic.AddInt64(&m.TotalRequests, 1)
}

// IncSuccessRequests 增加成功请求数
func (m *LeafMetrics) IncSuccessRequests() {
	atomic.AddInt64(&m.SuccessRequests, 1)
}

// IncFailedRequests 增加失败请求数
func (m *LeafMetrics) IncFailedRequests() {
	atomic.AddInt64(&m.FailedRequests, 1)
}

// IncSegmentLoads 增加号段加载次数
func (m *LeafMetrics) IncSegmentLoads() {
	atomic.AddInt64(&m.SegmentLoads, 1)
}

// IncBufferSwitches 增加缓冲区切换次数
func (m *LeafMetrics) IncBufferSwitches() {
	atomic.AddInt64(&m.BufferSwitches, 1)
}

// CalculateQPS 计算QPS
func (m *LeafMetrics) CalculateQPS() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	duration := now.Sub(m.LastUpdateTime).Seconds()
	if duration > 0 {
		totalReqs := atomic.LoadInt64(&m.TotalRequests)
		m.AverageQPS = float64(totalReqs) / duration
	}
	m.LastUpdateTime = now
}

// SuccessRate 计算成功率
func (m *LeafMetrics) SuccessRate() float64 {
	totalReqs := atomic.LoadInt64(&m.TotalRequests)
	successReqs := atomic.LoadInt64(&m.SuccessRequests)

	if totalReqs == 0 {
		return 0
	}
	return float64(successReqs) / float64(totalReqs)
}
