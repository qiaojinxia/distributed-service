package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/qiaojinxia/distributed-service/framework/component"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/transport/http"
)

// ================================
// 🚀 传输层包装器
// ================================

// HTTPTransport HTTP传输层包装器
type HTTPTransport struct {
	port     int
	mode     string
	handlers []HTTPHandler
	server   *http.Server
}

// Start 启动HTTP传输层
func (h *HTTPTransport) Start(ctx context.Context) error {
	// 创建HTTP服务器配置
	cfg := &http.Config{
		Port: h.port,
		Mode: h.mode,
	}

	// 创建HTTP服务器
	h.server = http.NewServer(cfg)

	// 注册所有路由处理器
	for _, handler := range h.handlers {
		handler(h.server.Engine())
	}

	// 启动服务器
	return h.server.Start(ctx)
}

// Stop 停止HTTP传输层
func (h *HTTPTransport) Stop(ctx context.Context) error {
	if h.server != nil {
		return h.server.Stop(ctx)
	}
	return nil
}

// Builder 应用构建器 - 提供流畅的链式配置API
type Builder struct {
	app *App

	// HTTP处理器
	httpHandlers []HTTPHandler

	// gRPC处理器
	grpcHandlers []GRPCHandler

	// 组件管理器
	componentManager *component.Manager

	// 自动检测标志
	autoDetect bool
}

// New 创建新的构建器
func New() *Builder {
	return &Builder{
		app:              NewApp(),
		componentManager: component.NewManager(),
		autoDetect:       false,
	}
}

// ================================
// 🛠️ 基础配置API
// ================================

// Port 设置端口
func (b *Builder) Port(port int) *Builder {
	b.app.opts.Port = port
	return b
}

// Mode 设置运行模式 (debug/release/test)
func (b *Builder) Mode(mode string) *Builder {
	b.app.opts.Mode = mode
	return b
}

// Name 设置应用名称
func (b *Builder) Name(name string) *Builder {
	b.app.opts.Name = name
	return b
}

// Version 设置应用版本
func (b *Builder) Version(version string) *Builder {
	b.app.opts.Version = version
	return b
}

// Config 设置配置文件路径
func (b *Builder) Config(path string) *Builder {
	b.app.opts.ConfigPath = path
	// 同时设置组件管理器的配置路径
	b.componentManager = component.NewManager(component.WithConfig(path))
	return b
}

// ================================
// 🧩 组件配置API (新增)
// ================================

// WithDatabase 配置数据库组件
func (b *Builder) WithDatabase(cfg *config.MySQLConfig) *Builder {
	b.componentManager = component.NewManager(component.WithDatabase(cfg))
	return b
}

// WithRedis 配置Redis组件
func (b *Builder) WithRedis(cfg *config.RedisConfig) *Builder {
	b.componentManager = component.NewManager(component.WithRedis(cfg))
	return b
}

// WithRedisCluster 配置Redis集群组件
func (b *Builder) WithRedisCluster(cfg *config.RedisClusterConfig) *Builder {
	b.componentManager = component.NewManager(component.WithRedisCluster(cfg))
	return b
}

// WithAuth 配置认证组件
func (b *Builder) WithAuth(cfg *config.JWTConfig) *Builder {
	b.componentManager = component.NewManager(component.WithAuth(cfg))
	return b
}

// WithRegistry 配置服务注册组件
func (b *Builder) WithRegistry(cfg *config.ConsulConfig) *Builder {
	b.componentManager = component.NewManager(component.WithRegistry(cfg))
	return b
}

// WithGRPCConfig 配置gRPC组件
func (b *Builder) WithGRPCConfig(cfg *config.GRPCConfig) *Builder {
	b.componentManager = component.NewManager(component.WithGRPC(cfg))
	return b
}

// WithMQ 配置消息队列组件
func (b *Builder) WithMQ(cfg *config.RabbitMQConfig) *Builder {
	b.componentManager = component.NewManager(component.WithMQ(cfg))
	return b
}

// WithMetrics 配置监控组件
func (b *Builder) WithMetrics(cfg *config.MetricsConfig) *Builder {
	b.componentManager = component.NewManager(component.WithMetrics(cfg))
	return b
}

// WithTracing 配置链路追踪组件
func (b *Builder) WithTracing(cfg *config.TracingConfig) *Builder {
	b.componentManager = component.NewManager(component.WithTracing(cfg))
	return b
}

// WithProtection 配置保护组件
func (b *Builder) WithProtection(cfg *config.ProtectionConfig) *Builder {
	b.componentManager = component.NewManager(component.WithProtection(cfg))
	return b
}

