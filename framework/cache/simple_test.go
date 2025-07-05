package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestSimpleRedisCache(t *testing.T) {
	// 创建一个模拟的Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// 测试连接
	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// 创建简单Redis缓存
	cache := NewSimpleRedisCache(client, "test")

	// 测试基本操作
	err = cache.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "value1" {
		t.Fatalf("Expected value1, got %v", value)
	}

	// 测试统计
	stats := cache.GetStats()
	if stats.Sets != 1 || stats.Hits != 1 {
		t.Fatalf("Stats incorrect: %+v", stats)
	}

	t.Logf("Simple Redis cache test passed")
}

func TestMemoryCache(t *testing.T) {
	config := MemoryConfig{
		MaxSize:         100,
		DefaultTTL:      time.Minute,
		CleanupInterval: time.Second * 10,
	}

	cache := NewMemoryCache(config)
	defer func(cache *MemoryCache) {
		_ = cache.Close()

	}(cache)

	ctx := context.Background()

	// 测试基本操作
	err := cache.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "value1" {
		t.Fatalf("Expected value1, got %v", value)
	}

	// 测试统计
	stats := cache.GetStats()
	if stats.Sets != 1 || stats.Hits != 1 {
		t.Fatalf("Stats incorrect: %+v", stats)
	}

	t.Logf("Memory cache test passed")
}

func TestCacheManager(t *testing.T) {
	manager := NewManager()

	// 注册构建器
	manager.RegisterBuilder(TypeMemory, &MemoryBuilder{})

	// 创建缓存
	config := Config{
		Type: TypeMemory,
		Name: "test-cache",
		Settings: map[string]interface{}{
			"max_size":    100,
			"default_ttl": "1m",
		},
	}

	err := manager.CreateCache(config)
	if err != nil {
		t.Fatalf("CreateCache failed: %v", err)
	}

	// 获取缓存
	cache, err := manager.GetCache("test-cache")
	if err != nil {
		t.Fatalf("GetCache failed: %v", err)
	}

	// 测试缓存
	ctx := context.Background()
	err = cache.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "value1" {
		t.Fatalf("Expected value1, got %v", value)
	}

	// 清理
	err = manager.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	t.Logf("Cache manager test passed")
}
