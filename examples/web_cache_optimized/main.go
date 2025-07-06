package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
	log.Println("🚀 启动优化后的Web缓存示例...")

	// 🎯 极简启动 - 一行代码启动包含缓存的Web服务
	err := core.New().
		Port(8080).
		Mode("debug").
		Name("web-cache-demo").
		Version("v1.0.0").
		WithCacheForWebApp(). // 自动配置Web应用缓存
		OnlyHTTP().
		HTTP(setupRoutes).
		Run()

	if err != nil {
		log.Fatalf("Web服务启动失败: %v", err)
	}
}

// setupRoutes 设置HTTP路由
func setupRoutes(r interface{}) {
	if engine, ok := r.(*gin.Engine); ok {
		// 📊 缓存演示路由组
		cache := engine.Group("/cache")
		{
			// 方式1: 使用全局缓存API（最简单）
			cache.GET("/simple/set/:key/:value", simpleSetCache)
			cache.GET("/simple/get/:key", simpleGetCache)

			// 方式2: 检查缓存中间件是否启用
			cache.GET("/middleware/check", checkCacheMiddleware)

			// 方式3: 直接使用特定缓存
			cache.GET("/users/set/:id/:name", setUser)
			cache.GET("/users/get/:id", getUser)

			// 方式4: 缓存统计信息
			cache.GET("/stats/:cache_name", getCacheStats)
			cache.GET("/stats", getAllCacheStats)
		}

		// 🏠 健康检查
		engine.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "web-cache-demo",
				"caches":  []string{"users", "sessions", "products", "configs"},
			})
		})

		// 📚 API文档
		engine.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"title":   "缓存框架API演示",
				"version": "v1.0.0",
				"apis": map[string]string{
					"简单缓存设置": "GET /cache/simple/set/:key/:value",
					"简单缓存获取": "GET /cache/simple/get/:key",
					"中间件检查":  "GET /cache/middleware/check",
					"用户缓存设置": "GET /cache/users/set/:id/:name",
					"用户缓存获取": "GET /cache/users/get/:id",
					"缓存统计信息": "GET /cache/stats/:cache_name",
					"所有缓存统计": "GET /cache/stats",
					"健康检查":   "GET /health",
				},
			})
		})
	}
}

// ================================
// 方式1: 全局缓存API（最简单）
// ================================

func simpleSetCache(c *gin.Context) {
	key := c.Param("key")
	value := c.Param("value")

	// 🎯 直接使用全局缓存API
	userCache := core.GetUserCache()
	if userCache == nil {
		c.JSON(500, gin.H{"error": "缓存服务不可用"})
		return
	}

	err := userCache.Set(context.Background(), key, value, time.Hour)
	if err != nil {
		c.JSON(500, gin.H{"error": "缓存设置失败", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "缓存设置成功",
		"method":  "全局API",
		"key":     key,
		"value":   value,
	})
}

func simpleGetCache(c *gin.Context) {
	key := c.Param("key")

	// 🎯 直接使用全局缓存API
	userCache := core.GetUserCache()
	if userCache == nil {
		c.JSON(500, gin.H{"error": "缓存服务不可用"})
		return
	}

	value, err := userCache.Get(context.Background(), key)
	if err != nil {
		c.JSON(404, gin.H{"error": "缓存未找到", "key": key})
		return
	}

	c.JSON(200, gin.H{
		"message": "缓存获取成功",
		"method":  "全局API",
		"key":     key,
		"value":   value,
	})
}

// ================================
// 方式2: 缓存中间件检查
// ================================

func checkCacheMiddleware(c *gin.Context) {
	// 检查缓存中间件是否启用
	if cacheEnabled, exists := c.Get("cache_enabled"); exists && cacheEnabled == true {
		c.JSON(200, gin.H{
			"message":        "缓存中间件已启用",
			"middleware":     "active",
			"recommendation": "建议使用 core.GetCache() 全局API获取缓存",
			"global_api": map[string]string{
				"获取用户缓存": "core.GetUserCache()",
				"获取任意缓存": "core.GetCache(name)",
				"检查缓存存在": "core.HasCache(name)",
			},
		})
	} else {
		c.JSON(500, gin.H{
			"message":    "缓存中间件未启用",
			"middleware": "inactive",
		})
	}
}

// ================================
// 方式3: 业务场景示例
// ================================

func setUser(c *gin.Context) {
	id := c.Param("id")
	name := c.Param("name")

	// 🎯 业务数据结构
	user := map[string]interface{}{
		"id":        id,
		"name":      name,
		"timestamp": time.Now().Unix(),
		"source":    "web-api",
	}

	// 使用专用的用户缓存
	userCache := core.GetUserCache()
	if userCache == nil {
		c.JSON(500, gin.H{"error": "用户缓存不可用"})
		return
	}

	err := userCache.Set(context.Background(), "user:"+id, user, time.Hour*6)
	if err != nil {
		c.JSON(500, gin.H{"error": "用户缓存失败", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "用户缓存成功",
		"user":    user,
		"ttl":     "6小时",
	})
}

func getUser(c *gin.Context) {
	id := c.Param("id")

	// 使用专用的用户缓存
	userCache := core.GetUserCache()
	if userCache == nil {
		c.JSON(500, gin.H{"error": "用户缓存不可用"})
		return
	}

	value, err := userCache.Get(context.Background(), "user:"+id)
	if err != nil {
		c.JSON(404, gin.H{"error": "用户未找到", "id": id})
		return
	}

	c.JSON(200, gin.H{
		"message": "用户获取成功",
		"user":    value,
	})
}

// ================================
// 方式4: 缓存统计监控
// ================================

func getCacheStats(c *gin.Context) {
	cacheName := c.Param("cache_name")

	stats, err := core.GetCacheStats(cacheName)
	if err != nil {
		c.JSON(404, gin.H{"error": "缓存统计获取失败", "cache": cacheName})
		return
	}

	c.JSON(200, gin.H{
		"cache": cacheName,
		"stats": stats,
	})
}

func getAllCacheStats(c *gin.Context) {
	cacheNames := []string{"users", "sessions", "products", "configs"}
	allStats := make(map[string]interface{})

	for _, name := range cacheNames {
		if stats, err := core.GetCacheStats(name); err == nil {
			allStats[name] = stats
		} else {
			allStats[name] = map[string]string{"error": "不可用"}
		}
	}

	c.JSON(200, gin.H{
		"message": "所有缓存统计",
		"stats":   allStats,
	})
}
