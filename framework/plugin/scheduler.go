package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskScheduler 任务调度器接口
type TaskScheduler interface {
	// 调度管理
	ScheduleTask(task *Task) error
	CancelTask(taskID string) error
	PauseTask(taskID string) error
	ResumeTask(taskID string) error

	// 任务查询
	GetTask(taskID string) *Task
	GetAllTasks() map[string]*Task
	GetTasksByStatus(status TaskStatus) []*Task

	// 生命周期
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 事件
	SetEventHandler(handler TaskEventHandler)
}

// Task 定时任务
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schedule    *Schedule              `json:"schedule"`
	Handler     TaskHandler            `json:"-"`
	Status      TaskStatus             `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`

	// 执行统计
	CreatedAt    time.Time `json:"created_at"`
	LastRunAt    time.Time `json:"last_run_at,omitempty"`
	NextRunAt    time.Time `json:"next_run_at,omitempty"`
	RunCount     int64     `json:"run_count"`
	FailureCount int64     `json:"failure_count"`

	// 内部状态
	cancelFunc context.CancelFunc
	mu         sync.RWMutex
}

// Schedule 调度配置
type Schedule struct {
	Type     ScheduleType  `json:"type"`
	Cron     string        `json:"cron,omitempty"`     // cron表达式
	Interval time.Duration `json:"interval,omitempty"` // 间隔时间
	Delay    time.Duration `json:"delay,omitempty"`    // 延迟时间
	RunOnce  bool          `json:"run_once,omitempty"` // 是否只执行一次
	MaxRuns  int64         `json:"max_runs,omitempty"` // 最大执行次数
}

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusPaused
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCanceled
)

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "pending"
	case TaskStatusRunning:
		return "running"
	case TaskStatusPaused:
		return "paused"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusFailed:
		return "failed"
	case TaskStatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

// ScheduleType 调度类型
type ScheduleType int

const (
	ScheduleTypeCron ScheduleType = iota
	ScheduleTypeInterval
	ScheduleTypeOnce
)

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *Task) error

// TaskEventHandler 任务事件处理器
type TaskEventHandler func(event *TaskEvent)

