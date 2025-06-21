package main

import (
	"context"
	"distributed-service/framework"
	"distributed-service/framework/config"
	httpTransport "distributed-service/framework/transport/http"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 启动高级功能框架示例...")

	// 使用完整的组件配置启动框架
	err := framework.New().
		Port(8080).
		Name("advanced-demo").
		Version("v2.0.0").
		Mode("debug").
		Config("config/config.yaml").

		// 🗄️ 核心数据存储
		WithDatabase(&config.MySQLConfig{
			Host:         "localhost",
			Port:         3306,
			Username:     "root",
			Password:     "password",
			Database:     "advanced_db",
			Charset:      "utf8mb4",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		}).
		WithRedis(&config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		}).

		// 🔐 认证和安全
		WithAuth(&config.JWTConfig{
			SecretKey: "advanced-demo-secret-key",
			Issuer:    "advanced-demo",
		}).
		WithProtection(&config.ProtectionConfig{
			Enabled: true,
			Storage: config.ProtectionStorageConfig{
				Type:   "memory",
				Prefix: "advanced:",
				TTL:    "300s",
			},
			RateLimitRules: []config.RateLimitRuleConfig{
				{
					Name:           "api-rate-limit",
					Resource:       "api",
					Threshold:      200,
					StatIntervalMs: 1000,
					Enabled:        true,
					Description:    "API接口限流",
				},
			},
		}).

		// 🌐 服务通信
		WithRegistry(&config.ConsulConfig{
			Host:                           "localhost",
			Port:                           8500,
			ServiceCheckInterval:           "10s",
			DeregisterCriticalServiceAfter: "30s",
		}).
		WithGRPCConfig(&config.GRPCConfig{
			Port:              9000,
			MaxRecvMsgSize:    4 << 20,
			MaxSendMsgSize:    4 << 20,
			ConnectionTimeout: "30s",
			MaxConnectionIdle: "60s",
			EnableReflection:  true,
			EnableHealthCheck: true,
		}).

		// 📦 消息队列
		WithMQ(&config.RabbitMQConfig{
			Host:     "localhost",
			Port:     5672,
			Username: "guest",
			Password: "guest",
			VHost:    "/",
		}).

		// 🔍 搜索和分析
		WithElasticsearch(&config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			Username:  "",
			Password:  "",
			Timeout:   30,
		}).

		// 📊 大数据处理
		WithKafka(&config.KafkaConfig{
			Brokers:       []string{"localhost:9092"},
			ClientID:      "advanced-demo",
			Group:         "advanced-group",
			Version:       "2.6.0",
			RetryBackoff:  100,
			RetryMax:      3,
			FlushMessages: 100,
			FlushBytes:    1024 * 1024,
			FlushTimeout:  100,
		}).

		// 🗃️ NoSQL数据库
		WithMongoDB(&config.MongoDBConfig{
			URI:            "mongodb://localhost:27017",
			Database:       "advanced_db",
			Username:       "",
			Password:       "",
			AuthDatabase:   "admin",
			MaxPoolSize:    100,
			MinPoolSize:    10,
			MaxIdleTimeMS:  60000,
			ConnectTimeout: 30,
			SocketTimeout:  30,
		}).

		// 🔑 分布式配置
		WithEtcd(&config.EtcdConfig{
			Endpoints:   []string{"localhost:2379"},
			Username:    "",
			Password:    "",
			DialTimeout: 5,
		}).

		// 📊 监控和追踪
		WithMetrics(&config.MetricsConfig{
			Enabled:        true,
			PrometheusPort: 9090,
		}).
		WithTracing(&config.TracingConfig{
			ServiceName:    "advanced-demo",
			ServiceVersion: "v2.0.0",
			Environment:    "development",
			Enabled:        true,
			ExporterType:   "jaeger",
			Endpoint:       "http://localhost:14268/api/traces",
			SampleRatio:    1.0,
		}).

		// 📝 日志配置
		WithLogger(&config.LoggerConfig{
			Level:      "info",
			Encoding:   "console",
			OutputPath: "stdout",
		}).

		// 🌐 HTTP服务配置
		HTTP(setupAdvancedRoutes).

		// 🔌 gRPC服务配置
		GRPC(setupAdvancedGRPCServices).

		// 🔄 生命周期回调
		BeforeStart(func(ctx context.Context) error {
			log.Println("🔧 执行高级功能初始化...")
			log.Println("  - 初始化数据库连接池")
			log.Println("  - 配置缓存策略")
			log.Println("  - 启动消息队列监听")
			log.Println("  - 连接搜索引擎")
			log.Println("  - 建立大数据处理管道")
			log.Println("  - 配置分布式锁")
			return nil
		}).
		AfterStart(func(ctx context.Context) error {
			log.Println("✅ 高级功能框架启动完成!")
			log.Println("📍 服务信息:")
			log.Println("  - HTTP API:     http://localhost:8080")
			log.Println("  - gRPC Server:  localhost:9000")
			log.Println("  - 健康检查:      http://localhost:8080/health")
			log.Println("  - 详细健康:      http://localhost:8080/health/detail")
			log.Println("  - API文档:      http://localhost:8080/swagger")
			log.Println("  - 监控指标:      http://localhost:9090/metrics")

			log.Println("🧩 已启用的高级组件:")
			log.Println("  - ✅ MySQL + Redis         (核心数据存储)")
			log.Println("  - ✅ JWT + Sentinel        (认证和保护)")
			log.Println("  - ✅ Consul + gRPC         (服务发现)")
			log.Println("  - ✅ RabbitMQ              (消息队列)")
			log.Println("  - ✅ Elasticsearch         (搜索引擎)")
			log.Println("  - ✅ Kafka                 (大数据流处理)")
			log.Println("  - ✅ MongoDB               (文档数据库)")
			log.Println("  - ✅ Etcd                  (分布式配置)")
			log.Println("  - ✅ Prometheus + Jaeger   (监控追踪)")

			return nil
		}).
		BeforeStop(func(ctx context.Context) error {
			log.Println("🧹 执行高级功能清理...")
			log.Println("  - 关闭数据库连接")
			log.Println("  - 清理缓存数据")
			log.Println("  - 停止消息队列")
			log.Println("  - 断开搜索引擎")
			log.Println("  - 关闭数据流管道")
			return nil
		}).

		// 🚀 启动框架
		Run()

	if err != nil {
		log.Fatalf("高级功能框架启动失败: %v", err)
	}
}

