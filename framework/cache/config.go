package cache

import (
	"github.com/qiaojinxia/distributed-service/framework/config"
	"time"
)

// 预设配置模板
var (
	// HighPerformanceHybridConfig 高性能混合缓存配置
	HighPerformanceHybridConfig = HybridConfig{
		L1Config: Config{
			Type: TypeMemory,
			Name: "l1",
			Settings: map[string]interface{}{
				"max_size":         50000,
				"default_ttl":      "30m",
				"cleanup_interval": "5m",
			},
		},
		L2Config: Config{
			Type: TypeRedis,
			Name: "l2",
			Settings: map[string]interface{}{
				"addr":      "localhost:6379",
				"password":  "",
				"db":        0,
				"pool_size": 20,
			},
		},
		SyncStrategy:       SyncStrategyWriteBack,
		WriteBackEnabled:   true,
		WriteBackInterval:  time.Minute * 2,
		WriteBackBatchSize: 200,
		L1TTL:              time.Minute * 30,
		L2TTL:              time.Hour * 12,
	}

	// LowMemoryHybridConfig 低内存混合缓存配置
	LowMemoryHybridConfig = HybridConfig{
		L1Config: Config{
			Type: TypeMemory,
			Name: "l1",
			Settings: map[string]interface{}{
				"max_size":         1000,
				"default_ttl":      "15m",
				"cleanup_interval": "5m",
			},
		},
		L2Config: Config{
			Type: TypeRedis,
			Name: "l2",
			Settings: map[string]interface{}{
				"addr":     "localhost:6379",
				"password": "",
				"db":       0,
			},
		},
		SyncStrategy:       SyncStrategyWriteAround,
		WriteBackEnabled:   false,
		WriteBackInterval:  0,
		WriteBackBatchSize: 0,
		L1TTL:              time.Minute * 15,
		L2TTL:              time.Hour * 48,
	}
)

// ConfigPresets 配置预设
type ConfigPresets struct{}

// GetDefaultHybridConfig 获取默认混合缓存配置
func (cp *ConfigPresets) GetDefaultHybridConfig() HybridConfig {
	return DefaultHybridConfig
}

// GetHighPerformanceHybridConfig 获取高性能混合缓存配置
func (cp *ConfigPresets) GetHighPerformanceHybridConfig() HybridConfig {
	return HighPerformanceHybridConfig
}

// GetLowMemoryHybridConfig 获取低内存混合缓存配置
func (cp *ConfigPresets) GetLowMemoryHybridConfig() HybridConfig {
	return LowMemoryHybridConfig
}

// CustomHybridConfig 自定义混合缓存配置构建器
type CustomHybridConfig struct {
	config HybridConfig
}

// NewCustomHybridConfig 创建自定义混合缓存配置构建器
func NewCustomHybridConfig() *CustomHybridConfig {
	return &CustomHybridConfig{
		config: DefaultHybridConfig,
	}
}

// WithL1Memory 设置L1为内存缓存
func (c *CustomHybridConfig) WithL1Memory(maxSize int, ttl time.Duration) *CustomHybridConfig {
	c.config.L1Config = Config{
		Type: TypeMemory,
		Name: "l1",
		Settings: map[string]interface{}{
			"max_size":         maxSize,
			"default_ttl":      ttl.String(),
			"cleanup_interval": "10m",
		},
	}
	c.config.L1TTL = ttl
	return c
}

// WithL2Redis 设置L2为Redis缓存
func (c *CustomHybridConfig) WithL2Redis(addr, password string, db int, ttl time.Duration) *CustomHybridConfig {
	c.config.L2Config = Config{
		Type: TypeRedis,
		Name: "l2",
		Settings: map[string]interface{}{
			"addr":     addr,
			"password": password,
			"db":       db,
		},
	}
	c.config.L2TTL = ttl
	return c
}

// WithSyncStrategy 设置同步策略
func (c *CustomHybridConfig) WithSyncStrategy(strategy SyncStrategy) *CustomHybridConfig {
	c.config.SyncStrategy = strategy
	return c
}

// WithWriteBack 设置写回配置
func (c *CustomHybridConfig) WithWriteBack(enabled bool, interval time.Duration, batchSize int) *CustomHybridConfig {
	c.config.WriteBackEnabled = enabled
	c.config.WriteBackInterval = interval
	c.config.WriteBackBatchSize = batchSize
	return c
}

// Build 构建配置
func (c *CustomHybridConfig) Build() HybridConfig {
	return c.config
}

// Presets 全局配置预设实例
var Presets = &ConfigPresets{}

