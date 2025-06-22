package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/app"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/middleware"
	httpTransport "github.com/qiaojinxia/distributed-service/framework/transport/http"
)

func main() {
	// 创建应用 - 优化的配置
	builder := app.New().
		Name("clean-logging-demo").
		Version("v1.0.0").
		EnableHTTP().
		EnableTracing().
		Port(8080).

		// 🔧 优化的追踪配置 - 不输出冗长的span信息
		WithTracing(&config.TracingConfig{
			ServiceName:    "clean-logging-demo",
			ServiceVersion: "v1.0.0",
			Environment:    "development",
			Enabled:        true,
			ExporterType:   "none", // 🚀 使用 "none" 避免输出span详情，只保留trace_id
			SampleRatio:    1.0,
		}).

		// 🔧 优化的日志配置
		WithLogger(&config.LoggerConfig{
			Level:      "info",
			Encoding:   "json", // JSON格式，便于分析，但不会有冗长的span信息
			OutputPath: "stdout",
		}).

		// 配置HTTP路由
		HTTP(setupRoutes)

	// 构建并运行应用
	if err := builder.Run(); err != nil {
		panic(err)
	}
}

func setupRoutes(r interface{}) {
	router := r.(*gin.Engine)

	// 添加优化的中间件
	router.Use(middleware.TraceContextMiddleware()) // 确保trace context（但不输出span详情）
	router.Use(optimizedLoggingMiddleware())        // 自定义简洁的日志中间件

	// API 路由组
	api := router.Group("/api/v1")
	{
		api.GET("/users/:id", getUserHandler)
		api.POST("/users", createUserHandler)
		api.GET("/health", healthHandler)
	}
}

// 🎯 优化的日志中间件 - 简洁但有用的日志
func optimizedLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取或创建 context
		ctx, exists := c.Get("ctx")
		if !exists {
			ctx = c.Request.Context()
		}
		requestCtx := ctx.(context.Context)

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 🚀 简洁的HTTP访问日志 - 一行包含关键信息
		if statusCode >= 400 {
			logger.Warn(requestCtx, "HTTP request completed with error",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.Int("status", statusCode),
				logger.Duration("latency", duration),
				logger.String("ip", c.ClientIP()),
			)
		} else {
			logger.Info(requestCtx, "HTTP request completed",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.Int("status", statusCode),
				logger.Duration("latency", duration),
				logger.String("ip", c.ClientIP()),
			)
		}
	}
}

// getUserHandler 获取用户处理器
func getUserHandler(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	userID := c.Param("id")

	// 🎯 简洁的业务日志 - 只记录关键信息
	logger.Info(ctx, "Processing get user request",
		logger.String("user_id", userID),
	)

	// 模拟业务逻辑
	time.Sleep(50 * time.Millisecond)

	if userID == "404" {
		logger.Warn(ctx, "User not found",
			logger.String("user_id", userID),
		)
		httpTransport.NotFound(c, "User not found")
		return
	}

	logger.Info(ctx, "User retrieved successfully",
		logger.String("user_id", userID),
	)

	httpTransport.Success(c, gin.H{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	})
}

// createUserHandler 创建用户处理器
func createUserHandler(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn(ctx, "Invalid request body",
			logger.Err(err),
		)
		httpTransport.BadRequest(c, "Invalid request body")
		return
	}

	logger.Info(ctx, "Creating user",
		logger.String("name", req.Name),
		logger.String("email", req.Email),
	)

	// 模拟创建用户
	time.Sleep(100 * time.Millisecond)
	userID := "user_" + time.Now().Format("20060102150405")

	logger.Info(ctx, "User created successfully",
		logger.String("user_id", userID),
	)

	httpTransport.Success(c, gin.H{
		"id":    userID,
		"name":  req.Name,
		"email": req.Email,
	})
}

// healthHandler 健康检查处理器
func healthHandler(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)

	logger.Debug(ctx, "Health check requested")

	httpTransport.Success(c, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}
