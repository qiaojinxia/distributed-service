package main

import (
	"context"
	"errors"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// 1. 初始化日志系统
	if err := logger.Init(&logger.Config{
		Level:      "info",
		Encoding:   "json",
		OutputPath: "stdout",
	}); err != nil {
		panic(err)
	}

	// 2. 初始化链路追踪
	tracingCfg := &tracing.Config{
		ServiceName:    "logger-demo",
		ServiceVersion: "v1.0.0",
		Environment:    "development",
		Enabled:        true,
		ExporterType:   "console",
		SampleRatio:    1.0,
	}
	tracingManager, err := tracing.NewTracingManager(context.Background(), tracingCfg)
	if err != nil {
		panic(err)
	}
	defer tracingManager.Shutdown(context.Background())

	// 获取 tracer
	tracer := otel.Tracer("logger-demo")

	// 演示不同的日志使用方式
	demonstrateLogging(tracer)
}

func demonstrateLogging(tracer interface{}) {
	// 创建带有 trace 的 context
	ctx, span := tracer.(interface {
		Start(ctx context.Context, spanName string, opts ...interface{}) (context.Context, interface{})
	}).Start(context.Background(), "demo-operation", attribute.String("operation", "logging_demo"))
	defer func() {
		if s, ok := span.(interface{ End() }); ok {
			s.End()
		}
	}()

	// 1. 基本日志记录（包级函数方式）
	logger.Info(ctx, "服务启动成功",
		logger.String("service", "logger-demo"),
		logger.String("version", "v1.0.0"),
		logger.Int("port", 8080),
	)

	// 2. 格式化日志
	logger.Infof(ctx, "用户 %s 登录成功，耗时 %v", "alice", time.Millisecond*250)

	// 3. 使用业务便捷字段
	logger.Info(ctx, "处理用户请求",
		logger.UserID("user123"),
		logger.RequestID("req456"),
		logger.Method("POST"),
		logger.Path("/api/users"),
		logger.StatusCode(201),
		logger.Latency(time.Millisecond*150),
	)

	// 4. 错误日志
	err := errors.New("数据库连接失败")
	logger.Error(ctx, "操作失败",
		logger.Err(err),
		logger.String("operation", "create_user"),
		logger.Database("user_db"),
	)

	// 5. 使用日志器实例
	userLogger := logger.Default().With(
		logger.Service("user-service"),
		logger.Version("v2.1.0"),
	)

	userLogger.Info(ctx, "用户服务初始化",
		logger.Count(100),
		logger.Duration("startup_time", time.Second*2),
	)

	// 6. 链式字段构建
	fields := logger.NewFields().
		String("module", "payment").
		Int("amount", 9999).
		Bool("is_test", true).
		Duration("processing_time", time.Millisecond*500).
		Build()

	logger.Info(ctx, "支付处理完成", fields...)

	// 7. 带上下文的子日志器
	paymentLogger := logger.Default().WithContext(ctx,
		logger.Service("payment-service"),
		logger.Environment("production"),
	)

	paymentLogger.Warn(ctx, "支付金额异常",
		logger.Int("amount", 0),
		logger.UserID("user789"),
	)

	// 8. 性能监控日志
	start := time.Now()
	simulateWork()
	duration := time.Since(start)

	logger.Info(ctx, "API调用完成",
		logger.Path("/api/orders"),
		logger.Method("GET"),
		logger.ResponseTime(duration),
		logger.ResponseSize(1024),
		logger.StatusCode(200),
	)

	// 9. 数据库操作日志
	logger.Debug(ctx, "执行SQL查询",
		logger.Database("orders"),
		logger.Table("order_items"),
		logger.SQL("SELECT * FROM order_items WHERE order_id = ?"),
		logger.Count(5),
	)

	// 10. 消息队列日志
	logger.Info(ctx, "消息发送成功",
		logger.Queue("order-events"),
		logger.Topic("order.created"),
		logger.String("message_id", "msg_123"),
		logger.Size(256),
	)
}

func simulateWork() {
	time.Sleep(time.Millisecond * 100)
}
