package cache

import (
	"time"

	"github.com/qiaojinxia/distributed-service/framework/config"
)

// DefaultHybridConfig 默认混合缓存配置
var DefaultHybridConfig = HybridConfig{
	L1Config: Config{
		Type: TypeMemory,
		Name: "l1",
		Settings: map[string]interface{}{
			"max_size":         10000,
			"default_ttl":      "1h",
			"cleanup_interval": "10m",
		},
	},
	L2Config: Config{
		Type: TypeRedis,
		Name: "l2",
		Settings: map[string]interface{}{
			"key_prefix": "hybrid",
		},
	},
	SyncStrategy:       SyncStrategyWriteThrough,
	WriteBackEnabled:   true,
	WriteBackInterval:  time.Minute * 5,
	WriteBackBatchSize: 100,
	L1TTL:              time.Hour,
	L2TTL:              time.Hour * 24,
}

// DefaultCacheStrategies 默认缓存策略
type DefaultCacheStrategies struct{}

// GetWebAppDefaults 获取Web应用默认缓存策略
func (d *DefaultCacheStrategies) GetWebAppDefaults() *config.CacheConfig {
	return &config.CacheConfig{
		Enabled:         true,
		DefaultType:     "hybrid",
		UseFramework:    true,
		GlobalKeyPrefix: "webapp",
		DefaultTTL:      "2h",
		Caches: map[string]config.CacheInstance{
			"users": {
				Type:      "redis",
				KeyPrefix: "users",
				TTL:       "6h",
				Settings: map[string]interface{}{
					"compress": true,
				},
			},
			"sessions": {
				Type:      "hybrid",
				KeyPrefix: "sessions",
				TTL:       "2h",
				Settings: map[string]interface{}{
					"l1_size": 5000,
					"l2_ttl":  "4h",
				},
			},
			"products": {
				Type:      "hybrid",
				KeyPrefix: "products",
				TTL:       "1h",
				Settings: map[string]interface{}{
					"l1_size": 10000,
					"l2_ttl":  "24h",
				},
			},
			"configs": {
				Type:      "redis",
				KeyPrefix: "configs",
				TTL:       "24h",
				Settings: map[string]interface{}{
					"compress": true,
				},
			},
		},
	}
}

// GetAPIDefaults 获取API服务默认缓存策略
func (d *DefaultCacheStrategies) GetAPIDefaults() *config.CacheConfig {
	return &config.CacheConfig{
		Enabled:         true,
		DefaultType:     "redis",
		UseFramework:    true,
		GlobalKeyPrefix: "api",
		DefaultTTL:      "1h",
		Caches: map[string]config.CacheInstance{
			"auth": {
				Type:      "redis",
				KeyPrefix: "auth",
				TTL:       "30m",
				Settings: map[string]interface{}{
					"compress": false, // JWT tokens don't compress well
				},
			},
			"rate_limit": {
				Type:      "redis",
				KeyPrefix: "rate",
				TTL:       "1h",
				Settings: map[string]interface{}{
					"compress": false,
				},
			},
			"data": {
				Type:      "redis",
				KeyPrefix: "data",
				TTL:       "2h",
				Settings: map[string]interface{}{
					"compress": true,
				},
			},
		},
	}
}

// GetMicroserviceDefaults 获取微服务默认缓存策略
func (d *DefaultCacheStrategies) GetMicroserviceDefaults() *config.CacheConfig {
	return &config.CacheConfig{
		Enabled:         true,
		DefaultType:     "hybrid",
		UseFramework:    true,
		GlobalKeyPrefix: "ms",
		DefaultTTL:      "30m",
		Caches: map[string]config.CacheInstance{
			"service_discovery": {
				Type:      "memory",
				KeyPrefix: "discovery",
				TTL:       "5m",
				Settings: map[string]interface{}{
					"max_size": 1000,
				},
			},
			"config": {
				Type:      "redis",
				KeyPrefix: "config",
				TTL:       "10m",
				Settings: map[string]interface{}{
					"compress": true,
				},
			},
			"metrics": {
				Type:      "memory",
				KeyPrefix: "metrics",
				TTL:       "1m",
				Settings: map[string]interface{}{
					"max_size": 5000,
				},
			},
		},
	}
}

// GetDevelopmentDefaults 获取开发环境默认缓存策略
func (d *DefaultCacheStrategies) GetDevelopmentDefaults() *config.CacheConfig {
	return &config.CacheConfig{
		Enabled:         true,
		DefaultType:     "memory",
		UseFramework:    false,
		GlobalKeyPrefix: "dev",
		DefaultTTL:      "5m", // 开发环境短TTL便于测试
		Caches: map[string]config.CacheInstance{
			"debug": {
				Type:      "memory",
				KeyPrefix: "debug",
				TTL:       "1m",
				Settings: map[string]interface{}{
					"max_size":         1000,
					"eviction_policy":  "lru",
					"enable_metrics":   true,
					"cleanup_interval": "30s",
				},
			},
			"test": {
				Type:      "memory",
				KeyPrefix: "test",
				TTL:       "30s",
				Settings: map[string]interface{}{
					"max_size":         500,
					"eviction_policy":  "ttl",
					"cleanup_interval": "10s",
				},
			},
			"performance": {
				Type:      "memory",
				KeyPrefix: "perf",
				TTL:       "2m",
				Settings: map[string]interface{}{
					"max_size":         2000,
					"eviction_policy":  "simple",
					"cleanup_interval": "30s",
				},
			},
		},
	}
}

// DefaultStrategies 全局默认策略实例
var DefaultStrategies = &DefaultCacheStrategies{}

// GetRecommendedStrategy 根据应用特征推荐策略
func GetRecommendedStrategy(appType string, mode string) *config.CacheConfig {
	switch mode {
	case "debug", "development", "dev":
		return DefaultStrategies.GetDevelopmentDefaults()
	}
	switch appType {
	case "webapp", "web", "frontend":
		return DefaultStrategies.GetWebAppDefaults()
	case "api", "rest", "backend":
		return DefaultStrategies.GetAPIDefaults()
	case "microservice", "ms", "service":
		return DefaultStrategies.GetMicroserviceDefaults()
	default:
		// 默认推荐Web应用策略
		return DefaultStrategies.GetWebAppDefaults()
	}
}
