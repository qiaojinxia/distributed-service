package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// 预定义错误
var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockTimeout     = errors.New("lock acquisition timeout")
	ErrLockNotOwned    = errors.New("lock not owned by this instance")
)

// RedisLock Redis分布式锁实现
type RedisLock struct {
	client *redis.Client
	prefix string
}

// NewRedisLock 创建Redis分布式锁
func NewRedisLock(client *redis.Client, prefix string) *RedisLock {
	if prefix == "" {
		prefix = "lock:"
	}
	return &RedisLock{
		client: client,
		prefix: prefix,
	}
}

// Lock 获取锁（阻塞直到获取成功或超时）
func (r *RedisLock) Lock(ctx context.Context, key string, ttl time.Duration) (*Handle, error) {
	return r.LockWithRetry(ctx, key, ttl, 100*time.Millisecond, -1) // -1表示无限重试
}

// TryLock 尝试获取锁（不阻塞）
func (r *RedisLock) TryLock(ctx context.Context, key string, ttl time.Duration) (*Handle, error) {
	lockKey := r.getLockKey(key)
	value := r.generateLockValue()

	// 使用SET命令的NX选项实现原子性加锁
	result, err := r.client.SetNX(ctx, lockKey, value, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !result {
		return nil, ErrLockNotAcquired
	}

	handle := &Handle{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		CreatedAt: time.Now(),
		locker:    r,
	}

	logger.Debug(ctx, "Lock acquired",
		zap.String("key", key),
		zap.String("value", value),
		zap.Duration("ttl", ttl))

	return handle, nil
}

// LockWithRetry 带重试的获取锁
func (r *RedisLock) LockWithRetry(ctx context.Context, key string, ttl time.Duration, retryInterval time.Duration, maxRetries int) (*Handle, error) {
	retries := 0
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for {
		handle, err := r.TryLock(ctx, key, ttl)
		if err == nil {
			return handle, nil
		}

		if !errors.Is(err, ErrLockNotAcquired) {
			return nil, err
		}

		// 检查重试次数
		if maxRetries > 0 && retries >= maxRetries {
			return nil, ErrLockTimeout
		}

		retries++

		// 等待重试或上下文取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

// unlock 释放锁（内部方法）
func (r *RedisLock) unlock(key, value string) error {
	lockKey := r.getLockKey(key)

	// 使用Lua脚本确保原子性：只有值匹配才删除
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	ctx := context.Background()
	result, err := r.client.Eval(ctx, luaScript, []string{lockKey}, value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotOwned
	}

	logger.Debug(ctx, "Lock released",
		zap.String("key", key),
		zap.String("value", value))

	return nil
}

// extend 延长锁的过期时间（内部方法）
func (r *RedisLock) extend(key, value string, ttl time.Duration) error {
	lockKey := r.getLockKey(key)

	// 使用Lua脚本确保原子性：只有值匹配才延期
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	ctx := context.Background()
	result, err := r.client.Eval(ctx, luaScript, []string{lockKey}, value, int(ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotOwned
	}

	logger.Debug(ctx, "Lock extended",
		zap.String("key", key),
		zap.String("value", value),
		zap.Duration("ttl", ttl))

	return nil
}

// StartAutoRenew 启动自动续期
func (r *RedisLock) StartAutoRenew(handle *Handle, renewInterval time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	handle.ctx = ctx
	handle.cancel = cancel

	go func() {
		ticker := time.NewTicker(renewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := handle.Extend(handle.TTL); err != nil {
					logger.Error(ctx, "Failed to auto-renew lock",
						zap.String("key", handle.Key),
						zap.Error(err))
					return
				}
				logger.Debug(ctx, "Lock auto-renewed",
					zap.String("key", handle.Key))
			}
		}
	}()
}

// getLockKey 获取完整的锁键名
func (r *RedisLock) getLockKey(key string) string {
	return r.prefix + key
}

// generateLockValue 生成锁的唯一值
func (r *RedisLock) generateLockValue() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
