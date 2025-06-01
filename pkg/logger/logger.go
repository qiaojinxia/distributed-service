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
