package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/app"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/middleware"
	"github.com/qiaojinxia/distributed-service/framework/tracing"
	httpTransport "github.com/qiaojinxia/distributed-service/framework/transport/http"
)

func main() {
	// 创建应用
	builder := app.New().
		Name("trace-logging-demo").
		Version("v1.0.0").
		EnableHTTP().
		EnableTracing().
		Port(8080).

		// 配置链路追踪
		WithTracing(&config.TracingConfig{
			ServiceName:    "trace-logging-demo",
			ServiceVersion: "v1.0.0",
			Environment:    "development",
			Enabled:        true,
			ExporterType:   "stdout", // 输出到控制台，便于查看
			SampleRatio:    1.0,
		}).

		// 配置日志
		WithLogger(&config.LoggerConfig{
			Level:      "info",
			Encoding:   "json", // 使用JSON格式，便于看到trace_id字段
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

	// 添加中间件
	router.Use(middleware.TraceContextMiddleware()) // 确保trace context
	router.Use(middleware.LoggingMiddleware())      // 自动记录请求日志（带trace id）

	// API 路由组
	api := router.Group("/api/v1")
	{
		api.GET("/users/:id", getUserHandler)
		api.POST("/users", createUserHandler)
		api.GET("/orders/:id", getOrderHandler)
		api.GET("/health", healthHandler)
	}
}

// getUserHandler 获取用户处理器 - 展示如何在业务逻辑中使用带trace id的日志
func getUserHandler(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	userID := c.Param("id")

	// 使用带上下文的日志记录，自动包含trace id
	logger.Info(ctx, "Getting user information",
		logger.String("user_id", userID),
		logger.String("operation", "get_user"),
	)

	// 模拟业务逻辑
	if err := simulateUserFetch(ctx, userID); err != nil {
		logger.Error(ctx, "Failed to fetch user",
			logger.String("user_id", userID),
			logger.Error_(err),
		)
		httpTransport.InternalError(c, "Failed to fetch user")
		return
	}

	logger.Info(ctx, "User information retrieved successfully",
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
			logger.Error_(err),
		)
		httpTransport.BadRequest(c, "Invalid request body")
		return
	}

	logger.Info(ctx, "Creating new user",
		logger.String("name", req.Name),
		logger.String("email", req.Email),
	)

	// 模拟创建用户
	userID, err := simulateUserCreate(ctx, req.Name, req.Email)
	if err != nil {
		logger.Error(ctx, "Failed to create user",
			logger.String("name", req.Name),
			logger.String("email", req.Email),
			logger.Error_(err),
		)
		httpTransport.InternalError(c, "Failed to create user")
		return
	}

	logger.Info(ctx, "User created successfully",
		logger.String("user_id", userID),
		logger.String("name", req.Name),
	)

	httpTransport.Success(c, gin.H{
		"id":    userID,
		"name":  req.Name,
		"email": req.Email,
	})
}

// getOrderHandler 获取订单处理器 - 展示跨服务调用的trace id传播
func getOrderHandler(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	orderID := c.Param("id")

	logger.Info(ctx, "Processing order request",
		logger.String("order_id", orderID),
	)

	// 模拟跨服务调用
	if err := simulateServiceCall(ctx, "payment-service", orderID); err != nil {
		logger.Error(ctx, "Payment service call failed",
			logger.String("order_id", orderID),
			logger.String("service", "payment-service"),
			logger.Error_(err),
		)
		httpTransport.InternalError(c, "Payment processing failed")
		return
	}

	if err := simulateServiceCall(ctx, "inventory-service", orderID); err != nil {
		logger.Error(ctx, "Inventory service call failed",
			logger.String("order_id", orderID),
			logger.String("service", "inventory-service"),
			logger.Error_(err),
		)
		httpTransport.InternalError(c, "Inventory check failed")
		return
	}

	logger.Info(ctx, "Order processed successfully",
		logger.String("order_id", orderID),
	)

	httpTransport.Success(c, gin.H{
		"id":     orderID,
		"status": "completed",
		"amount": 99.99,
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

// 模拟函数

func simulateUserFetch(ctx context.Context, userID string) error {
	// 创建子span
	ctx, span := tracing.StartSpan(ctx, "db.query.users")
	defer span.End()

	logger.Debug(ctx, "Querying user from database",
		logger.String("user_id", userID),
		logger.String("table", "users"),
	)

	// 模拟数据库查询延迟
	time.Sleep(50 * time.Millisecond)

	if userID == "404" {
		return &UserNotFoundError{UserID: userID}
	}

	logger.Debug(ctx, "User found in database",
		logger.String("user_id", userID),
	)

	return nil
}

func simulateUserCreate(ctx context.Context, name, email string) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "db.insert.users")
	defer span.End()

	logger.Debug(ctx, "Inserting user into database",
		logger.String("name", name),
		logger.String("email", email),
	)

	// 模拟数据库插入延迟
	time.Sleep(100 * time.Millisecond)

	userID := "user_123456"

	logger.Debug(ctx, "User inserted successfully",
		logger.String("user_id", userID),
	)

	return userID, nil
}

func simulateServiceCall(ctx context.Context, serviceName, orderID string) error {
	ctx, span := tracing.StartSpan(ctx, "http.client."+serviceName)
	defer span.End()

	logger.Info(ctx, "Calling external service",
		logger.String("service", serviceName),
		logger.String("order_id", orderID),
	)

	// 模拟网络延迟
	time.Sleep(200 * time.Millisecond)

	// 模拟偶发错误
	if serviceName == "payment-service" && orderID == "error" {
		return &ServiceError{Service: serviceName, Message: "Payment declined"}
	}

	logger.Info(ctx, "Service call completed successfully",
		logger.String("service", serviceName),
		logger.String("order_id", orderID),
	)

	return nil
}

// 自定义错误类型

type UserNotFoundError struct {
	UserID string
}

func (e *UserNotFoundError) Error() string {
	return "user not found: " + e.UserID
}

type ServiceError struct {
	Service string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Service + ": " + e.Message
}
