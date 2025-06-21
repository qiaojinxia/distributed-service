package redis_cluster

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"

	"github.com/go-redis/redis/v8"
)

// Client Redis Cluster客户端
type Client struct {
	client *redis.ClusterClient
	config *Config
	logger logger.Logger
}

// Config Redis Cluster配置
type Config struct {
	Addrs      []string `yaml:"addrs" json:"addrs"`             // 集群节点地址
	Password   string   `yaml:"password" json:"password"`       // 密码
	MaxRetries int      `yaml:"max_retries" json:"max_retries"` // 最大重试次数

	// 连接池配置
	PoolSize           int           `yaml:"pool_size" json:"pool_size"`
	MinIdleConns       int           `yaml:"min_idle_conns" json:"min_idle_conns"`
	MaxConnAge         time.Duration `yaml:"max_conn_age" json:"max_conn_age"`
	PoolTimeout        time.Duration `yaml:"pool_timeout" json:"pool_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	IdleCheckFrequency time.Duration `yaml:"idle_check_frequency" json:"idle_check_frequency"`

	// 集群配置
	MaxRedirects   int  `yaml:"max_redirects" json:"max_redirects"`       // 最大重定向次数
	ReadOnly       bool `yaml:"read_only" json:"read_only"`               // 只读模式
	RouteByLatency bool `yaml:"route_by_latency" json:"route_by_latency"` // 按延迟路由
	RouteRandomly  bool `yaml:"route_randomly" json:"route_randomly"`     // 随机路由

	// 超时配置
	DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
}

// NewClient 创建Redis Cluster客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis cluster config is required")
	}

	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("redis cluster addresses are required")
	}

	// 创建集群客户端配置
	rdbOpts := &redis.ClusterOptions{
		Addrs:      cfg.Addrs,
		Password:   cfg.Password,
		MaxRetries: cfg.MaxRetries,

		// 连接池配置
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,

		// 集群配置
		MaxRedirects:   cfg.MaxRedirects,
		ReadOnly:       cfg.ReadOnly,
		RouteByLatency: cfg.RouteByLatency,
		RouteRandomly:  cfg.RouteRandomly,

		// 超时配置
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 设置默认值
	if rdbOpts.PoolSize == 0 {
		rdbOpts.PoolSize = 10
	}
	if rdbOpts.MaxRetries == 0 {
		rdbOpts.MaxRetries = 3
	}
	if rdbOpts.MaxRedirects == 0 {
		rdbOpts.MaxRedirects = 8
	}

	client := redis.NewClusterClient(rdbOpts)

	return &Client{
		client: client,
		config: cfg,
		logger: logger.GetLogger(),
	}, nil
}

// Ping 检查集群连接
func (c *Client) Ping(ctx context.Context) error {
	result := c.client.Ping(ctx)
	if result.Err() != nil {
		return fmt.Errorf("redis cluster ping failed: %w", result.Err())
	}
	return nil
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result := c.client.Get(ctx, key)
	if errors.Is(result.Err(), redis.Nil) {
		return "", nil
	}
	return result.Result()
}

// Set 设置值
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	result := c.client.Set(ctx, key, value, expiration)
	return result.Err()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	result := c.client.Del(ctx, keys...)
	return result.Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	result := c.client.Exists(ctx, keys...)
	return result.Result()
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	result := c.client.Expire(ctx, key, expiration)
	return result.Err()
}

// TTL 获取过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	result := c.client.TTL(ctx, key)
	return result.Result()
}

// HSet 哈希表设置
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	result := c.client.HSet(ctx, key, values...)
	return result.Err()
}

// HGet 哈希表获取
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	result := c.client.HGet(ctx, key, field)
	if errors.Is(result.Err(), redis.Nil) {
		return "", nil
	}
	return result.Result()
}

// HGetAll 获取哈希表所有字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result := c.client.HGetAll(ctx, key)
	return result.Result()
}

// HDel 删除哈希表字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	result := c.client.HDel(ctx, key, fields...)
	return result.Err()
}

// LPush 列表左推
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	result := c.client.LPush(ctx, key, values...)
	return result.Err()
}

// RPush 列表右推
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	result := c.client.RPush(ctx, key, values...)
	return result.Err()
}

// LPop 列表左弹出
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	result := c.client.LPop(ctx, key)
	if errors.Is(result.Err(), redis.Nil) {
		return "", nil
	}
	return result.Result()
}

// RPop 列表右弹出
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	result := c.client.RPop(ctx, key)
	if errors.Is(result.Err(), redis.Nil) {
		return "", nil
	}
	return result.Result()
}

// SAdd 集合添加
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	result := c.client.SAdd(ctx, key, members...)
	return result.Err()
}

// SMembers 获取集合所有成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	result := c.client.SMembers(ctx, key)
	return result.Result()
}

// SRem 移除集合成员
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	result := c.client.SRem(ctx, key, members...)
	return result.Err()
}

// ZAdd 有序集合添加
func (c *Client) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	result := c.client.ZAdd(ctx, key, members...)
	return result.Err()
}

// ZRange 有序集合范围查询
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result := c.client.ZRange(ctx, key, start, stop)
	return result.Result()
}

// ZRem 移除有序集合成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	result := c.client.ZRem(ctx, key, members...)
	return result.Err()
}

// Incr 递增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	result := c.client.Incr(ctx, key)
	return result.Result()
}

// Decr 递减
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	result := c.client.Decr(ctx, key)
	return result.Result()
}

// Pipeline 管道操作
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// Transaction 事务操作
func (c *Client) Transaction(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return c.client.Watch(ctx, fn, keys...)
}

// ClusterInfo 获取集群信息
func (c *Client) ClusterInfo(ctx context.Context) (string, error) {
	result := c.client.ClusterInfo(ctx)
	return result.Result()
}

// ClusterNodes 获取集群节点信息
func (c *Client) ClusterNodes(ctx context.Context) (string, error) {
	result := c.client.ClusterNodes(ctx)
	return result.Result()
}

// ForEachShard 对每个分片执行操作
func (c *Client) ForEachShard(ctx context.Context, fn func(ctx context.Context, shard *redis.Client) error) error {
	return c.client.ForEachShard(ctx, fn)
}

// ForEachMaster 对每个主节点执行操作
func (c *Client) ForEachMaster(ctx context.Context, fn func(ctx context.Context, master *redis.Client) error) error {
	return c.client.ForEachMaster(ctx, fn)
}

// Close 关闭连接
func (c *Client) Close() error {
	err := c.client.Close()
	if err != nil {
		c.logger.Errorf("Redis cluster close failed: %v", err)
		return err
	}
	c.logger.Info("Redis cluster client closed")
	return nil
}

// GetClient 获取原生客户端
func (c *Client) GetClient() *redis.ClusterClient {
	return c.client
}

// ConvertConfig 转换配置格式
func ConvertConfig(cfg *config.RedisClusterConfig) (*Config, error) {
	return &Config{
		Addrs:              cfg.Addrs,
		Password:           cfg.Password,
		MaxRetries:         cfg.MaxRetries,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         time.Duration(cfg.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(cfg.PoolTimeout) * time.Second,
		IdleTimeout:        time.Duration(cfg.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency) * time.Second,
		MaxRedirects:       cfg.MaxRedirects,
		ReadOnly:           cfg.ReadOnly,
		RouteByLatency:     cfg.RouteByLatency,
		RouteRandomly:      cfg.RouteRandomly,
		DialTimeout:        time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:        time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(cfg.WriteTimeout) * time.Second,
	}, nil
}

// 全局Redis Cluster客户端实例
var globalClient *Client

// InitRedisCluster 初始化Redis Cluster客户端
func InitRedisCluster(ctx context.Context, cfg *Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return err
	}

	// 测试连接
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("redis cluster connection test failed: %w", err)
	}

	globalClient = client
	logger.GetLogger().Info("Redis Cluster client initialized successfully")
	return nil
}

// GetClient 获取全局Redis Cluster客户端
func GetClient() *Client {
	return globalClient
}
