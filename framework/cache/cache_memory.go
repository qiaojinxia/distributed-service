package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	gocache "github.com/patrickmn/go-cache"
)

// EvictionPolicy 淘汰策略类型
type EvictionPolicy string

const (
	EvictionPolicyLRU    EvictionPolicy = "lru"    // 最近最少使用
	EvictionPolicyTTL    EvictionPolicy = "ttl"    // 基于过期时间
	EvictionPolicySimple EvictionPolicy = "simple" // 简单策略(go-cache)
)

type MemoryCache struct {
	// 使用现成的库实现不同的淘汰策略
	lruCache    *lru.Cache[string, interface{}]     // LRU策略
	expireCache *expirable.LRU[string, interface{}] // TTL策略
	goCache     *gocache.Cache                      // 支持TTL的简单缓存

	// 配置
	config    MemoryConfig
	stats     Stats
	mutex     sync.RWMutex
	onEvicted func(string, interface{})
}

type MemoryConfig struct {
	MaxSize         int            `json:"max_size" yaml:"max_size"`                 // 最大缓存大小
	DefaultTTL      time.Duration  `json:"default_ttl" yaml:"default_ttl"`           // 默认过期时间
	CleanupInterval time.Duration  `json:"cleanup_interval" yaml:"cleanup_interval"` // 清理间隔
	EvictionPolicy  EvictionPolicy `json:"eviction_policy" yaml:"eviction_policy"`   // 淘汰策略
}

func NewMemoryCache(config MemoryConfig) (*MemoryCache, error) {
	// 设置默认值
	if config.MaxSize <= 0 {
		config.MaxSize = 1000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = time.Hour
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = time.Minute * 10
	}
	if config.EvictionPolicy == "" {
		config.EvictionPolicy = EvictionPolicyLRU
	}

	mc := &MemoryCache{
		config: config,
		stats:  Stats{LastUpdated: time.Now()},
	}

	// 根据策略初始化相应的缓存库
	switch config.EvictionPolicy {
	case EvictionPolicyLRU:
		lruCache, err := lru.NewWithEvict(config.MaxSize, mc.onLRUEvicted)
		if err != nil {
			return nil, err
		}
		mc.lruCache = lruCache

	case EvictionPolicyTTL:
		mc.expireCache = expirable.NewLRU[string, interface{}](
			config.MaxSize,
			mc.onExpirableEvicted,
			config.DefaultTTL,
		)

	case EvictionPolicySimple:
		mc.goCache = gocache.New(config.DefaultTTL, config.CleanupInterval)
		if mc.onEvicted != nil {
			mc.goCache.OnEvicted(mc.onEvicted)
		}

	default:
		return nil, fmt.Errorf("unsupported eviction policy: %s", config.EvictionPolicy)
	}

	return mc, nil
}

// 回调函数实现
func (m *MemoryCache) onLRUEvicted(key string, value interface{}) {
	m.mutex.Lock()
	m.stats.Evictions++
	m.mutex.Unlock()
	if m.onEvicted != nil {
		m.onEvicted(key, value)
	}
}

func (m *MemoryCache) onExpirableEvicted(key string, value interface{}) {
	m.mutex.Lock()
	m.stats.Evictions++
	m.mutex.Unlock()
	if m.onEvicted != nil {
		m.onEvicted(key, value)
	}
}

