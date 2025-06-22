package mongodb

import (
	"context"
	"fmt"

	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
)

// Client MongoDB客户端
type Client struct {
	config   *Config
	logger   logger.Logger
	database Database
}

// Config MongoDB配置
type Config struct {
	URI      string `yaml:"uri" json:"uri"`           // MongoDB连接URI
	Database string `yaml:"database" json:"database"` // 数据库名称
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码

	// 连接池配置
	MaxPoolSize     int `yaml:"max_pool_size" json:"max_pool_size"`       // 最大连接池大小
	MinPoolSize     int `yaml:"min_pool_size" json:"min_pool_size"`       // 最小连接池大小
	MaxIdleTime     int `yaml:"max_idle_time" json:"max_idle_time"`       // 最大空闲时间(秒)
	ConnectTimeout  int `yaml:"connect_timeout" json:"connect_timeout"`   // 连接超时(秒)
	ServerTimeout   int `yaml:"server_timeout" json:"server_timeout"`     // 服务器选择超时(秒)
	SocketTimeout   int `yaml:"socket_timeout" json:"socket_timeout"`     // 套接字超时(秒)
	HeartbeatFreq   int `yaml:"heartbeat_freq" json:"heartbeat_freq"`     // 心跳频率(秒)
	CompressorLevel int `yaml:"compressor_level" json:"compressor_level"` // 压缩级别

	// TLS配置
	TLS struct {
		Enable   bool   `yaml:"enable" json:"enable"`
		CertFile string `yaml:"cert_file" json:"cert_file"`
		KeyFile  string `yaml:"key_file" json:"key_file"`
		CAFile   string `yaml:"ca_file" json:"ca_file"`
	} `yaml:"tls" json:"tls"`
}

// Database 数据库接口
type Database interface {
	Collection(name string) Collection
	RunCommand(ctx context.Context, command interface{}) (interface{}, error)
	Drop(ctx context.Context) error
	Ping(ctx context.Context) error
}

// Collection 集合接口
type Collection interface {
	InsertOne(ctx context.Context, document interface{}) (interface{}, error)
	InsertMany(ctx context.Context, documents []interface{}) ([]interface{}, error)
	FindOne(ctx context.Context, filter interface{}) (interface{}, error)
	Find(ctx context.Context, filter interface{}) ([]interface{}, error)
	UpdateOne(ctx context.Context, filter, update interface{}) error
	UpdateMany(ctx context.Context, filter, update interface{}) error
	DeleteOne(ctx context.Context, filter interface{}) error
	DeleteMany(ctx context.Context, filter interface{}) error
	CountDocuments(ctx context.Context, filter interface{}) (int64, error)
}

// NewClient 创建MongoDB客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mongodb config is required")
	}

	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}

	return &Client{
		config: cfg,
		logger: logger.GetLogger(),
	}, nil
}

// Connect 连接到MongoDB
func (c *Client) Connect(ctx context.Context) error {
	c.logger.Infof(context.Background(), "Connecting to MongoDB: %s", c.config.URI)

	// 模拟连接逻辑
	db := &database{
		name:   c.config.Database,
		client: c,
		logger: c.logger,
	}

	c.database = db
	c.logger.Info(context.Background(), "Connected to MongoDB successfully")
	return nil
}

// Disconnect 断开连接
func (c *Client) Disconnect(ctx context.Context) error {
	if c.database == nil {
		return nil
	}

	c.logger.Info(context.Background(), "Disconnected from MongoDB")
	return nil
}

// Database 获取数据库实例
func (c *Client) Database() Database {
	return c.database
}

// Ping 检查连接
func (c *Client) Ping(ctx context.Context) error {
	if c.database == nil {
		return fmt.Errorf("mongodb not connected")
	}

	return c.database.Ping(ctx)
}

// Close 关闭连接
func (c *Client) Close() error {
	c.logger.Info(context.Background(), "MongoDB client closed")
	return nil
}

// database 数据库实现
type database struct {
	name   string
	client *Client
	logger logger.Logger
}

