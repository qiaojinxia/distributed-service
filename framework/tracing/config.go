package tracing

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config 追踪配置
type Config struct {
	ServiceName    string  `mapstructure:"service_name" yaml:"service_name"`
	ServiceVersion string  `mapstructure:"service_version" yaml:"service_version"`
	Environment    string  `mapstructure:"environment" yaml:"environment"`
	Enabled        bool    `mapstructure:"enabled" yaml:"enabled"`
	ExporterType   string  `mapstructure:"exporter_type" yaml:"exporter_type"` // "otlp", "stdout", "jaeger"
	Endpoint       string  `mapstructure:"endpoint" yaml:"endpoint"`
	SampleRatio    float64 `mapstructure:"sample_ratio" yaml:"sample_ratio"`
}

// Manager TracingManager 追踪管理器
type Manager struct {
	tracer   *sdktrace.TracerProvider
	config   *Config
	shutdown func(context.Context) error
}

// NewTracingManager 创建新的追踪管理器
func NewTracingManager(ctx context.Context, config *Config) (*Manager, error) {
	if !config.Enabled {
		return &Manager{
			config:   config,
			shutdown: func(context.Context) error { return nil },
		}, nil
	}

	// 创建资源
	res, err := newResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建导出器
	exporter, err := newExporter(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRatio)),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Manager{
		tracer: tp,
		config: config,
		shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}

// Shutdown 关闭追踪管理器
func (tm *Manager) Shutdown(ctx context.Context) error {
	if tm.shutdown != nil {
		return tm.shutdown(ctx)
	}
	return nil
}

// newResource 创建资源
func newResource(config *Config) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
			attribute.String("library.language", "go"),
		),
	)
}

// newExporter 创建导出器
func newExporter(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	switch config.ExporterType {
	case "otlp":
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(config.Endpoint),
			otlptracehttp.WithInsecure(), // 在生产环境中应该使用 TLS
		)
	case "stdout":
		return stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	default:
		return stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	}
}
