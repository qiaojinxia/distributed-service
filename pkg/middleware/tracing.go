package middleware

import (
	"context"
	"distributed-service/pkg/tracing"

	"github.com/gin-gonic/gin"
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
			c.Next()
			return
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
