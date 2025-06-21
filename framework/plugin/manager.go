package plugin

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// DefaultManager 默认插件管理器实现
type DefaultManager struct {
	registry           Registry
	configProvider     ConfigProvider
	eventBus           EventBus
	loader             Loader
	factory            PluginFactory
	dependencyResolver DependencyResolver

	// 插件状态管理
	pluginStates map[string]Status
	pluginHealth map[string]HealthStatus

	// 生命周期管理
	initOrder  []string // 初始化顺序
	startOrder []string // 启动顺序

	// 同步控制
	mu sync.RWMutex

	// 配置
	config *ManagerConfig
	logger Logger

	// 状态
	started bool
	stopped bool
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	EnableAutoLoad        bool          // 是否启用自动加载
	AutoLoadDirectory     string        // 自动加载目录
	EnableHotSwap         bool          // 是否启用热插拔
	HealthCheckInterval   time.Duration // 健康检查间隔
	EnableDependencyCheck bool          // 是否启用依赖检查
	MaxStartupTime        time.Duration // 最大启动时间
	EnableMetrics         bool          // 是否启用指标收集
}

// NewDefaultManager 创建默认管理器
func NewDefaultManager(config *ManagerConfig) *DefaultManager {
	if config == nil {
		config = &ManagerConfig{
			EnableAutoLoad:        false,
			EnableHotSwap:         false,
			HealthCheckInterval:   30 * time.Second,
			EnableDependencyCheck: true,
			MaxStartupTime:        60 * time.Second,
			EnableMetrics:         false,
		}
	}

	return &DefaultManager{
		registry:       NewDefaultRegistry(),
		configProvider: NewDefaultConfigProvider(),
		eventBus:       NewDefaultEventBus(),
		pluginStates:   make(map[string]Status),
		pluginHealth:   make(map[string]HealthStatus),
		config:         config,
	}
}

// SetRegistry 设置注册表
func (m *DefaultManager) SetRegistry(registry Registry) {
	m.registry = registry
}

// SetConfigProvider 设置配置提供者
func (m *DefaultManager) SetConfigProvider(provider ConfigProvider) {
	m.configProvider = provider
}

// SetEventBus 设置事件总线
func (m *DefaultManager) SetEventBus(eventBus EventBus) {
	m.eventBus = eventBus
}

// SetLoader 设置加载器
func (m *DefaultManager) SetLoader(loader Loader) {
	m.loader = loader
}

// SetFactory 设置工厂
func (m *DefaultManager) SetFactory(factory PluginFactory) {
	m.factory = factory
}

// SetDependencyResolver 设置依赖解析器
func (m *DefaultManager) SetDependencyResolver(resolver DependencyResolver) {
	m.dependencyResolver = resolver
}

// SetLogger 设置日志记录器
func (m *DefaultManager) SetLogger(logger Logger) {
	m.logger = logger
	if eb, ok := m.eventBus.(*DefaultEventBus); ok {
		eb.SetLogger(logger)
	}
}

// LoadPlugin 加载插件
func (m *DefaultManager) LoadPlugin(path string) error {
	if m.loader == nil {
		return fmt.Errorf("plugin loader not set")
	}

	plugin, err := m.loader.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin from %s: %w", path, err)
	}

	// 注册插件
	if err := m.registry.Register(plugin); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	// 初始化状态
	m.mu.Lock()
	m.pluginStates[plugin.Name()] = StatusInitialized
	m.pluginHealth[plugin.Name()] = HealthStatus{
		Healthy:   true,
		Message:   "Plugin loaded",
		Timestamp: time.Now(),
	}
	m.mu.Unlock()

	// 发布事件
	event := NewPluginEvent(EventPluginLoaded, plugin.Name(), plugin)
	if err := m.eventBus.Publish(event); err != nil && m.logger != nil {
		m.logger.Error("Failed to publish plugin loaded event", "error", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin loaded", "name", plugin.Name(), "version", plugin.Version())
	}

	return nil
}

