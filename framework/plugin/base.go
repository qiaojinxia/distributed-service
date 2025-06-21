package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BasePlugin 插件基础类 - 提供默认实现
type BasePlugin struct {
	name         string
	version      string
	description  string
	dependencies []string

	config  Config
	status  Status
	health  HealthStatus
	context *Context

	mu sync.RWMutex

	// 生命周期钩子
	onInitialize func(ctx context.Context, config Config) error
	onStart      func(ctx context.Context) error
	onStop       func(ctx context.Context) error
	onDestroy    func(ctx context.Context) error
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name, version, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		status:      StatusUnknown,
		health: HealthStatus{
			Healthy:   false,
			Message:   "Not initialized",
			Timestamp: time.Now(),
		},
	}
}

// Name 插件名称
func (p *BasePlugin) Name() string {
	return p.name
}

// Version 插件版本
func (p *BasePlugin) Version() string {
	return p.version
}

// Description 插件描述
func (p *BasePlugin) Description() string {
	return p.description
}

// Dependencies 插件依赖
func (p *BasePlugin) Dependencies() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回副本
	deps := make([]string, len(p.dependencies))
	copy(deps, p.dependencies)
	return deps
}

// SetDependencies 设置依赖
func (p *BasePlugin) SetDependencies(deps []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dependencies = make([]string, len(deps))
	copy(p.dependencies, deps)
}

// Initialize 初始化插件
func (p *BasePlugin) Initialize(ctx context.Context, config Config) error {
	p.mu.Lock()
	p.status = StatusInitializing
	p.config = config
	p.mu.Unlock()

	// 调用自定义初始化逻辑
	if p.onInitialize != nil {
		if err := p.onInitialize(ctx, config); err != nil {
			p.mu.Lock()
			p.status = StatusFailed
			p.health = HealthStatus{
				Healthy:   false,
				Message:   fmt.Sprintf("Initialize failed: %v", err),
				Timestamp: time.Now(),
			}
			p.mu.Unlock()
			return err
		}
	}

	p.mu.Lock()
	p.status = StatusInitialized
	p.health = HealthStatus{
		Healthy:   true,
		Message:   "Initialized",
		Timestamp: time.Now(),
	}
	p.mu.Unlock()

	return nil
}

// Start 启动插件
func (p *BasePlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.status != StatusInitialized && p.status != StatusStopped {
		p.mu.Unlock()
		return fmt.Errorf("cannot start plugin from status: %s", p.status)
	}
	p.status = StatusStarting
	p.mu.Unlock()

	// 调用自定义启动逻辑
	if p.onStart != nil {
		if err := p.onStart(ctx); err != nil {
			p.mu.Lock()
			p.status = StatusFailed
			p.health = HealthStatus{
				Healthy:   false,
				Message:   fmt.Sprintf("Start failed: %v", err),
				Timestamp: time.Now(),
			}
			p.mu.Unlock()
			return err
		}
	}

	p.mu.Lock()
	p.status = StatusRunning
	p.health = HealthStatus{
		Healthy:   true,
		Message:   "Running",
		Timestamp: time.Now(),
	}
	p.mu.Unlock()

	return nil
}

// Stop 停止插件
func (p *BasePlugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	if p.status != StatusRunning {
		p.mu.Unlock()
		return fmt.Errorf("cannot stop plugin from status: %s", p.status)
	}
	p.status = StatusStopping
	p.mu.Unlock()

	// 调用自定义停止逻辑
	if p.onStop != nil {
		if err := p.onStop(ctx); err != nil {
			p.mu.Lock()
			p.status = StatusFailed
			p.health = HealthStatus{
				Healthy:   false,
				Message:   fmt.Sprintf("Stop failed: %v", err),
				Timestamp: time.Now(),
			}
			p.mu.Unlock()
			return err
		}
	}

	p.mu.Lock()
	p.status = StatusStopped
	p.health = HealthStatus{
		Healthy:   true,
		Message:   "Stopped",
		Timestamp: time.Now(),
	}
	p.mu.Unlock()

	return nil
}

