package main

import (
	"context"
	"distributed-service/framework"
	"distributed-service/framework/config"
	"log"
)

func main() {
	log.Println("🧩 启动组件化框架示例...")

	// 使用选项模式配置各种组件
	err := framework.New().
		Port(8080).                   // 基础配置
		Name("component-demo").       // 应用名称
		Version("v1.0.0").            // 应用版本
		Mode("debug").                // 开发模式
		Config("config/config.yaml"). // 配置文件

		// 🗄️ 数据库组件配置
		WithDatabase(&config.MySQLConfig{
			Host:         "localhost",
			Port:         3306,
			Username:     "root",
			Password:     "password",
			Database:     "demo_db",
			Charset:      "utf8mb4",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		}).

		// 🗃️ Redis组件配置
		WithRedis(&config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		}).

		// 🔐 JWT认证组件配置
		WithAuth(&config.JWTConfig{
			SecretKey: "your-super-secret-key",
			Issuer:    "component-demo",
		}).

		// 📡 服务注册组件配置
		WithRegistry(&config.ConsulConfig{
			Host:                           "localhost",
			Port:                           8500,
			ServiceCheckInterval:           "10s",
			DeregisterCriticalServiceAfter: "30s",
		}).

		// 🔌 gRPC组件配置
		WithGRPCConfig(&config.GRPCConfig{
			Port:              9000,
			MaxRecvMsgSize:    4 << 20, // 4MB
			MaxSendMsgSize:    4 << 20, // 4MB
			ConnectionTimeout: "30s",
			MaxConnectionIdle: "60s",
			EnableReflection:  true,
			EnableHealthCheck: true,
		}).

		// 🐰 消息队列组件配置
		WithMQ(&config.RabbitMQConfig{
			Host:     "localhost",
			Port:     5672,
			Username: "guest",
			Password: "guest",
			VHost:    "/",
		}).

		// 📊 监控组件配置
		WithMetrics(&config.MetricsConfig{
			Enabled:        true,
			PrometheusPort: 9090,
		}).

		// 🔍 链路追踪组件配置
		WithTracing(&config.TracingConfig{
			ServiceName:    "component-demo",
			ServiceVersion: "v1.0.0",
			Environment:    "development",
			Enabled:        true,
			ExporterType:   "jaeger",
			Endpoint:       "http://localhost:14268/api/traces",
			SampleRatio:    1.0,
		}).

		// 🛡️ 保护组件配置
		WithProtection(&config.ProtectionConfig{
			Enabled: true,
			Storage: config.ProtectionStorageConfig{
				Type:   "memory",
				Prefix: "demo:",
				TTL:    "300s",
			},
			RateLimitRules: []config.RateLimitRuleConfig{
				{
					Name:           "api-rate-limit",
					Resource:       "api",
					Threshold:      100,
					StatIntervalMs: 1000,
					Enabled:        true,
					Description:    "API接口限流",
				},
			},
		}).

		// 📝 日志组件配置
		WithLogger(&config.LoggerConfig{
			Level:      "info",
			Encoding:   "console",
			OutputPath: "stdout",
		}).

		// 🌐 HTTP路由配置
		HTTP(setupHTTPRoutes).

		// 🔌 gRPC服务配置
		GRPC(setupGRPCServices).

		// 🔄 生命周期回调
		BeforeStart(func(ctx context.Context) error {
			log.Println("🔧 执行启动前初始化...")
			log.Println("  - 检查数据库连接")
			log.Println("  - 初始化缓存")
			log.Println("  - 准备消息队列")
			return nil
		}).
		AfterStart(func(ctx context.Context) error {
			log.Println("✅ 所有组件启动完成!")
			log.Println("📍 HTTP服务: http://localhost:8080")
			log.Println("📍 gRPC服务: localhost:9000")
			log.Println("📍 监控指标: http://localhost:9090/metrics")
			log.Println("📍 健康检查: http://localhost:8080/health")

			// 展示组件访问
			log.Println("🧩 可访问的组件:")
			log.Println("  - 数据库: ✅ MySQL + Redis")
			log.Println("  - 认证: ✅ JWT Manager")
			log.Println("  - 注册: ✅ Consul Registry")
			log.Println("  - 消息: ✅ RabbitMQ")
			log.Println("  - 监控: ✅ Prometheus Metrics")
			log.Println("  - 追踪: ✅ Jaeger Tracing")
			log.Println("  - 保护: ✅ Sentinel Protection")

			return nil
		}).
		BeforeStop(func(ctx context.Context) error {
			log.Println("🧹 执行停止前清理...")
			log.Println("  - 关闭数据库连接")
			log.Println("  - 清理缓存")
			log.Println("  - 关闭消息队列")
			return nil
		}).

		// 🚀 启动框架
		Run()

	if err != nil {
		log.Fatalf("组件化框架启动失败: %v", err)
	}
}

// setupHTTPRoutes 设置HTTP路由
func setupHTTPRoutes(r interface{}) {
	log.Println("📡 设置HTTP路由...")
	log.Println("  - GET  /health        健康检查")
	log.Println("  - GET  /info          服务信息")
	log.Println("  - POST /api/auth      用户认证")
	log.Println("  - GET  /api/users     用户列表")
	log.Println("  - GET  /metrics       监控指标")
}

// setupGRPCServices 设置gRPC服务
func setupGRPCServices(s interface{}) {
	log.Println("🔌 注册gRPC服务...")
	log.Println("  - UserService         用户服务")
	log.Println("  - OrderService        订单服务")
	log.Println("  - NotificationService 通知服务")
}
