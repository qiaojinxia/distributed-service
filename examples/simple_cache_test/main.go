package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
)

func main() {
	log.Println("🔍 直接测试缓存创建...")

	// 创建框架缓存服务
	cacheService := cache.NewFrameworkCacheService()

	// 初始化缓存服务
	ctx := context.Background()
	if err := cacheService.Initialize(ctx); err != nil {
		log.Fatalf("缓存服务初始化失败: %v", err)
	}

	fmt.Println("✅ 缓存服务初始化成功")

	// 创建内存缓存配置
	memoryConfig := cache.Config{
		Type: cache.TypeMemory,
		Name: "users",
		Settings: map[string]interface{}{
			"max_size":         1000,
			"eviction_policy":  "lru",
			"default_ttl":      "2h",
			"cleanup_interval": "10m",
		},
	}

	// 创建缓存实例
	if err := cacheService.Manager.CreateCache(memoryConfig); err != nil {
		log.Fatalf("创建缓存实例失败: %v", err)
	}

	fmt.Println("✅ 缓存实例创建成功")

	// 测试获取缓存
	userCache, err := cacheService.GetNamedCache("users")
	if err != nil {
		log.Fatalf("获取缓存失败: %v", err)
	}

	if userCache == nil {
		log.Fatalf("获取的缓存为nil")
	}

	fmt.Println("✅ 缓存获取成功")

	// 测试缓存操作
	if err := userCache.Set(ctx, "test_key", "test_value", time.Minute); err != nil {
		log.Fatalf("设置缓存失败: %v", err)
	}

	value, err := userCache.Get(ctx, "test_key")
	if err != nil {
		log.Fatalf("获取缓存失败: %v", err)
	}

	fmt.Printf("✅ 缓存操作成功: %v\n", value)

	// 测试框架集成
	fmt.Println("\n🔗 测试框架集成...")

	// 模拟框架初始化流程
	testFrameworkIntegration(cacheService)
}

func testFrameworkIntegration(cacheService *cache.FrameworkCacheService) {
	// 模拟component.manager中的registerCacheToGlobalSystem

	// 创建一个简单的回调函数来模拟core包的initGlobalCacheSystem
	callback := func(fcs *cache.FrameworkCacheService) error {
		fmt.Println("  🔄 模拟全局缓存系统初始化...")

		// 模拟core包中的frameworkCacheService赋值
		// frameworkCacheService = fcs

		// 测试GetNamedCache
		testCache, err := fcs.GetNamedCache("users")
		if err != nil {
			return fmt.Errorf("获取命名缓存失败: %w", err)
		}

		if testCache == nil {
			return fmt.Errorf("获取的命名缓存为nil")
		}

		fmt.Println("  ✅ 全局缓存系统初始化成功")
		return nil
	}

	// 调用回调函数
	if err := callback(cacheService); err != nil {
		log.Fatalf("框架集成测试失败: %v", err)
	}

	fmt.Println("✅ 框架集成测试成功")
}