// Get 接口实现
func (m *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	switch m.config.EvictionPolicy {
	case EvictionPolicyLRU:
		if m.lruCache == nil {
			return nil, ErrKeyNotFound
		}
		value, ok := m.lruCache.Get(key)
		if !ok {
			m.mutex.Lock()
			m.stats.Misses++
			m.mutex.Unlock()
			return nil, ErrKeyNotFound
		}
		m.mutex.Lock()
		m.stats.Hits++
		m.mutex.Unlock()
		return value, nil

	case EvictionPolicyTTL:
		if m.expireCache == nil {
			return nil, ErrKeyNotFound
		}
		value, ok := m.expireCache.Get(key)
		if !ok {
			m.mutex.Lock()
			m.stats.Misses++
			m.mutex.Unlock()
			return nil, ErrKeyNotFound
		}
		m.mutex.Lock()
		m.stats.Hits++
		m.mutex.Unlock()
		return value, nil

	case EvictionPolicySimple:
		if m.goCache == nil {
			return nil, ErrKeyNotFound
		}
		value, found := m.goCache.Get(key)
		if !found {
			m.mutex.Lock()
			m.stats.Misses++
			m.mutex.Unlock()
			return nil, ErrKeyNotFound
		}
		m.mutex.Lock()
		m.stats.Hits++
		m.mutex.Unlock()
		return value, nil

	default:
		return nil, fmt.Errorf("unsupported eviction policy: %s", m.config.EvictionPolicy)
	}
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	switch m.config.EvictionPolicy {
	case EvictionPolicyLRU:
		if m.lruCache == nil {
			return fmt.Errorf("LRU cache not initialized")
		}
		m.lruCache.Add(key, value)
		m.mutex.Lock()
		m.stats.Sets++
		m.mutex.Unlock()
		return nil

	case EvictionPolicyTTL:
		if m.expireCache == nil {
			return fmt.Errorf("TTL cache not initialized")
		}
		ttl := expiration
		if ttl <= 0 {
			ttl = m.config.DefaultTTL
		}
		// 如果需要不同的TTL，需要使用 AddWithTTL 方法
		m.expireCache.Add(key, value)
		m.mutex.Lock()
		m.stats.Sets++
		m.mutex.Unlock()
		return nil

	case EvictionPolicySimple:
		if m.goCache == nil {
			return fmt.Errorf("simple cache not initialized")
		}
		ttl := expiration
		if ttl <= 0 {
			ttl = m.config.DefaultTTL
		}
		m.goCache.Set(key, value, ttl)
		m.mutex.Lock()
		m.stats.Sets++
		m.mutex.Unlock()
		return nil

	default:
		return fmt.Errorf("unsupported eviction policy: %s", m.config.EvictionPolicy)
	}
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	switch m.config.EvictionPolicy {
	case EvictionPolicyLRU:
		if m.lruCache != nil {
			m.lruCache.Remove(key)
		}
	case EvictionPolicyTTL:
		if m.expireCache != nil {
			m.expireCache.Remove(key)
		}
	case EvictionPolicySimple:
		if m.goCache != nil {
			m.goCache.Delete(key)
		}
	}

	m.mutex.Lock()
	m.stats.Deletes++
	m.mutex.Unlock()

	return nil
}

func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	switch m.config.EvictionPolicy {
	case EvictionPolicyLRU:
		if m.lruCache == nil {
			return false, nil
		}
		return m.lruCache.Contains(key), nil

	case EvictionPolicyTTL:
		if m.expireCache == nil {
			return false, nil
		}
		return m.expireCache.Contains(key), nil

	case EvictionPolicySimple:
		if m.goCache == nil {
			return false, nil
		}
		_, found := m.goCache.Get(key)
		return found, nil

	default:
		return false, fmt.Errorf("unsupported eviction policy: %s", m.config.EvictionPolicy)
	}
}

func (m *MemoryCache) Clear(ctx context.Context) error {
	switch m.config.EvictionPolicy {
	case EvictionPolicyLRU:
		if m.lruCache != nil {
			m.lruCache.Purge()
		}
	case EvictionPolicyTTL:
		if m.expireCache != nil {
			m.expireCache.Purge()
		}
	case EvictionPolicySimple:
		if m.goCache != nil {
			m.goCache.Flush()
		}
	}

	return nil
}

func (m *MemoryCache) Close() error {
	return nil
}

func (m *MemoryCache) GetStats() Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.stats
}

func (m *MemoryCache) ResetStats() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.stats = Stats{LastUpdated: time.Now()}
}

// SetEvictionCallback 设置淘汰回调函数
func (m *MemoryCache) SetEvictionCallback(callback func(string, interface{})) {
	m.onEvicted = callback
}

type MemoryBuilder struct{}

func (b *MemoryBuilder) Build(config Config) (Cache, error) {
	memConfig := MemoryConfig{
		MaxSize:         1000,
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute * 10,
		EvictionPolicy:  EvictionPolicyLRU,
	}

	if settings := config.Settings; settings != nil {
		if data, err := json.Marshal(settings); err == nil {
			if err := json.Unmarshal(data, &memConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal memory config: %w", err)
			}
		}
	}

	return NewMemoryCache(memConfig)
}
