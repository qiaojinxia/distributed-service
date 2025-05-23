package api

import (
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/middleware"
	"net/http"

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
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		authHandler := NewAuthHandler(userService, jwtManager)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
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
			usersProtected.POST("", userHandler.Create)       // Only authenticated users can create users
			usersProtected.DELETE("/:id", userHandler.Delete) // Only authenticated users can delete users
		}
	}
}
