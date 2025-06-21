package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/logger"
)

// App 应用实例 - 框架的核心运行时
type App struct {
	ctx     context.Context
	cancel  context.CancelFunc
	name    string
	version string

	// 配置
	opts *Options

	// 组件
	transports []Transport
	components []Component

	// 生命周期回调
	beforeStart []func(context.Context) error
	afterStart  []func(context.Context) error
	beforeStop  []func(context.Context) error
	afterStop   []func(context.Context) error
}

// Options 应用配置选项
type Options struct {
	// 基础配置
	Name    string
	Version string
	Mode    string
	Port    int

	// 组件开关
	EnableHTTP    bool
	EnableGRPC    bool
	EnableMetrics bool
	EnableTracing bool

	// 配置文件
	ConfigPath string

	// 其他选项
	ShutdownTimeout time.Duration
}

// Transport 传输层接口
type Transport interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// Component 组件接口
type Component interface {
	Name() string
	Init(context.Context) error
	Start(context.Context) error
	Stop(context.Context) error
}

// HTTPHandler HTTP路由处理器
type HTTPHandler func(interface{})

// GRPCHandler gRPC服务处理器
type GRPCHandler func(interface{})

// NewApp 创建新的应用实例
func NewApp(opts ...Option) *App {
	// 默认配置
	options := &Options{
		Name:            "distributed-service",
		Version:         "v1.0.0",
		Mode:            "release",
		Port:            8080,
		EnableHTTP:      true,
		EnableGRPC:      true,
		EnableMetrics:   true,
		EnableTracing:   true,
		ShutdownTimeout: 30 * time.Second,
	}

	// 应用选项
	for _, opt := range opts {
		opt(options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		ctx:    ctx,
		cancel: cancel,
		name:   options.Name,
		opts:   options,
	}
}

// Option 配置选项函数
type Option func(*Options)

// Name 设置应用名称
func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// Version 设置应用版本
func Version(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

// Port 设置端口
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// Mode 设置运行模式
func Mode(mode string) Option {
	return func(o *Options) {
		o.Mode = mode
	}
}

// Run 运行应用 - 阻塞直到收到停止信号
func (a *App) Run() error {
	// 1. 执行启动前回调
	for _, fn := range a.beforeStart {
		if err := fn(a.ctx); err != nil {
			return fmt.Errorf("before start callback failed: %w", err)
		}
	}

	// 2. 初始化组件
	for _, comp := range a.components {
		if err := comp.Init(a.ctx); err != nil {
			return fmt.Errorf("component %s init failed: %w", comp.Name(), err)
		}
	}

	// 3. 启动组件
	for _, comp := range a.components {
		if err := comp.Start(a.ctx); err != nil {
			return fmt.Errorf("component %s start failed: %w", comp.Name(), err)
		}
	}

	// 4. 启动传输层
	for _, transport := range a.transports {
		if err := transport.Start(a.ctx); err != nil {
			return fmt.Errorf("transport start failed: %w", err)
		}
	}

	// 5. 执行启动后回调
	for _, fn := range a.afterStart {
		if err := fn(a.ctx); err != nil {
			return fmt.Errorf("after start callback failed: %w", err)
		}
	}

	log := logger.GetLogger()
	log.Info("🚀 Application started",
		logger.String("name", a.opts.Name),
		logger.String("version", a.opts.Version),
		logger.Int("port", a.opts.Port),
		logger.String("mode", a.opts.Mode))

	// 6. 等待停止信号
	return a.waitForShutdown()
}

// Stop 停止应用
func (a *App) Stop() error {
	log := logger.GetLogger()
	log.Info("🛑 Shutting down server...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), a.opts.ShutdownTimeout)
	defer cancel()

	// 执行停止前回调
	for _, fn := range a.beforeStop {
		if err := fn(ctx); err != nil {
			log.Error("Before stop callback failed", logger.Any("error", err))
		}
	}

	// 停止传输层
	for _, transport := range a.transports {
		if err := transport.Stop(ctx); err != nil {
			log.Error("Transport stop failed", logger.Any("error", err))
		}
	}

	// 停止组件
	for i := len(a.components) - 1; i >= 0; i-- {
		if err := a.components[i].Stop(ctx); err != nil {
			log.Error("Component stop failed",
				logger.String("component", a.components[i].Name()),
				logger.Any("error", err))
		}
	}

	// 执行停止后回调
	for _, fn := range a.afterStop {
		if err := fn(ctx); err != nil {
			log.Error("After stop callback failed", logger.Any("error", err))
		}
	}

	a.cancel()
	log.Info("✅ Server stopped gracefully")
	return nil
}

// waitForShutdown 等待停止信号
func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	return a.Stop()
}

// Context 获取应用上下文
func (a *App) Context() context.Context {
	return a.ctx
}

// Options 获取应用配置
func (a *App) Options() *Options {
	return a.opts
}

// AddTransport 添加传输层
func (a *App) AddTransport(transport Transport) {
	a.transports = append(a.transports, transport)
}

// AddComponent 添加组件
func (a *App) AddComponent(component Component) {
	a.components = append(a.components, component)
}

// BeforeStart 添加启动前回调
func (a *App) BeforeStart(fn func(context.Context) error) {
	a.beforeStart = append(a.beforeStart, fn)
}

// AfterStart 添加启动后回调
func (a *App) AfterStart(fn func(context.Context) error) {
	a.afterStart = append(a.afterStart, fn)
}

// BeforeStop 添加停止前回调
func (a *App) BeforeStop(fn func(context.Context) error) {
	a.beforeStop = append(a.beforeStop, fn)
}

// AfterStop 添加停止后回调
func (a *App) AfterStop(fn func(context.Context) error) {
	a.afterStop = append(a.afterStop, fn)
}
