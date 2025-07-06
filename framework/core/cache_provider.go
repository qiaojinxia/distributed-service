package core

import (
	"github.com/qiaojinxia/distributed-service/framework/cache"
)

// Provider 实现 middleware.CacheProvider 接口
type Provider struct{}

// GetUserCache 获取用户缓存
func (cp *Provider) GetUserCache() cache.Cache {
	return GetUserCache()
}

// GetSessionCache 获取会话缓存
func (cp *Provider) GetSessionCache() cache.Cache {
	return GetSessionCache()
}

// GetProductCache 获取产品缓存
func (cp *Provider) GetProductCache() cache.Cache {
	return GetProductCache()
}

// GetConfigCache 获取配置缓存
func (cp *Provider) GetConfigCache() cache.Cache {
	return GetConfigCache()
}

// GetCache 获取指定名称的缓存
func (cp *Provider) GetCache(name string) cache.Cache {
	return GetCache(name)
}

// 全局缓存提供者实例
var defaultCacheProvider = &Provider{}

// GetCacheProvider 获取缓存提供者
func GetCacheProvider() *Provider {
	return defaultCacheProvider
}
