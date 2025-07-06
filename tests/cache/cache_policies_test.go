package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
)

func TestCachePolicies(t *testing.T) {
	t.Log("🎯 缓存策略功能测试")

	ctx := context.Background()

	t.Run("LRUPolicy", func(t *testing.T) {
		t.Log("📋 LRU策略测试")
		testLRU(ctx, t)
	})

	t.Run("TTLPolicy", func(t *testing.T) {
		t.Log("⏰ TTL策略测试")
		testTTL(ctx, t)
	})

	t.Run("SimplePolicy", func(t *testing.T) {
		t.Log("🔧 Simple策略测试")
		testSimple(ctx, t)
	})
}

func testLRU(ctx context.Context, t *testing.T) {
	config := cache.MemoryConfig{
		MaxSize:        3,
		DefaultTTL:     time.Minute,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}

	lruCache, err := cache.NewMemoryCache(config)
	if err != nil {
		t.Fatalf("LRU缓存创建失败: %v", err)
	}
	t.Log("✅ LRU缓存创建成功")

	// 填满缓存
	lruCache.Set(ctx, "key1", "value1", 0)
	lruCache.Set(ctx, "key2", "value2", 0)
	lruCache.Set(ctx, "key3", "value3", 0)

	// 访问key1使其变为最近使用
	lruCache.Get(ctx, "key1")

	// 添加第4个键，应该淘汰key2
	lruCache.Set(ctx, "key4", "value4", 0)

	exists1, _ := lruCache.Exists(ctx, "key1")
	exists2, _ := lruCache.Exists(ctx, "key2")
	exists4, _ := lruCache.Exists(ctx, "key4")

	if !exists1 {
		t.Error("key1应该存在（最近访问过）")
	}
	if exists2 {
		t.Error("key2应该被淘汰（最久未使用）")
	}
	if !exists4 {
		t.Error("key4应该存在（新添加的）")
	}
	t.Log("✅ LRU淘汰机制工作正常")
}

func testTTL(ctx context.Context, t *testing.T) {
	config := cache.MemoryConfig{
		MaxSize:         10,
		DefaultTTL:      time.Second * 2,
		CleanupInterval: time.Millisecond * 100,
		EvictionPolicy:  cache.EvictionPolicyTTL,
	}

	ttlCache, err := cache.NewMemoryCache(config)
	if err != nil {
		t.Fatalf("TTL缓存创建失败: %v", err)
	}
	t.Log("✅ TTL缓存创建成功")

	// 测试默认TTL
	ttlCache.Set(ctx, "default_key", "default_value", 0)
	value, err := ttlCache.Get(ctx, "default_key")
	if err != nil {
		t.Errorf("默认TTL获取失败: %v", err)
	}
	if value != "default_value" {
		t.Errorf("默认TTL值不正确: 期望 'default_value', 得到 '%v'", value)
	}
	t.Log("✅ 默认TTL正常工作")

	// 测试自定义短TTL
	ttlCache.Set(ctx, "short_key", "short_value", time.Millisecond*300)
	value, err = ttlCache.Get(ctx, "short_key")
	if err != nil {
		t.Errorf("自定义TTL获取失败: %v", err)
	}
	if value != "short_value" {
		t.Errorf("自定义TTL值不正确: 期望 'short_value', 得到 '%v'", value)
	}
	t.Log("✅ 自定义TTL设置成功")

	// 等待自定义TTL过期
	t.Log("⏳ 等待500ms直到数据过期...")
	time.Sleep(time.Millisecond * 500)
	_, err = ttlCache.Get(ctx, "short_key")
	if err == nil {
		t.Error("自定义TTL应该过期")
	}
	t.Log("✅ 自定义TTL过期正常")

	// 验证默认TTL仍存在
	value, err = ttlCache.Get(ctx, "default_key")
	if err != nil {
		t.Errorf("默认TTL数据获取失败: %v", err)
	}
	if value != "default_value" {
		t.Errorf("默认TTL数据应该仍存在: 期望 'default_value', 得到 '%v'", value)
	}
	t.Log("✅ 默认TTL数据仍然存在")
}

func testSimple(ctx context.Context, t *testing.T) {
	config := cache.MemoryConfig{
		MaxSize:         5,
		DefaultTTL:      time.Second * 3,
		CleanupInterval: time.Second,
		EvictionPolicy:  cache.EvictionPolicySimple,
	}

	simpleCache, err := cache.NewMemoryCache(config)
	if err != nil {
		t.Fatalf("Simple缓存创建失败: %v", err)
	}
	t.Log("✅ Simple缓存创建成功")

	// 基本操作测试
	simpleCache.Set(ctx, "config1", "value1", 0)
	simpleCache.Set(ctx, "config2", "value2", 0)

	value1, err1 := simpleCache.Get(ctx, "config1")
	value2, err2 := simpleCache.Get(ctx, "config2")

	if err1 != nil {
		t.Errorf("config1获取失败: %v", err1)
	}
	if err2 != nil {
		t.Errorf("config2获取失败: %v", err2)
	}
	if value1 != "value1" {
		t.Errorf("config1值不正确: 期望 'value1', 得到 '%v'", value1)
	}
	if value2 != "value2" {
		t.Errorf("config2值不正确: 期望 'value2', 得到 '%v'", value2)
	}
	t.Log("✅ Simple缓存基本操作成功")

	// 自定义TTL测试
	simpleCache.Set(ctx, "temp", "temp_value", time.Millisecond*400)
	value, err := simpleCache.Get(ctx, "temp")
	if err != nil {
		t.Errorf("Simple自定义TTL获取失败: %v", err)
	}
	if value != "temp_value" {
		t.Errorf("Simple自定义TTL值不正确: 期望 'temp_value', 得到 '%v'", value)
	}
	t.Log("✅ Simple自定义TTL设置成功")

	// 等待过期
	t.Log("⏳ 等待600ms直到数据过期...")
	time.Sleep(time.Millisecond * 600)
	_, err = simpleCache.Get(ctx, "temp")
	if err == nil {
		t.Error("Simple TTL过期应该生效")
	}
	t.Log("✅ Simple TTL过期正常")
}