// WithLogger 配置日志组件
func (b *Builder) WithLogger(cfg *config.LoggerConfig) *Builder {
	b.componentManager = component.NewManager(component.WithLogger(cfg))
	return b
}

// WithElasticsearch 配置Elasticsearch组件
func (b *Builder) WithElasticsearch(cfg *config.ElasticsearchConfig) *Builder {
	b.componentManager = component.NewManager(component.WithElasticsearch(cfg))
	return b
}

// WithKafka 配置Kafka组件
func (b *Builder) WithKafka(cfg *config.KafkaConfig) *Builder {
	b.componentManager = component.NewManager(component.WithKafka(cfg))
	return b
}

// WithMongoDB 配置MongoDB组件
func (b *Builder) WithMongoDB(cfg *config.MongoDBConfig) *Builder {
	b.componentManager = component.NewManager(component.WithMongoDB(cfg))
	return b
}

// WithEtcd 配置Etcd组件
func (b *Builder) WithEtcd(cfg *config.EtcdConfig) *Builder {
	b.componentManager = component.NewManager(component.WithEtcd(cfg))
	return b
}

// DisableComponents 禁用指定组件
func (b *Builder) DisableComponents(components ...string) *Builder {
	b.componentManager = component.NewManager(component.DisableComponent(components...))
	return b
}

// ================================
// 🎯 快捷模式配置
// ================================

// Dev 开发模式 - 8080端口，debug模式
func (b *Builder) Dev() *Builder {
	return b.Port(8080).Mode("debug")
}

// Prod 生产模式 - 80端口，release模式
func (b *Builder) Prod() *Builder {
	return b.Port(80).Mode("release")
}

// Test 测试模式 - 随机端口，test模式，只启用HTTP
func (b *Builder) Test() *Builder {
	return b.Port(0).Mode("test").OnlyHTTP()
}

// ================================
// 🔧 组件控制
// ================================

// EnableHTTP 启用HTTP服务
func (b *Builder) EnableHTTP() *Builder {
	b.app.opts.EnableHTTP = true
	return b
}

// EnableGRPC 启用gRPC服务
func (b *Builder) EnableGRPC() *Builder {
	b.app.opts.EnableGRPC = true
	return b
}

// EnableMetrics 启用监控指标
func (b *Builder) EnableMetrics() *Builder {
	b.app.opts.EnableMetrics = true
	return b
}

// EnableTracing 启用链路追踪
func (b *Builder) EnableTracing() *Builder {
	b.app.opts.EnableTracing = true
	return b
}

// DisableHTTP 禁用HTTP服务
func (b *Builder) DisableHTTP() *Builder {
	b.app.opts.EnableHTTP = false
	return b
}

// DisableGRPC 禁用gRPC服务
func (b *Builder) DisableGRPC() *Builder {
	b.app.opts.EnableGRPC = false
	return b
}

// DisableMetrics 禁用监控指标
func (b *Builder) DisableMetrics() *Builder {
	b.app.opts.EnableMetrics = false
	return b
}

// DisableTracing 禁用链路追踪
func (b *Builder) DisableTracing() *Builder {
	b.app.opts.EnableTracing = false
	return b
}

// OnlyHTTP 只启用HTTP服务
func (b *Builder) OnlyHTTP() *Builder {
	// 禁用应用级别的gRPC
	b.EnableHTTP().DisableGRPC()

	// 同时禁用组件管理器中的gRPC
	b.componentManager = component.NewManager(component.DisableComponent("grpc"))

	return b
}

// OnlyGRPC 只启用gRPC服务
func (b *Builder) OnlyGRPC() *Builder {
	// 禁用应用级别的HTTP
	b.EnableGRPC().DisableHTTP()

	// 确保组件管理器中gRPC是启用的（保持默认）
	// 这里不需要特殊处理，因为默认配置就是启用gRPC的

	return b
}

// EnableAll 启用所有组件
func (b *Builder) EnableAll() *Builder {
	return b.EnableHTTP().EnableGRPC().EnableMetrics().EnableTracing()
}

// DisableAll 禁用所有组件
func (b *Builder) DisableAll() *Builder {
	return b.DisableHTTP().DisableGRPC().DisableMetrics().DisableTracing()
}

// Enable 启用指定组件
func (b *Builder) Enable(components ...string) *Builder {
	for _, comp := range components {
		switch strings.ToLower(comp) {
		case "http":
			b.EnableHTTP()
		case "grpc":
			b.EnableGRPC()
		case "metrics":
			b.EnableMetrics()
		case "tracing":
			b.EnableTracing()
		}
	}
	return b
}

