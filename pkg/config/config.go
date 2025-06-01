package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	GRPC       GRPCConfig       `mapstructure:"grpc"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Consul     ConsulConfig     `mapstructure:"consul"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	Redis      RedisConfig      `mapstructure:"redis"`
	RabbitMQ   RabbitMQConfig   `mapstructure:"rabbitmq"`
	Tracing    TracingConfig    `mapstructure:"tracing"`
	Protection ProtectionConfig `mapstructure:"protection"`
}

type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Tags    string `mapstructure:"tags"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Issuer    string `mapstructure:"issuer"`
}

type ConsulConfig struct {
	Host                           string `mapstructure:"host"`
	Port                           int    `mapstructure:"port"`
	ServiceCheckInterval           string `mapstructure:"service_check_interval"`
	DeregisterCriticalServiceAfter string `mapstructure:"deregister_critical_service_after"`
}

type MetricsConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	PrometheusPort int  `mapstructure:"prometheus_port"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

type TracingConfig struct {
	ServiceName    string  `mapstructure:"service_name"`
	ServiceVersion string  `mapstructure:"service_version"`
	Environment    string  `mapstructure:"environment"`
	Enabled        bool    `mapstructure:"enabled"`
	ExporterType   string  `mapstructure:"exporter_type"` // "otlp", "stdout", "jaeger"
	Endpoint       string  `mapstructure:"endpoint"`
	SampleRatio    float64 `mapstructure:"sample_ratio"`
}

type GRPCConfig struct {
	Port                  int    `mapstructure:"port"`
	MaxRecvMsgSize        int    `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize        int    `mapstructure:"max_send_msg_size"`
	ConnectionTimeout     string `mapstructure:"connection_timeout"`
	MaxConnectionIdle     string `mapstructure:"max_connection_idle"`
	MaxConnectionAge      string `mapstructure:"max_connection_age"`
	MaxConnectionAgeGrace string `mapstructure:"max_connection_age_grace"`
	Time                  string `mapstructure:"time"`
	Timeout               string `mapstructure:"timeout"`
	EnableReflection      bool   `mapstructure:"enable_reflection"`
	EnableHealthCheck     bool   `mapstructure:"enable_health_check"`
}

var GlobalConfig Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// ProtectionConfig 保护配置
type ProtectionConfig struct {
	Enabled         bool                       `mapstructure:"enabled"`
	Storage         ProtectionStorageConfig    `mapstructure:"storage"`
	RateLimitRules  []RateLimitRuleConfig      `mapstructure:"rate_limit_rules"`
	CircuitBreakers []CircuitBreakerRuleConfig `mapstructure:"circuit_breakers"`
}

// ProtectionStorageConfig 保护存储配置
type ProtectionStorageConfig struct {
	Type   string      `mapstructure:"type"`   // memory, redis
	Prefix string      `mapstructure:"prefix"` // 键前缀
	TTL    string      `mapstructure:"ttl"`    // 数据过期时间
	Redis  RedisConfig `mapstructure:"redis"`  // Redis 配置
	Memory struct {
		MaxEntries  int    `mapstructure:"max_entries"`  // 最大条目数
		CleanupTick string `mapstructure:"cleanup_tick"` // 清理间隔
	} `mapstructure:"memory"` // 内存配置
}

// RateLimitRuleConfig 限流规则配置 - 统一命名规范
type RateLimitRuleConfig struct {
	Name           string  `mapstructure:"name"`             // 限流规则名称
	Resource       string  `mapstructure:"resource"`         // 资源标识 (原key字段)
	Threshold      float64 `mapstructure:"threshold"`        // 统计窗口内允许的最大请求数量
	StatIntervalMs uint32  `mapstructure:"stat_interval_ms"` // 统计窗口时间(毫秒)，QPS = (Threshold × 1000) / StatIntervalMs
	Enabled        bool    `mapstructure:"enabled"`
	Description    string  `mapstructure:"description"`
}

// CircuitBreakerRuleConfig 熔断器规则配置
type CircuitBreakerRuleConfig struct {
	Name                         string  `mapstructure:"name"`
	Resource                     string  `mapstructure:"resource"` // 资源名称，如果为空则使用name
	Strategy                     string  `mapstructure:"strategy"` // 熔断策略: "ErrorRatio", "ErrorCount", "SlowRequestRatio"
	Enabled                      bool    `mapstructure:"enabled"`
	RetryTimeoutMs               uint32  `mapstructure:"retry_timeout_ms"`                 // 熔断后重试超时时间(毫秒)
	MinRequestAmount             uint64  `mapstructure:"min_request_amount"`               // 触发熔断的最小请求数
	StatIntervalMs               uint32  `mapstructure:"stat_interval_ms"`                 // 统计时间窗口(毫秒)
	StatSlidingWindowBucketCount uint32  `mapstructure:"stat_sliding_window_bucket_count"` // 滑动窗口桶数
	MaxAllowedRtMs               uint64  `mapstructure:"max_allowed_rt_ms"`                // 最大允许响应时间(毫秒)，仅慢调用策略有效
	Threshold                    float64 `mapstructure:"threshold"`                        // 熔断阈值，根据策略不同含义不同
	ProbeNum                     uint64  `mapstructure:"probe_num"`                        // 半开状态探测请求数量
	Description                  string  `mapstructure:"description"`
}
