package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
	log.Println("🔍 调试缓存初始化流程...")

	// 创建纯内存缓存配置
	cacheConfig := &config.CacheConfig{
		Enabled:         true,
		DefaultType:     "memory",
		UseFramework:    false,
		GlobalKeyPrefix: "debug",
		DefaultTTL:      "1h",
		Caches: map[string]config.CacheInstance{
			"users": {
				Type:      "memory",
				KeyPrefix: "users",
				TTL:       "2h",
				Settings: map[string]interface{}{
					"max_size":         1000,
					"eviction_policy":  "lru",
					"default_ttl":      "2h",
					"cleanup_interval": "10m",
				},
			},
		},
	}

	// 在goroutine中启动框架
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("框架启动panic: %v", r)
			}
		}()

		err := core.New().
			Port(8085).
			Mode("debug").
			Name("debug-cache").
			WithCache(cacheConfig).
			OnlyHTTP().
			Run()
		if err != nil {
			log.Printf("框架启动失败: %v", err)
		}
	}()

	// 等待初始化
	fmt.Println("等待框架初始化...")
	for i := 1; i <= 10; i++ {
		time.Sleep(time.Second)
		fmt.Printf("第%d秒 - 检查缓存可用性:\n", i)

		// 检查各种缓存API
		userCache := core.GetUserCache()
		fmt.Printf("  GetUserCache(): %v\n", userCache != nil)

		hasUsers := core.HasCache("users")
		fmt.Printf("  HasCache('users'): %v\n", hasUsers)

		genericCache := core.GetCache("users")
		fmt.Printf("  GetCache('users'): %v\n", genericCache != nil)

		if userCache != nil {
			fmt.Println("  ✅ 缓存系统已初始化！")

			// 测试基本操作
			ctx := context.Background()
			err := userCache.Set(ctx, "test_key", "test_value", time.Minute)
			if err != nil {
				fmt.Printf("  设置测试失败: %v\n", err)
			} else {
				value, err := userCache.Get(ctx, "test_key")
				if err != nil {
					fmt.Printf("  获取测试失败: %v\n", err)
				} else {
					fmt.Printf("  ✅ 缓存测试成功: %v\n", value)
				}
			}
			break
		}

		if i == 10 {
			fmt.Println("  ❌ 10秒后缓存仍未初始化")
		}
	}
}
