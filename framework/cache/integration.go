package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/database"
)

// IntegrationManager 集成管理器，用于管理框架中的缓存实例
type IntegrationManager struct {
	*Manager
	frameworkConfig *config.Config
}

// NewIntegrationManager 创建集成管理器
func NewIntegrationManager(frameworkConfig *config.Config) *IntegrationManager {
	manager := NewManager()

	// 注册标准构建器
	manager.RegisterBuilder(TypeMemory, &MemoryBuilder{})
	
	// 注册框架Redis构建器（使用框架的Redis客户端）
	if database.RedisClient != nil {
		manager.RegisterBuilder(TypeRedis, NewSimpleRedisBuilder(database.RedisClient))
		manager.RegisterBuilder("framework-redis", NewSimpleRedisBuilder(database.RedisClient))
	}
	
	// 注册混合缓存构建器
	manager.RegisterBuilder(TypeHybrid, NewHybridBuilder(manager))

	return &IntegrationManager{
		Manager:         manager,
		frameworkConfig: frameworkConfig,
	}
}

// CreateFrameworkRedisCache 创建使用框架Redis的缓存
func (im *IntegrationManager) CreateFrameworkRedisCache(name, keyPrefix string) error {
	cfg := Config{
		Type: "framework-redis",
		Name: name,
		Settings: map[string]interface{}{
			"key_prefix": keyPrefix,
		},
	}

	return im.CreateCache(cfg)
}

// CreateHybridCacheWithFrameworkRedis 创建使用框架Redis作为L2的混合缓存
func (im *IntegrationManager) CreateHybridCacheWithFrameworkRedis(name string, hybridConfig HybridConfig) error {
	// 确保L2使用框架Redis
	hybridConfig.L2Config = Config{
		Type: "framework-redis",
		Name: "l2-" + name,
		Settings: map[string]interface{}{
			"key_prefix": hybridConfig.L2Config.Settings["key_prefix"],
		},
	}

	cfg := Config{
		Type: TypeHybrid,
		Name: name,
		Settings: map[string]interface{}{
			"l1_config":             hybridConfig.L1Config,
			"l2_config":             hybridConfig.L2Config,
			"sync_strategy":         hybridConfig.SyncStrategy,
			"write_back_enabled":    hybridConfig.WriteBackEnabled,
			"write_back_interval":   hybridConfig.WriteBackInterval,
			"write_back_batch_size": hybridConfig.WriteBackBatchSize,
			"l1_ttl":                hybridConfig.L1TTL,
			"l2_ttl":                hybridConfig.L2TTL,
		},
	}

	return im.CreateCache(cfg)
}

// CreateCacheFromFrameworkConfig 从框架配置创建缓存
func (im *IntegrationManager) CreateCacheFromFrameworkConfig() error {
	if im.frameworkConfig == nil {
		return fmt.Errorf("framework config not provided")
	}

	// 创建默认的框架Redis缓存
	if database.RedisClient != nil {
		err := im.CreateFrameworkRedisCache("default", "app")
		if err != nil {
			return fmt.Errorf("failed to create default framework redis cache: %w", err)
		}
	}

	// 创建默认的混合缓存
	err := im.CreateHybridCacheWithFrameworkRedis("hybrid", DefaultHybridConfig)
	if err != nil {
		return fmt.Errorf("failed to create default hybrid cache: %w", err)
	}

	return nil
}

// GetOrCreateFrameworkCache 获取或创建框架缓存
func (im *IntegrationManager) GetOrCreateFrameworkCache(name, keyPrefix string) (Cache, error) {
	// 先尝试获取现有缓存
	cache, err := im.GetCache(name)
	if err == nil {
		return cache, nil
	}

	// 如果不存在，创建新的框架Redis缓存
	err = im.CreateFrameworkRedisCache(name, keyPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create framework cache %s: %w", name, err)
	}

	return im.GetCache(name)
}

// ManagerOptions  缓存管理器选项
type ManagerOptions struct {
	DefaultKeyPrefix  string        // 默认键前缀
	EnableAutoCreate  bool          // 是否启用自动创建
	DefaultTTL        time.Duration // 默认过期时间
	EnableMetrics     bool          // 是否启用指标收集
	EnableHybridCache bool          // 是否启用混合缓存
	HybridCacheConfig HybridConfig  // 混合缓存配置
}

// DefaultCacheManagerOptions 默认缓存管理器选项
func DefaultCacheManagerOptions() ManagerOptions {
	return ManagerOptions{
		DefaultKeyPrefix:  "app",
		EnableAutoCreate:  true,
		DefaultTTL:        time.Hour,
		EnableMetrics:     true,
		EnableHybridCache: true,
		HybridCacheConfig: DefaultHybridConfig,
	}
}

// Service 缓存服务，提供高级缓存操作
type Service struct {
	manager *IntegrationManager
	options ManagerOptions
}

// NewCacheService 创建缓存服务
func NewCacheService(frameworkConfig *config.Config, options ...ManagerOptions) *Service {
	opts := DefaultCacheManagerOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	return &Service{
		manager: NewIntegrationManager(frameworkConfig),
		options: opts,
	}
}

// Initialize 初始化缓存服务
func (cs *Service) Initialize(ctx context.Context) error {
	// 创建默认缓存配置
	err := cs.manager.CreateCacheFromFrameworkConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize cache service: %w", err)
	}

	return nil
}

// GetCache 获取指定名称的缓存
func (cs *Service) GetCache(name string) (Cache, error) {
	cache, err := cs.manager.GetCache(name)
	if err != nil && cs.options.EnableAutoCreate {
		// 自动创建缓存
		return cs.manager.GetOrCreateFrameworkCache(name, cs.options.DefaultKeyPrefix+":"+name)
	}
	return cache, err
}

// GetDefaultCache 获取默认缓存
func (cs *Service) GetDefaultCache() (Cache, error) {
	return cs.GetCache("default")
}

// GetHybridCache 获取混合缓存
func (cs *Service) GetHybridCache() (Cache, error) {
	return cs.GetCache("hybrid")
}

// CreateNamespaceCache 创建命名空间缓存
func (cs *Service) CreateNamespaceCache(namespace string) (Cache, error) {
	keyPrefix := cs.options.DefaultKeyPrefix + ":" + namespace
	return cs.manager.GetOrCreateFrameworkCache(namespace, keyPrefix)
}

// WithRedisClient 使用外部Redis客户端创建缓存
func (cs *Service) WithRedisClient(name string, client *redis.Client, keyPrefix string) error {
	cache := NewSimpleRedisCache(client, keyPrefix)
	cs.manager.caches[name] = cache
	return nil
}

// Close 关闭缓存服务
func (cs *Service) Close() error {
	return cs.manager.Close()
}

// GetStats 获取所有缓存的统计信息
func (cs *Service) GetStats() map[string]Stats {
	stats := make(map[string]Stats)

	for _, name := range cs.manager.ListCaches() {
		if cache, err := cs.manager.GetCache(name); err == nil {
			if statsCache, ok := cache.(interface{ GetStats() Stats }); ok {
				stats[name] = statsCache.GetStats()
			}
		}
	}

	return stats
}