// NewHybridConfigFromFramework 从框架配置创建混合缓存配置
func NewHybridConfigFromFramework(frameworkConfig *config.Config, keyPrefix string) HybridConfig {
	hybridConfig := DefaultHybridConfig

	// 使用框架Redis配置
	hybridConfig.L2Config = Config{
		Type: "framework-redis",
		Name: "l2",
		Settings: map[string]interface{}{
			"key_prefix": keyPrefix,
		},
	}

	return hybridConfig
}

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	UseFramework bool   `json:"use_framework" yaml:"use_framework"` // 使用框架Redis
	KeyPrefix    string `json:"key_prefix" yaml:"key_prefix"`       // 键前缀
	// 如果不使用框架Redis，可以单独配置
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
}

// NewRedisConfigFromFramework 从框架配置创建Redis缓存配置
func NewRedisConfigFromFramework(frameworkConfig *config.Config, keyPrefix string) RedisCacheConfig {
	return RedisCacheConfig{
		UseFramework: true,
		KeyPrefix:    keyPrefix,
	}
}

// ConfigBuilder 缓存配置构建器
type ConfigBuilder struct {
	config Config
}

// NewCacheConfigBuilder 创建缓存配置构建器
func NewCacheConfigBuilder(cacheType Type, name string) *ConfigBuilder {
	return &ConfigBuilder{
		config: Config{
			Type:     cacheType,
			Name:     name,
			Settings: make(map[string]interface{}),
		},
	}
}

// WithSetting 设置配置项
func (b *ConfigBuilder) WithSetting(key string, value interface{}) *ConfigBuilder {
	b.config.Settings[key] = value
	return b
}

// WithKeyPrefix 设置键前缀
func (b *ConfigBuilder) WithKeyPrefix(prefix string) *ConfigBuilder {
	return b.WithSetting("key_prefix", prefix)
}

// WithRedisAddr 设置Redis地址
func (b *ConfigBuilder) WithRedisAddr(addr string) *ConfigBuilder {
	return b.WithSetting("addr", addr)
}

// WithRedisDB 设置Redis数据库
func (b *ConfigBuilder) WithRedisDB(db int) *ConfigBuilder {
	return b.WithSetting("db", db)
}

// WithRedisPassword 设置Redis密码
func (b *ConfigBuilder) WithRedisPassword(password string) *ConfigBuilder {
	return b.WithSetting("password", password)
}

// WithMemoryMaxSize 设置内存缓存最大大小
func (b *ConfigBuilder) WithMemoryMaxSize(size int) *ConfigBuilder {
	return b.WithSetting("max_size", size)
}

// WithTTL 设置TTL
func (b *ConfigBuilder) WithTTL(ttl time.Duration) *ConfigBuilder {
	return b.WithSetting("default_ttl", ttl.String())
}

// UseFrameworkRedis 使用框架Redis
func (b *ConfigBuilder) UseFrameworkRedis() *ConfigBuilder {
	b.config.Type = "framework-redis"
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() Config {
	return b.config
}

// Factory  缓存工厂
type Factory struct{}

// NewCacheFactory 创建缓存工厂
func NewCacheFactory() *Factory {
	return &Factory{}
}

// CreateMemoryCache 创建内存缓存配置
func (f *Factory) CreateMemoryCache(name string, maxSize int, ttl time.Duration) Config {
	return NewCacheConfigBuilder(TypeMemory, name).
		WithMemoryMaxSize(maxSize).
		WithTTL(ttl).
		Build()
}

// CreateRedisCache 创建Redis缓存配置
func (f *Factory) CreateRedisCache(name, addr, password string, db int, keyPrefix string) Config {
	return NewCacheConfigBuilder(TypeRedis, name).
		WithRedisAddr(addr).
		WithRedisPassword(password).
		WithRedisDB(db).
		WithKeyPrefix(keyPrefix).
		Build()
}

// CreateFrameworkRedisCache 创建框架Redis缓存配置
func (f *Factory) CreateFrameworkRedisCache(name, keyPrefix string) Config {
	return NewCacheConfigBuilder("framework-redis", name).
		WithKeyPrefix(keyPrefix).
		Build()
}

// CreateHybridCache 创建混合缓存配置
func (f *Factory) CreateHybridCache(name string, l1Config, l2Config Config, strategy SyncStrategy) Config {
	return Config{
		Type: TypeHybrid,
		Name: name,
		Settings: map[string]interface{}{
			"l1_config":             l1Config,
			"l2_config":             l2Config,
			"sync_strategy":         strategy,
			"write_back_enabled":    true,
			"write_back_interval":   time.Minute * 5,
			"write_back_batch_size": 100,
			"l1_ttl":                time.Hour,
			"l2_ttl":                time.Hour * 24,
		},
	}
}
