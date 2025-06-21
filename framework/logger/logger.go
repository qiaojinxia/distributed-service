package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

type ctxKey struct{}

// Config represents logger configuration
type Config struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

// InitLogger initializes the global logger
func InitLogger(cfg *Config) error {
	// Create encoder config
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

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Create core
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

	// Create logger
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// WithContext adds logger to context
func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, globalLogger)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return logger
	}
	return globalLogger
}

// With creates a child logger with additional fields
func With(fields ...zapcore.Field) *zap.Logger {
	return globalLogger.With(fields...)
}

// Debug logs a debug message with context
func Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// Info logs an info message with context
func Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// Warn logs a warning message with context
func Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// Error logs an error message with context
func Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Error(msg, fields...)
}

// Fatal logs a fatal message with context and exits
func Fatal(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Fatal(msg, fields...)
}

// String Field creators
func String(key, value string) zapcore.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zapcore.Field {
	return zap.Int64(key, value)
}

func Bool(key string, value bool) zapcore.Field {
	return zap.Bool(key, value)
}

func Duration(key string, value time.Duration) zapcore.Field {
	return zap.Duration(key, value)
}

func Any(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}

func Error_(err error) zapcore.Field {
	return zap.Error(err)
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	With(fields ...zapcore.Field) Logger
}

// zapLogger zap日志器的包装实现
type zapLogger struct {
	logger *zap.Logger
}

// Debug 实现Logger接口
func (l *zapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// Info 实现Logger接口
func (l *zapLogger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn 实现Logger接口
func (l *zapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error 实现Logger接口
func (l *zapLogger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 实现Logger接口
func (l *zapLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

// Debugf 格式化debug日志
func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.logger.Sugar().Debugf(template, args...)
}

// Infof 格式化info日志
func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.logger.Sugar().Infof(template, args...)
}

// Warnf 格式化warn日志
func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.logger.Sugar().Warnf(template, args...)
}

// Errorf 格式化error日志
func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.logger.Sugar().Errorf(template, args...)
}

// Fatalf 格式化fatal日志
func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.logger.Sugar().Fatalf(template, args...)
}

// With 创建带字段的子日志器
func (l *zapLogger) With(fields ...zapcore.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// GetLogger 获取全局日志器
func GetLogger() Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		InitLogger(&Config{
			Level:      "info",
			Encoding:   "console",
			OutputPath: "stdout",
		})
	}
	return &zapLogger{logger: globalLogger}
}
