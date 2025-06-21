package etcd

import (
	"context"
	"fmt"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Client Etcd客户端
type Client struct {
	client *clientv3.Client
	config *Config
	logger logger.Logger
}

// Config Etcd配置
type Config struct {
	Endpoints   []string `yaml:"endpoints" json:"endpoints"`       // Etcd集群端点
	Username    string   `yaml:"username" json:"username"`         // 用户名
	Password    string   `yaml:"password" json:"password"`         // 密码
	DialTimeout int      `yaml:"dial_timeout" json:"dial_timeout"` // 连接超时(秒)

	// TLS配置
	TLS struct {
		Enable   bool   `yaml:"enable" json:"enable"`
		CertFile string `yaml:"cert_file" json:"cert_file"`
		KeyFile  string `yaml:"key_file" json:"key_file"`
		CAFile   string `yaml:"ca_file" json:"ca_file"`
	} `yaml:"tls" json:"tls"`

	// 高级配置
	AutoSyncInterval     time.Duration `yaml:"auto_sync_interval" json:"auto_sync_interval"`           // 自动同步间隔
	DialKeepAliveTime    time.Duration `yaml:"dial_keep_alive_time" json:"dial_keep_alive_time"`       // 保活时间
	DialKeepAliveTimeout time.Duration `yaml:"dial_keep_alive_timeout" json:"dial_keep_alive_timeout"` // 保活超时
	MaxCallSendMsgSize   int           `yaml:"max_call_send_msg_size" json:"max_call_send_msg_size"`   // 最大发送消息大小
	MaxCallRecvMsgSize   int           `yaml:"max_call_recv_msg_size" json:"max_call_recv_msg_size"`   // 最大接收消息大小
	CompactionMode       string        `yaml:"compaction_mode" json:"compaction_mode"`                 // 压缩模式
	RejectOldCluster     bool          `yaml:"reject_old_cluster" json:"reject_old_cluster"`           // 拒绝旧集群
}

// WatchEvent 监听事件
type WatchEvent struct {
	Type      string // PUT, DELETE
	Key       string
	Value     []byte
	PrevValue []byte
}

// WatchCallback 监听回调函数
type WatchCallback func(event *WatchEvent) error

// NewClient 创建Etcd客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("etcd config is required")
	}

	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints are required")
	}

	// 创建客户端配置
	clientCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	}

	// 设置默认值
	if clientCfg.DialTimeout == 0 {
		clientCfg.DialTimeout = 5 * time.Second
	}

	// 高级配置
	if cfg.AutoSyncInterval > 0 {
		clientCfg.AutoSyncInterval = cfg.AutoSyncInterval
	}
	if cfg.DialKeepAliveTime > 0 {
		clientCfg.DialKeepAliveTime = cfg.DialKeepAliveTime
	}
	if cfg.DialKeepAliveTimeout > 0 {
		clientCfg.DialKeepAliveTimeout = cfg.DialKeepAliveTimeout
	}
	if cfg.MaxCallSendMsgSize > 0 {
		clientCfg.MaxCallSendMsgSize = cfg.MaxCallSendMsgSize
	}
	if cfg.MaxCallRecvMsgSize > 0 {
		clientCfg.MaxCallRecvMsgSize = cfg.MaxCallRecvMsgSize
	}
	clientCfg.RejectOldCluster = cfg.RejectOldCluster

	// TLS配置
	if cfg.TLS.Enable {
		// TODO: 配置TLS证书
		// tlsConfig, err := loadTLSConfig(cfg.TLS.CertFile, cfg.TLS.KeyFile, cfg.TLS.CAFile)
		// if err != nil {
		//     return nil, fmt.Errorf("failed to load TLS config: %w", err)
		// }
		// clientCfg.TLS = tlsConfig
	}

	// 创建客户端
	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
		logger: logger.GetLogger(),
	}, nil
}

// Put 设置键值对
func (c *Client) Put(ctx context.Context, key, value string) error {
	_, err := c.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}
	return nil
}

// PutWithTTL 设置带TTL的键值对
func (c *Client) PutWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	// 创建租约
	lease, err := c.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}

	// 设置键值对并绑定租约
	_, err = c.client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to put key %s with TTL: %w", key, err)
	}

	return nil
}

// Get 获取单个键的值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if len(resp.Kvs) == 0 {
		return "", nil // 键不存在
	}

	return string(resp.Kvs[0].Value), nil
}

// GetWithPrefix 根据前缀获取多个键值对
func (c *Client) GetWithPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := c.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get keys with prefix %s: %w", prefix, err)
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

// Delete 删除键
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// DeleteWithPrefix 根据前缀删除多个键
func (c *Client) DeleteWithPrefix(ctx context.Context, prefix string) (int64, error) {
	resp, err := c.client.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return 0, fmt.Errorf("failed to delete keys with prefix %s: %w", prefix, err)
	}
	return resp.Deleted, nil
}

