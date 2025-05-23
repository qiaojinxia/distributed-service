package middleware

import (
	"context"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth creates a JWT authentication middleware
func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("ctx").(context.Context)

		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn(ctx, "Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn(ctx, "Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			logger.Warn(ctx, "Empty token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Empty token"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := jwtManager.ValidateToken(ctx, tokenString)
		if err != nil {
			logger.Error(ctx, "Token validation failed", logger.Error_(err))
			if err == auth.ErrTokenExpired {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// Update context with user info
		userCtx := context.WithValue(ctx, "user_id", claims.UserID)
		userCtx = context.WithValue(userCtx, "username", claims.Username)
		c.Set("ctx", userCtx)

		logger.Info(ctx, "User authenticated successfully",
			logger.String("username", claims.Username),
			logger.Int("user_id", int(claims.UserID)),
		)

		c.Next()
	}
}

// OptionalJWTAuth creates an optional JWT authentication middleware
// If token is provided, it validates it, but doesn't require authentication
func OptionalJWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("ctx").(context.Context)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.Next()
			return
		}

		// Try to validate the token
		claims, err := jwtManager.ValidateToken(ctx, tokenString)
		if err == nil {
			// Token is valid, add user info to context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)

			userCtx := context.WithValue(ctx, "user_id", claims.UserID)
			userCtx = context.WithValue(userCtx, "username", claims.Username)
			c.Set("ctx", userCtx)
		}

		c.Next()
	}
}