// Disable 禁用指定组件
func (b *Builder) Disable(components ...string) *Builder {
	for _, comp := range components {
		switch strings.ToLower(comp) {
		case "http":
			b.DisableHTTP()
		case "grpc":
			b.DisableGRPC()
		case "metrics":
			b.DisableMetrics()
		case "tracing":
			b.DisableTracing()
		}
	}
	return b
}

// ================================
// 🔄 生命周期钩子
// ================================

// OnInit 初始化钩子
func (b *Builder) OnInit(callback func() error) *Builder {
	b.app.BeforeStart(func(ctx context.Context) error {
		return callback()
	})
	return b
}

// OnReady 就绪钩子
func (b *Builder) OnReady(callback func() error) *Builder {
	b.app.AfterStart(func(ctx context.Context) error {
		return callback()
	})
	return b
}

// OnStop 停止钩子
func (b *Builder) OnStop(callback func() error) *Builder {
	b.app.BeforeStop(func(ctx context.Context) error {
		return callback()
	})
	return b
}

// BeforeStart 启动前回调
func (b *Builder) BeforeStart(callback func(context.Context) error) *Builder {
	b.app.BeforeStart(callback)
	return b
}

// AfterStart 启动后回调
func (b *Builder) AfterStart(callback func(context.Context) error) *Builder {
	b.app.AfterStart(callback)
	return b
}

// BeforeStop 停止前回调
func (b *Builder) BeforeStop(callback func(context.Context) error) *Builder {
	b.app.BeforeStop(callback)
	return b
}

// AfterStop 停止后回调
func (b *Builder) AfterStop(callback func(context.Context) error) *Builder {
	b.app.AfterStop(callback)
	return b
}

// ================================
// 🌐 传输层配置
// ================================

// HTTP 添加HTTP路由处理器
func (b *Builder) HTTP(handler HTTPHandler) *Builder {
	b.httpHandlers = append(b.httpHandlers, handler)
	b.EnableHTTP()
	return b
}

// GRPC 添加gRPC服务处理器
func (b *Builder) GRPC(handler GRPCHandler) *Builder {
	b.grpcHandlers = append(b.grpcHandlers, handler)
	b.EnableGRPC()
	return b
}

// ================================
// 🤖 智能检测
// ================================

// AutoDetect 自动检测环境和配置
func (b *Builder) AutoDetect() *Builder {
	b.autoDetect = true

	// 自动检测端口
	if port := os.Getenv("PORT"); port != "" {
		if p := parseInt(port); p > 0 {
			b.Port(p)
		}
	}

	// 自动检测运行模式
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		b.Mode(mode)
	} else if os.Getenv("ENV") == "production" {
		b.Mode("release")
	} else {
		b.Mode("debug")
	}

	// 自动检测配置文件
	configFiles := []string{
		os.Getenv("CONFIG_PATH"),
		"config/config.yaml",
		"config/config.yml",
		"config.yaml",
		"config.yml",
		"app.yaml",
		"app.yml",
	}

	for _, file := range configFiles {
		if file != "" && fileExists(file) {
			b.Config(file)
			break
		}
	}

	return b
}

// WithEnv 从环境变量设置配置
func (b *Builder) WithEnv() *Builder {
	// 从环境变量读取配置
	if port := getEnvInt("PORT", 0); port > 0 {
		b.Port(port)
	}

	if mode := getEnvString("GIN_MODE", ""); mode != "" {
		b.Mode(mode)
	}

	if cfg := getEnvString("CONFIG_PATH", ""); cfg != "" {
		b.Config(cfg)
	}

	if name := getEnvString("APP_NAME", ""); name != "" {
		b.Name(name)
	}

	if version := getEnvString("APP_VERSION", ""); version != "" {
		b.Version(version)
	}

	// 组件开关
	if getEnvBool("DISABLE_HTTP", false) {
		b.DisableHTTP()
	}

	if getEnvBool("DISABLE_GRPC", false) {
		b.DisableGRPC()
	}

	if getEnvBool("DISABLE_METRICS", false) {
		b.DisableMetrics()
	}

	if getEnvBool("DISABLE_TRACING", false) {
		b.DisableTracing()
	}

	return b
}

// ================================
// 🚀 启动方法
// ================================

// Run 构建并启动应用
func (b *Builder) Run() error {
	// 设置默认值
	b.setupDefaults()

	// 构建应用
	if err := b.build(); err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}

	// 启动应用
	return b.app.Run()
}

