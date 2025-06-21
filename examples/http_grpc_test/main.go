package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework"
	"github.com/qiaojinxia/distributed-service/framework/logger"
)

func main() {
	log := logger.GetLogger()

	log.Info("🚀 启动HTTP + gRPC集成服务...")

	// 同时启动HTTP和gRPC服务
	err := framework.New().
		Port(8080).                // HTTP端口
		Name("http-grpc-service"). // 服务名称
		Version("v1.0.0").         // 版本
		Mode("debug").             // 运行模式
		EnableAll().               // 启用所有服务 (HTTP + gRPC + Metrics + Tracing)
		HTTP(setupHTTPRoutes).     // 注册HTTP路由
		GRPC(setupGRPCServices).   // 注册gRPC服务
		BeforeStart(func(ctx context.Context) error {
			log.Info("🔧 初始化HTTP + gRPC服务...")
			return nil
		}).
		AfterStart(func(ctx context.Context) error {
			log.Info("✅ HTTP + gRPC服务启动完成!")
			log.Info("🌐 HTTP服务监听: http://localhost:8080")
			log.Info("🔌 gRPC服务监听: localhost:9093")
			log.Info("📋 可用的HTTP端点:")
			log.Info("  - GET  http://localhost:8080/health")
			log.Info("  - GET  http://localhost:8080/ping")
			log.Info("  - GET  http://localhost:8080/api/users")
			log.Info("🔌 可用的gRPC服务:")
			log.Info("  - HealthCheck (grpc.health.v1.Health)")
			log.Info("  - Server Reflection")
			return nil
		}).
		Run()

	if err != nil {
		log.Fatal("服务启动失败", logger.Any("error", err))
	}
}

// setupHTTPRoutes 设置HTTP路由
func setupHTTPRoutes(r interface{}) {
	router := r.(*gin.Engine)
	log := logger.GetLogger()

	log.Info("🌐 注册HTTP路由:")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "http-test",
			"version": "v1.0.0",
			"message": "HTTP service is working!",
		})
	})
	log.Info("  ✅ GET /health")

	// Ping端点
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "pong",
			"timestamp": "2024-01-20T10:00:00Z",
		})
	})
	log.Info("  ✅ GET /ping")

	// 简化的用户API
	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "Alice", "email": "alice@example.com"},
				{"id": 2, "name": "Bob", "email": "bob@example.com"},
			},
			"total": 2,
		})
	})
	log.Info("  ✅ GET /api/users")

	// 用户详情
	router.GET("/api/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"id":    id,
			"name":  "User " + id,
			"email": "user" + id + "@example.com",
		})
	})
	log.Info("  ✅ GET /api/users/:id")
}

// setupGRPCServices 设置gRPC服务
func setupGRPCServices(s interface{}) {
	// gRPC服务器实例
	// server := s.(*grpc.Server)

	log := logger.GetLogger()
	log.Info("🔌 注册gRPC服务:")
	log.Info("  ✅ HealthCheck服务 (自动注册)")
	log.Info("  ✅ Server Reflection (自动注册)")

	// 这里可以注册自定义的gRPC服务
	// 例如：
	// pb.RegisterUserServiceServer(server, &userService{})
	// pb.RegisterOrderServiceServer(server, &orderService{})
}
