// Package middleware provides HTTP common middleware
package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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

// HTTPRecoveryMiddleware HTTP恢复中间件
func HTTPRecoveryMiddleware() gin.HandlerFunc {
	return GinRecovery()
}

// HTTPLoggingMiddleware HTTP日志中间件
func HTTPLoggingMiddleware() gin.HandlerFunc {
	return GinLogger()
}

// HTTPCORSMiddleware CORS中间件
func HTTPCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// HTTPMetricsMiddleware HTTP指标中间件
func HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// 记录指标
		duration := time.Since(start)

		// 这里可以集成Prometheus指标
		// metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path, fmt.Sprintf("%d", c.Writer.Status())).Observe(duration.Seconds())
		// metrics.HTTPRequestTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, fmt.Sprintf("%d", c.Writer.Status())).Inc()

		// 临时记录到日志
		fmt.Printf("[METRICS] %s %s - %d - %v\n", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}
