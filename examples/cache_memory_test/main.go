package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
	log.Println("🔍 纯内存缓存测试（无Redis依赖）...")

	// 启动框架，使用纯内存缓存配置
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("框架启动出现panic: %v", r)
			}
		}()

		err := core.New().
			Port(8084).
			Mode("debug").
			Name("memory-cache-test").
			// 使用纯内存缓存配置，避免Redis依赖
			WithCache(getMemoryCacheConfig()).
			OnlyHTTP().
			HTTP(setupRoutes).
			Run()
		if err != nil {
			log.Printf("框架启动失败: %v", err)
		}
	}()

	// 等待框架启动
	fmt.Println("⏳ 等待框架启动...")
	for i := 1; i <= 5; i++ {
		time.Sleep(time.Second)
		fmt.Printf("等待第%d秒...\n", i)

		// 检查API是否可用
		if testAPIsAvailable() {
			fmt.Printf("✅ 第%d秒检测到API可用！\n", i)
			break
		}

		if i == 5 {
			fmt.Println("继续等待...")
		}
	}

	// 额外等待确保完全初始化
	time.Sleep(time.Second * 2)

	// 最终测试
	fmt.Println("\n📊 纯内存缓存API测试")
	testMemoryCacheAPIs()
}

// getMemoryCacheConfig 获取纯内存缓存配置
func getMemoryCacheConfig() *config.CacheConfig {
	return &config.CacheConfig{
		Enabled:         true,
		DefaultType:     "memory", // 使用内存类型
		UseFramework:    false,    // 不使用框架Redis
		GlobalKeyPrefix: "test",
		DefaultTTL:      "1h",
		Caches: map[string]config.CacheInstance{
			"users": {
				Type:      "memory", // 纯内存
				KeyPrefix: "users",
				TTL:       "2h",
				Settings: map[string]interface{}{
					"max_size":         1000,
					"eviction_policy":  "lru",
					"default_ttl":      time.Hour * 2,
					"cleanup_interval": time.Minute * 10,
				},
			},
			"sessions": {
				Type:      "memory", // 纯内存
				KeyPrefix: "sessions",
				TTL:       "30m",
				Settings: map[string]interface{}{
					"max_size":         500,
					"eviction_policy":  "ttl",
					"default_ttl":      time.Minute * 30,
					"cleanup_interval": time.Minute * 5,
				},
			},
			"products": {
				Type:      "memory", // 纯内存
				KeyPrefix: "products",
				TTL:       "1h",
				Settings: map[string]interface{}{
					"max_size":         2000,
					"eviction_policy":  "simple",
					"default_ttl":      time.Hour,
					"cleanup_interval": time.Minute * 15,
				},
			},
			"configs": {
				Type:      "memory", // 纯内存
				KeyPrefix: "configs",
				TTL:       "24h",
				Settings: map[string]interface{}{
					"max_size":         100,
					"eviction_policy":  "lru",
					"default_ttl":      time.Hour * 24,
					"cleanup_interval": time.Hour,
				},
			},
		},
	}
}

func setupRoutes(r interface{}) {
	if engine, ok := r.(*gin.Engine); ok {
		engine.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		engine.GET("/cache-test", func(c *gin.Context) {
			userCache := core.GetUserCache()
			if userCache == nil {
				c.JSON(500, gin.H{"error": "缓存不可用"})
				return
			}

			ctx := context.Background()

			// 测试设置
			err := userCache.Set(ctx, "api_test", "api_value", time.Minute)
			if err != nil {
				c.JSON(500, gin.H{"error": "设置失败", "details": err.Error()})
				return
			}

			// 测试获取
			value, err := userCache.Get(ctx, "api_test")
			if err != nil {
				c.JSON(500, gin.H{"error": "获取失败", "details": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"message":    "缓存API测试成功",
				"value":      value,
				"cache_type": "memory",
			})
		})
	}
}

func testAPIsAvailable() bool {
	userCache := core.GetUserCache()
	return userCache != nil
}

func testMemoryCacheAPIs() {
	fmt.Println("纯内存缓存API测试:")

	userCache := core.GetUserCache()
	fmt.Printf("  GetUserCache(): %v\n", userCache != nil)
	if userCache != nil {
		fmt.Println("    ✅ 用户缓存可用")

		// 测试完整的CRUD操作
		ctx := context.Background()

		// 设置
		err := userCache.Set(ctx, "test_user", "test_value", time.Minute)
		if err != nil {
			fmt.Printf("    ❌ 设置失败: %v\n", err)
		} else {
			fmt.Println("    ✅ 设置成功")

			// 获取
			value, err := userCache.Get(ctx, "test_user")
			if err != nil {
				fmt.Printf("    ❌ 获取失败: %v\n", err)
			} else if value == "test_value" {
				fmt.Println("    ✅ 获取成功")

				// 存在性检查
				exists, err := userCache.Exists(ctx, "test_user")
				if err != nil {
					fmt.Printf("    ❌ 存在性检查失败: %v\n", err)
				} else if exists {
					fmt.Println("    ✅ 存在性检查成功")

					// 删除
					err = userCache.Delete(ctx, "test_user")
					if err != nil {
						fmt.Printf("    ❌ 删除失败: %v\n", err)
					} else {
						fmt.Println("    ✅ 删除成功")

						// 验证删除
						exists, err = userCache.Exists(ctx, "test_user")
						if err != nil {
							fmt.Printf("    ❌ 删除验证失败: %v\n", err)
						} else if !exists {
							fmt.Println("    ✅ 删除验证成功")
						} else {
							fmt.Println("    ❌ 删除验证失败：键仍然存在")
						}
					}
				} else {
					fmt.Println("    ❌ 存在性检查失败：返回false")
				}
			} else {
				fmt.Printf("    ❌ 值不匹配: 期望'test_value', 得到'%v'\n", value)
			}
		}
	} else {
		fmt.Println("    ❌ 用户缓存不可用")
	}

	// 测试其他缓存
	sessionCache := core.GetSessionCache()
	fmt.Printf("  GetSessionCache(): %v\n", sessionCache != nil)

	productCache := core.GetProductCache()
	fmt.Printf("  GetProductCache(): %v\n", productCache != nil)

	configCache := core.GetConfigCache()
	fmt.Printf("  GetConfigCache(): %v\n", configCache != nil)

	// 测试通用API
	hasUsers := core.HasCache("users")
	fmt.Printf("  HasCache('users'): %v\n", hasUsers)

	// 测试多缓存操作
	if sessionCache != nil && productCache != nil {
		fmt.Println("\n🔄 测试多缓存操作:")
		ctx := context.Background()

		// 并行设置多个缓存
		_ = sessionCache.Set(ctx, "session1", "session_data", time.Minute)
		_ = productCache.Set(ctx, "product1", "product_data", time.Minute)

		sessionVal, err1 := sessionCache.Get(ctx, "session1")
		productVal, err2 := productCache.Get(ctx, "product1")

		if err1 == nil && err2 == nil {
			fmt.Printf("    ✅ 多缓存操作成功: session=%v, product=%v\n", sessionVal, productVal)
		} else {
			fmt.Printf("    ❌ 多缓存操作失败: err1=%v, err2=%v\n", err1, err2)
		}
	}
}
