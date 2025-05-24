package api

import (
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/circuitbreaker"
	"distributed-service/pkg/config"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/middleware"
	"distributed-service/pkg/ratelimit"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// RegisterRoutes registers all API routes
// @title Distributed Service API
// @version 1.0
// @description This is a distributed service server with JWT authentication.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func RegisterRoutes(r *gin.Engine, userService service.UserService, jwtManager *auth.JWTManager) {
	// 初始化限流器和熔断器
	rateLimiter, err := ratelimit.NewRateLimiterFromConfig(config.GlobalConfig.RateLimit)
	if err != nil {
		logger.Error(nil, "Failed to initialize rate limiter", zap.Error(err))
		// 使用一个禁用的限流器
		rateLimiter, _ = ratelimit.NewRateLimiter(config.RateLimitConfig{Enabled: false})
	}
	circuitBreaker := circuitbreaker.NewCircuitBreaker()

	// 初始化默认熔断器配置
	circuitbreaker.InitDefaultCircuitBreakers()

	// 配置API专用熔断器
	circuitbreaker.ConfigureCommand("auth_register", circuitbreaker.Config{
		Timeout:                3000, // 3秒超时
		MaxConcurrentRequests:  20,   // 最大20个并发
		RequestVolumeThreshold: 10,   // 10个请求后开始统计
		SleepWindow:            5000, // 5秒休眠窗口
		ErrorPercentThreshold:  30,   // 30%错误率
	})

	circuitbreaker.ConfigureCommand("auth_login", circuitbreaker.Config{
		Timeout:                3000,
		MaxConcurrentRequests:  30,
		RequestVolumeThreshold: 15,
		SleepWindow:            5000,
		ErrorPercentThreshold:  25,
	})

	circuitbreaker.ConfigureCommand("user_get", circuitbreaker.Config{
		Timeout:                2000, // 用户查询较快
		MaxConcurrentRequests:  100,
		RequestVolumeThreshold: 20,
		SleepWindow:            3000,
		ErrorPercentThreshold:  40,
	})
	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 熔断器指标流端点
	r.GET("/hystrix", circuitbreaker.MetricsStreamHandler())

	// 熔断器状态查看端点
	r.GET("/circuit-breaker/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, circuitbreaker.HealthCheck())
	})

	// Health check endpoint (with rate limiting)
	r.GET("/health", rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("health_check")), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"services": gin.H{
				"mysql":    "connected",
				"redis":    "connected",
				"rabbitmq": "connected",
				"consul":   "connected",
			},
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required, with rate limiting)
		authHandler := NewAuthHandler(userService, jwtManager)
		authBase := v1.Group("/auth")
		authBase.Use(rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("auth_public"))) // 认证公开端点限流
		{
			authBase.POST("/register",
				circuitBreaker.Middleware("auth_register", nil),
				authHandler.Register)
			authBase.POST("/login",
				circuitBreaker.Middleware("auth_login", nil),
				authHandler.Login)
			authBase.POST("/refresh",
				circuitBreaker.Middleware("auth_refresh", nil),
				authHandler.RefreshToken)
		}

		// Protected authentication routes
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.JWTAuth(jwtManager))
		authProtected.Use(rateLimiter.UserRateLimit(rateLimiter.GetConfiguredLimit("auth_protected"))) // 认证保护端点限流
		{
			authProtected.POST("/change-password",
				circuitBreaker.Middleware("auth_change_password", nil),
				authHandler.ChangePassword)
		}

		// User routes - some protected, some public
		userHandler := NewUserHandler(userService)
		users := v1.Group("/users")
		users.Use(rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("user_public"))) // 用户公开API限流
		{
			// Public routes
			users.GET("/:id",
				circuitBreaker.Middleware("user_get", nil),
				userHandler.GetByID) // Anyone can view user profiles
		}

		// Protected user routes
		usersProtected := v1.Group("/users")
		usersProtected.Use(middleware.JWTAuth(jwtManager))
		usersProtected.Use(rateLimiter.UserRateLimit(rateLimiter.GetConfiguredLimit("user_protected"))) // 用户保护API限流
		{
			usersProtected.GET("/me",
				circuitBreaker.Middleware("user_me", nil),
				userHandler.GetMe) // Get current user info
			usersProtected.POST("",
				circuitBreaker.Middleware("user_create", nil),
				userHandler.Create) // Only authenticated users can create users
			usersProtected.DELETE("/:id",
				circuitBreaker.Middleware("user_delete", nil),
				userHandler.Delete) // Only authenticated users can delete users
		}
	}
}
