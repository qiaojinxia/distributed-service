package api

import (
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/middleware"
	"net/http"

	"context"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	ctx := context.Background()

	logger.Info(ctx, "API routes initialized with Sentinel protection")

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"services": gin.H{
				"mysql":    "connected",
				"redis":    "connected",
				"rabbitmq": "connected",
				"consul":   "connected",
			},
			"protection": gin.H{
				"enabled": true,
				"type":    "sentinel",
			},
		})
	})

	// Sentinel protection status endpoint
	r.GET("/protection/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"sentinel": gin.H{
				"enabled": true,
				"type":    "flow_control_and_circuit_breaker",
			},
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		authHandler := NewAuthHandler(userService, jwtManager)
		authBase := v1.Group("/auth")
		{
			authBase.POST("/register", authHandler.Register)
			authBase.POST("/login", authHandler.Login)
			authBase.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected authentication routes
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.JWTAuth(jwtManager))
		{
			authProtected.POST("/change-password", authHandler.ChangePassword)
		}

		// User routes - some protected, some public
		userHandler := NewUserHandler(userService)
		users := v1.Group("/users")
		{
			// Public routes
			users.GET("/:id", userHandler.GetByID) // Anyone can view user profiles
		}

		// Protected user routes
		usersProtected := v1.Group("/users")
		usersProtected.Use(middleware.JWTAuth(jwtManager))
		{
			usersProtected.GET("/me", userHandler.GetByID)    // Get current user info
			usersProtected.POST("", userHandler.Create)       // Create new user
			usersProtected.DELETE("/:id", userHandler.Delete) // Delete user
		}
	}
}
