package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/database"
	"github.com/qiaojinxia/distributed-service/framework/logger"
)

// FrameworkCacheService 框架缓存服务，使用框架的Redis客户端
type FrameworkCacheService struct {
	Manager *Manager
}

// NewFrameworkCacheService 创建框架缓存服务
func NewFrameworkCacheService() *FrameworkCacheService {
	manager := NewManager()
	
	// 注册标准构建器
	manager.RegisterBuilder(TypeMemory, &MemoryBuilder{})
	
	service := &FrameworkCacheService{
		Manager: manager,
	}
	
	// 注册需要manager的构建器
	manager.RegisterBuilder(TypeHybrid, NewHybridBuilder(manager))

	return service
}

// Initialize 初始化缓存服务
func (fcs *FrameworkCacheService) Initialize(ctx context.Context) error {
	// 注册内存缓存构建器（始终可用）
	fcs.Manager.RegisterBuilder(TypeMemory, &MemoryBuilder{})
	fcs.Manager.RegisterBuilder("memory", &MemoryBuilder{})
	
	// 如果Redis客户端可用，注册Redis构建器
	if database.RedisClient != nil {
		fcs.Manager.RegisterBuilder(TypeRedis, NewSimpleRedisBuilder(database.RedisClient))
		fcs.Manager.RegisterBuilder("redis", NewSimpleRedisBuilder(database.RedisClient))
		logger.Info(ctx, "✅ Framework cache service initialized with Redis support")
	} else {
		logger.Warn(ctx, "⚠️ Redis client not available, using memory-only caching")
	}
	
	// 注册混合缓存构建器
	fcs.Manager.RegisterBuilder(TypeHybrid, &HybridBuilder{})
	fcs.Manager.RegisterBuilder("hybrid", &HybridBuilder{})

	return nil
}

// CreateDefaultCaches 创建默认缓存实例（内存缓存）
func (fcs *FrameworkCacheService) CreateDefaultCaches(ctx context.Context) error {
	logger.Info(ctx, "🔧 Creating default memory-based caches...")
	
	// 默认缓存实例配置 - 使用time.Duration而非字符串
	defaultCaches := map[string]map[string]interface{}{
		"users": {
			"max_size":         1000,
			"eviction_policy":  "lru",
			"default_ttl":      time.Hour * 2,       // 2小时
			"cleanup_interval": time.Minute * 10,    // 10分钟
		},
		"sessions": {
			"max_size":         500,
			"eviction_policy":  "ttl", 
			"default_ttl":      time.Minute * 30,    // 30分钟
			"cleanup_interval": time.Minute * 5,     // 5分钟 - 恢复合理值
		},
		"products": {
			"max_size":         2000,
			"eviction_policy":  "simple",
			"default_ttl":      time.Hour,           // 1小时
			"cleanup_interval": time.Minute * 15,    // 15分钟 - 恢复合理值
		},
		"configs": {
			"max_size":         100,
			"eviction_policy":  "lru",
			"default_ttl":      time.Hour * 24,      // 24小时
			"cleanup_interval": time.Hour,           // 1小时
		},
	}
	
	// 创建默认缓存实例
	for name, settings := range defaultCaches {
		config := Config{
			Type:     TypeMemory,
			Name:     name,
			Settings: settings,
		}
		
		if err := fcs.Manager.CreateCache(config); err != nil {
			logger.Error(ctx, "Failed to create default cache", 
				logger.String("name", name), 
				logger.Err(err))
			continue
		}
		
		logger.Info(ctx, "✅ Default cache created", 
			logger.String("name", name), 
			logger.String("type", "memory"))
	}
	
	return nil
}

// CreateUserCache 创建用户缓存
func (fcs *FrameworkCacheService) CreateUserCache() error {
	return fcs.CreateRedisCache("users", "app:users")
}

// CreateSessionCache 创建会话缓存
func (fcs *FrameworkCacheService) CreateSessionCache() error {
	return fcs.CreateRedisCache("sessions", "app:sessions")
}

// CreateProductCache 创建产品缓存
func (fcs *FrameworkCacheService) CreateProductCache() error {
	return fcs.CreateRedisCache("products", "app:products")
}

// CreateRedisCache 创建Redis缓存
func (fcs *FrameworkCacheService) CreateRedisCache(name, keyPrefix string) error {
	config := Config{
		Type: TypeRedis,
		Name: name,
		Settings: map[string]interface{}{
			"key_prefix": keyPrefix,
		},
	}

	return fcs.Manager.CreateCache(config)
}

// CreateHybridCache 创建混合缓存（L1内存 + L2Redis）
func (fcs *FrameworkCacheService) CreateHybridCache(name string, l1Config Config, keyPrefix string, syncStrategy SyncStrategy) error {
	l2Config := Config{
		Type: TypeRedis,
		Name: "l2-" + name,
		Settings: map[string]interface{}{
			"key_prefix": keyPrefix,
		},
	}

	config := Config{
		Type: TypeHybrid,
		Name: name,
		Settings: map[string]interface{}{
			"l1_config":             l1Config,
			"l2_config":             l2Config,
			"sync_strategy":         syncStrategy,
			"write_back_enabled":    true,
			"write_back_interval":   time.Minute * 5,
			"write_back_batch_size": 100,
			"l1_ttl":                time.Hour,
			"l2_ttl":                time.Hour * 24,
		},
	}

	return fcs.Manager.CreateCache(config)
}

// GetUserCache 获取用户缓存
func (fcs *FrameworkCacheService) GetUserCache() (Cache, error) {
	return fcs.Manager.GetCache("users")
}

// GetSessionCache 获取会话缓存
func (fcs *FrameworkCacheService) GetSessionCache() (Cache, error) {
	return fcs.Manager.GetCache("sessions")
}

// GetProductCache 获取产品缓存
func (fcs *FrameworkCacheService) GetProductCache() (Cache, error) {
	return fcs.Manager.GetCache("products")
}

// GetCache 获取指定名称的缓存
func (fcs *FrameworkCacheService) GetCache(name string) (Cache, error) {
	return fcs.Manager.GetCache(name)
}

// Close 关闭缓存服务
func (fcs *FrameworkCacheService) Close() error {
	return fcs.Manager.Close()
}

// GetStats 获取所有缓存的统计信息
func (fcs *FrameworkCacheService) GetStats() map[string]Stats {
	stats := make(map[string]Stats)
	cacheNames := fcs.Manager.ListCaches()

	for _, name := range cacheNames {
		if cache, err := fcs.Manager.GetCache(name); err == nil {
			if statsCache, ok := cache.(interface{ GetStats() Stats }); ok {
				stats[name] = statsCache.GetStats()
			}
		}
	}

	return stats
}

// GetNamedCache 获取指定名称的缓存（包装版本）
func (fcs *FrameworkCacheService) GetNamedCache(name string) (Cache, error) {
	cache := fcs.Manager.GetNamedCache(name)
	if cache == nil {
		return nil, ErrCacheNotFound
	}
	return cache, nil
}

// CreateCacheWithFrameworkRedis 便捷函数：使用框架Redis创建缓存
func CreateCacheWithFrameworkRedis(name, keyPrefix string) (Cache, error) {
	if database.RedisClient == nil {
		return nil, fmt.Errorf("framework Redis client not initialized")
	}

	cache := NewSimpleRedisCache(database.RedisClient, keyPrefix)
	return cache, nil
}