// Destroy 销毁插件
func (p *BasePlugin) Destroy(ctx context.Context) error {
	// 调用自定义销毁逻辑
	if p.onDestroy != nil {
		if err := p.onDestroy(ctx); err != nil {
			p.mu.Lock()
			p.status = StatusFailed
			p.health = HealthStatus{
				Healthy:   false,
				Message:   fmt.Sprintf("Destroy failed: %v", err),
				Timestamp: time.Now(),
			}
			p.mu.Unlock()
			return err
		}
	}

	p.mu.Lock()
	p.status = StatusDestroyed
	p.health = HealthStatus{
		Healthy:   false,
		Message:   "Destroyed",
		Timestamp: time.Now(),
	}
	p.mu.Unlock()

	return nil
}

// Status 获取状态
func (p *BasePlugin) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// Health 获取健康状态
func (p *BasePlugin) Health() HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.health
}

// GetConfig 获取配置
func (p *BasePlugin) GetConfig() Config {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// SetContext 设置上下文
func (p *BasePlugin) SetContext(ctx *Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.context = ctx
}

// GetContext 获取上下文
func (p *BasePlugin) GetContext() *Context {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.context
}

// 生命周期钩子设置方法

// OnInitialize 设置初始化钩子
func (p *BasePlugin) OnInitialize(fn func(ctx context.Context, config Config) error) {
	p.onInitialize = fn
}

// OnStart 设置启动钩子
func (p *BasePlugin) OnStart(fn func(ctx context.Context) error) {
	p.onStart = fn
}

// OnStop 设置停止钩子
func (p *BasePlugin) OnStop(fn func(ctx context.Context) error) {
	p.onStop = fn
}

// OnDestroy 设置销毁钩子
func (p *BasePlugin) OnDestroy(fn func(ctx context.Context) error) {
	p.onDestroy = fn
}

// BaseServicePlugin 服务插件基础类
type BaseServicePlugin struct {
	*BasePlugin
	service   interface{}
	endpoints []Endpoint
}

// NewBaseServicePlugin 创建基础服务插件
func NewBaseServicePlugin(name, version, description string) *BaseServicePlugin {
	return &BaseServicePlugin{
		BasePlugin: NewBasePlugin(name, version, description),
		endpoints:  make([]Endpoint, 0),
	}
}

// GetService 获取服务实例
func (p *BaseServicePlugin) GetService() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.service
}

// SetService 设置服务实例
func (p *BaseServicePlugin) SetService(service interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.service = service
}

// GetEndpoints 获取服务端点
func (p *BaseServicePlugin) GetEndpoints() []Endpoint {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 返回副本
	endpoints := make([]Endpoint, len(p.endpoints))
	copy(endpoints, p.endpoints)
	return endpoints
}

// AddEndpoint 添加端点
func (p *BaseServicePlugin) AddEndpoint(endpoint Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.endpoints = append(p.endpoints, endpoint)
}

// BaseMiddlewarePlugin 中间件插件基础类
type BaseMiddlewarePlugin struct {
	*BasePlugin
	middleware interface{}
	priority   int
}

// NewBaseMiddlewarePlugin 创建基础中间件插件
func NewBaseMiddlewarePlugin(name, version, description string, priority int) *BaseMiddlewarePlugin {
	return &BaseMiddlewarePlugin{
		BasePlugin: NewBasePlugin(name, version, description),
		priority:   priority,
	}
}

// GetMiddleware 获取中间件实例
func (p *BaseMiddlewarePlugin) GetMiddleware() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.middleware
}

// SetMiddleware 设置中间件实例
func (p *BaseMiddlewarePlugin) SetMiddleware(middleware interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.middleware = middleware
}

// Priority 获取优先级
func (p *BaseMiddlewarePlugin) Priority() int {
	return p.priority
}

// BaseTransportPlugin 传输层插件基础类
type BaseTransportPlugin struct {
	*BasePlugin
	transport interface{}
	protocol  string
}

