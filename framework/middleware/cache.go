package middleware

import (
	"github.com/gin-gonic/gin"
)

// CacheMiddleware 缓存中间件 - 简化版本，避免循环依赖
func CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 中间件仅提供标记，表示支持缓存
		// 实际缓存获取由用户代码中的 core.GetCache() 完成
		c.Set("cache_enabled", true)
		c.Next()
	}
}

// GetCacheFromContext 从gin context获取缓存（简化版本）
func GetCacheFromContext(c *gin.Context, name string) interface{} {
	// 简化版本：由于避免循环依赖，不再自动注入缓存
	// 用户应该直接使用 core.GetCache(name) 获取缓存
	if cacheEnabled, exists := c.Get("cache_enabled"); exists && cacheEnabled == true {
		// 返回一个提示，告知用户应该使用全局API
		return map[string]string{
			"message": "请使用 core.GetCache(\"" + name + "\") 获取缓存",
		}
	}
	return nil
}