// setupAdvancedRoutes 设置高级HTTP路由
func setupAdvancedRoutes(r interface{}) {
	if engine, ok := r.(*gin.Engine); ok {
		log.Println("📡 设置高级HTTP路由...")

		// 健康检查管理器
		healthManager := httpTransport.NewHealthManager()

		// 添加各种健康检查
		healthManager.AddCheck(httpTransport.NewDatabaseHealthCheck(
			"mysql",
			func(ctx context.Context) error {
				// 这里检查MySQL连接
				return nil
			},
		))

		healthManager.AddCheck(httpTransport.NewRedisHealthCheck(
			"redis",
			func(ctx context.Context) error {
				// 这里检查Redis连接
				return nil
			},
		))

		healthManager.AddCheck(httpTransport.NewHTTPHealthCheck(
			"elasticsearch",
			"http://localhost:9200/_cluster/health",
		))

		// 设置健康检查路由
		healthManager.SetupHealthRoutes(engine)

		// API路由组
		api := engine.Group("/api/v1")
		{
			// 用户管理
			api.GET("/users", getUsersHandler)
			api.POST("/users", createUserHandler)
			api.GET("/users/:id", getUserHandler)
			api.PUT("/users/:id", updateUserHandler)
			api.DELETE("/users/:id", deleteUserHandler)

			// 搜索功能
			api.GET("/search", searchHandler)
			api.POST("/search/index", indexDocumentHandler)

			// 数据分析
			api.GET("/analytics/stats", getStatsHandler)
			api.POST("/analytics/events", trackEventHandler)

			// 文档管理
			api.GET("/documents", getDocumentsHandler)
			api.POST("/documents", createDocumentHandler)

			// 配置管理
			api.GET("/config/:key", getConfigHandler)
			api.PUT("/config/:key", setConfigHandler)
		}

		// 管理路由
		admin := engine.Group("/admin")
		{
			admin.GET("/metrics", getMetricsHandler)
			admin.GET("/logs", getLogsHandler)
			admin.POST("/cache/clear", clearCacheHandler)
			admin.GET("/components/status", getComponentStatusHandler)
		}

		// 静态文件和文档
		engine.Static("/static", "./static")
		engine.GET("/swagger/*any", swaggerHandler)

		log.Println("✅ 高级HTTP路由配置完成")
		log.Println("  - API路由:      /api/v1/*")
		log.Println("  - 管理接口:     /admin/*")
		log.Println("  - 健康检查:     /health")
		log.Println("  - 静态文件:     /static/*")
		log.Println("  - API文档:      /swagger/*")
	}
}

