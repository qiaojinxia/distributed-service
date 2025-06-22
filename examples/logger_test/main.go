package main

import (
	"context"
	"errors"

	"github.com/qiaojinxia/distributed-service/framework/logger"
)

func main() {
	// 初始化日志系统
	if err := logger.Init(&logger.Config{
		Level:      "debug", // 设置为debug级别以显示所有日志
		Encoding:   "json",
		OutputPath: "stdout",
	}); err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 测试不同级别的日志是否显示 caller 信息
	testLogLevels(ctx)
}

func testLogLevels(ctx context.Context) {
	logger.Debug(ctx, "这是 Debug 级别日志",
		logger.String("level", "debug"),
		logger.String("test", "caller_info"),
	)

	logger.Info(ctx, "这是 Info 级别日志",
		logger.String("level", "info"),
		logger.String("test", "caller_info"),
	)

	logger.Warn(ctx, "这是 Warn 级别日志",
		logger.String("level", "warn"),
		logger.String("test", "caller_info"),
	)

	// 创建一个错误用于测试
	err := errors.New("测试错误")
	logger.Error(ctx, "这是 Error 级别日志",
		logger.String("level", "error"),
		logger.String("test", "caller_info"),
		logger.Err(err),
	)

	// 测试格式化日志
	logger.Debugf(ctx, "格式化 Debug 日志: %s", "测试")
	logger.Infof(ctx, "格式化 Info 日志: %s", "测试")
	logger.Warnf(ctx, "格式化 Warn 日志: %s", "测试")
	logger.Errorf(ctx, "格式化 Error 日志: %s", "测试")

	// 测试子日志器
	subLogger := logger.Default().With(
		logger.String("component", "test-component"),
		logger.String("module", "caller-test"),
	)

	subLogger.Warn(ctx, "子日志器的 Warn 级别日志",
		logger.String("sub_logger", "true"),
	)
}