// Async 异步启动
func (b *Builder) Async() error {
	go func() {
		if err := b.Run(); err != nil {
			fmt.Printf("Framework error: %v\n", err)
		}
	}()
	return nil
}

// Build 构建应用（不启动）
func (b *Builder) Build() (*App, error) {
	b.setupDefaults()
	if err := b.build(); err != nil {
		return nil, err
	}
	return b.app, nil
}

// ================================
// 🔧 内部方法
// ================================

// setupDefaults 设置默认值
func (b *Builder) setupDefaults() {
	// 如果没有设置任何HTTP处理器，添加默认的健康检查
	if len(b.httpHandlers) == 0 && b.app.opts.EnableHTTP {
		b.HTTP(defaultHTTPHandler)
	}
}

// build 构建应用
func (b *Builder) build() error {

	logger.Info(context.Background(), "🔧 Building app",
		logger.String("name", b.app.opts.Name),
		logger.String("version", b.app.opts.Version))
	logger.Info(context.Background(), "📡 Service configuration",
		logger.Bool("HTTP", b.app.opts.EnableHTTP),
		logger.Bool("gRPC", b.app.opts.EnableGRPC),
		logger.Bool("Metrics", b.app.opts.EnableMetrics),
		logger.Bool("Tracing", b.app.opts.EnableTracing))

	// 初始化组件管理器
	logger.Info(context.Background(), "🔧 Initializing components...")
	if err := b.componentManager.Init(b.app.ctx); err != nil {
		return fmt.Errorf("failed to init components: %w", err)
	}

	// 将组件管理器添加到应用
	b.app.AddComponent(&ComponentWrapper{manager: b.componentManager})

	// 初始化HTTP传输层
	if b.app.opts.EnableHTTP {
		if err := b.setupHTTPTransport(); err != nil {
			return fmt.Errorf("failed to setup HTTP transport: %w", err)
		}
	}

	// 初始化gRPC传输层
	if b.app.opts.EnableGRPC {
		if err := b.setupGRPCTransport(); err != nil {
			return fmt.Errorf("failed to setup gRPC transport: %w", err)
		}
	}

	return nil
}

// setupHTTPTransport 设置HTTP传输层
func (b *Builder) setupHTTPTransport() error {

	// 导入HTTP包
	httpTransport := &HTTPTransport{
		port:     b.app.opts.Port,
		mode:     b.app.opts.Mode,
		handlers: b.httpHandlers,
	}

	b.app.AddTransport(httpTransport)
	logger.Info(context.Background(), "✅ HTTP transport configured")
	return nil
}

// setupGRPCTransport 设置gRPC传输层
func (b *Builder) setupGRPCTransport() error {

	// 将 gRPC 处理器传递给组件管理器
	if len(b.grpcHandlers) > 0 {
		// 转换处理器类型
		var handlers []component.GRPCHandler
		for _, h := range b.grpcHandlers {
			handlers = append(handlers, component.GRPCHandler(h))
		}

		// 设置 gRPC 处理器
		b.componentManager.SetGRPCHandlers(handlers)
	}

	// gRPC已经在组件管理器中处理
	logger.Info(context.Background(), "✅ gRPC transport configured (via component manager)")
	return nil
}

// defaultHTTPHandler 默认HTTP处理器
func defaultHTTPHandler(r interface{}) {

	// 这里会在 transport/http 中实现具体的路由
	logger.Info(context.Background(), "📡 Setting up default HTTP routes...")
}

// ================================
// 🔧 组件访问器 (新增)
// ================================

// GetComponentManager 获取组件管理器
func (b *Builder) GetComponentManager() *component.Manager {
	return b.componentManager
}

// ================================
// 🔧 工具函数
// ================================

// parseInt 安全的字符串转整数
func parseInt(s string) int {
	var result int
	_, _ = fmt.Sscanf(s, "%d", &result)
	return result
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	if filename == "" {
		return false
	}
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// getEnvString 获取环境变量字符串
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取环境变量整数
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvBool 获取环境变量布尔值
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// ================================
// 🔗 组件包装器
// ================================

// ComponentWrapper 组件管理器的包装器，实现Component接口
type ComponentWrapper struct {
	manager *component.Manager
}

func (c *ComponentWrapper) Name() string {
	return "ComponentManager"
}

func (c *ComponentWrapper) Init(ctx context.Context) error {
	return c.manager.Init(ctx)
}

func (c *ComponentWrapper) Start(ctx context.Context) error {
	return c.manager.Start(ctx)
}

func (c *ComponentWrapper) Stop(ctx context.Context) error {
	return c.manager.Stop(ctx)
}
