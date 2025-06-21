package main

import (
	"context"
	"distributed-service/internal/api"
	grpcService "distributed-service/internal/grpc"
	"distributed-service/internal/model"
	"distributed-service/internal/repository"
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/config"
	"distributed-service/pkg/database"
	grpcServer "distributed-service/pkg/grpc"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/middleware"
	"distributed-service/pkg/mq"
	"distributed-service/pkg/registry"
	"distributed-service/pkg/tracing"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "distributed-service/docs" // This is for swagger

	pb "distributed-service/api/proto/user"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// @title Distributed Service API
// @version 1.0
// @description This is a distributed service server with JWT authentication.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Create base context
	ctx := context.Background()

	// Load configuration
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.InitLogger(&logger.Config{
		Level:      config.GlobalConfig.Logger.Level,
		Encoding:   config.GlobalConfig.Logger.Encoding,
		OutputPath: config.GlobalConfig.Logger.OutputPath,
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize tracing
	tracingConfig := &tracing.Config{
		ServiceName:    config.GlobalConfig.Tracing.ServiceName,
		ServiceVersion: config.GlobalConfig.Tracing.ServiceVersion,
		Environment:    config.GlobalConfig.Tracing.Environment,
		Enabled:        config.GlobalConfig.Tracing.Enabled,
		ExporterType:   config.GlobalConfig.Tracing.ExporterType,
		Endpoint:       config.GlobalConfig.Tracing.Endpoint,
		SampleRatio:    config.GlobalConfig.Tracing.SampleRatio,
	}

	tracingManager, err := tracing.NewTracingManager(ctx, tracingConfig)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize tracing", logger.Error_(err))
	}
	defer func() {
		if err := tracingManager.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Failed to shutdown tracing", logger.Error_(err))
		}
	}()

	// Add service info to context
	ctx = context.WithValue(ctx, "service_name", config.GlobalConfig.Server.Name)
	ctx = context.WithValue(ctx, "service_version", config.GlobalConfig.Server.Version)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		config.GlobalConfig.JWT.SecretKey,
		config.GlobalConfig.JWT.Issuer,
	)

	// Initialize service registry
	serviceRegistry, err := registry.NewServiceRegistry(ctx, &config.GlobalConfig.Consul)
	if err != nil {
		logger.Fatal(ctx, "Failed to create service registry", logger.Error_(err))
	}

	// Initialize MySQL
	if err := database.InitMySQL(ctx, &config.GlobalConfig.MySQL); err != nil {
		logger.Fatal(ctx, "Failed to initialize MySQL", logger.Error_(err))
	}

	// Auto migrate database
	if err := database.DB.AutoMigrate(&model.User{}); err != nil {
		logger.Fatal(ctx, "Failed to migrate database", logger.Error_(err))
	}

	// Initialize Redis
	if err := database.InitRedis(ctx, &config.GlobalConfig.Redis); err != nil {
		logger.Fatal(ctx, "Failed to initialize Redis", logger.Error_(err))
	}

	// Initialize RabbitMQ
	if err := mq.InitRabbitMQ(ctx, &config.GlobalConfig.RabbitMQ); err != nil {
		logger.Fatal(ctx, "Failed to initialize RabbitMQ", logger.Error_(err))
	}
	defer mq.CloseRabbitMQ(ctx)

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.DB)

	// Initialize services
	userService := service.NewUserService(userRepo)

	// Initialize gRPC server
	grpcConfig, err := grpcServer.ConvertConfig(&config.GlobalConfig.GRPC)
	if err != nil {
		logger.Fatal(ctx, "Failed to convert gRPC config", logger.Error_(err))
	}

	// 创建gRPC拦截器链
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// 1. 基础中间件（来自 common.go）
	unaryInterceptors = append(unaryInterceptors,
		middleware.GRPCRecoveryInterceptor(), // 恢复中间件（最先执行）
		middleware.GRPCLoggingInterceptor(),  // 日志中间件
		middleware.GRPCMetricsInterceptor(),  // 指标中间件
	)
	streamInterceptors = append(streamInterceptors,
		middleware.GRPCStreamRecoveryInterceptor(), // 流恢复中间件
		middleware.GRPCStreamLoggingInterceptor(),  // 流日志中间件
	)

	// 2. Sentinel保护中间件
	sentinelMiddleware, err := middleware.NewSentinelProtectionMiddleware(ctx, &config.GlobalConfig.Protection)
	if err != nil {
		logger.Error(ctx, "Failed to initialize Sentinel middleware", logger.Error_(err))
	} else if sentinelMiddleware.IsEnabled() {
		unaryInterceptors = append(unaryInterceptors, sentinelMiddleware.GRPCUnaryInterceptor())
		streamInterceptors = append(streamInterceptors, sentinelMiddleware.GRPCStreamInterceptor())
		logger.Info(ctx, "Sentinel protection enabled for gRPC")
	}

	// 3. 链路追踪中间件（如果启用）
	if config.GlobalConfig.Tracing.Enabled {
		unaryInterceptors = append(unaryInterceptors, middleware.GRPCTracingInterceptor())
		streamInterceptors = append(streamInterceptors, middleware.GRPCStreamTracingInterceptor())
	}

	// 使用拦截器创建gRPC服务器
	grpcSrv, err := grpcServer.NewServerWithInterceptors(ctx, grpcConfig, unaryInterceptors, streamInterceptors)
	if err != nil {
		logger.Fatal(ctx, "Failed to create gRPC server", logger.Error_(err))
	}

	// 记录gRPC中间件配置
	logger.Info(ctx, "gRPC middleware configured",
		logger.Int("unary_interceptors", len(unaryInterceptors)),
		logger.Int("stream_interceptors", len(streamInterceptors)),
		logger.Bool("tracing_enabled", config.GlobalConfig.Tracing.Enabled),
		logger.String("middleware_chain", "recovery->logging->metrics->tracing->protection"))

	// Register gRPC services
	userGRPCService := grpcService.NewUserServiceServer(userService, jwtManager)
	pb.RegisterUserServiceServer(grpcSrv.GetServer(), userGRPCService)

	// Set health status for user service
	grpcSrv.SetHealthStatus("user.v1.UserService", 1) // SERVING

	// Set Gin mode
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// Initialize router with custom middleware
	r := gin.New() // 使用空的路由器，手动添加中间件

	// 1. 基础中间件（来自 common.go）
	r.Use(middleware.GinRecovery()) // 恢复中间件（最先执行）
	r.Use(middleware.GinLogger())   // 日志中间件

	// 2. 自定义上下文中间件
	r.Use(func(c *gin.Context) {
		// Create request-scoped context with timeout
		reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Add request information
		reqCtx = context.WithValue(reqCtx, "request_id", c.GetHeader("X-Request-ID"))
		reqCtx = context.WithValue(reqCtx, "user_agent", c.GetHeader("User-Agent"))

		// Store context in Gin
		c.Set("ctx", reqCtx)
		c.Next()
	})

	// 3. Sentinel保护中间件
	if sentinelMiddleware != nil && sentinelMiddleware.IsEnabled() {
		r.Use(sentinelMiddleware.HTTPMiddleware())
		logger.Info(ctx, "Sentinel protection enabled for HTTP")
	}

	// 4. 链路追踪中间件（如果启用）
	if config.GlobalConfig.Tracing.Enabled {
		r.Use(middleware.TracingMiddleware(config.GlobalConfig.Tracing.ServiceName))
		r.Use(middleware.CustomTracingMiddleware())
	}

	// 5. 指标中间件
	r.Use(middleware.MetricsMiddleware())

	// 记录HTTP中间件配置
	logger.Info(ctx, "HTTP middleware configured",
		logger.Bool("tracing_enabled", config.GlobalConfig.Tracing.Enabled),
		logger.String("middleware_chain", "recovery->logging->context->protection->tracing->metrics"))

	// Register routes with JWT manager
	api.RegisterRoutes(r, userService, jwtManager, &config.GlobalConfig)

	// Register service with Consul
	if err := serviceRegistry.RegisterService(ctx, &config.GlobalConfig.Server); err != nil {
		logger.Fatal(ctx, "Failed to register service", logger.Error_(err))
	}

	// Setup Prometheus metrics endpoint
	if config.GlobalConfig.Metrics.Enabled {
		go func() {
			metricsServer := &http.Server{
				Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Metrics.PrometheusPort),
				Handler: promhttp.Handler(),
			}
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error(ctx, "Metrics server error", logger.Error_(err))
			}
		}()
	}

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		logger.Info(ctx, "Starting HTTP server",
			logger.String("address", srv.Addr),
			logger.String("mode", config.GlobalConfig.Server.Mode),
		)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, "Failed to start HTTP server", logger.Error_(err))
		}
	}()

	// Start gRPC server
	go func() {
		logger.Info(ctx, "Starting gRPC server",
			logger.Int("port", config.GlobalConfig.GRPC.Port),
		)

		if err := grpcSrv.Start(ctx); err != nil {
			logger.Fatal(ctx, "Failed to start gRPC server", logger.Error_(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "Shutting down server...")

	// Create shutdown context
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Deregister service
	if err := serviceRegistry.DeregisterService(shutdownCtx, &config.GlobalConfig.Server); err != nil {
		logger.Error(shutdownCtx, "Failed to deregister service", logger.Error_(err))
	}

	// Shutdown gRPC server
	if err := grpcSrv.Stop(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "Failed to shutdown gRPC server", logger.Error_(err))
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal(shutdownCtx, "HTTP server forced to shutdown", logger.Error_(err))
	}

	logger.Info(shutdownCtx, "Server exiting")
}
