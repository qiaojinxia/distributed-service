package component

import (
	"context"
	"fmt"

	"github.com/qiaojinxia/distributed-service/framework/auth"
	"github.com/qiaojinxia/distributed-service/framework/cache"
	"github.com/qiaojinxia/distributed-service/framework/common/idgen"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/database"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/middleware"
	"github.com/qiaojinxia/distributed-service/framework/tracing"
	localgrpc "github.com/qiaojinxia/distributed-service/framework/transport/grpc"
	"github.com/qiaojinxia/distributed-service/pkg/etcd"
	"github.com/qiaojinxia/distributed-service/pkg/kafka"
	"github.com/qiaojinxia/distributed-service/pkg/mq"
	"github.com/qiaojinxia/distributed-service/pkg/redis_cluster"
	"github.com/qiaojinxia/distributed-service/pkg/registry"

	"google.golang.org/grpc"
)

// ================================
// 🚀 传输层包装器
// ================================

// GRPCHandler gRPC服务处理器
type GRPCHandler func(interface{})

// IDGenService ID生成器服务接口
type IDGenService interface {
	Initialize(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	NextID(ctx context.Context, bizTag string) (int64, error)
	BatchNextID(ctx context.Context, bizTag string, count int) ([]int64, error)
}

// Manager  组件管理器 - 统一管理所有组件的生命周期
type Manager struct {
	// 核心组件
	config     *config.Config
	auth       *auth.JWTManager
	registry   *registry.ServiceRegistry
	grpcServer *localgrpc.Server

	// 中间件和保护
	protection *middleware.SentinelProtectionMiddleware

	// 监控和追踪
	tracing *tracing.Manager

	// 缓存管理器
	cacheService *cache.FrameworkCacheService

	// ID生成器
	idGenService IDGenService

	// gRPC处理器
	grpcHandlers []GRPCHandler

	// 组件配置
	opts *Options

	// 状态
	initialized bool
	started     bool
}

// Options 组件配置选项
type Options struct {
	// 配置文件
	ConfigPath string

	// 组件开关
	EnableConfig        bool
	EnableLogger        bool
	EnableDatabase      bool
	EnableRedis         bool
	EnableRedisCluster  bool
	EnableAuth          bool
	EnableRegistry      bool
	EnableGRPC          bool
	EnableMQ            bool
	EnableMetrics       bool
	EnableTracing       bool
	EnableProtection    bool
	EnableElasticsearch bool
	EnableKafka         bool
	EnableMongoDB       bool
	EnableEtcd          bool
	EnableCache         bool
	EnableIDGen         bool

	// 组件配置
	DatabaseConfig      *config.MySQLConfig
	RedisConfig         *config.RedisConfig
	RedisClusterConfig  *config.RedisClusterConfig
	AuthConfig          *config.JWTConfig
	RegistryConfig      *config.ConsulConfig
	GRPCConfig          *config.GRPCConfig
	MQConfig            *config.RabbitMQConfig
	MetricsConfig       *config.MetricsConfig
	TracingConfig       *config.TracingConfig
	ProtectionConfig    *config.ProtectionConfig
	LoggerConfig        *config.LoggerConfig
	ElasticsearchConfig *config.ElasticsearchConfig
	KafkaConfig         *config.KafkaConfig
	MongoDBConfig       *config.MongoDBConfig
	EtcdConfig          *config.EtcdConfig
	CacheConfig         *config.CacheConfig
	IDGenConfig         *config.IDGenConfig
}

// Option 组件配置选项函数
type Option func(*Options)

// NewManager 创建组件管理器
func NewManager(opts ...Option) *Manager {
	// 默认配置
	options := &Options{
		ConfigPath:          "config/config.yaml",
		EnableConfig:        true,
		EnableLogger:        true,
		EnableDatabase:      true,
		EnableRedis:         true,
		EnableRedisCluster:  false,
		EnableAuth:          true,
		EnableRegistry:      true,
		EnableGRPC:          true,
		EnableMQ:            true,
		EnableMetrics:       true,
		EnableTracing:       true,
		EnableProtection:    true,
		EnableElasticsearch: false, // 默认禁用，按需启用
		EnableKafka:         false, // 默认禁用，按需启用
		EnableMongoDB:       false, // 默认禁用，按需启用
		EnableEtcd:          false, // 默认禁用，按需启用
		EnableCache:         true,  // 默认启用缓存
		EnableIDGen:         false, // 默认禁用，按需启用
	}

	// 应用选项
	for _, opt := range opts {
		opt(options)
	}

	return &Manager{
		opts: options,
	}
}

// SetGRPCHandlers 设置 gRPC 处理器
func (m *Manager) SetGRPCHandlers(handlers []GRPCHandler) {
	m.grpcHandlers = handlers
}

// ================================
// 🛠️ 配置选项
// ================================

// WithConfig 配置文件选项
func WithConfig(path string) Option {
	return func(o *Options) {
		o.ConfigPath = path
	}
}

// WithDatabase 数据库配置
func WithDatabase(cfg *config.MySQLConfig) Option {
	return func(o *Options) {
		o.DatabaseConfig = cfg
		o.EnableDatabase = true
	}
}

// WithRedis Redis配置
func WithRedis(cfg *config.RedisConfig) Option {
	return func(o *Options) {
		o.RedisConfig = cfg
		o.EnableRedis = true
	}
}

// WithRedisCluster Redis集群配置
func WithRedisCluster(cfg *config.RedisClusterConfig) Option {
	return func(o *Options) {
		o.RedisClusterConfig = cfg
		o.EnableRedisCluster = true
	}
}

// WithAuth 认证配置
func WithAuth(cfg *config.JWTConfig) Option {
	return func(o *Options) {
		o.AuthConfig = cfg
		o.EnableAuth = true
	}
}

// WithRegistry 服务注册配置
func WithRegistry(cfg *config.ConsulConfig) Option {
	return func(o *Options) {
		o.RegistryConfig = cfg
		o.EnableRegistry = true
	}
}

// WithGRPC gRPC配置
func WithGRPC(cfg *config.GRPCConfig) Option {
	return func(o *Options) {
		o.GRPCConfig = cfg
		o.EnableGRPC = true
	}
}

// WithMQ 消息队列配置
func WithMQ(cfg *config.RabbitMQConfig) Option {
	return func(o *Options) {
		o.MQConfig = cfg
		o.EnableMQ = true
	}
}

// WithMetrics 监控配置
func WithMetrics(cfg *config.MetricsConfig) Option {
	return func(o *Options) {
		o.MetricsConfig = cfg
		o.EnableMetrics = true
	}
}

// WithTracing 链路追踪配置
func WithTracing(cfg *config.TracingConfig) Option {
	return func(o *Options) {
		o.TracingConfig = cfg
		o.EnableTracing = true
	}
}

// WithProtection 保护组件配置
func WithProtection(cfg *config.ProtectionConfig) Option {
	return func(o *Options) {
		o.ProtectionConfig = cfg
		o.EnableProtection = true
	}
}

// WithLogger 日志配置
func WithLogger(cfg *config.LoggerConfig) Option {
	return func(o *Options) {
		o.LoggerConfig = cfg
		o.EnableLogger = true
	}
}

// WithElasticsearch Elasticsearch配置
func WithElasticsearch(cfg *config.ElasticsearchConfig) Option {
	return func(o *Options) {
		o.ElasticsearchConfig = cfg
		o.EnableElasticsearch = true
	}
}

// WithKafka Kafka配置
func WithKafka(cfg *config.KafkaConfig) Option {
	return func(o *Options) {
		o.KafkaConfig = cfg
		o.EnableKafka = true
	}
}

// WithMongoDB MongoDB配置
func WithMongoDB(cfg *config.MongoDBConfig) Option {
	return func(o *Options) {
		o.MongoDBConfig = cfg
		o.EnableMongoDB = true
	}
}

// WithEtcd Etcd配置
func WithEtcd(cfg *config.EtcdConfig) Option {
	return func(o *Options) {
		o.EtcdConfig = cfg
		o.EnableEtcd = true
	}
}

// WithCache 缓存配置
func WithCache(cfg *config.CacheConfig) Option {
	return func(o *Options) {
		o.CacheConfig = cfg
		o.EnableCache = true
	}
}

// WithIDGen ID生成器配置
func WithIDGen(cfg *config.IDGenConfig) Option {
	return func(o *Options) {
		o.IDGenConfig = cfg
		o.EnableIDGen = true
	}
}

// DisableComponent 禁用指定组件
func DisableComponent(components ...string) Option {
	return func(o *Options) {
		for _, comp := range components {
			switch comp {
			case "config":
				o.EnableConfig = false
			case "logger":
				o.EnableLogger = false
			case "database":
				o.EnableDatabase = false
			case "redis":
				o.EnableRedis = false
			case "redis_cluster":
				o.EnableRedisCluster = false
			case "auth":
				o.EnableAuth = false
			case "registry":
				o.EnableRegistry = false
			case "grpc":
				o.EnableGRPC = false
			case "mq":
				o.EnableMQ = false
			case "metrics":
				o.EnableMetrics = false
			case "tracing":
				o.EnableTracing = false
			case "protection":
				o.EnableProtection = false
			case "elasticsearch":
				o.EnableElasticsearch = false
			case "kafka":
				o.EnableKafka = false
			case "mongodb":
				o.EnableMongoDB = false
			case "etcd":
				o.EnableEtcd = false
			case "cache":
				o.EnableCache = false
			case "idgen":
				o.EnableIDGen = false
			}
		}
	}
}

// ================================
// 🔄 生命周期管理
// ================================

// Init 初始化所有启用的组件
func (m *Manager) Init(ctx context.Context) error {
	if m.initialized {
		return nil
	}

	logger.Info(ctx, "🔧 Initializing components...")

	// 1. 初始化配置
	if m.opts.EnableConfig {
		if err := m.initConfig(ctx); err != nil {
			return fmt.Errorf("failed to init config: %w", err)
		}
	}

	// 2. 初始化日志
	if m.opts.EnableLogger {
		if err := m.initLogger(ctx); err != nil {
			return fmt.Errorf("failed to init logger: %w", err)
		}
	}

	// 3. 初始化数据库
	if m.opts.EnableDatabase {
		if err := m.initDatabase(ctx); err != nil {
			return fmt.Errorf("failed to init database: %w", err)
		}
	}

	// 4. 初始化Redis
	if m.opts.EnableRedis {
		if err := m.initRedis(ctx); err != nil {
			return fmt.Errorf("failed to init redis: %w", err)
		}
	}

	// 5. 初始化Redis集群
	if m.opts.EnableRedisCluster {
		if err := m.initRedisCluster(ctx); err != nil {
			return fmt.Errorf("failed to init redis cluster: %w", err)
		}
	}

	// 6. 初始化认证
	if m.opts.EnableAuth {
		if err := m.initAuth(ctx); err != nil {
			return fmt.Errorf("failed to init auth: %w", err)
		}
	}

	// 7. 初始化链路追踪
	if m.opts.EnableTracing {
		if err := m.initTracing(ctx); err != nil {
			return fmt.Errorf("failed to init tracing: %w", err)
		}
	}

	// 8. 初始化监控指标
	if m.opts.EnableMetrics {
		if err := m.initMetrics(ctx); err != nil {
			return fmt.Errorf("failed to init metrics: %w", err)
		}
	}

	// 9. 初始化保护组件
	if m.opts.EnableProtection {
		if err := m.initProtection(ctx); err != nil {
			return fmt.Errorf("failed to init protection: %w", err)
		}
	}

	// 10. 初始化消息队列
	if m.opts.EnableMQ {
		if err := m.initMQ(ctx); err != nil {
			return fmt.Errorf("failed to init mq: %w", err)
		}
	}

	// 11. 初始化Kafka
	if m.opts.EnableKafka {
		if err := m.initKafka(ctx); err != nil {
			return fmt.Errorf("failed to init kafka: %w", err)
		}
	}

	// 12. 初始化Etcd
	if m.opts.EnableEtcd {
		if err := m.initEtcd(ctx); err != nil {
			return fmt.Errorf("failed to init etcd: %w", err)
		}
	}

	// 13. 初始化服务注册
	if m.opts.EnableRegistry {
		if err := m.initRegistry(ctx); err != nil {
			return fmt.Errorf("failed to init registry: %w", err)
		}
	}

	// 14. 初始化gRPC服务器
	if m.opts.EnableGRPC {
		if err := m.initGRPCServer(ctx); err != nil {
			return fmt.Errorf("failed to init grpc: %w", err)
		}
	}

	// 15. 初始化Elasticsearch
	if m.opts.EnableElasticsearch {
		if err := m.initElasticsearch(ctx); err != nil {
			return fmt.Errorf("failed to init elasticsearch: %w", err)
		}
	}

	// 16. 初始化MongoDB
	if m.opts.EnableMongoDB {
		if err := m.initMongoDB(ctx); err != nil {
			return fmt.Errorf("failed to init mongodb: %w", err)
		}
	}

	// 17. 初始化缓存
	if m.opts.EnableCache {
		if err := m.initCache(ctx); err != nil {
			return fmt.Errorf("failed to init cache: %w", err)
		}
	}

	// 18. 初始化ID生成器
	if m.opts.EnableIDGen {
		if err := m.initIDGen(ctx); err != nil {
			return fmt.Errorf("failed to init idgen: %w", err)
		}
	}

	m.initialized = true
	logger.Info(ctx, "✅ All components initialized")
	return nil
}

// Start 启动所有组件
func (m *Manager) Start(ctx context.Context) error {
	if !m.initialized {
		return fmt.Errorf("components not initialized")
	}

	if m.started {
		return nil
	}

	logger.Info(ctx, "🚀 Starting components...")

	// 启动各个组件
	if m.grpcServer != nil {
		// 在启动 gRPC 服务器之前注册用户提供的服务
		if len(m.grpcHandlers) > 0 {
			logger.Info(ctx, "🔌 Registering gRPC services...", logger.Int("handler_count", len(m.grpcHandlers)))
			for i, handler := range m.grpcHandlers {
				logger.Info(ctx, "  📝 Calling gRPC handler", logger.Int("handler_index", i+1))
				handler(m.grpcServer)
			}
			logger.Info(ctx, "✅ All gRPC services registered")
		} else {
			logger.Warn(ctx, "⚠️ No gRPC handlers found - no services will be registered")
		}

		if err := m.grpcServer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start grpc server: %w", err)
		}
	}

	// 启动ID生成器
	if m.idGenService != nil {
		if err := m.idGenService.Start(ctx); err != nil {
			return fmt.Errorf("failed to start idgen service: %w", err)
		}
	}

	m.started = true
	logger.Info(ctx, "✅ All components started")
	return nil
}

