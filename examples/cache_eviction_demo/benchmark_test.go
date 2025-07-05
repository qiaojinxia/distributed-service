package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
)

// BenchmarkLRUCache 基准测试LRU缓存
func BenchmarkLRUCache(b *testing.B) {
	config := cache.MemoryConfig{
		MaxSize:        10000,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}
	
	lruCache, err := cache.NewMemoryCache(config)
	if err != nil {
		b.Fatal(err)
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			lruCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}
	})
	
	b.Run("Get", func(b *testing.B) {
		// 先填充一些数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			lruCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			lruCache.Get(ctx, key)
		}
	})
	
	b.Run("Mixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			if i%2 == 0 {
				lruCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
			} else {
				lruCache.Get(ctx, key)
			}
		}
	})
}

// BenchmarkTTLCache 基准测试TTL缓存
func BenchmarkTTLCache(b *testing.B) {
	config := cache.MemoryConfig{
		MaxSize:        10000,
		DefaultTTL:     time.Hour,
		EvictionPolicy: cache.EvictionPolicyTTL,
	}
	
	ttlCache, err := cache.NewMemoryCache(config)
	if err != nil {
		b.Fatal(err)
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			ttlCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}
	})
	
	b.Run("Get", func(b *testing.B) {
		// 先填充一些数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			ttlCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			ttlCache.Get(ctx, key)
		}
	})
}

// BenchmarkSimpleCache 基准测试Simple缓存
func BenchmarkSimpleCache(b *testing.B) {
	config := cache.MemoryConfig{
		MaxSize:         10000,
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute,
		EvictionPolicy:  cache.EvictionPolicySimple,
	}
	
	simpleCache, err := cache.NewMemoryCache(config)
	if err != nil {
		b.Fatal(err)
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			simpleCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}
	})
	
	b.Run("Get", func(b *testing.B) {
		// 先填充一些数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			simpleCache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			simpleCache.Get(ctx, key)
		}
	})
}

// BenchmarkComparePolicies 对比不同策略的性能
func BenchmarkComparePolicies(b *testing.B) {
	ctx := context.Background()
	
	policies := []struct {
		name   string
		policy cache.EvictionPolicy
	}{
		{"LRU", cache.EvictionPolicyLRU},
		{"TTL", cache.EvictionPolicyTTL},
		{"Simple", cache.EvictionPolicySimple},
	}
	
	for _, p := range policies {
		b.Run(p.name, func(b *testing.B) {
			config := cache.MemoryConfig{
				MaxSize:         1000,
				DefaultTTL:      time.Hour,
				CleanupInterval: time.Minute,
				EvictionPolicy:  p.policy,
			}
			
			testCache, err := cache.NewMemoryCache(config)
			if err != nil {
				b.Fatal(err)
			}
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key_%d", i%500)
				value := fmt.Sprintf("value_%d", i)
				
				// 80% 写入，20% 读取
				if i%5 == 0 {
					testCache.Get(ctx, key)
				} else {
					testCache.Set(ctx, key, value, 0)
				}
			}
		})
	}
}