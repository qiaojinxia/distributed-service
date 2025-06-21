package lock

import (
	"context"
	"time"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	// Lock 获取锁
	Lock(ctx context.Context, key string, ttl time.Duration) (*Handle, error)

	// TryLock 尝试获取锁，不阻塞
	TryLock(ctx context.Context, key string, ttl time.Duration) (*Handle, error)

	// LockWithRetry 带重试的获取锁
	LockWithRetry(ctx context.Context, key string, ttl time.Duration, retryInterval time.Duration, maxRetries int) (*Handle, error)
}

// Handle LockHandle 锁句柄
type Handle struct {
	Key       string
	Value     string
	TTL       time.Duration
	CreatedAt time.Time
	locker    DistributedLock
	ctx       context.Context
	cancel    context.CancelFunc
}

// Unlock 释放锁
func (h *Handle) Unlock() error {
	if h.cancel != nil {
		h.cancel()
	}
	if redisLock, ok := h.locker.(*RedisLock); ok {
		return redisLock.unlock(h.Key, h.Value)
	}
	return nil
}

// Extend 延长锁的过期时间
func (h *Handle) Extend(ttl time.Duration) error {
	if redisLock, ok := h.locker.(*RedisLock); ok {
		return redisLock.extend(h.Key, h.Value, ttl)
	}
	return nil
}

// Options LockOptions 锁选项
type Options struct {
	TTL           time.Duration // 锁的过期时间
	RetryInterval time.Duration // 重试间隔
	MaxRetries    int           // 最大重试次数
	AutoRenew     bool          // 是否自动续期
	RenewInterval time.Duration // 续期间隔
}

// DefaultLockOptions 默认锁选项
func DefaultLockOptions() *Options {
	return &Options{
		TTL:           30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxRetries:    10,
		AutoRenew:     false,
		RenewInterval: 10 * time.Second,
	}
}
