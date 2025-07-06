package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/core"
)

func TestCacheIntegration(t *testing.T) {
	t.Log("🔍 缓存框架集成测试")

	// 启动框架
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("框架启动panic: %v", r)
			}
		}()

		err := core.New().
			Port(8088).
			Mode("release").
			Name("cache-integration-test").
			OnlyHTTP().
			Run()
		if err != nil {
			t.Logf("框架启动失败: %v", err)
		}
	}()

	// 等待框架初始化
	t.Log("⏳ 等待框架初始化...")
	time.Sleep(time.Second * 3)

	t.Run("CacheAPIAvailability", func(t *testing.T) {
		// 测试各种缓存API
		userCache := core.GetUserCache()
		if userCache == nil {
			t.Error("GetUserCache() 应该返回可用实例")
		}

		sessionCache := core.GetSessionCache()
		if sessionCache == nil {
			t.Error("GetSessionCache() 应该返回可用实例")
		}

		productCache := core.GetProductCache()
		if productCache == nil {
			t.Error("GetProductCache() 应该返回可用实例")
		}

		configCache := core.GetConfigCache()
		if configCache == nil {
			t.Error("GetConfigCache() 应该返回可用实例")
		}

		hasUsers := core.HasCache("users")
		if !hasUsers {
			t.Error("HasCache('users') 应该返回true")
		}

		hasNonExistent := core.HasCache("nonexistent")
		if hasNonExistent {
			t.Error("HasCache('nonexistent') 应该返回false")
		}
	})

	t.Run("BasicCacheOperations", func(t *testing.T) {
		userCache := core.GetUserCache()
		if userCache == nil {
			t.Fatal("用户缓存不可用，无法进行基本操作测试")
		}

		ctx := context.Background()

		// 测试Set操作
		err := userCache.Set(ctx, "test_key", "test_value", time.Minute)
		if err != nil {
			t.Errorf("缓存设置失败: %v", err)
		}

		// 测试Get操作
		value, err := userCache.Get(ctx, "test_key")
		if err != nil {
			t.Errorf("缓存获取失败: %v", err)
		}
		if value != "test_value" {
			t.Errorf("获取的值不正确: 期望 'test_value', 得到 '%v'", value)
		}

		// 测试Exists操作
		exists, err := userCache.Exists(ctx, "test_key")
		if err != nil {
			t.Errorf("存在性检查失败: %v", err)
		}
		if !exists {
			t.Error("存在性检查应该返回true")
		}

		// 测试Delete操作
		err = userCache.Delete(ctx, "test_key")
		if err != nil {
			t.Errorf("缓存删除失败: %v", err)
		}

		// 验证删除后不存在
		exists, err = userCache.Exists(ctx, "test_key")
		if err != nil {
			t.Errorf("删除验证失败: %v", err)
		}
		if exists {
			t.Error("删除后键不应该存在")
		}
	})

	t.Run("CacheIsolation", func(t *testing.T) {
		userCache := core.GetUserCache()
		sessionCache := core.GetSessionCache()

		if userCache == nil || sessionCache == nil {
			t.Skip("缓存实例不可用，跳过隔离测试")
		}

		ctx := context.Background()

		// 测试缓存隔离
		userCache.Set(ctx, "same_key", "user_value", time.Minute)
		sessionCache.Set(ctx, "same_key", "session_value", time.Minute)

		userVal, _ := userCache.Get(ctx, "same_key")
		sessionVal, _ := sessionCache.Get(ctx, "same_key")

		if userVal != "user_value" {
			t.Errorf("用户缓存值不正确: 期望 'user_value', 得到 '%v'", userVal)
		}
		if sessionVal != "session_value" {
			t.Errorf("会话缓存值不正确: 期望 'session_value', 得到 '%v'", sessionVal)
		}
	})
}