// Stop 停止所有组件
func (m *Manager) Stop(ctx context.Context) error {
	if !m.started {
		return nil
	}

	logger.Info(ctx, "🛑 Stopping components...")

	// 停止各个组件
	if m.grpcServer != nil {
		if err := m.grpcServer.Stop(ctx); err != nil {
			logger.Error(ctx, "Failed to stop grpc server", logger.Err(err))
		}
	}

	if m.tracing != nil {
		if err := m.tracing.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Failed to stop tracing", logger.Err(err))
		}
	}

	if m.cacheService != nil {
		if err := m.cacheService.Close(); err != nil {
			logger.Error(ctx, "Failed to stop cache service", logger.Err(err))
		}
	}

	if m.idGenService != nil {
		if err := m.idGenService.Stop(ctx); err != nil {
			logger.Error(ctx, "Failed to stop idgen service", logger.Err(err))
		}
	}

	m.started = false
	logger.Info(ctx, "✅ All components stopped")
	return nil
}

// ================================
// 🔧 组件初始化方法
// ================================

// initConfig 初始化配置
func (m *Manager) initConfig(ctx context.Context) error {
	if err := config.LoadConfig(m.opts.ConfigPath); err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}
	m.config = &config.GlobalConfig

	logger.Info(ctx, "✅ Config loaded")
	return nil
}

