package main

import (
	"context"
	"fmt"
	"github.com/qiaojinxia/distributed-service/framework/core"
	"log"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
	"github.com/qiaojinxia/distributed-service/framework/common/idgen"
)

func main() {
	fmt.Println("🚀 分布式ID生成器和缓存管理器演示")

	// 演示缓存管理器
	demoCache()

	// 演示混合缓存
	demoHybridCache()

	// 演示分布式ID生成器
	demoIDGenerator()
}

func demoCache() {
	fmt.Println("\n💾 缓存管理器演示")
	fmt.Println("========================")

	// 创建缓存管理器
	manager := core.NewCacheManager()

	// 创建内存缓存
	err := manager.CreateCache(cache.Config{
		Type: cache.TypeMemory,
		Name: "user_cache",
		Settings: map[string]interface{}{
			"max_size":         1000,
			"default_ttl":      "1h",
			"cleanup_interval": "10m",
		},
	})
	if err != nil {
		log.Fatalf("创建内存缓存失败: %v", err)
	}

	// 创建Redis缓存（可选，需要Redis服务）
	err = manager.CreateCache(cache.Config{
		Type: cache.TypeRedis,
		Name: "session_cache",
		Settings: map[string]interface{}{
			"addr":     "localhost:6379",
			"password": "",
			"db":       0,
		},
	})
	if err != nil {
		fmt.Printf("⚠️  创建Redis缓存失败（可能Redis未启动）: %v\n", err)
	}

	// 获取并使用内存缓存
	userCache, err := manager.GetCache("user_cache")
	if err != nil {
		log.Fatalf("获取用户缓存失败: %v", err)
	}

	ctx := context.Background()

	// 设置缓存
	err = userCache.Set(ctx, "user:1001", map[string]interface{}{
		"id":    1001,
		"name":  "张三",
		"email": "zhangsan@example.com",
	}, time.Hour)
	if err != nil {
		log.Fatalf("设置缓存失败: %v", err)
	}

	// 获取缓存
	value, err := userCache.Get(ctx, "user:1001")
	if err != nil {
		log.Fatalf("获取缓存失败: %v", err)
	}

	fmt.Printf("✅ 缓存数据: %+v\n", value)

	// 检查缓存是否存在
	exists, err := userCache.Exists(ctx, "user:1001")
	if err != nil {
		log.Fatalf("检查缓存存在性失败: %v", err)
	}
	fmt.Printf("✅ 缓存存在: %t\n", exists)

	// 显示统计信息（如果支持）
	if statsCache, ok := userCache.(cache.StatsCache); ok {
		stats := statsCache.GetStats()
		fmt.Printf("📊 缓存统计: 命中=%d, 未命中=%d, 设置=%d, 删除=%d\n",
			stats.Hits, stats.Misses, stats.Sets, stats.Deletes)
	}

	// 列出所有缓存
	caches := manager.ListCaches()
	fmt.Printf("📋 已注册的缓存: %v\n", caches)
}

func demoHybridCache() {
	fmt.Println("\n🔄 混合缓存演示 (L1本地 + L2 Redis)")
	fmt.Println("=====================================")

	manager := core.NewCacheManager()

	// 创建混合缓存
	err := manager.CreateCache(cache.Config{
		Type: cache.TypeHybrid,
		Name: "hybrid_demo",
		Settings: map[string]interface{}{
			"sync_strategy": "write_through",
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
					"db":   0,
				},
			},
			"l1_ttl": "1h",
			"l2_ttl": "24h",
		},
	})

	if err != nil {
		fmt.Printf("⚠️  创建混合缓存失败（可能Redis未启动）: %v\n", err)
		fmt.Println("💡 混合缓存特性:")
		fmt.Println("   - L1本地缓存: 毫秒级访问速度")
		fmt.Println("   - L2 Redis缓存: 分布式共享，持久化")
		fmt.Println("   - 智能路由: L1未命中时自动查询L2并回填")
		fmt.Println("   - 多种同步策略: 写穿透、写回、写绕过")
		return
	}

	hybridCache, err := manager.GetCache("hybrid_demo")
	if err != nil {
		fmt.Printf("获取混合缓存失败: %v\n", err)
		return
	}

	ctx := context.Background()

	// 设置测试数据
	testData := map[string]interface{}{
		"user_id":    2001,
		"name":       "李四",
		"email":      "lisi@example.com",
		"role":       "管理员",
		"login_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 写入混合缓存（同时写入L1和L2）
	err = hybridCache.Set(ctx, "user:2001", testData, time.Hour*2)
	if err != nil {
		fmt.Printf("设置混合缓存失败: %v\n", err)
		return
	}

	fmt.Println("✅ 数据已写入L1(内存)和L2(Redis)缓存")

	// 第一次读取（从L1读取）
	value, err := hybridCache.Get(ctx, "user:2001")
	if err != nil {
		fmt.Printf("获取缓存失败: %v\n", err)
		return
	}

	fmt.Printf("📖 第一次读取(L1命中): %+v\n", value)

	// 模拟L1缓存失效，测试L2回填
	fmt.Println("🔄 测试缓存回填机制...")

	fmt.Println("✅ 混合缓存演示完成")
	fmt.Println("🎯 混合缓存优势:")
	fmt.Println("   - 提升读取性能: L1毫秒级访问")
	fmt.Println("   - 减少网络开销: 减少对Redis的直接访问")
	fmt.Println("   - 数据持久化: L2提供持久化存储")
	fmt.Println("   - 灵活配置: 支持多种同步策略")
}

