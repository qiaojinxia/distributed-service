package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
	"github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
	fmt.Println("🚀 缓存模块演示程序")
	fmt.Println("==========================================")
	
	// 演示1: 直接使用内存缓存
	fmt.Println("\n1️⃣ 直接使用内存缓存")
	directCacheDemo()
	
	// 演示2: 框架集成缓存
	fmt.Println("\n2️⃣ 框架集成缓存")
	frameworkCacheDemo()
	
	// 演示3: 淘汰策略对比
	fmt.Println("\n3️⃣ 淘汰策略对比")
	evictionPolicyDemo()
	
	fmt.Println("\n✅ 演示完成")
}

// 直接使用内存缓存演示
func directCacheDemo() {
	ctx := context.Background()
	
	// 创建LRU缓存
	config := cache.MemoryConfig{
		MaxSize:        100,
		DefaultTTL:     time.Hour,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}
	
	lruCache, err := cache.NewMemoryCache(config)
	if err != nil {
		fmt.Printf("❌ 创建缓存失败: %v\n", err)
		return
	}
	
	// 存储数据
	userInfo := map[string]interface{}{
		"id":   123,
		"name": "张三",
		"role": "管理员",
	}
	
	err = lruCache.Set(ctx, "user:123", userInfo, time.Minute*30)
	if err != nil {
		fmt.Printf("❌ 存储失败: %v\n", err)
		return
	}
	fmt.Println("✅ 用户信息存储成功")
	
	// 获取数据
	data, err := lruCache.Get(ctx, "user:123")
	if err != nil {
		fmt.Printf("❌ 获取失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 获取用户信息: %+v\n", data)
	
	// 检查存在性
	exists, _ := lruCache.Exists(ctx, "user:123")
	fmt.Printf("✅ 用户缓存存在: %v\n", exists)
}

// 框架集成缓存演示
func frameworkCacheDemo() {
	// 启动框架
	go func() {
		err := core.New().
			Port(8081).
			Name("cache-demo").
			OnlyHTTP().
			Run()
		if err != nil {
			fmt.Printf("框架启动失败: %v\n", err)
		}
	}()
	
	// 等待框架初始化
	fmt.Println("⏳ 等待框架初始化...")
	time.Sleep(time.Second * 3)
	
	ctx := context.Background()
	
	// 测试用户缓存
	fmt.Println("--- 用户缓存测试 (LRU策略) ---")
	userCache := core.GetUserCache()
	if userCache != nil {
		user := User{
			ID:    456,
			Name:  "李四",
			Email: "lisi@example.com",
		}
		err := userCache.Set(ctx, "user:456", user, time.Hour)
		if err != nil {
			fmt.Printf("❌ 用户缓存设置失败: %v\n", err)
		} else {
			fmt.Println("✅ 用户缓存设置成功")
			
			// 获取数据验证
			if data, err := userCache.Get(ctx, "user:456"); err == nil {
				fmt.Printf("✅ 用户缓存获取成功: %+v\n", data)
			}
		}
	} else {
		fmt.Println("❌ 用户缓存不可用")
	}
	
	// 测试会话缓存
	fmt.Println("--- 会话缓存测试 (TTL策略) ---")
	sessionCache := core.GetSessionCache()
	if sessionCache != nil {
		session := Session{
			ID:         "sess_demo_123",
			UserID:     456,
			CreatedAt:  time.Now(),
			LastAccess: time.Now(),
		}
		err := sessionCache.Set(ctx, "session:demo", session, time.Minute*5)
		if err != nil {
			fmt.Printf("❌ 会话缓存设置失败: %v\n", err)
		} else {
			fmt.Println("✅ 会话缓存设置成功")
			
			// 验证存在性
			if exists, _ := sessionCache.Exists(ctx, "session:demo"); exists {
				fmt.Println("✅ 会话缓存存在验证通过")
			}
		}
	} else {
		fmt.Println("❌ 会话缓存不可用")
	}
	
	// 测试产品缓存
	fmt.Println("--- 产品缓存测试 (Simple策略) ---")
	productCache := core.GetProductCache()
	if productCache != nil {
		products := []Product{
			{ID: 1, Name: "iPhone 15", Price: 5999.00},
			{ID: 2, Name: "MacBook Pro", Price: 12999.00},
			{ID: 3, Name: "iPad Air", Price: 3999.00},
		}
		err := productCache.Set(ctx, "hot_products", products, time.Hour*2)
		if err != nil {
			fmt.Printf("❌ 产品缓存设置失败: %v\n", err)
		} else {
			fmt.Println("✅ 产品缓存设置成功")
			
			// 获取并显示产品列表
			if data, err := productCache.Get(ctx, "hot_products"); err == nil {
				if productList, ok := data.([]Product); ok {
					fmt.Printf("✅ 热门产品列表: %+v\n", productList)
				}
			}
		}
	} else {
		fmt.Println("❌ 产品缓存不可用")
	}
}

// 淘汰策略对比演示
func evictionPolicyDemo() {
	ctx := context.Background()
	
	// LRU策略演示
	fmt.Println("--- LRU策略演示 ---")
	lruConfig := cache.MemoryConfig{
		MaxSize:        3, // 限制为3个条目
		DefaultTTL:     time.Hour,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}
	
	lruCache, _ := cache.NewMemoryCache(lruConfig)
	
	// 填满缓存
	lruCache.Set(ctx, "key1", "value1", 0)
	lruCache.Set(ctx, "key2", "value2", 0)
	lruCache.Set(ctx, "key3", "value3", 0)
	fmt.Println("✅ LRU缓存已填满 (key1, key2, key3)")
	
	// 访问key1，使其变为最近使用
	lruCache.Get(ctx, "key1")
	fmt.Println("📖 访问了key1，使其变为最近使用")
	
	// 添加新键，应该淘汰最久未使用的key2
	lruCache.Set(ctx, "key4", "value4", 0)
	fmt.Println("➕ 添加key4，应该淘汰最久未使用的key2")
	
	// 检查结果
	exists1, _ := lruCache.Exists(ctx, "key1")
	exists2, _ := lruCache.Exists(ctx, "key2")
	exists3, _ := lruCache.Exists(ctx, "key3")
	exists4, _ := lruCache.Exists(ctx, "key4")
	
	fmt.Printf("🔍 LRU结果: key1=%v, key2=%v, key3=%v, key4=%v\n", exists1, exists2, exists3, exists4)
	if exists1 && !exists2 && exists3 && exists4 {
		fmt.Println("✅ LRU策略工作正常：key2被正确淘汰")
	} else {
		fmt.Println("❌ LRU策略可能存在问题")
	}
	
	// TTL策略演示
	fmt.Println("\n--- TTL策略演示 ---")
	ttlConfig := cache.MemoryConfig{
		MaxSize:         10,
		DefaultTTL:      time.Second * 3,
		CleanupInterval: time.Millisecond * 100,
		EvictionPolicy:  cache.EvictionPolicyTTL,
	}
	
	ttlCache, _ := cache.NewMemoryCache(ttlConfig)
	
	// 设置不同TTL的数据
	ttlCache.Set(ctx, "short_lived", "短期数据", time.Millisecond*800)
	ttlCache.Set(ctx, "long_lived", "长期数据", time.Second*5)
	fmt.Println("✅ 设置了短期数据(800ms)和长期数据(5s)")
	
	// 立即检查
	shortExists, _ := ttlCache.Exists(ctx, "short_lived")
	longExists, _ := ttlCache.Exists(ctx, "long_lived")
	fmt.Printf("🔍 立即检查: 短期=%v, 长期=%v\n", shortExists, longExists)
	
	// 等待短期数据过期
	fmt.Println("⏳ 等待1.2秒...")
	time.Sleep(time.Millisecond * 1200)
	
	shortExists, _ = ttlCache.Exists(ctx, "short_lived")
	longExists, _ = ttlCache.Exists(ctx, "long_lived")
	fmt.Printf("🔍 1.2秒后检查: 短期=%v, 长期=%v\n", shortExists, longExists)
	
	if !shortExists && longExists {
		fmt.Println("✅ TTL策略工作正常：短期数据已过期，长期数据仍存在")
	} else {
		fmt.Println("❌ TTL策略可能存在问题")
	}
}

// 数据模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Session struct {
	ID         string    `json:"id"`
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	LastAccess time.Time `json:"last_access"`
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}