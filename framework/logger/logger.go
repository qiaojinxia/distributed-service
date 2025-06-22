package logger

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger  *zap.Logger
	defaultConfig = &Config{
		Level:      "info",
		Encoding:   "console",
		OutputPath: "stdout",
	}
)

// Config 日志配置
type Config struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

// Logger 日志接口 - 统一且简洁
type Logger interface {
	// 带上下文的日志方法（推荐使用，自动包含 trace id）
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)

	// 格式化日志方法
	Debugf(ctx context.Context, template string, args ...interface{})
	Infof(ctx context.Context, template string, args ...interface{})
	Warnf(ctx context.Context, template string, args ...interface{})
	Errorf(ctx context.Context, template string, args ...interface{})
	Fatalf(ctx context.Context, template string, args ...interface{})

	// 创建子日志器
	With(fields ...Field) Logger
	WithContext(ctx context.Context, fields ...Field) Logger
}

// Field 日志字段类型别名，简化使用
type Field = zapcore.Field

// logger 内部实现
type logger struct {
	zap *zap.Logger
}

// 实现 Logger 接口

func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.logWithTraceLevel(ctx, "debug", msg, fields...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.logWithTraceLevel(ctx, "info", msg, fields...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.logWithTraceLevel(ctx, "warn", msg, fields...)
}

func (l *logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.logWithTraceLevel(ctx, "error", msg, fields...)
}

func (l *logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.logWithTraceLevel(ctx, "fatal", msg, fields...)
}

func (l *logger) Debugf(ctx context.Context, template string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		sugar := l.zap.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		).Sugar()
		sugar.Debugf(template, args...)
	} else {
		l.zap.Sugar().Debugf(template, args...)
	}
}

func (l *logger) Infof(ctx context.Context, template string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		sugar := l.zap.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		).Sugar()
		sugar.Infof(template, args...)
	} else {
		l.zap.Sugar().Infof(template, args...)
	}
}

func (l *logger) Warnf(ctx context.Context, template string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		sugar := l.zap.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		).Sugar()
		sugar.Warnf(template, args...)
	} else {
		l.zap.Sugar().Warnf(template, args...)
	}
}

func (l *logger) Errorf(ctx context.Context, template string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		sugar := l.zap.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		).Sugar()
		sugar.Errorf(template, args...)
	} else {
		l.zap.Sugar().Errorf(template, args...)
	}
}

func (l *logger) Fatalf(ctx context.Context, template string, args ...interface{}) {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		sugar := l.zap.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		).Sugar()
		sugar.Fatalf(template, args...)
	} else {
		l.zap.Sugar().Fatalf(template, args...)
	}
}

func (l *logger) With(fields ...Field) Logger {
	return &logger{zap: l.zap.With(fields...)}
}

func (l *logger) WithContext(ctx context.Context, fields ...Field) Logger {
	fields = l.addTraceFields(ctx, fields)
	return &logger{zap: l.zap.With(fields...)}
}

// 内部方法

func (l *logger) logWithTraceLevel(ctx context.Context, level string, msg string, fields ...Field) {
	fields = l.addTraceFields(ctx, fields)

	// 为了正确显示调用位置，我们需要跳过额外的调用层
	// 调用链: main.go -> logger.Warn() -> Default().Warn() -> l.Warn() -> logWithTraceLevel -> zap.Warn
	// 需要跳过: logWithTraceLevel -> l.Warn() -> Default().Warn() -> logger.Warn() = 4层
	loggerWithSkip := l.zap.WithOptions(zap.AddCallerSkip(2))

	// 根据级别调用对应的方法
	switch level {
	case "debug":
		loggerWithSkip.Debug(msg, fields...)
	case "info":
		loggerWithSkip.Info(msg, fields...)
	case "warn":
		loggerWithSkip.Warn(msg, fields...)
	case "error":
		loggerWithSkip.Error(msg, fields...)
	case "fatal":
		loggerWithSkip.Fatal(msg, fields...)
	default:
		loggerWithSkip.Info(msg, fields...)
	}
}

func (l *logger) addTraceFields(ctx context.Context, fields []Field) []Field {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		fields = append(fields,
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return fields
}

// 包级初始化和工厂函数

// Init 初始化全局日志器
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = defaultConfig
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", cfg.Level, err)
	}

	// 创建核心
	var core zapcore.Core
	if cfg.Encoding == "console" {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
	}

	// 创建全局日志器
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// New 创建新的日志器实例
func New(zapLogger *zap.Logger) Logger {
	return &logger{zap: zapLogger}
}

// Default 获取默认日志器
func Default() Logger {
	if globalLogger == nil {
		_ = Init(nil) // 使用默认配置
	}
	return &logger{zap: globalLogger}
}

// 包级便捷函数 - 使用全局日志器

func Debug(ctx context.Context, msg string, fields ...Field) {
	Default().Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	Default().Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	Default().Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	Default().Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	Default().Fatal(ctx, msg, fields...)
}

func Debugf(ctx context.Context, template string, args ...interface{}) {
	Default().Debugf(ctx, template, args...)
}

func Infof(ctx context.Context, template string, args ...interface{}) {
	Default().Infof(ctx, template, args...)
}

func Warnf(ctx context.Context, template string, args ...interface{}) {
	Default().Warnf(ctx, template, args...)
}

func Errorf(ctx context.Context, template string, args ...interface{}) {
	Default().Errorf(ctx, template, args...)
}

func Fatalf(ctx context.Context, template string, args ...interface{}) {
	Default().Fatalf(ctx, template, args...)
}

func With(fields ...Field) Logger {
	return Default().With(fields...)
}

func WithContext(ctx context.Context, fields ...Field) Logger {
	return Default().WithContext(ctx, fields...)
}

// 兼容性函数 - 保持向后兼容

// InitLogger 兼容旧的初始化函数名
func InitLogger(cfg *Config) error {
	return Init(cfg)
}

// GetLogger 兼容旧的获取函数名
func GetLogger() Logger {
	return Default()
}
