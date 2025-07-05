package main

import (
	"context"
	"fmt"
	"github.com/qiaojinxia/distributed-service/framework/core"
	"log"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
)

func main() {
	fmt.Println("🔄 混合缓存（L1本地 + L2 Redis）演示")
	fmt.Println("=====================================")

	// 演示不同的缓存策略
	demoWriteThroughCache()
	demoWriteBackCache()
	demoCustomHybridCache()
	demoConfigPresets()
}

func demoWriteThroughCache() {
	fmt.Println("\n📝 写穿透策略演示 (Write-Through)")
	fmt.Println("==================================")

	manager := core.NewCacheManager()

	// 创建写穿透混合缓存
	err := manager.CreateCache(cache.Config{
		Type: cache.TypeHybrid,
		Name: "write_through_cache",
		Settings: map[string]interface{}{
			"sync_strategy": "write_through",
			"l1_config": map[string]interface{}{
				"type": "memory",
				"settings": map[string]interface{}{
					"max_size":    1000,
					"default_ttl": "30m",
				},
			},
			"l2_config": map[string]interface{}{
				"type": "redis",
				"settings": map[string]interface{}{
					"addr": "localhost:6379",
					"db":   0,
				},
			},
			"l1_ttl": "30m",
			"l2_ttl": "2h",
		},
	})

	if err != nil {
		fmt.Printf("⚠️  创建写穿透缓存失败（可能Redis未启动）: %v\n", err)
		return
	}

	hybridCache, err := manager.GetCache("write_through_cache")
	if err != nil {
		log.Fatalf("获取混合缓存失败: %v", err)
	}

	ctx := context.Background()

	// 写入数据（同时写入L1和L2）
	err = hybridCache.Set(ctx, "user:1001", map[string]interface{}{
		"id":    1001,
		"name":  "张三",
		"email": "zhangsan@example.com",
		"dept":  "技术部",
	}, time.Hour)

	if err != nil {
		log.Fatalf("设置缓存失败: %v", err)
	}

	fmt.Println("✅ 数据已写入L1和L2缓存")

	// 读取数据（优先从L1读取）
	value, err := hybridCache.Get(ctx, "user:1001")
	if err != nil {
		log.Fatalf("获取缓存失败: %v", err)
	}

	fmt.Printf("📖 读取到的数据: %+v\n", value)

	// 显示统计信息
	if hybridCacheImpl, ok := hybridCache.(*cache.HybridCache); ok {
		stats := hybridCacheImpl.GetStats()
		fmt.Printf("📊 缓存统计: L1命中=%d, L1未命中=%d, L2命中=%d, L2未命中=%d\n",
			stats.L1Hits, stats.L1Misses, stats.L2Hits, stats.L2Misses)
	}
}

func demoWriteBackCache() {
	fmt.Println("\n🔄 写回策略演示 (Write-Back)")
	fmt.Println("=============================")

	manager := core.NewCacheManager()

	// 创建写回混合缓存
	err := manager.CreateCache(cache.Config{
		Type: cache.TypeHybrid,
		Name: "write_back_cache",
		Settings: map[string]interface{}{
			"sync_strategy":         "write_back",
			"write_back_enabled":    true,
			"write_back_interval":   "5s",
			"write_back_batch_size": 10,
			"l1_config": map[string]interface{}{
				"type": "memory",
				"settings": map[string]interface{}{
					"max_size":    1000,
					"default_ttl": "1h",
				},
			},
			"l2_config": map[string]interface{}{
				"type": "redis",
				"settings": map[string]interface{}{
					"addr": "localhost:6379",
					"db":   1,
				},
			},
			"l1_ttl": "1h",
			"l2_ttl": "24h",
		},
	})

	if err != nil {
		fmt.Printf("⚠️  创建写回缓存失败（可能Redis未启动）: %v\n", err)
		return
	}

	hybridCache, err := manager.GetCache("write_back_cache")
	if err != nil {
		log.Fatalf("获取混合缓存失败: %v", err)
	}

	ctx := context.Background()

	// 批量写入数据（先写L1，定时写回L2）
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("product:%d", i)
		value := map[string]interface{}{
			"id":    i,
			"name":  fmt.Sprintf("商品%d", i),
			"price": 99.99 + float64(i),
			"stock": 100 + i*10,
		}

		err = hybridCache.Set(ctx, key, value, 0)
		if err != nil {
			log.Printf("设置缓存失败: %v", err)
			continue
		}

		fmt.Printf("✅ 商品%d已写入L1缓存\n", i)
	}

	fmt.Println("⏳ 等待写回Redis...")
	time.Sleep(6 * time.Second)

	// 显示统计信息
	if hybridCacheImpl, ok := hybridCache.(*cache.HybridCache); ok {
		stats := hybridCacheImpl.GetStats()
		fmt.Printf("📊 写回统计: L1设置=%d, L2设置=%d, 写回次数=%d\n",
			stats.L1Sets, stats.L2Sets, stats.Writebacks)
	}

	fmt.Println("🎯 写回策略：优先写L1，定时批量写回L2，提升写性能")
}

