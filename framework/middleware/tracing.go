package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware 创建追踪中间件
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	// 使用 OpenTelemetry 官方的 Gin 中间件
	return otelgin.Middleware(serviceName)
}

// CustomTracingMiddleware 自定义追踪中间件，提供更多控制
func CustomTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取 context
		ctx, exists := c.Get("ctx")
		if !exists {
			ctx = c.Request.Context()
		}

		requestCtx := ctx.(context.Context)

		// 开始 span
		spanName := c.Request.Method + " " + c.FullPath()
		requestCtx, span := tracing.StartSpan(requestCtx, spanName)
		defer span.End()

		// 更新 context
		c.Set("ctx", requestCtx)

		// 添加请求属性
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.remote_addr", c.ClientIP()),
		)

		// 获取 trace id 并设置到 gin context
		if span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			c.Set("trace_id", traceID)
			c.Header("X-Trace-ID", traceID)
		}

		// 处理请求
		c.Next()

		// 记录响应信息
		statusCode := c.Writer.Status()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// 追踪HTTP请求
		tracing.TraceHTTPRequest(requestCtx, c.Request.Method, c.FullPath(), statusCode)
	}
}

// RequestIDMiddleware 请求ID中间件，为每个请求生成唯一ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 如果没有请求ID，从当前span生成一个
			if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
				requestID = span.SpanContext().TraceID().String()
			}
		}

		// 设置请求ID到header和context
		if requestID != "" {
			c.Header("X-Request-ID", requestID)
			if ctx, exists := c.Get("ctx"); exists {
				requestCtx := ctx.(context.Context)
				tracing.AddSpanAttributes(requestCtx, attribute.String("request.id", requestID))
			}
		}

		c.Next()
	}
}

// LoggingMiddleware HTTP 请求日志中间件，自动记录 trace id
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取或创建 context
		ctx, exists := c.Get("ctx")
		if !exists {
			ctx = c.Request.Context()
		}
		requestCtx := ctx.(context.Context)

		// 记录请求开始
		logger.Info(requestCtx, "HTTP request started",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("query", c.Request.URL.RawQuery),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
		)

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		// 记录请求完成
		if statusCode >= 400 {
			logger.Error(requestCtx, "HTTP request completed with error",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.Int("status_code", statusCode),
				logger.Duration("duration", duration),
				logger.Int("response_size", responseSize),
			)
		} else {
			logger.Info(requestCtx, "HTTP request completed",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.Int("status_code", statusCode),
				logger.Duration("duration", duration),
				logger.Int("response_size", responseSize),
			)
		}
	}
}

// TraceContextMiddleware 确保每个请求都有正确的 trace context
func TraceContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前请求的 context
		ctx := c.Request.Context()

		// 如果 context 中没有 span，创建一个
		span := trace.SpanFromContext(ctx)
		if !span.SpanContext().IsValid() {
			ctx, span = tracing.StartSpan(ctx, "http.request")
			defer span.End()
		}

		// 将 context 设置到 gin
		c.Set("ctx", ctx)

		// 如果有有效的 span，设置 trace id
		if span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			c.Set("trace_id", traceID)
			c.Header("X-Trace-ID", traceID)
		}

		c.Next()
	}
}