// initLogger 初始化日志
func (m *Manager) initLogger(ctx context.Context) error {
	var cfg *logger.Config
	if m.opts.LoggerConfig != nil {
		cfg = &logger.Config{
			Level:      m.opts.LoggerConfig.Level,
			Encoding:   m.opts.LoggerConfig.Encoding,
			OutputPath: m.opts.LoggerConfig.OutputPath,
		}
	} else if m.config != nil {
		cfg = &logger.Config{
			Level:      m.config.Logger.Level,
			Encoding:   m.config.Logger.Encoding,
			OutputPath: m.config.Logger.OutputPath,
		}
	} else {
		cfg = &logger.Config{
			Level:      "info",
			Encoding:   "console",
			OutputPath: "stdout",
		}
	}

	err := logger.InitLogger(cfg)
	if err != nil {
		return err
	}

	logger.Info(ctx, "✅ Logger initialized")
	return nil
}

// initDatabase 初始化数据库
func (m *Manager) initDatabase(ctx context.Context) error {
	var cfg *config.MySQLConfig
	if m.opts.DatabaseConfig != nil {
		cfg = m.opts.DatabaseConfig
	} else if m.config != nil {
		cfg = &m.config.MySQL
	} else {
		return fmt.Errorf("database config not found")
	}

	if err := database.InitMySQL(ctx, cfg); err != nil {
		return err
	}
	logger.Info(ctx, "✅ Database initialized")
	return nil
}