// UnloadPlugin 卸载插件
func (m *DefaultManager) UnloadPlugin(name string) error {
	plugin := m.registry.Get(name)
	if plugin == nil {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// 检查依赖关系
	if m.config.EnableDependencyCheck {
		dependents := m.registry.(*DefaultRegistry).GetDependents(name)
		if len(dependents) > 0 {
			var dependentNames []string
			for _, dep := range dependents {
				dependentNames = append(dependentNames, dep.Name())
			}
			return fmt.Errorf("cannot unload plugin '%s': other plugins depend on it: %v", name, dependentNames)
		}
	}

	// 停止插件
	if m.GetPluginStatus(name) == StatusRunning {
		if err := m.StopPlugin(name); err != nil {
			m.logger.Warn("Failed to stop plugin during unload", "name", name, "error", err)
		}
	}

	// 销毁插件
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := plugin.Destroy(ctx); err != nil && m.logger != nil {
		m.logger.Warn("Plugin destroy failed", "name", name, "error", err)
	}

	// 卸载插件（如果支持）
	if m.loader != nil {
		if err := m.loader.Unload(plugin); err != nil && m.logger != nil {
			m.logger.Warn("Plugin unload failed", "name", name, "error", err)
		}
	}

	// 从注册表移除
	if err := m.registry.Unregister(name); err != nil {
		return fmt.Errorf("failed to unregister plugin: %w", err)
	}

	// 清理状态
	m.mu.Lock()
	delete(m.pluginStates, name)
	delete(m.pluginHealth, name)
	m.mu.Unlock()

	// 发布事件
	event := NewPluginEvent(EventPluginUnloaded, name, plugin)
	if err := m.eventBus.Publish(event); err != nil && m.logger != nil {
		m.logger.Error("Failed to publish plugin unloaded event", "error", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin unloaded", "name", name)
	}

	return nil
}

// GetPlugin 获取插件
func (m *DefaultManager) GetPlugin(name string) Plugin {
	return m.registry.Get(name)
}

// GetAllPlugins 获取所有插件
func (m *DefaultManager) GetAllPlugins() map[string]Plugin {
	return m.registry.GetAll()
}

// InitializePlugin 初始化插件
func (m *DefaultManager) InitializePlugin(name string, config Config) error {
	plugin := m.registry.Get(name)
	if plugin == nil {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// 检查状态
	currentStatus := m.GetPluginStatus(name)
	if currentStatus != StatusInitialized {
		return fmt.Errorf("plugin '%s' is not in initialized state, current: %s", name, currentStatus)
	}

	// 更新状态
	m.updatePluginStatus(name, StatusInitializing)

	// 初始化插件
	ctx, cancel := context.WithTimeout(context.Background(), m.config.MaxStartupTime)
	defer cancel()

	if config == nil {
		config = m.configProvider.GetPluginConfig(name)
	}

	err := plugin.Initialize(ctx, config)
	if err != nil {
		m.updatePluginStatus(name, StatusFailed)
		m.updatePluginHealth(name, HealthStatus{
			Healthy:   false,
			Message:   fmt.Sprintf("Initialization failed: %v", err),
			Timestamp: time.Now(),
		})

		// 发布失败事件
		event := NewPluginEvent(EventPluginFailed, name, err)
		m.eventBus.Publish(event)

		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// 更新状态
	m.updatePluginStatus(name, StatusInitialized)
	m.updatePluginHealth(name, HealthStatus{
		Healthy:   true,
		Message:   "Initialized successfully",
		Timestamp: time.Now(),
	})

	// 发布事件
	event := NewPluginEvent(EventPluginInitialized, name, plugin)
	if err := m.eventBus.Publish(event); err != nil && m.logger != nil {
		m.logger.Error("Failed to publish plugin initialized event", "error", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin initialized", "name", name)
	}

	return nil
}

// StartPlugin 启动插件
func (m *DefaultManager) StartPlugin(name string) error {
	plugin := m.registry.Get(name)
	if plugin == nil {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// 检查状态
	currentStatus := m.GetPluginStatus(name)
	if currentStatus != StatusInitialized && currentStatus != StatusStopped {
		return fmt.Errorf("plugin '%s' cannot be started from state: %s", name, currentStatus)
	}

	// 检查依赖
	if m.config.EnableDependencyCheck {
		if err := m.checkDependencies(name); err != nil {
			return fmt.Errorf("dependency check failed: %w", err)
		}
	}

	// 更新状态
	m.updatePluginStatus(name, StatusStarting)

	// 启动插件
	ctx, cancel := context.WithTimeout(context.Background(), m.config.MaxStartupTime)
	defer cancel()

	err := plugin.Start(ctx)
	if err != nil {
		m.updatePluginStatus(name, StatusFailed)
		m.updatePluginHealth(name, HealthStatus{
			Healthy:   false,
			Message:   fmt.Sprintf("Start failed: %v", err),
			Timestamp: time.Now(),
		})

		// 发布失败事件
		event := NewPluginEvent(EventPluginFailed, name, err)
		m.eventBus.Publish(event)

		return fmt.Errorf("plugin start failed: %w", err)
	}

	// 更新状态
	m.updatePluginStatus(name, StatusRunning)
	m.updatePluginHealth(name, HealthStatus{
		Healthy:   true,
		Message:   "Running",
		Timestamp: time.Now(),
	})

	// 发布事件
	event := NewPluginEvent(EventPluginStarted, name, plugin)
	if err := m.eventBus.Publish(event); err != nil && m.logger != nil {
		m.logger.Error("Failed to publish plugin started event", "error", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin started", "name", name)
	}

	return nil
}

// StopPlugin 停止插件
func (m *DefaultManager) StopPlugin(name string) error {
	plugin := m.registry.Get(name)
	if plugin == nil {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// 检查状态
	currentStatus := m.GetPluginStatus(name)
	if currentStatus != StatusRunning {
		return fmt.Errorf("plugin '%s' is not running, current state: %s", name, currentStatus)
	}

	// 更新状态
	m.updatePluginStatus(name, StatusStopping)

	// 停止插件
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := plugin.Stop(ctx)
	if err != nil {
		m.updatePluginStatus(name, StatusFailed)
		m.updatePluginHealth(name, HealthStatus{
			Healthy:   false,
			Message:   fmt.Sprintf("Stop failed: %v", err),
			Timestamp: time.Now(),
		})

		return fmt.Errorf("plugin stop failed: %w", err)
	}

	// 更新状态
	m.updatePluginStatus(name, StatusStopped)
	m.updatePluginHealth(name, HealthStatus{
		Healthy:   true,
		Message:   "Stopped",
		Timestamp: time.Now(),
	})

	// 发布事件
	event := NewPluginEvent(EventPluginStopped, name, plugin)
	if err := m.eventBus.Publish(event); err != nil && m.logger != nil {
		m.logger.Error("Failed to publish plugin stopped event", "error", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin stopped", "name", name)
	}

	return nil
}

// RestartPlugin 重启插件
func (m *DefaultManager) RestartPlugin(name string) error {
	if err := m.StopPlugin(name); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	if err := m.StartPlugin(name); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin restarted", "name", name)
	}

	return nil
}

// GetPluginStatus 获取插件状态
func (m *DefaultManager) GetPluginStatus(name string) Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, exists := m.pluginStates[name]; exists {
		return status
	}
	return StatusUnknown
}

// GetPluginHealth 获取插件健康状态
func (m *DefaultManager) GetPluginHealth(name string) HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if health, exists := m.pluginHealth[name]; exists {
		return health
	}
	return HealthStatus{
		Healthy:   false,
		Message:   "Plugin not found",
		Timestamp: time.Now(),
	}
}

// GetPluginsStatus 获取所有插件状态
func (m *DefaultManager) GetPluginsStatus() map[string]Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Status)
	for name, status := range m.pluginStates {
		result[name] = status
	}
	return result
}

// PublishEvent 发布事件
func (m *DefaultManager) PublishEvent(event *Event) error {
	return m.eventBus.Publish(event)
}

// SubscribeEvent 订阅事件
func (m *DefaultManager) SubscribeEvent(eventType string, handler EventHandler) error {
	return m.eventBus.Subscribe(eventType, handler)
}

// UnsubscribeEvent 取消订阅事件
func (m *DefaultManager) UnsubscribeEvent(eventType string, handler EventHandler) error {
	return m.eventBus.Unsubscribe(eventType, handler)
}

// StartAll 启动所有插件
func (m *DefaultManager) StartAll() error {
	plugins := m.registry.GetAll()

	// 解析依赖顺序
	var pluginList []Plugin
	for _, plugin := range plugins {
		pluginList = append(pluginList, plugin)
	}

	if m.dependencyResolver != nil {
		orderedPlugins, err := m.dependencyResolver.Resolve(pluginList)
		if err != nil {
			return fmt.Errorf("dependency resolution failed: %w", err)
		}
		pluginList = orderedPlugins
	}

	// 按顺序启动插件
	for _, plugin := range pluginList {
		if err := m.StartPlugin(plugin.Name()); err != nil {
			if m.logger != nil {
				m.logger.Error("Failed to start plugin", "name", plugin.Name(), "error", err)
			}
			// 继续启动其他插件
		}
	}

	m.started = true
	return nil
}

// StopAll 停止所有插件
func (m *DefaultManager) StopAll() error {
	plugins := m.registry.GetAll()

	// 反向停止插件
	var pluginNames []string
	for name := range plugins {
		pluginNames = append(pluginNames, name)
	}

	// 按字母序倒序停止
	sort.Sort(sort.Reverse(sort.StringSlice(pluginNames)))

	for _, name := range pluginNames {
		if err := m.StopPlugin(name); err != nil {
			if m.logger != nil {
				m.logger.Error("Failed to stop plugin", "name", name, "error", err)
			}
			// 继续停止其他插件
		}
	}

	m.stopped = true
	return nil
}

// 辅助方法

// updatePluginStatus 更新插件状态
func (m *DefaultManager) updatePluginStatus(name string, status Status) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pluginStates[name] = status
}

// updatePluginHealth 更新插件健康状态
func (m *DefaultManager) updatePluginHealth(name string, health HealthStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pluginHealth[name] = health
}

// checkDependencies 检查插件依赖
func (m *DefaultManager) checkDependencies(pluginName string) error {
	plugin := m.registry.Get(pluginName)
	if plugin == nil {
		return fmt.Errorf("plugin not found")
	}

	for _, depName := range plugin.Dependencies() {
		depPlugin := m.registry.Get(depName)
		if depPlugin == nil {
			return fmt.Errorf("dependency '%s' not found", depName)
		}

		depStatus := m.GetPluginStatus(depName)
		if depStatus != StatusRunning {
			return fmt.Errorf("dependency '%s' is not running (status: %s)", depName, depStatus)
		}
	}

	return nil
}

// GetRegistry 获取插件注册表
func (m *DefaultManager) GetRegistry() Registry {
	return m.registry
}

// IsStarted 检查管理器是否已启动
func (m *DefaultManager) IsStarted() bool {
	return m.started
}

// IsStopped 检查管理器是否已停止
func (m *DefaultManager) IsStopped() bool {
	return m.stopped
}
