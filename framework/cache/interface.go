package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Close() error
}

type BatchCache interface {
	Cache
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	MSet(ctx context.Context, keyValues map[string]interface{}, expiration time.Duration) error
	MDelete(ctx context.Context, keys []string) error
}

type SerializableCache interface {
	Cache
	GetObject(ctx context.Context, key string, obj interface{}) error
	SetObject(ctx context.Context, key string, obj interface{}, expiration time.Duration) error
}

type Stats struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Deletes     int64
	Errors      int64
	Evictions   int64
	LastUpdated time.Time
}

type StatsCache interface {
	Cache
	GetStats() Stats
	ResetStats()
}

type Type string

const (
	TypeMemory    Type = "memory"
	TypeRedis     Type = "redis"
	TypeMemcached Type = "memcached"
	TypeHybrid    Type = "hybrid"
)
