package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/database"
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

// Initialize 初始化缓存服务，使用框架的Redis客户端
func (fcs *FrameworkCacheService) Initialize(ctx context.Context) error {
	// 检查框架Redis客户端是否已初始化
	if database.RedisClient == nil {
		return fmt.Errorf("framework Redis client not initialized, please call database.InitRedis() first")
	}

	// 注册Redis构建器，使用框架的Redis客户端
	fcs.Manager.RegisterBuilder(TypeRedis, NewSimpleRedisBuilder(database.RedisClient))
	fcs.Manager.RegisterBuilder("redis", NewSimpleRedisBuilder(database.RedisClient))

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

// CreateCacheWithFrameworkRedis 便捷函数：使用框架Redis创建缓存
func CreateCacheWithFrameworkRedis(name, keyPrefix string) (Cache, error) {
	if database.RedisClient == nil {
		return nil, fmt.Errorf("framework Redis client not initialized")
	}

	cache := NewSimpleRedisCache(database.RedisClient, keyPrefix)
	return cache, nil
}