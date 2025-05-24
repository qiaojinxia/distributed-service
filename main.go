package main

import (
	"context"
	"distributed-service/internal/api"
	"distributed-service/internal/model"
	"distributed-service/internal/repository"
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/config"
	"distributed-service/pkg/database"
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

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Set Gin mode
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// Initialize router
	r := gin.Default()

	// Add context middleware
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

	// Add tracing middleware
	if config.GlobalConfig.Tracing.Enabled {
		r.Use(middleware.TracingMiddleware(config.GlobalConfig.Tracing.ServiceName))
		r.Use(middleware.CustomTracingMiddleware())
	}

	// Add metrics middleware
	r.Use(middleware.MetricsMiddleware())

	// Register routes with JWT manager
	api.RegisterRoutes(r, userService, jwtManager)

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
		logger.Info(ctx, "Starting server",
			logger.String("address", srv.Addr),
			logger.String("mode", config.GlobalConfig.Server.Mode),
		)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, "Failed to start server", logger.Error_(err))
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

	// Shutdown server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal(shutdownCtx, "Server forced to shutdown", logger.Error_(err))
	}

	logger.Info(shutdownCtx, "Server exiting")
}