// Watch 监听键的变化
func (c *Client) Watch(ctx context.Context, key string, callback WatchCallback) error {
	watchChan := c.client.Watch(ctx, key)

	go func() {
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				watchEvent := &WatchEvent{
					Key:   string(event.Kv.Key),
					Value: event.Kv.Value,
				}

				switch event.Type {
				case clientv3.EventTypePut:
					watchEvent.Type = "PUT"
				case clientv3.EventTypeDelete:
					watchEvent.Type = "DELETE"
					if event.PrevKv != nil {
						watchEvent.PrevValue = event.PrevKv.Value
					}
				}

				if err := callback(watchEvent); err != nil {
					c.logger.Errorf("Watch callback error: %v", err)
				}
			}
		}
	}()

	return nil
}

// WatchWithPrefix 监听前缀匹配的键变化
func (c *Client) WatchWithPrefix(ctx context.Context, prefix string, callback WatchCallback) error {
	watchChan := c.client.Watch(ctx, prefix, clientv3.WithPrefix())

	go func() {
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				watchEvent := &WatchEvent{
					Key:   string(event.Kv.Key),
					Value: event.Kv.Value,
				}

				switch event.Type {
				case clientv3.EventTypePut:
					watchEvent.Type = "PUT"
				case clientv3.EventTypeDelete:
					watchEvent.Type = "DELETE"
					if event.PrevKv != nil {
						watchEvent.PrevValue = event.PrevKv.Value
					}
				}

				if err := callback(watchEvent); err != nil {
					c.logger.Errorf("Watch callback error: %v", err)
				}
			}
		}
	}()

	return nil
}

// Lock 分布式锁
func (c *Client) Lock(ctx context.Context, key string, ttl time.Duration) (*clientv3.TxnResponse, error) {
	// 创建租约
	lease, err := c.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to grant lease for lock: %w", err)
	}

	// 尝试获取锁
	txn := c.client.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "", clientv3.WithLease(lease.ID))).
		Else(clientv3.OpGet(key))

	resp, err := txn.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return resp, nil
}

// Unlock 释放分布式锁
func (c *Client) Unlock(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// Transaction 事务操作
func (c *Client) Transaction(ctx context.Context, conditions []clientv3.Cmp, thenOps []clientv3.Op, elseOps []clientv3.Op) (*clientv3.TxnResponse, error) {
	txn := c.client.Txn(ctx).If(conditions...).Then(thenOps...).Else(elseOps...)
	resp, err := txn.Commit()
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}
	return resp, nil
}

// Lease 租约管理
func (c *Client) GrantLease(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
	resp, err := c.client.Grant(ctx, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to grant lease: %w", err)
	}
	return resp, nil
}

// KeepAlive 保持租约活跃
func (c *Client) KeepAlive(ctx context.Context, leaseID clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ch, err := c.client.KeepAlive(ctx, leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to keep alive lease: %w", err)
	}
	return ch, nil
}

// RevokeLease 撤销租约
func (c *Client) RevokeLease(ctx context.Context, leaseID clientv3.LeaseID) error {
	_, err := c.client.Revoke(ctx, leaseID)
	if err != nil {
		return fmt.Errorf("failed to revoke lease: %w", err)
	}
	return nil
}

// MemberList 获取集群成员列表
func (c *Client) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	resp, err := c.client.MemberList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get member list: %w", err)
	}
	return resp, nil
}

// Status 获取服务器状态
func (c *Client) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	resp, err := c.client.Status(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	return resp, nil
}

// Compact 压缩历史版本
func (c *Client) Compact(ctx context.Context, rev int64) (*clientv3.CompactResponse, error) {
	resp, err := c.client.Compact(ctx, rev)
	if err != nil {
		return nil, fmt.Errorf("failed to compact: %w", err)
	}
	return resp, nil
}

// Ping 检查连接
func (c *Client) Ping(ctx context.Context) error {
	// 尝试获取一个不存在的键来测试连接
	_, err := c.client.Get(ctx, "/__health_check__")
	if err != nil {
		return fmt.Errorf("etcd ping failed: %w", err)
	}
	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	err := c.client.Close()
	if err != nil {
		c.logger.Errorf("Etcd client close failed: %v", err)
		return err
	}
	c.logger.Info("Etcd client closed")
	return nil
}

// GetClient 获取原生客户端
func (c *Client) GetClient() *clientv3.Client {
	return c.client
}

// ConvertConfig 转换配置格式
func ConvertConfig(cfg *config.EtcdConfig) (*Config, error) {
	result := &Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: cfg.DialTimeout,
	}

	result.TLS.Enable = cfg.TLS.Enable
	result.TLS.CertFile = cfg.TLS.CertFile
	result.TLS.KeyFile = cfg.TLS.KeyFile
	result.TLS.CAFile = cfg.TLS.CAFile

	return result, nil
}

// 全局Etcd客户端实例
var globalClient *Client

// InitEtcd 初始化Etcd客户端
func InitEtcd(ctx context.Context, cfg *Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return err
	}

	// 测试连接
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("etcd connection test failed: %w", err)
	}

	globalClient = client
	logger.GetLogger().Info("Etcd client initialized successfully")
	return nil
}

// GetClient 获取全局Etcd客户端
func GetClient() *Client {
	return globalClient
}