// NewBaseTransportPlugin 创建基础传输层插件
func NewBaseTransportPlugin(name, version, description, protocol string) *BaseTransportPlugin {
	return &BaseTransportPlugin{
		BasePlugin: NewBasePlugin(name, version, description),
		protocol:   protocol,
	}
}

// GetTransport 获取传输层实例
func (p *BaseTransportPlugin) GetTransport() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.transport
}

// SetTransport 设置传输层实例
func (p *BaseTransportPlugin) SetTransport(transport interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.transport = transport
}

// GetProtocol 获取协议类型
func (p *BaseTransportPlugin) GetProtocol() string {
	return p.protocol
}

// PluginBuilder 插件构建器
type PluginBuilder struct {
	plugin *BasePlugin
}

// NewPluginBuilder 创建插件构建器
func NewPluginBuilder(name, version, description string) *PluginBuilder {
	return &PluginBuilder{
		plugin: NewBasePlugin(name, version, description),
	}
}

// Dependencies 设置依赖
func (b *PluginBuilder) Dependencies(deps []string) *PluginBuilder {
	b.plugin.SetDependencies(deps)
	return b
}

// OnInitialize 设置初始化钩子
func (b *PluginBuilder) OnInitialize(fn func(ctx context.Context, config Config) error) *PluginBuilder {
	b.plugin.OnInitialize(fn)
	return b
}

// OnStart 设置启动钩子
func (b *PluginBuilder) OnStart(fn func(ctx context.Context) error) *PluginBuilder {
	b.plugin.OnStart(fn)
	return b
}

// OnStop 设置停止钩子
func (b *PluginBuilder) OnStop(fn func(ctx context.Context) error) *PluginBuilder {
	b.plugin.OnStop(fn)
	return b
}

// OnDestroy 设置销毁钩子
func (b *PluginBuilder) OnDestroy(fn func(ctx context.Context) error) *PluginBuilder {
	b.plugin.OnDestroy(fn)
	return b
}

// Build 构建插件
func (b *PluginBuilder) Build() Plugin {
	return b.plugin
}

// ServicePluginBuilder 服务插件构建器
type ServicePluginBuilder struct {
	plugin *BaseServicePlugin
}

// NewServicePluginBuilder 创建服务插件构建器
func NewServicePluginBuilder(name, version, description string) *ServicePluginBuilder {
	return &ServicePluginBuilder{
		plugin: NewBaseServicePlugin(name, version, description),
	}
}

// Dependencies 设置依赖
func (b *ServicePluginBuilder) Dependencies(deps []string) *ServicePluginBuilder {
	b.plugin.SetDependencies(deps)
	return b
}

// Service 设置服务实例
func (b *ServicePluginBuilder) Service(service interface{}) *ServicePluginBuilder {
	b.plugin.SetService(service)
	return b
}

// Endpoint 添加端点
func (b *ServicePluginBuilder) Endpoint(endpoint Endpoint) *ServicePluginBuilder {
	b.plugin.AddEndpoint(endpoint)
	return b
}

// OnInitialize 设置初始化钩子
func (b *ServicePluginBuilder) OnInitialize(fn func(ctx context.Context, config Config) error) *ServicePluginBuilder {
	b.plugin.OnInitialize(fn)
	return b
}

// OnStart 设置启动钩子
func (b *ServicePluginBuilder) OnStart(fn func(ctx context.Context) error) *ServicePluginBuilder {
	b.plugin.OnStart(fn)
	return b
}

// OnStop 设置停止钩子
func (b *ServicePluginBuilder) OnStop(fn func(ctx context.Context) error) *ServicePluginBuilder {
	b.plugin.OnStop(fn)
	return b
}

// OnDestroy 设置销毁钩子
func (b *ServicePluginBuilder) OnDestroy(fn func(ctx context.Context) error) *ServicePluginBuilder {
	b.plugin.OnDestroy(fn)
	return b
}

// Build 构建服务插件
func (b *ServicePluginBuilder) Build() ServicePlugin {
	return b.plugin
}
