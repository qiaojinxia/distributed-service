package tracing

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes" // gRPC status codes
)

const (
	TracerName = "distributed-service"
)

// StartSpan 开始一个新的 span
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer(TracerName)
	return tracer.Start(ctx, spanName, opts...)
}

// SpanFromContext 从上下文中获取当前 span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanAttributes 向当前 span 添加属性
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError 记录错误到当前 span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanStatus 设置 span 状态
func SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(code, description)
}

// TraceHTTPRequest 追踪 HTTP 请求的辅助函数
func TraceHTTPRequest(ctx context.Context, method, path string, statusCode int) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.Int("http.status_code", statusCode),
	)

	// 根据状态码设置 span 状态
	if statusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// TraceDatabase 追踪数据库操作的辅助函数
func TraceDatabase(ctx context.Context, operation, table string, rowsAffected int64) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
		attribute.Int64("db.rows_affected", rowsAffected),
		attribute.String("db.system", "mysql"),
	)
}

// TraceCache 追踪缓存操作的辅助函数
func TraceCache(ctx context.Context, operation, key string, hit bool) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cache.operation", operation),
		attribute.String("cache.key", key),
		attribute.Bool("cache.hit", hit),
		attribute.String("cache.system", "redis"),
	)
}

// TraceMessageQueue 追踪消息队列操作的辅助函数
func TraceMessageQueue(ctx context.Context, operation, queue string, messageCount int) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("mq.operation", operation),
		attribute.String("mq.queue", queue),
		attribute.Int("mq.message_count", messageCount),
		attribute.String("mq.system", "rabbitmq"),
	)
}

// TraceGRPC 追踪 gRPC 调用的辅助函数
func TraceGRPC(ctx context.Context, method, service string, statusCode grpccodes.Code) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", service),
		attribute.String("rpc.method", method),
		attribute.String("rpc.grpc.status_code", statusCode.String()),
	)

	// Set span status based on gRPC status code
	if statusCode != grpccodes.OK {
		span.SetStatus(codes.Error, fmt.Sprintf("gRPC %s", statusCode.String()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// TraceGRPCClient 追踪 gRPC 客户端调用的辅助函数
func TraceGRPCClient(ctx context.Context, target, method string, statusCode grpccodes.Code) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.target", target),
		attribute.String("rpc.method", method),
		attribute.String("rpc.grpc.status_code", statusCode.String()),
		attribute.String("component", "grpc-client"),
	)

	// Set span status based on gRPC status code
	if statusCode != grpccodes.OK {
		span.SetStatus(codes.Error, fmt.Sprintf("gRPC %s", statusCode.String()))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// WithSpan 执行函数并自动管理 span 生命周期的辅助函数
func WithSpan(ctx context.Context, spanName string, fn func(ctx context.Context) error) error {
	ctx, span := StartSpan(ctx, spanName)
	defer span.End()

	if err := fn(ctx); err != nil {
		RecordError(ctx, err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// WithSpanResult 执行函数并自动管理 span 生命周期的辅助函数（支持返回值）
func WithSpanResult[T any](ctx context.Context, spanName string, fn func(ctx context.Context) (T, error)) (T, error) {
	ctx, span := StartSpan(ctx, spanName)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		RecordError(ctx, err)
		var zero T
		return zero, err
	}

	span.SetStatus(codes.Ok, "")
	return result, nil
}
