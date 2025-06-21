// Package middleware provides gRPC common middleware
package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"distributed-service/pkg/logger"
	"distributed-service/pkg/metrics"
	"distributed-service/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// ===== gRPC Common Middleware =====

// GRPCLoggingInterceptor logs gRPC requests
func GRPCLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Get client IP
		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		logger.Info(ctx, "gRPC request started",
			zap.String("method", info.FullMethod),
			zap.String("client_ip", clientIP))

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			logger.Error(ctx, "gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("client_ip", clientIP),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			logger.Info(ctx, "gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.String("client_ip", clientIP),
				zap.Duration("duration", duration))
		}

		return resp, err
	}
}

// GRPCStreamLoggingInterceptor logs gRPC stream requests
func GRPCStreamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx := stream.Context()

		// Get client IP
		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		logger.Info(ctx, "gRPC stream started",
			zap.String("method", info.FullMethod),
			zap.String("client_ip", clientIP),
			zap.Bool("client_stream", info.IsClientStream),
			zap.Bool("server_stream", info.IsServerStream))

		err := handler(srv, stream)

		duration := time.Since(start)
		if err != nil {
			logger.Error(ctx, "gRPC stream failed",
				zap.String("method", info.FullMethod),
				zap.String("client_ip", clientIP),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			logger.Info(ctx, "gRPC stream completed",
				zap.String("method", info.FullMethod),
				zap.String("client_ip", clientIP),
				zap.Duration("duration", duration))
		}

		return err
	}
}

// GRPCRecoveryInterceptor recovers from panics in gRPC handlers
func GRPCRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(ctx, "gRPC panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())))

				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// GRPCStreamRecoveryInterceptor recovers from panics in gRPC stream handlers
func GRPCStreamRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(stream.Context(), "gRPC stream panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())))

				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()

		return handler(srv, stream)
	}
}

// GRPCMetricsInterceptor records metrics for gRPC requests
func GRPCMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Record metrics
		methodName := getMethodName(info.FullMethod)
		status := "success"
		if err != nil {
			status = "error"
		}

		metrics.GRPCRequestsTotal.WithLabelValues(methodName, status).Inc()
		metrics.GRPCRequestDuration.WithLabelValues(methodName, status).Observe(duration.Seconds())

		return resp, err
	}
}

// GRPCTracingInterceptor adds tracing to gRPC requests
func GRPCTracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		serviceName := getServiceName(info.FullMethod)
		methodName := getMethodName(info.FullMethod)

		spanName := fmt.Sprintf("grpc.%s/%s", serviceName, methodName)

		ctx, span := tracing.StartSpan(ctx, spanName)
		defer span.End()

		// Add attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", serviceName),
			attribute.String("rpc.method", methodName),
			attribute.String("rpc.grpc.method", info.FullMethod),
		)

		// Get client info
		if p, ok := peer.FromContext(ctx); ok {
			span.SetAttributes(attribute.String("net.peer.name", p.Addr.String()))
		}

		resp, err := handler(ctx, req)

		if err != nil {
			tracing.RecordError(ctx, err)
			span.SetAttributes(attribute.String("rpc.grpc.status_code", status.Code(err).String()))
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", codes.OK.String()))
		}

		return resp, err
	}
}

// GRPCStreamTracingInterceptor adds tracing to gRPC stream requests
func GRPCStreamTracingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		serviceName := getServiceName(info.FullMethod)
		methodName := getMethodName(info.FullMethod)

		spanName := fmt.Sprintf("grpc.%s/%s", serviceName, methodName)

		ctx, span := tracing.StartSpan(ctx, spanName)
		defer span.End()

		// Add attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", serviceName),
			attribute.String("rpc.method", methodName),
			attribute.String("rpc.grpc.method", info.FullMethod),
			attribute.Bool("rpc.grpc.client_stream", info.IsClientStream),
			attribute.Bool("rpc.grpc.server_stream", info.IsServerStream),
		)

		// Wrap stream with context
		wrappedStream := WrapServerStream(stream, ctx)

		err := handler(srv, wrappedStream)

		if err != nil {
			tracing.RecordError(ctx, err)
			span.SetAttributes(attribute.String("rpc.grpc.status_code", status.Code(err).String()))
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", codes.OK.String()))
		}

		return err
	}
}

// ===== Helper Types and Functions =====

// wrappedServerStream wraps grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the custom context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// WrapServerStream wraps a grpc.ServerStream with a custom context
func WrapServerStream(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &wrappedServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}
