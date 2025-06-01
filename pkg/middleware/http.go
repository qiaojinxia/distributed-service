// Package middleware provides HTTP common middleware
package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ===== HTTP Common Middleware =====

// GinLogger returns a gin.HandlerFunc (middleware) that logs requests using custom format.
func GinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom format with more details
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// GinRecovery returns a gin.HandlerFunc (middleware) that recovers from any panics.
func GinRecovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(500, gin.H{
				"error":   "Internal Server Error",
				"message": err,
			})
		}
		c.AbortWithStatus(500)
	})
}