// initRedis 初始化Redis
func (m *Manager) initRedis(ctx context.Context) error {
	var cfg *config.RedisConfig
	if m.opts.RedisConfig != nil {
		cfg = m.opts.RedisConfig
	} else if m.config != nil {
		cfg = &m.config.Redis
	} else {
		return fmt.Errorf("redis config not found")
	}

	if err := database.InitRedis(ctx, cfg); err != nil {
		return err
	}

	logger.Info(ctx, "✅ Redis initialized")
	return nil
}

// initRedisCluster 初始化Redis集群
func (m *Manager) initRedisCluster(ctx context.Context) error {
	var cfg *config.RedisClusterConfig
	if m.opts.RedisClusterConfig != nil {
		cfg = m.opts.RedisClusterConfig
	} else if m.config != nil {
		cfg = &m.config.RedisCluster
	} else {
		return fmt.Errorf("redis cluster config not found")
	}

	clusterCfg, err := redis_cluster.ConvertConfig(cfg)
	if err != nil {
		return err
	}

	if err := redis_cluster.InitRedisCluster(ctx, clusterCfg); err != nil {
		return err
	}

	logger.Info(ctx, "✅ Redis Cluster initialized")
	return nil
}

// initAuth 初始化认证
func (m *Manager) initAuth(ctx context.Context) error {
	var secretKey, issuer string
	if m.opts.AuthConfig != nil {
		secretKey = m.opts.AuthConfig.SecretKey
		issuer = m.opts.AuthConfig.Issuer
	} else if m.config != nil {
		secretKey = m.config.JWT.SecretKey
		issuer = m.config.JWT.Issuer
	} else {
		secretKey = "default-secret-key"
		issuer = "distributed-service"
	}

	m.auth = auth.NewJWTManager(secretKey, issuer)

	logger.Info(ctx, "✅ Auth initialized")
	return nil
}