// TaskEvent 任务事件
type TaskEvent struct {
	Type      TaskEventType          `json:"type"`
	TaskID    string                 `json:"task_id"`
	TaskName  string                 `json:"task_name"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     error                  `json:"error,omitempty"`
}

// TaskEventType 任务事件类型
type TaskEventType int

const (
	TaskEventScheduled TaskEventType = iota
	TaskEventStarted
	TaskEventCompleted
	TaskEventFailed
	TaskEventCanceled
	TaskEventPaused
	TaskEventResumed
)

func (t TaskEventType) String() string {
	switch t {
	case TaskEventScheduled:
		return "scheduled"
	case TaskEventStarted:
		return "started"
	case TaskEventCompleted:
		return "completed"
	case TaskEventFailed:
		return "failed"
	case TaskEventCanceled:
		return "canceled"
	case TaskEventPaused:
		return "paused"
	case TaskEventResumed:
		return "resumed"
	default:
		return "unknown"
	}
}

// DefaultTaskScheduler 默认任务调度器实现
type DefaultTaskScheduler struct {
	tasks        map[string]*Task
	running      bool
	eventHandler TaskEventHandler
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	logger       Logger
}

// NewDefaultTaskScheduler 创建默认任务调度器
func NewDefaultTaskScheduler() *DefaultTaskScheduler {
	return &DefaultTaskScheduler{
		tasks: make(map[string]*Task),
	}
}

// SetLogger 设置日志记录器
func (s *DefaultTaskScheduler) SetLogger(logger Logger) {
	s.logger = logger
}

// Start 启动调度器
func (s *DefaultTaskScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	if s.logger != nil {
		s.logger.Info("Task scheduler started")
	}

	return nil
}

// Stop 停止调度器
func (s *DefaultTaskScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// 取消所有任务
	for _, task := range s.tasks {
		if task.cancelFunc != nil {
			task.cancelFunc()
		}
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.running = false

	if s.logger != nil {
		s.logger.Info("Task scheduler stopped")
	}

	return nil
}

// IsRunning 检查是否运行中
func (s *DefaultTaskScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// ScheduleTask 调度任务
func (s *DefaultTaskScheduler) ScheduleTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	// 检查任务是否已存在
	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID '%s' already exists", task.ID)
	}

	// 设置任务状态
	task.Status = TaskStatusPending
	task.CreatedAt = time.Now()

	// 计算下次运行时间
	nextRun, err := s.calculateNextRun(task)
	if err != nil {
		return fmt.Errorf("failed to calculate next run time: %w", err)
	}
	task.NextRunAt = nextRun

	// 注册任务
	s.tasks[task.ID] = task

	// 启动任务
	go s.runTask(task)

	// 发送事件
	s.sendEvent(&TaskEvent{
		Type:      TaskEventScheduled,
		TaskID:    task.ID,
		TaskName:  task.Name,
		Timestamp: time.Now(),
	})

	if s.logger != nil {
		s.logger.Info("Task scheduled", "id", task.ID, "name", task.Name)
	}

	return nil
}

// CancelTask 取消任务
func (s *DefaultTaskScheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID '%s' not found", taskID)
	}

	task.mu.Lock()
	if task.cancelFunc != nil {
		task.cancelFunc()
	}
	task.Status = TaskStatusCanceled
	task.mu.Unlock()

	delete(s.tasks, taskID)

	// 发送事件
	s.sendEvent(&TaskEvent{
		Type:      TaskEventCanceled,
		TaskID:    task.ID,
		TaskName:  task.Name,
		Timestamp: time.Now(),
	})

	if s.logger != nil {
		s.logger.Info("Task canceled", "id", taskID)
	}

	return nil
}

// PauseTask 暂停任务
func (s *DefaultTaskScheduler) PauseTask(taskID string) error {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task with ID '%s' not found", taskID)
	}

	task.mu.Lock()
	if task.Status == TaskStatusRunning {
		task.Status = TaskStatusPaused
	}
	task.mu.Unlock()

	// 发送事件
	s.sendEvent(&TaskEvent{
		Type:      TaskEventPaused,
		TaskID:    task.ID,
		TaskName:  task.Name,
		Timestamp: time.Now(),
	})

	return nil
}

// ResumeTask 恢复任务
func (s *DefaultTaskScheduler) ResumeTask(taskID string) error {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task with ID '%s' not found", taskID)
	}

	task.mu.Lock()
	if task.Status == TaskStatusPaused {
		task.Status = TaskStatusPending
	}
	task.mu.Unlock()

	// 发送事件
	s.sendEvent(&TaskEvent{
		Type:      TaskEventResumed,
		TaskID:    task.ID,
		TaskName:  task.Name,
		Timestamp: time.Now(),
	})

	return nil
}

// GetTask 获取任务
func (s *DefaultTaskScheduler) GetTask(taskID string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, exists := s.tasks[taskID]; exists {
		// 返回副本
		taskCopy := *task
		return &taskCopy
	}
	return nil
}

// GetAllTasks 获取所有任务
func (s *DefaultTaskScheduler) GetAllTasks() map[string]*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*Task)
	for id, task := range s.tasks {
		taskCopy := *task
		result[id] = &taskCopy
	}
	return result
}

// GetTasksByStatus 根据状态获取任务
func (s *DefaultTaskScheduler) GetTasksByStatus(status TaskStatus) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, task := range s.tasks {
		task.mu.RLock()
		if task.Status == status {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
		task.mu.RUnlock()
	}
	return result
}

// SetEventHandler 设置事件处理器
func (s *DefaultTaskScheduler) SetEventHandler(handler TaskEventHandler) {
	s.eventHandler = handler
}

// 私有方法

// runTask 运行任务
func (s *DefaultTaskScheduler) runTask(task *Task) {
	taskCtx, taskCancel := context.WithCancel(s.ctx)
	task.cancelFunc = taskCancel

	for {
		select {
		case <-taskCtx.Done():
			return
		case <-time.After(time.Until(task.NextRunAt)):
			task.mu.RLock()
			if task.Status == TaskStatusPaused {
				task.mu.RUnlock()
				time.Sleep(time.Second) // 暂停时短暂等待
				continue
			}
			task.mu.RUnlock()

			// 执行任务
			s.executeTask(taskCtx, task)

			// 计算下次运行时间
			if task.Schedule.RunOnce ||
				(task.Schedule.MaxRuns > 0 && task.RunCount >= task.Schedule.MaxRuns) {
				task.mu.Lock()
				task.Status = TaskStatusCompleted
				task.mu.Unlock()
				return
			}

			nextRun, err := s.calculateNextRun(task)
			if err != nil {
				if s.logger != nil {
					s.logger.Error("Failed to calculate next run", "task", task.ID, "error", err)
				}
				task.mu.Lock()
				task.Status = TaskStatusFailed
				task.mu.Unlock()
				return
			}

			task.mu.Lock()
			task.NextRunAt = nextRun
			task.mu.Unlock()
		}
	}
}

// executeTask 执行任务
func (s *DefaultTaskScheduler) executeTask(ctx context.Context, task *Task) {
	task.mu.Lock()
	task.Status = TaskStatusRunning
	task.LastRunAt = time.Now()
	task.RunCount++
	task.mu.Unlock()

	// 发送开始事件
	s.sendEvent(&TaskEvent{
		Type:      TaskEventStarted,
		TaskID:    task.ID,
		TaskName:  task.Name,
		Timestamp: time.Now(),
	})

	// 执行任务处理器
	err := task.Handler(ctx, task)

	if err != nil {
		task.mu.Lock()
		task.Status = TaskStatusPending
		task.FailureCount++
		task.mu.Unlock()

		// 发送失败事件
		s.sendEvent(&TaskEvent{
			Type:      TaskEventFailed,
			TaskID:    task.ID,
			TaskName:  task.Name,
			Timestamp: time.Now(),
			Error:     err,
		})

		if s.logger != nil {
			s.logger.Error("Task execution failed", "task", task.ID, "error", err)
		}
	} else {
		task.mu.Lock()
		task.Status = TaskStatusPending
		task.mu.Unlock()

		// 发送完成事件
		s.sendEvent(&TaskEvent{
			Type:      TaskEventCompleted,
			TaskID:    task.ID,
			TaskName:  task.Name,
			Timestamp: time.Now(),
		})

		if s.logger != nil {
			s.logger.Debug("Task completed", "task", task.ID)
		}
	}
}

// calculateNextRun 计算下次运行时间
func (s *DefaultTaskScheduler) calculateNextRun(task *Task) (time.Time, error) {
	now := time.Now()

	switch task.Schedule.Type {
	case ScheduleTypeOnce:
		if task.RunCount == 0 {
			return now.Add(task.Schedule.Delay), nil
		}
		return time.Time{}, fmt.Errorf("one-time task already executed")

	case ScheduleTypeInterval:
		if task.RunCount == 0 {
			return now.Add(task.Schedule.Delay), nil
		}
		return now.Add(task.Schedule.Interval), nil

	case ScheduleTypeCron:
		// 简单的cron解析（这里可以集成第三方cron库）
		return s.parseCron(task.Schedule.Cron, now)

	default:
		return time.Time{}, fmt.Errorf("unsupported schedule type")
	}
}

// parseCron 解析cron表达式（简化版本）
func (s *DefaultTaskScheduler) parseCron(cronExpr string, from time.Time) (time.Time, error) {
	// 这里是一个简化的实现，实际项目中建议使用专业的cron库如github.com/robfig/cron
	switch cronExpr {
	case "@every 1m":
		return from.Add(time.Minute), nil
	case "@every 5m":
		return from.Add(5 * time.Minute), nil
	case "@every 1h":
		return from.Add(time.Hour), nil
	case "@daily":
		next := from.Add(24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported cron expression: %s", cronExpr)
	}
}

// sendEvent 发送事件
func (s *DefaultTaskScheduler) sendEvent(event *TaskEvent) {
	if s.eventHandler != nil {
		go s.eventHandler(event)
	}
}

// TaskBuilder 任务构建器
type TaskBuilder struct {
	task *Task
}

// NewTaskBuilder 创建任务构建器
func NewTaskBuilder(id, name string) *TaskBuilder {
	return &TaskBuilder{
		task: &Task{
			ID:       id,
			Name:     name,
			Metadata: make(map[string]interface{}),
			Schedule: &Schedule{},
		},
	}
}

// Description 设置描述
func (b *TaskBuilder) Description(desc string) *TaskBuilder {
	b.task.Description = desc
	return b
}

// Handler 设置处理器
func (b *TaskBuilder) Handler(handler TaskHandler) *TaskBuilder {
	b.task.Handler = handler
	return b
}

// Cron 设置cron调度
func (b *TaskBuilder) Cron(cronExpr string) *TaskBuilder {
	b.task.Schedule.Type = ScheduleTypeCron
	b.task.Schedule.Cron = cronExpr
	return b
}

// Interval 设置间隔调度
func (b *TaskBuilder) Interval(interval time.Duration) *TaskBuilder {
	b.task.Schedule.Type = ScheduleTypeInterval
	b.task.Schedule.Interval = interval
	return b
}

// Once 设置一次性调度
func (b *TaskBuilder) Once(delay time.Duration) *TaskBuilder {
	b.task.Schedule.Type = ScheduleTypeOnce
	b.task.Schedule.Delay = delay
	b.task.Schedule.RunOnce = true
	return b
}

// Delay 设置延迟
func (b *TaskBuilder) Delay(delay time.Duration) *TaskBuilder {
	b.task.Schedule.Delay = delay
	return b
}

// MaxRuns 设置最大运行次数
func (b *TaskBuilder) MaxRuns(max int64) *TaskBuilder {
	b.task.Schedule.MaxRuns = max
	return b
}

// Metadata 设置元数据
func (b *TaskBuilder) Metadata(key string, value interface{}) *TaskBuilder {
	b.task.Metadata[key] = value
	return b
}

// Build 构建任务
func (b *TaskBuilder) Build() *Task {
	return b.task
}
