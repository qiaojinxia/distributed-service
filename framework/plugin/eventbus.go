package plugin

import (
	"fmt"
	"sync"
	"time"
)

// DefaultEventBus 默认事件总线实现
type DefaultEventBus struct {
	subscribers map[string][]EventHandler
	mu          sync.RWMutex
	logger      Logger
}

// NewDefaultEventBus 创建默认事件总线
func NewDefaultEventBus() *DefaultEventBus {
	return &DefaultEventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

// SetLogger 设置日志记录器
func (eb *DefaultEventBus) SetLogger(logger Logger) {
	eb.logger = logger
}

// Publish 发布事件
func (eb *DefaultEventBus) Publish(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 设置时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.mu.RLock()
	handlers, exists := eb.subscribers[event.Type]
	if !exists {
		eb.mu.RUnlock()
		// 没有订阅者，不是错误
		if eb.logger != nil {
			eb.logger.Debug("No subscribers for event type", "type", event.Type)
		}
		return nil
	}

	// 复制处理器列表，避免在执行过程中发生修改
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	eb.mu.RUnlock()

	// 异步执行所有处理器
	for _, handler := range handlersCopy {
		go func(h EventHandler) {
			if err := h(event); err != nil && eb.logger != nil {
				eb.logger.Error("Event handler error",
					"type", event.Type,
					"source", event.Source,
					"error", err)
			}
		}(handler)
	}

	if eb.logger != nil {
		eb.logger.Debug("Event published",
			"type", event.Type,
			"source", event.Source,
			"handlers", len(handlersCopy))
	}

	return nil
}

// PublishSync 同步发布事件
func (eb *DefaultEventBus) PublishSync(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 设置时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.mu.RLock()
	handlers, exists := eb.subscribers[event.Type]
	if !exists {
		eb.mu.RUnlock()
		// 没有订阅者，不是错误
		if eb.logger != nil {
			eb.logger.Debug("No subscribers for event type", "type", event.Type)
		}
		return nil
	}

	// 复制处理器列表
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	eb.mu.RUnlock()

	// 同步执行所有处理器
	var errors []error
	for _, handler := range handlersCopy {
		if err := handler(event); err != nil {
			errors = append(errors, err)
			if eb.logger != nil {
				eb.logger.Error("Event handler error",
					"type", event.Type,
					"source", event.Source,
					"error", err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("event handling errors: %v", errors)
	}

	if eb.logger != nil {
		eb.logger.Debug("Event published synchronously",
			"type", event.Type,
			"source", event.Source,
			"handlers", len(handlersCopy))
	}

	return nil
}

// Subscribe 订阅事件
func (eb *DefaultEventBus) Subscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("event handler cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)

	if eb.logger != nil {
		eb.logger.Debug("Event subscription added", "type", eventType)
	}

	return nil
}

// Unsubscribe 取消订阅事件
func (eb *DefaultEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("event handler cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers, exists := eb.subscribers[eventType]
	if !exists {
		return fmt.Errorf("no subscribers for event type: %s", eventType)
	}

	// 移除处理器（基于指针比较）
	for i, h := range handlers {
		if &h == &handler {
			eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)

			// 如果没有处理器了，删除条目
			if len(eb.subscribers[eventType]) == 0 {
				delete(eb.subscribers, eventType)
			}

			if eb.logger != nil {
				eb.logger.Debug("Event subscription removed", "type", eventType)
			}

			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// GetSubscribers 获取订阅者
func (eb *DefaultEventBus) GetSubscribers(eventType string) []EventHandler {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, exists := eb.subscribers[eventType]
	if !exists {
		return nil
	}

	// 返回副本
	result := make([]EventHandler, len(handlers))
	copy(result, handlers)
	return result
}

// GetEventTypes 获取所有已订阅的事件类型
func (eb *DefaultEventBus) GetEventTypes() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var types []string
	for eventType := range eb.subscribers {
		types = append(types, eventType)
	}
	return types
}

// GetSubscriberCount 获取指定事件类型的订阅者数量
func (eb *DefaultEventBus) GetSubscriberCount(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if handlers, exists := eb.subscribers[eventType]; exists {
		return len(handlers)
	}
	return 0
}

// Clear 清空所有订阅
func (eb *DefaultEventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers = make(map[string][]EventHandler)

	if eb.logger != nil {
		eb.logger.Debug("Event bus cleared")
	}
}

// HasSubscribers 检查是否有订阅者
func (eb *DefaultEventBus) HasSubscribers(eventType string) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, exists := eb.subscribers[eventType]
	return exists && len(handlers) > 0
}

// 预定义的事件类型常量
const (
	// 插件生命周期事件
	EventPluginLoaded      = "plugin.loaded"
	EventPluginUnloaded    = "plugin.unloaded"
	EventPluginInitialized = "plugin.initialized"
	EventPluginStarted     = "plugin.started"
	EventPluginStopped     = "plugin.stopped"
	EventPluginFailed      = "plugin.failed"
	EventPluginHealthCheck = "plugin.health_check"

	// 系统事件
	EventSystemStarted       = "system.started"
	EventSystemStopped       = "system.stopped"
	EventSystemConfigChanged = "system.config_changed"

	// 服务事件
	EventServiceRegistered   = "service.registered"
	EventServiceUnregistered = "service.unregistered"
	EventServiceRequest      = "service.request"
	EventServiceResponse     = "service.response"

	// 自定义事件前缀
	EventCustomPrefix = "custom."
)

// EventBuilder 事件构建器
type EventBuilder struct {
	event *Event
}

// NewEventBuilder 创建事件构建器
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		event: &Event{
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
	}
}

// Type 设置事件类型
func (eb *EventBuilder) Type(eventType string) *EventBuilder {
	eb.event.Type = eventType
	return eb
}

// Source 设置事件源
func (eb *EventBuilder) Source(source string) *EventBuilder {
	eb.event.Source = source
	return eb
}

// Target 设置事件目标
func (eb *EventBuilder) Target(target string) *EventBuilder {
	eb.event.Target = target
	return eb
}

// Data 设置事件数据
func (eb *EventBuilder) Data(data interface{}) *EventBuilder {
	eb.event.Data = data
	return eb
}

// Metadata 设置元数据
func (eb *EventBuilder) Metadata(key string, value interface{}) *EventBuilder {
	if eb.event.Metadata == nil {
		eb.event.Metadata = make(map[string]interface{})
	}
	eb.event.Metadata[key] = value
	return eb
}

// Timestamp 设置时间戳
func (eb *EventBuilder) Timestamp(timestamp time.Time) *EventBuilder {
	eb.event.Timestamp = timestamp
	return eb
}

// Build 构建事件
func (eb *EventBuilder) Build() *Event {
	return eb.event
}

// 便捷方法创建预定义事件

// NewPluginEvent 创建插件事件
func NewPluginEvent(eventType, pluginName string, data interface{}) *Event {
	return NewEventBuilder().
		Type(eventType).
		Source(pluginName).
		Data(data).
		Build()
}

// NewSystemEvent 创建系统事件
func NewSystemEvent(eventType string, data interface{}) *Event {
	return NewEventBuilder().
		Type(eventType).
		Source("system").
		Data(data).
		Build()
}

// NewServiceEvent 创建服务事件
func NewServiceEvent(eventType, serviceName string, data interface{}) *Event {
	return NewEventBuilder().
		Type(eventType).
		Source(serviceName).
		Data(data).
		Build()
}
