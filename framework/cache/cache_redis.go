package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// SimpleRedisCache 简化的Redis缓存实现，只支持外部注入的Redis客户端
type SimpleRedisCache struct {
	client    *redis.Client
	stats     Stats
	keyPrefix string
}

// NewSimpleRedisCache 使用外部Redis客户端创建缓存实例
func NewSimpleRedisCache(client *redis.Client, keyPrefix string) *SimpleRedisCache {
	return &SimpleRedisCache{
		client:    client,
		stats:     Stats{LastUpdated: time.Now()},
		keyPrefix: keyPrefix,
	}
}

// addKeyPrefix 添加键前缀
func (r *SimpleRedisCache) addKeyPrefix(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	return r.keyPrefix + ":" + key
}

// addKeyPrefixToSlice 为键切片添加前缀
func (r *SimpleRedisCache) addKeyPrefixToSlice(keys []string) []string {
	if r.keyPrefix == "" {
		return keys
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.keyPrefix + ":" + key
	}
	return prefixedKeys
}

// Get 获取值
func (r *SimpleRedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	prefixedKey := r.addKeyPrefix(key)
	result, err := r.client.Get(ctx, prefixedKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.stats.Misses++
			return nil, ErrKeyNotFound
		}
		r.stats.Errors++
		return nil, err
	}

	r.stats.Hits++
	return result, nil
}

// Set 设置值
func (r *SimpleRedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	prefixedKey := r.addKeyPrefix(key)
	err := r.client.Set(ctx, prefixedKey, value, expiration).Err()
	if err != nil {
		r.stats.Errors++
		return err
	}

	r.stats.Sets++
	return nil
}

// Delete 删除键
func (r *SimpleRedisCache) Delete(ctx context.Context, key string) error {
	prefixedKey := r.addKeyPrefix(key)
	err := r.client.Del(ctx, prefixedKey).Err()
	if err != nil {
		r.stats.Errors++
		return err
	}

	r.stats.Deletes++
	return nil
}

// Exists 检查键是否存在
func (r *SimpleRedisCache) Exists(ctx context.Context, key string) (bool, error) {
	prefixedKey := r.addKeyPrefix(key)
	result, err := r.client.Exists(ctx, prefixedKey).Result()
	if err != nil {
		r.stats.Errors++
		return false, err
	}

	return result > 0, nil
}

// Clear 清空缓存（注意：这会清空整个数据库）
func (r *SimpleRedisCache) Clear(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		r.stats.Errors++
		return err
	}
	return nil
}

// Close 关闭连接（不关闭外部注入的客户端）
func (r *SimpleRedisCache) Close() error {
	// 不关闭外部注入的Redis客户端
	return nil
}

// MGet 批量获取
func (r *SimpleRedisCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	prefixedKeys := r.addKeyPrefixToSlice(keys)
	results, err := r.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		r.stats.Errors++
		return nil, err
	}

	data := make(map[string]interface{})
	for i, result := range results {
		if result != nil {
			data[keys[i]] = result // 返回原始键名，不包含前缀
			r.stats.Hits++
		} else {
			r.stats.Misses++
		}
	}

	return data, nil
}

// MSet 批量设置
func (r *SimpleRedisCache) MSet(ctx context.Context, keyValues map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()

	for key, value := range keyValues {
		prefixedKey := r.addKeyPrefix(key)
		pipe.Set(ctx, prefixedKey, value, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.stats.Errors++
		return err
	}

	r.stats.Sets += int64(len(keyValues))
	return nil
}

// MDelete 批量删除
func (r *SimpleRedisCache) MDelete(ctx context.Context, keys []string) error {
	prefixedKeys := r.addKeyPrefixToSlice(keys)
	err := r.client.Del(ctx, prefixedKeys...).Err()
	if err != nil {
		r.stats.Errors++
		return err
	}

	r.stats.Deletes += int64(len(keys))
	return nil
}

// GetStats 获取统计信息
func (r *SimpleRedisCache) GetStats() Stats {
	return r.stats
}

// ResetStats 重置统计信息
func (r *SimpleRedisCache) ResetStats() {
	r.stats = Stats{LastUpdated: time.Now()}
}

// GetObject 获取对象
func (r *SimpleRedisCache) GetObject(ctx context.Context, key string, obj interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	if str, ok := data.(string); ok {
		return json.Unmarshal([]byte(str), obj)
	}

	return fmt.Errorf("data is not a string")
}

// SetObject 设置对象
func (r *SimpleRedisCache) SetObject(ctx context.Context, key string, obj interface{}, expiration time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return r.Set(ctx, key, string(data), expiration)
}

// SimpleRedisBuilder 简化的Redis构建器
type SimpleRedisBuilder struct {
	redisClient *redis.Client
}

// NewSimpleRedisBuilder 创建简化的Redis构建器
func NewSimpleRedisBuilder(redisClient *redis.Client) *SimpleRedisBuilder {
	return &SimpleRedisBuilder{
		redisClient: redisClient,
	}
}

// Build 构建缓存实例
func (b *SimpleRedisBuilder) Build(config Config) (Cache, error) {
	if b.redisClient == nil {
		return nil, fmt.Errorf("redis client not provided")
	}

	keyPrefix := ""
	if settings := config.Settings; settings != nil {
		if prefix, ok := settings["key_prefix"].(string); ok {
			keyPrefix = prefix
		}
	}

	return NewSimpleRedisCache(b.redisClient, keyPrefix), nil
}

// RedisBuilder 标准Redis构建器（不使用外部客户端）
type RedisBuilder struct{}

func (b *RedisBuilder) Build(_ Config) (Cache, error) {
	// 这个构建器不创建Redis连接，只是为了兼容
	return nil, fmt.Errorf("RedisBuilder is deprecated, use SimpleRedisBuilder with injected client instead")
}
