package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/core"
)

func TestTTLBehavior(t *testing.T) {
	t.Log("⏰ TTL行为测试 - 验证TTL修复")

	// 启动框架
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("框架启动panic: %v", r)
			}
		}()

		err := core.New().
			Port(8089).
			Mode("release").
			Name("ttl-test").
			OnlyHTTP().
			Run()
		if err != nil {
			t.Logf("框架启动失败: %v", err)
		}
	}()

	// 等待框架初始化
	t.Log("⏳ 等待框架初始化...")
	time.Sleep(time.Second * 3)

	t.Run("CacheAvailability", func(t *testing.T) {
		sessionCache := core.GetSessionCache()
		if sessionCache == nil {
			t.Fatal("会话缓存（TTL策略）应该可用")
		}
	})

	t.Run("DefaultTTL", func(t *testing.T) {
		sessionCache := core.GetSessionCache()
		if sessionCache == nil {
			t.Skip("会话缓存不可用")
		}

		ctx := context.Background()

		err := sessionCache.Set(ctx, "default_key", "default_value", 0)
		if err != nil {
			t.Errorf("默认TTL设置失败: %v", err)
		}

		value, err := sessionCache.Get(ctx, "default_key")
		if err != nil {
			t.Errorf("默认TTL获取失败: %v", err)
		}
		if value != "default_value" {
			t.Errorf("默认TTL值不正确: 期望 'default_value', 得到 '%v'", value)
		}
	})

	t.Run("CustomShortTTL", func(t *testing.T) {
		sessionCache := core.GetSessionCache()
		if sessionCache == nil {
			t.Skip("会话缓存不可用")
		}

		ctx := context.Background()

		// 设置500ms过期的数据
		err := sessionCache.Set(ctx, "short_key", "short_value", time.Millisecond*500)
		if err != nil {
			t.Errorf("自定义TTL设置失败: %v", err)
		}

		// 立即获取应该成功
		value, err := sessionCache.Get(ctx, "short_key")
		if err != nil {
			t.Errorf("立即获取失败: %v", err)
		}
		if value != "short_value" {
			t.Errorf("立即获取值不正确: 期望 'short_value', 得到 '%v'", value)
		}

		// 立即检查存在性
		exists, err := sessionCache.Exists(ctx, "short_key")
		if err != nil {
			t.Errorf("立即存在性检查失败: %v", err)
		}
		if !exists {
			t.Error("立即检查应该存在")
		}

		// 等待过期
		t.Log("⏳ 等待800ms直到数据过期...")
		time.Sleep(time.Millisecond * 800)

		// 过期后获取应该失败
		_, err = sessionCache.Get(ctx, "short_key")
		if err == nil {
			t.Error("过期后获取应该失败")
		}

		// 过期后存在性检查应该返回false
		exists, err = sessionCache.Exists(ctx, "short_key")
		if err != nil {
			t.Errorf("过期后存在性检查错误: %v", err)
		}
		if exists {
			t.Error("过期后应该不存在")
		}
	})

	t.Run("DefaultTTLStillExists", func(t *testing.T) {
		sessionCache := core.GetSessionCache()
		if sessionCache == nil {
			t.Skip("会话缓存不可用")
		}

		ctx := context.Background()

		// 验证默认TTL数据仍然存在
		value, err := sessionCache.Get(ctx, "default_key")
		if err != nil {
			t.Errorf("默认TTL数据获取失败: %v", err)
		}
		if value != "default_value" {
			t.Errorf("默认TTL数据应该仍存在: 期望 'default_value', 得到 '%v'", value)
		}
	})

	t.Run("MultipleTTLCoexistence", func(t *testing.T) {
		sessionCache := core.GetSessionCache()
		if sessionCache == nil {
			t.Skip("会话缓存不可用")
		}

		ctx := context.Background()

		// 设置不同过期时间的数据
		sessionCache.Set(ctx, "ttl_1s", "1秒数据", time.Second)
		sessionCache.Set(ctx, "ttl_2s", "2秒数据", time.Second*2)

		// 立即检查都存在
		exists1, _ := sessionCache.Exists(ctx, "ttl_1s")
		exists2, _ := sessionCache.Exists(ctx, "ttl_2s")

		if !exists1 || !exists2 {
			t.Error("不同TTL数据都应该立即存在")
		}

		// 等待1.5秒，1秒的应该过期
		time.Sleep(time.Millisecond * 1500)

		exists1, _ = sessionCache.Exists(ctx, "ttl_1s")
		exists2, _ = sessionCache.Exists(ctx, "ttl_2s")

		if exists1 {
			t.Error("1.5秒后1秒TTL数据应该过期")
		}
		if !exists2 {
			t.Error("1.5秒后2秒TTL数据应该仍存在")
		}
	})
}