func demoCustomHybridCache() {
	fmt.Println("\n🛠️  自定义混合缓存配置演示")
	fmt.Println("=============================")

	// 使用配置构建器创建自定义配置
	customConfig := cache.NewCustomHybridConfig().
		WithL1Memory(5000, time.Minute*45).
		WithL2Redis("localhost:6379", "", 2, time.Hour*6).
		WithSyncStrategy(cache.SyncStrategyWriteBack).
		WithWriteBack(true, time.Minute*3, 50).
		Build()

	fmt.Printf("🔧 自定义配置:\n")
	fmt.Printf("   L1缓存: 内存，最大5000条，TTL=45分钟\n")
	fmt.Printf("   L2缓存: Redis，DB=2，TTL=6小时\n")
	fmt.Printf("   同步策略: 写回模式\n")
	fmt.Printf("   写回间隔: 3分钟，批量大小=50\n")

	// 创建混合缓存
	hybridCache, err := cache.NewHybridCache(customConfig)
	if err != nil {
		fmt.Printf("⚠️  创建自定义混合缓存失败: %v\n", err)
		return
	}
	defer hybridCache.Close()

	ctx := context.Background()

	// 测试缓存操作
	testKey := "custom:test:1"
	testValue := map[string]interface{}{
		"message": "这是自定义混合缓存测试",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	}

	err = hybridCache.Set(ctx, testKey, testValue, 0)
	if err != nil {
		fmt.Printf("⚠️  设置缓存失败: %v\n", err)
		return
	}

	value, err := hybridCache.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("⚠️  获取缓存失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 自定义缓存测试成功: %+v\n", value)
}

func demoConfigPresets() {
	fmt.Println("\n📋 配置预设演示")
	fmt.Println("================")

	// 展示不同的预设配置
	presets := []struct {
		name   string
		config cache.HybridConfig
		desc   string
	}{
		{
			name:   "默认配置",
			config: cache.Presets.GetDefaultHybridConfig(),
			desc:   "平衡性能和内存使用",
		},
		{
			name:   "高性能配置",
			config: cache.Presets.GetHighPerformanceHybridConfig(),
			desc:   "高性能，大内存，写回模式",
		},
		{
			name:   "低内存配置",
			config: cache.Presets.GetLowMemoryHybridConfig(),
			desc:   "节省内存，写绕过模式",
		},
	}

	for _, preset := range presets {
		fmt.Printf("\n🎨 %s (%s):\n", preset.name, preset.desc)
		fmt.Printf("   同步策略: %s\n", preset.config.SyncStrategy)
		fmt.Printf("   L1 TTL: %v\n", preset.config.L1TTL)
		fmt.Printf("   L2 TTL: %v\n", preset.config.L2TTL)
		fmt.Printf("   写回启用: %t\n", preset.config.WriteBackEnabled)
		if preset.config.WriteBackEnabled {
			fmt.Printf("   写回间隔: %v\n", preset.config.WriteBackInterval)
		}
	}

	fmt.Println("\n💡 使用提示:")
	fmt.Println("   • 写穿透(Write-Through): 同时写L1和L2，数据一致性好")
	fmt.Println("   • 写回(Write-Back): 先写L1再写L2，写性能好")
	fmt.Println("   • 写绕过(Write-Around): 只写L2，适合写多读少场景")
	fmt.Println("   • L1缓存提供快速访问，L2缓存提供持久化")
}

// 真实环境使用示例
func realWorldExample() {
	fmt.Println("\n🌍 真实环境使用示例")
	fmt.Println("=====================")

	// 创建缓存管理器
	manager := core.NewCacheManager()

	// 为不同业务场景创建不同的混合缓存

	// 1. 用户会话缓存 - 高性能配置
	manager.CreateCache(cache.Config{
		Type:     cache.TypeHybrid,
		Name:     "user_session",
		Settings: cache.Presets.GetHighPerformanceHybridConfig(),
	})

	// 2. 商品信息缓存 - 默认配置
	manager.CreateCache(cache.Config{
		Type:     cache.TypeHybrid,
		Name:     "product_info",
		Settings: cache.Presets.GetDefaultHybridConfig(),
	})

	// 3. 统计数据缓存 - 低内存配置
	manager.CreateCache(cache.Config{
		Type:     cache.TypeHybrid,
		Name:     "statistics",
		Settings: cache.Presets.GetLowMemoryHybridConfig(),
	})

	fmt.Println("✅ 已创建多个业务缓存实例")

	// 使用缓存
	userSessionCache, _ := manager.GetCache("user_session")
	productInfoCache, _ := manager.GetCache("product_info")
	statisticsCache, _ := manager.GetCache("statistics")

	ctx := context.Background()

	// 用户会话
	userSessionCache.Set(ctx, "session:abc123", map[string]interface{}{
		"user_id":    1001,
		"login_time": time.Now(),
	}, time.Hour*2)

	// 商品信息
	productInfoCache.Set(ctx, "product:123", map[string]interface{}{
		"name":  "智能手表",
		"price": 1299.00,
	}, time.Hour*12)

	// 统计数据
	statisticsCache.Set(ctx, "daily_stats:2023-12-01", map[string]interface{}{
		"pv": 10000,
		"uv": 5000,
	}, time.Hour*48)

	fmt.Println("✅ 各业务缓存数据设置完成")
}