// initTracing 初始化链路追踪
func (m *Manager) initTracing(ctx context.Context) error {
	var cfg *tracing.Config
	if m.opts.TracingConfig != nil {
		cfg = &tracing.Config{
			ServiceName:    m.opts.TracingConfig.ServiceName,
			ServiceVersion: m.opts.TracingConfig.ServiceVersion,
			Environment:    m.opts.TracingConfig.Environment,
			Enabled:        m.opts.TracingConfig.Enabled,
			ExporterType:   m.opts.TracingConfig.ExporterType,
			Endpoint:       m.opts.TracingConfig.Endpoint,
			SampleRatio:    m.opts.TracingConfig.SampleRatio,
		}
	} else if m.config != nil {
		cfg = &tracing.Config{
			ServiceName:    m.config.Tracing.ServiceName,
			ServiceVersion: m.config.Tracing.ServiceVersion,
			Environment:    m.config.Tracing.Environment,
			Enabled:        m.config.Tracing.Enabled,
			ExporterType:   m.config.Tracing.ExporterType,
			Endpoint:       m.config.Tracing.Endpoint,
			SampleRatio:    m.config.Tracing.SampleRatio,
		}
	} else {
		cfg = &tracing.Config{
			ServiceName:    "distributed-service",
			ServiceVersion: "v1.0.0",
			Environment:    "development",
			Enabled:        true,
			ExporterType:   "jaeger",
			SampleRatio:    1.0,
		}
	}

	tracingManager, err := tracing.NewTracingManager(ctx, cfg)
	if err != nil {
		return err
	}
	m.tracing = tracingManager

	logger.Info(ctx, "✅ Tracing initialized")
	return nil
}

