package plugin

import (
	"context"
	"time"
)

// Plugin 插件接口 - 所有插件必须实现的核心接口
type Plugin interface {
	// 插件基本信息
	Name() string           // 插件名称
	Version() string        // 插件版本
	Description() string    // 插件描述
	Dependencies() []string // 依赖的插件列表

	// 插件生命周期
	Initialize(ctx context.Context, config Config) error // 初始化插件
	Start(ctx context.Context) error                     // 启动插件
	Stop(ctx context.Context) error                      // 停止插件
	Destroy(ctx context.Context) error                   // 销毁插件

	// 插件状态
	Status() Status       // 获取插件状态
	Health() HealthStatus // 获取健康状态
}

// ServicePlugin 服务插件接口 - 提供具体服务能力的插件
type ServicePlugin interface {
	Plugin

	// 服务相关方法
	GetService() interface{}  // 获取服务实例
	GetEndpoints() []Endpoint // 获取服务端点
}

// MiddlewarePlugin 中间件插件接口
type MiddlewarePlugin interface {
	Plugin

	// 中间件相关方法
	GetMiddleware() interface{} // 获取中间件实例
	Priority() int              // 中间件优先级
}

// TransportPlugin 传输层插件接口
type TransportPlugin interface {
	Plugin

	// 传输层相关方法
	GetTransport() interface{} // 获取传输层实例
	GetProtocol() string       // 获取协议类型
}

// Status 插件状态枚举
type Status int

const (
	StatusUnknown Status = iota
	StatusInitializing
	StatusInitialized
	StatusStarting
	StatusRunning
	StatusStopping
	StatusStopped
	StatusFailed
	StatusDestroyed
)

func (s Status) String() string {
	switch s {
	case StatusInitializing:
		return "initializing"
	case StatusInitialized:
		return "initialized"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	case StatusStopped:
		return "stopped"
	case StatusFailed:
		return "failed"
	case StatusDestroyed:
		return "destroyed"
	default:
		return "unknown"
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Healthy   bool                   `json:"healthy"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Config 插件配置接口
type Config interface {
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	Set(key string, value interface{})
	All() map[string]interface{}
}

// Endpoint 服务端点信息
type Endpoint struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Event 插件事件
type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventHandler 事件处理器
type EventHandler func(event *Event) error

// EventListener 事件监听器接口
type EventListener interface {
	OnEvent(event *Event) error
	GetEventTypes() []string
}

// Loader 插件加载器接口
type Loader interface {
	Load(path string) (Plugin, error)
	Unload(plugin Plugin) error
	Scan(directory string) ([]string, error)
}

// Registry 插件注册表接口
type Registry interface {
	Register(plugin Plugin) error
	Unregister(name string) error
	Get(name string) Plugin
	GetAll() map[string]Plugin
	GetByType(pluginType string) []Plugin
	Exists(name string) bool
}

// Manager 插件管理器接口
type Manager interface {
	// 插件管理
	LoadPlugin(path string) error
	UnloadPlugin(name string) error
	GetPlugin(name string) Plugin
	GetAllPlugins() map[string]Plugin

	// 生命周期管理
	InitializePlugin(name string, config Config) error
	StartPlugin(name string) error
	StopPlugin(name string) error
	RestartPlugin(name string) error

	// 状态查询
	GetPluginStatus(name string) Status
	GetPluginHealth(name string) HealthStatus
	GetPluginsStatus() map[string]Status

	// 事件系统
	PublishEvent(event *Event) error
	SubscribeEvent(eventType string, handler EventHandler) error
	UnsubscribeEvent(eventType string, handler EventHandler) error

	// 插件注册表访问
	GetRegistry() Registry
}

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	GetPluginConfig(pluginName string) Config
	SetPluginConfig(pluginName string, config Config) error
	LoadConfig(path string) error
	SaveConfig(path string) error
}

// PluginFactory 插件工厂接口
type PluginFactory interface {
	CreatePlugin(pluginType string, config Config) (Plugin, error)
	GetSupportedTypes() []string
}

// Context 插件上下文
type Context struct {
	PluginName string
	Manager    Manager
	Registry   Registry
	EventBus   EventBus
	Logger     Logger
	Config     Config
	Metadata   map[string]interface{}
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(event *Event) error
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
	GetSubscribers(eventType string) []EventHandler
}

// Logger 插件日志接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

// DependencyResolver 依赖解析器接口
type DependencyResolver interface {
	Resolve(plugins []Plugin) ([]Plugin, error)
	ValidateDependencies(plugin Plugin) error
	GetDependencyGraph() map[string][]string
}

// HotSwap 热插拔接口
type HotSwap interface {
	CanHotSwap() bool
	PrepareSwap(newPlugin Plugin) error
	PerformSwap(oldPlugin, newPlugin Plugin) error
	RollbackSwap() error
}

// Metrics 插件指标接口
type Metrics interface {
	RecordMetric(name string, value float64, tags map[string]string)
	IncrementCounter(name string, tags map[string]string)
	RecordLatency(name string, duration time.Duration, tags map[string]string)
	GetMetrics() map[string]interface{}
}

// Security 插件安全接口
type Security interface {
	ValidatePlugin(plugin Plugin) error
	GetPermissions(plugin Plugin) []string
	CheckPermission(plugin Plugin, action string) bool
	IsolatePlugin(plugin Plugin) error
}
