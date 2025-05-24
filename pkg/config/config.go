package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Consul   ConsulConfig   `mapstructure:"consul"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
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