// initMetrics 初始化监控指标
func (m *Manager) initMetrics(ctx context.Context) error {
	// 监控指标通常在框架级别自动初始化

	logger.Info(ctx, "✅ Metrics initialized")
	return nil
}

// initProtection 初始化保护组件
func (m *Manager) initProtection(ctx context.Context) error {
	var cfg *config.ProtectionConfig
	if m.opts.ProtectionConfig != nil {
		cfg = m.opts.ProtectionConfig
	} else if m.config != nil {
		cfg = &m.config.Protection
	} else {
		return fmt.Errorf("protection config not found")
	}

	protectionMiddleware, err := middleware.NewSentinelProtectionMiddleware(ctx, cfg)
	if err != nil {
		return err
	}
	m.protection = protectionMiddleware

	logger.Info(ctx, "✅ Protection initialized")
	return nil
}

// initMQ 初始化消息队列
func (m *Manager) initMQ(ctx context.Context) error {
	var cfg *config.RabbitMQConfig
	if m.opts.MQConfig != nil {
		cfg = m.opts.MQConfig
	} else if m.config != nil {
		cfg = &m.config.RabbitMQ
	} else {
		return fmt.Errorf("mq config not found")
	}

	if err := mq.InitRabbitMQ(ctx, cfg); err != nil {
		return err
	}

	logger.Info(ctx, "✅ Message Queue initialized")
	return nil
}

// initRegistry 初始化服务注册
func (m *Manager) initRegistry(ctx context.Context) error {
	var cfg *config.ConsulConfig
	if m.opts.RegistryConfig != nil {
		cfg = m.opts.RegistryConfig
	} else if m.config != nil {
		cfg = &m.config.Consul
	} else {
		return fmt.Errorf("registry config not found")
	}

	registryInstance, err := registry.NewServiceRegistry(ctx, cfg)
	if err != nil {
		return err
	}
	m.registry = registryInstance

	logger.Info(ctx, "✅ Service Registry initialized")
	return nil
}

// initGRPCServer 初始化gRPC服务器
func (m *Manager) initGRPCServer(ctx context.Context) error {
	var cfg *localgrpc.Config
	if m.opts.GRPCConfig != nil {
		convertedCfg, err := localgrpc.ConvertConfig(m.opts.GRPCConfig)
		if err != nil {
			return err
		}
		cfg = convertedCfg
	} else if m.config != nil {
		convertedCfg, err := localgrpc.ConvertConfig(&m.config.GRPC)
		if err != nil {
			return err
		}
		cfg = convertedCfg
	} else {
		// 默认配置
		defaultConfig := &config.GRPCConfig{
			Port: 9000,
		}
		convertedCfg, err := localgrpc.ConvertConfig(defaultConfig)
		if err != nil {
			return err
		}
		cfg = convertedCfg
	}

	// 创建拦截器链
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// 添加基础中间件
	unaryInterceptors = append(unaryInterceptors,
		middleware.GRPCRecoveryInterceptor(),
		middleware.GRPCLoggingInterceptor(),
		middleware.GRPCMetricsInterceptor(),
	)
	streamInterceptors = append(streamInterceptors,
		middleware.GRPCStreamRecoveryInterceptor(),
		middleware.GRPCStreamLoggingInterceptor(),
	)

	// 添加保护中间件
	if m.protection != nil && m.protection.IsEnabled() {
		unaryInterceptors = append(unaryInterceptors, m.protection.GRPCUnaryInterceptor())
		streamInterceptors = append(streamInterceptors, m.protection.GRPCStreamInterceptor())
	}

	// 添加链路追踪中间件
	if m.tracing != nil {
		unaryInterceptors = append(unaryInterceptors, middleware.GRPCTracingInterceptor())
		streamInterceptors = append(streamInterceptors, middleware.GRPCStreamTracingInterceptor())
	}

	grpcSrv, err := localgrpc.NewServerWithInterceptors(ctx, cfg, unaryInterceptors, streamInterceptors)
	if err != nil {
		return err
	}
	m.grpcServer = grpcSrv

	logger.Info(ctx, "✅ gRPC Server initialized")
	return nil
}