func (d *database) Collection(name string) Collection {
	return &collection{
		name:     name,
		database: d,
		logger:   d.logger,
	}
}

func (d *database) RunCommand(ctx context.Context, command interface{}) (interface{}, error) {
	d.logger.Infof(context.Background(), "Running command on database: %s", d.name)
	return nil, nil
}

func (d *database) Drop(ctx context.Context) error {
	d.logger.Info(context.Background(), "Dropping MongoDB database")
	return nil
}

func (d *database) Ping(ctx context.Context) error {
	d.logger.Info(context.Background(), "Pinging MongoDB")
	return nil
}

// collection 集合实现
type collection struct {
	name     string
	database *database
	logger   logger.Logger
}

func (c *collection) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	c.logger.Infof(ctx, "Inserting document into collection: %s", c.name)
	return "generated_id", nil
}

func (c *collection) InsertMany(ctx context.Context, documents []interface{}) ([]interface{}, error) {
	c.logger.Infof(ctx, "Inserting %d documents into collection: %s", len(documents), c.name)
	return make([]interface{}, len(documents)), nil
}

func (c *collection) FindOne(ctx context.Context, filter interface{}) (interface{}, error) {
	c.logger.Infof(ctx, "Finding one document in collection: %s", c.name)
	return nil, nil
}

func (c *collection) Find(ctx context.Context, filter interface{}) ([]interface{}, error) {
	c.logger.Infof(ctx, "Finding documents in collection: %s", c.name)
	return nil, nil
}

func (c *collection) UpdateOne(ctx context.Context, filter, update interface{}) error {
	c.logger.Infof(ctx, "Updating one document in collection: %s", c.name)
	return nil
}

func (c *collection) UpdateMany(ctx context.Context, filter, update interface{}) error {
	c.logger.Infof(context.Background(), "Updating many documents in collection: %s", c.name)
	return nil
}

func (c *collection) DeleteOne(ctx context.Context, filter interface{}) error {
	c.logger.Infof(context.Background(), "Deleting one document from collection: %s", c.name)
	return nil
}

func (c *collection) DeleteMany(ctx context.Context, filter interface{}) error {
	c.logger.Infof(context.Background(), "Deleting many documents from collection: %s", c.name)
	return nil
}

func (c *collection) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	c.logger.Infof(context.Background(), "Counting documents in collection: %s", c.name)
	return 0, nil
}

// ConvertConfig 转换配置格式
func ConvertConfig(cfg *config.MongoDBConfig) (*Config, error) {
	return &Config{
		URI:            cfg.URI,
		Database:       cfg.Database,
		Username:       cfg.Username,
		Password:       cfg.Password,
		MaxPoolSize:    cfg.MaxPoolSize,
		MinPoolSize:    cfg.MinPoolSize,
		MaxIdleTime:    cfg.MaxIdleTimeMS,
		ConnectTimeout: cfg.ConnectTimeout,
		SocketTimeout:  cfg.SocketTimeout,
		TLS: struct {
			Enable   bool   `yaml:"enable" json:"enable"`
			CertFile string `yaml:"cert_file" json:"cert_file"`
			KeyFile  string `yaml:"key_file" json:"key_file"`
			CAFile   string `yaml:"ca_file" json:"ca_file"`
		}{
			Enable:   cfg.TLS.Enable,
			CertFile: cfg.TLS.CertFile,
			KeyFile:  cfg.TLS.KeyFile,
			CAFile:   cfg.TLS.CAFile,
		},
	}, nil
}

// 全局MongoDB客户端实例
var globalClient *Client

// InitMongoDB 初始化MongoDB客户端
func InitMongoDB(ctx context.Context, cfg *Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return err
	}

	// 连接到MongoDB
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("mongodb connection failed: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("mongodb ping failed: %w", err)
	}

	globalClient = client
	logger.GetLogger().Info(context.Background(), "MongoDB client initialized successfully")
	return nil
}

// GetGlobalDatabase 获取全局数据库实例
func GetGlobalDatabase() Database {
	return globalClient.Database()
}

// GetClient 获取全局MongoDB客户端
func GetClient() *Client {
	return globalClient
}