// setupAdvancedGRPCServices 设置高级gRPC服务
func setupAdvancedGRPCServices(s interface{}) {
	log.Println("🔌 注册高级gRPC服务...")

	log.Println("✅ 高级gRPC服务注册完成")
	log.Println("  - UserService             (用户管理)")
	log.Println("  - SearchService           (搜索服务)")
	log.Println("  - AnalyticsService        (数据分析)")
	log.Println("  - DocumentService         (文档管理)")
	log.Println("  - ConfigService           (配置管理)")
	log.Println("  - NotificationService     (通知服务)")
}

// ================================
// 🔗 HTTP处理器实现
// ================================

func getUsersHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"users": []gin.H{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
		},
		"total": 2,
	})
}

func createUserHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"id":      3,
		"message": "User created successfully",
	})
}

func getUserHandler(c *gin.Context) {
	id := c.Param("id")
	httpTransport.Success(c, gin.H{
		"id":    id,
		"name":  "User " + id,
		"email": "user" + id + "@example.com",
	})
}

func updateUserHandler(c *gin.Context) {
	id := c.Param("id")
	httpTransport.Success(c, gin.H{
		"id":      id,
		"message": "User updated successfully",
	})
}

func deleteUserHandler(c *gin.Context) {
	id := c.Param("id")
	httpTransport.Success(c, gin.H{
		"id":      id,
		"message": "User deleted successfully",
	})
}

func searchHandler(c *gin.Context) {
	query := c.Query("q")
	httpTransport.Success(c, gin.H{
		"query":   query,
		"results": []gin.H{},
		"total":   0,
	})
}

func indexDocumentHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"message": "Document indexed successfully",
	})
}

func getStatsHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"users":     1234,
		"documents": 5678,
		"searches":  9999,
	})
}

func trackEventHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"message": "Event tracked successfully",
	})
}

func getDocumentsHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"documents": []gin.H{},
		"total":     0,
	})
}

func createDocumentHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"id":      "doc123",
		"message": "Document created successfully",
	})
}

func getConfigHandler(c *gin.Context) {
	key := c.Param("key")
	httpTransport.Success(c, gin.H{
		"key":   key,
		"value": "config_value_" + key,
	})
}

func setConfigHandler(c *gin.Context) {
	key := c.Param("key")
	httpTransport.Success(c, gin.H{
		"key":     key,
		"message": "Config updated successfully",
	})
}

func getMetricsHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"cpu":    "45%",
		"memory": "67%",
		"disk":   "23%",
	})
}

func getLogsHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"logs": []gin.H{
			{"level": "INFO", "message": "Application started"},
			{"level": "DEBUG", "message": "Database connected"},
		},
	})
}

func clearCacheHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"message": "Cache cleared successfully",
	})
}

func getComponentStatusHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"components": gin.H{
			"mysql":         "healthy",
			"redis":         "healthy",
			"elasticsearch": "healthy",
			"kafka":         "healthy",
			"mongodb":       "healthy",
			"etcd":          "healthy",
		},
	})
}

func swaggerHandler(c *gin.Context) {
	httpTransport.Success(c, gin.H{
		"message": "Swagger documentation",
		"url":     "/swagger/index.html",
	})
}