// initElasticsearch 初始化Elasticsearch
func (m *Manager) initElasticsearch(ctx context.Context) error {
	var cfg *config.ElasticsearchConfig
	if m.opts.ElasticsearchConfig != nil {
		cfg = m.opts.ElasticsearchConfig
	} else if m.config != nil {
		cfg = &m.config.Elasticsearch
	} else {
		return fmt.Errorf("elasticsearch config not found")
	}

	// 这里可以初始化Elasticsearch
	_ = cfg // 暂时忽略配置

	logger.Info(ctx, "✅ Elasticsearch initialized")
	return nil
}

// initKafka 初始化Kafka
func (m *Manager) initKafka(ctx context.Context) error {
	var cfg *config.KafkaConfig
	if m.opts.KafkaConfig != nil {
		cfg = m.opts.KafkaConfig
	} else if m.config != nil {
		cfg = &m.config.Kafka
	} else {
		return fmt.Errorf("kafka config not found")
	}

	kafkaCfg, err := kafka.ConvertConfig(cfg)
	if err != nil {
		return err
	}

	if err := kafka.InitKafka(ctx, kafkaCfg); err != nil {
		return err
	}

	logger.Info(ctx, "✅ Kafka initialized")
	return nil
}

// initMongoDB 初始化MongoDB
func (m *Manager) initMongoDB(ctx context.Context) error {
	var cfg *config.MongoDBConfig
	if m.opts.MongoDBConfig != nil {
		cfg = m.opts.MongoDBConfig
	} else if m.config != nil {
		cfg = &m.config.MongoDB
	} else {
		return fmt.Errorf("mongodb config not found")
	}

	// 这里可以初始化MongoDB
	_ = cfg // 暂时忽略配置

	logger.Info(ctx, "✅ MongoDB initialized")
	return nil
}

// initEtcd 初始化Etcd
func (m *Manager) initEtcd(ctx context.Context) error {
	var cfg *config.EtcdConfig
	if m.opts.EtcdConfig != nil {
		cfg = m.opts.EtcdConfig
	} else if m.config != nil {
		cfg = &m.config.Etcd
	} else {
		return fmt.Errorf("etcd config not found")
	}

	etcdCfg, err := etcd.ConvertConfig(cfg)
	if err != nil {
		return err
	}

	if err := etcd.InitEtcd(ctx, etcdCfg); err != nil {
		return err
	}

	logger.Info(ctx, "✅ Etcd initialized")
	return nil
}

// initCache 初始化缓存
func (m *Manager) initCache(ctx context.Context) error {
	// 初始化缓存服务
	m.cacheService = cache.NewFrameworkCacheService()

	// 初始化缓存服务
	if err := m.cacheService.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize cache service: %w", err)
	}

	// 如果有配置，创建预定义的缓存实例
	var cacheConfig *config.CacheConfig
	if m.opts.CacheConfig != nil {
		cacheConfig = m.opts.CacheConfig
	} else if m.config != nil {
		cacheConfig = &m.config.Cache
	}

	if cacheConfig != nil && cacheConfig.Enabled {
		// 根据配置创建缓存实例
		for name, instanceCfg := range cacheConfig.Caches {
			if err := m.createCacheFromConfig(name, instanceCfg, cacheConfig); err != nil {
				logger.Error(ctx, "Failed to create cache instance", 
					logger.String("name", name), 
					logger.Err(err))
				continue
			}
			logger.Info(ctx, "✅ Cache instance created", 
				logger.String("name", name), 
				logger.String("type", instanceCfg.Type))
		}
	}

	logger.Info(ctx, "✅ Cache initialized")
	return nil
}