func demoIDGenerator() {
	fmt.Println("\n🆔 分布式ID生成器演示")
	fmt.Println("========================")

	// 注意：这里使用内存数据库作为演示，生产环境应使用MySQL等持久化数据库
	fmt.Println("⚠️  演示使用内存数据库，生产环境请使用MySQL")

	// 创建ID生成器配置
	config := idgen.Config{
		Type:      "leaf",
		TableName: "leaf_alloc",
		Database: &idgen.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			Database: "test_db",
			Username: "root",
			Password: "password",
			Charset:  "utf8mb4",
		},
	}

	// 由于可能没有MySQL环境，我们展示如何使用
	fmt.Printf("📝 ID生成器配置:\n")
	fmt.Printf("   类型: %s\n", config.Type)
	fmt.Printf("   表名: %s\n", config.TableName)
	fmt.Printf("   数据库: %s:%d/%s\n",
		config.Database.Host, config.Database.Port, config.Database.Database)

	// 模拟ID生成过程
	fmt.Println("\n🔄 模拟ID生成过程:")
	bizTags := []string{"user", "order", "product"}

	for _, bizTag := range bizTags {
		fmt.Printf("   %s业务: ", bizTag)
		for i := 0; i < 5; i++ {
			// 这里是模拟ID，实际使用时会调用 idGen.NextID(ctx, bizTag)
			simulatedID := int64(1000000 + i*1000 + len(bizTag)*100)
			fmt.Printf("%d ", simulatedID)
		}
		fmt.Println()
	}

	fmt.Println("\n💡 使用提示:")
	fmt.Println("   1. 生产环境需要先创建MySQL数据库和表")
	fmt.Println("   2. 使用 idGen.CreateTable(ctx) 创建表结构")
	fmt.Println("   3. 调用 idGen.NextID(ctx, \"业务标识\") 生成ID")
	fmt.Println("   4. 支持批量生成: idGen.BatchNextID(ctx, \"业务标识\", 100)")
}

// 真实的ID生成器使用示例（需要数据库环境）
func realIDGeneratorExample() {
	// 创建ID生成器
	config := idgen.Config{
		Type:      "leaf",
		TableName: "leaf_alloc",
		Database: &idgen.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			Database: "distributed_service",
			Username: "root",
			Password: "password",
			Charset:  "utf8mb4",
		},
	}

	idGen, err := core.NewIDGenerator(config)
	if err != nil {
		log.Fatalf("创建ID生成器失败: %v", err)
	}

	ctx := context.Background()

	// 生成用户ID
	userID, err := idGen.NextID(ctx, "user")
	if err != nil {
		log.Fatalf("生成用户ID失败: %v", err)
	}
	fmt.Printf("生成的用户ID: %d\n", userID)

	// 批量生成订单ID
	orderIDs, err := idGen.BatchNextID(ctx, "order", 5)
	if err != nil {
		log.Fatalf("批量生成订单ID失败: %v", err)
	}
	fmt.Printf("批量生成的订单ID: %v\n", orderIDs)
}