// createCacheFromConfig 根据配置创建缓存实例
func (m *Manager) createCacheFromConfig(name string, instanceCfg config.CacheInstance, globalCfg *config.CacheConfig) error {
	keyPrefix := instanceCfg.KeyPrefix
	if keyPrefix == "" && globalCfg.GlobalKeyPrefix != "" {
		keyPrefix = globalCfg.GlobalKeyPrefix + ":" + name
	}

	switch instanceCfg.Type {
	case "memory":
		// 创建内存缓存配置
		memoryConfig := cache.Config{
			Type: cache.TypeMemory,
			Name: name,
			Settings: instanceCfg.Settings,
		}
		
		// 确保有默认TTL
		if memoryConfig.Settings == nil {
			memoryConfig.Settings = make(map[string]interface{})
		}
		if _, exists := memoryConfig.Settings["default_ttl"]; !exists && instanceCfg.TTL != "" {
			memoryConfig.Settings["default_ttl"] = instanceCfg.TTL
		}
		
		return m.cacheService.Manager.CreateCache(memoryConfig)

	case "redis":
		// 创建Redis缓存
		return m.cacheService.CreateRedisCache(name, keyPrefix)

	case "hybrid":
		// 创建混合缓存
		l1Config := cache.Config{
			Type: cache.TypeMemory,
			Name: "l1-" + name,
			Settings: map[string]interface{}{
				"max_size":    1000,
				"default_ttl": "5m",
			},
		}
		return m.cacheService.CreateHybridCache(name, l1Config, keyPrefix, cache.SyncStrategyWriteThrough)

	default:
		return fmt.Errorf("unsupported cache type: %s", instanceCfg.Type)
	}
}

// initIDGen 初始化ID生成器
func (m *Manager) initIDGen(ctx context.Context) error {
	logger.Info(ctx, "🆔 Initializing ID generator...")

	// 优先使用选项配置
	var idGenCfg *config.IDGenConfig
	if m.opts.IDGenConfig != nil {
		idGenCfg = m.opts.IDGenConfig
	} else if m.config != nil {
		idGenCfg = &m.config.IDGen
	} else {
		// 使用默认配置
		idGenCfg = &config.IDGenConfig{
			Enabled:      true,
			Type:         "leaf",
			UseFramework: true,
			DefaultStep:  1000,
		}
	}

	// 检查是否启用
	if !idGenCfg.Enabled {
		logger.Info(ctx, "ID generator is disabled, skipping initialization")
		return nil
	}

	// 创建框架ID生成器服务
	m.idGenService = idgen.NewFrameworkIDGenService()

	// 初始化服务
	if err := m.idGenService.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize ID generator service: %w", err)
	}

	logger.Info(ctx, "✅ ID generator initialized")
	return nil
}

// ================================
// 🔍 组件访问器
// ================================

// GetConfig 获取配置
func (m *Manager) GetConfig() *config.Config {
	return m.config
}

// GetAuth 获取认证管理器
func (m *Manager) GetAuth() *auth.JWTManager {
	return m.auth
}

// GetRegistry 获取服务注册器
func (m *Manager) GetRegistry() *registry.ServiceRegistry {
	return m.registry
}

// GetGRPCServer 获取gRPC服务器
func (m *Manager) GetGRPCServer() *localgrpc.Server {
	return m.grpcServer
}

// GetProtection 获取保护中间件
func (m *Manager) GetProtection() *middleware.SentinelProtectionMiddleware {
	return m.protection
}

// GetTracing 获取链路追踪
func (m *Manager) GetTracing() *tracing.Manager {
	return m.tracing
}

// GetCacheService 获取缓存服务
func (m *Manager) GetCacheService() *cache.FrameworkCacheService {
	return m.cacheService
}

// GetIDGenService 获取ID生成器服务
func (m *Manager) GetIDGenService() IDGenService {
	return m.idGenService
}

// IsInitialized 检查是否已初始化
func (m *Manager) IsInitialized() bool {
	return m.initialized
}

// IsStarted 检查是否已启动
func (m *Manager) IsStarted() bool {
	return m.started
}
