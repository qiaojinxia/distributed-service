package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	HTTP          HTTPConfig          `mapstructure:"http"`
	GRPC          GRPCConfig          `mapstructure:"grpc"`
	Logger        LoggerConfig        `mapstructure:"logger"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Consul        ConsulConfig        `mapstructure:"consul"`
	Metrics       MetricsConfig       `mapstructure:"metrics"`
	MySQL         MySQLConfig         `mapstructure:"mysql"`
	Redis         RedisConfig         `mapstructure:"redis"`
	RedisCluster  RedisClusterConfig  `mapstructure:"redis_cluster"`
	RabbitMQ      RabbitMQConfig      `mapstructure:"rabbitmq"`
	Tracing       TracingConfig       `mapstructure:"tracing"`
	Protection    ProtectionConfig    `mapstructure:"protection"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	MongoDB       MongoDBConfig       `mapstructure:"mongodb"`
	Etcd          EtcdConfig          `mapstructure:"etcd"`
	Cache         CacheConfig         `mapstructure:"cache"`
	IDGen         IDGenConfig         `mapstructure:"idgen"`
}

type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Tags    string `mapstructure:"tags"`
}

type HTTPConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`  // 秒
	WriteTimeout int    `mapstructure:"write_timeout"` // 秒
	IdleTimeout  int    `mapstructure:"idle_timeout"`  // 秒
	EnableTLS    bool   `mapstructure:"enable_tls"`
	CertFile     string `mapstructure:"cert_file"`
	KeyFile      string `mapstructure:"key_file"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Issuer    string `mapstructure:"issuer"`
}

type ConsulConfig struct {
	Host                           string `mapstructure:"host"`
	Port                           int    `mapstructure:"port"`
	ServiceCheckInterval           string `mapstructure:"service_check_interval"`
	DeregisterCriticalServiceAfter string `mapstructure:"deregister_critical_service_after"`
}

type MetricsConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	PrometheusPort int  `mapstructure:"prometheus_port"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

type TracingConfig struct {
	ServiceName    string  `mapstructure:"service_name"`
	ServiceVersion string  `mapstructure:"service_version"`
	Environment    string  `mapstructure:"environment"`
	Enabled        bool    `mapstructure:"enabled"`
	ExporterType   string  `mapstructure:"exporter_type"` // "otlp", "stdout", "jaeger"
	Endpoint       string  `mapstructure:"endpoint"`
	SampleRatio    float64 `mapstructure:"sample_ratio"`
}

type GRPCConfig struct {
	Port                  int    `mapstructure:"port"`
	MaxRecvMsgSize        int    `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize        int    `mapstructure:"max_send_msg_size"`
	ConnectionTimeout     string `mapstructure:"connection_timeout"`
	MaxConnectionIdle     string `mapstructure:"max_connection_idle"`
	MaxConnectionAge      string `mapstructure:"max_connection_age"`
	MaxConnectionAgeGrace string `mapstructure:"max_connection_age_grace"`
	Time                  string `mapstructure:"time"`
	Timeout               string `mapstructure:"timeout"`
	EnableReflection      bool   `mapstructure:"enable_reflection"`
	EnableHealthCheck     bool   `mapstructure:"enable_health_check"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	CACert    string   `mapstructure:"ca_cert"`
	Timeout   int      `mapstructure:"timeout"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers       []string `mapstructure:"brokers"`
	ClientID      string   `mapstructure:"client_id"`
	Group         string   `mapstructure:"group"`
	Version       string   `mapstructure:"version"`
	RetryBackoff  int      `mapstructure:"retry_backoff"` // 毫秒
	RetryMax      int      `mapstructure:"retry_max"`
	FlushMessages int      `mapstructure:"flush_messages"`
	FlushBytes    int      `mapstructure:"flush_bytes"`
	FlushTimeout  int      `mapstructure:"flush_timeout"` // 毫秒
	SASL          struct {
		Enable    bool   `mapstructure:"enable"`
		Mechanism string `mapstructure:"mechanism"`
		Username  string `mapstructure:"username"`
		Password  string `mapstructure:"password"`
	} `mapstructure:"sasl"`
	TLS struct {
		Enable   bool   `mapstructure:"enable"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
		CAFile   string `mapstructure:"ca_file"`
	} `mapstructure:"tls"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI            string `mapstructure:"uri"`
	Database       string `mapstructure:"database"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	AuthDatabase   string `mapstructure:"auth_database"`
	MaxPoolSize    int    `mapstructure:"max_pool_size"`
	MinPoolSize    int    `mapstructure:"min_pool_size"`
	MaxIdleTimeMS  int    `mapstructure:"max_idle_time_ms"`
	ConnectTimeout int    `mapstructure:"connect_timeout"` // 秒
	SocketTimeout  int    `mapstructure:"socket_timeout"`  // 秒
	TLS            struct {
		Enable   bool   `mapstructure:"enable"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
		CAFile   string `mapstructure:"ca_file"`
	} `mapstructure:"tls"`
}

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
	DialTimeout int      `mapstructure:"dial_timeout"` // 秒
	TLS         struct {
		Enable   bool   `mapstructure:"enable"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
		CAFile   string `mapstructure:"ca_file"`
	} `mapstructure:"tls"`
}

var GlobalConfig Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// ProtectionConfig 保护配置
type ProtectionConfig struct {
	Enabled         bool                       `mapstructure:"enabled"`
	Storage         ProtectionStorageConfig    `mapstructure:"storage"`
	RateLimitRules  []RateLimitRuleConfig      `mapstructure:"rate_limit_rules"`
	CircuitBreakers []CircuitBreakerRuleConfig `mapstructure:"circuit_breakers"`
}

// ProtectionStorageConfig 保护存储配置
type ProtectionStorageConfig struct {
	Type   string      `mapstructure:"type"`   // memory, redis
	Prefix string      `mapstructure:"prefix"` // 键前缀
	TTL    string      `mapstructure:"ttl"`    // 数据过期时间
	Redis  RedisConfig `mapstructure:"redis"`  // Redis 配置
	Memory struct {
		MaxEntries  int    `mapstructure:"max_entries"`  // 最大条目数
		CleanupTick string `mapstructure:"cleanup_tick"` // 清理间隔
	} `mapstructure:"memory"` // 内存配置
}

// RateLimitRuleConfig 限流规则配置 - 统一命名规范
type RateLimitRuleConfig struct {
	Name           string  `mapstructure:"name"`             // 限流规则名称
	Resource       string  `mapstructure:"resource"`         // 资源标识 (原key字段)
	Threshold      float64 `mapstructure:"threshold"`        // 统计窗口内允许的最大请求数量
	StatIntervalMs uint32  `mapstructure:"stat_interval_ms"` // 统计窗口时间(毫秒)，QPS = (Threshold × 1000) / StatIntervalMs
	Enabled        bool    `mapstructure:"enabled"`
	Description    string  `mapstructure:"description"`
}

// CircuitBreakerRuleConfig 熔断器规则配置
type CircuitBreakerRuleConfig struct {
	Name                         string  `mapstructure:"name"`
	Resource                     string  `mapstructure:"resource"` // 资源名称，如果为空则使用name
	Strategy                     string  `mapstructure:"strategy"` // 熔断策略: "ErrorRatio", "ErrorCount", "SlowRequestRatio"
	Enabled                      bool    `mapstructure:"enabled"`
	RetryTimeoutMs               uint32  `mapstructure:"retry_timeout_ms"`                 // 熔断后重试超时时间(毫秒)
	MinRequestAmount             uint64  `mapstructure:"min_request_amount"`               // 触发熔断的最小请求数
	StatIntervalMs               uint32  `mapstructure:"stat_interval_ms"`                 // 统计时间窗口(毫秒)
	StatSlidingWindowBucketCount uint32  `mapstructure:"stat_sliding_window_bucket_count"` // 滑动窗口桶数
	MaxAllowedRtMs               uint64  `mapstructure:"max_allowed_rt_ms"`                // 最大允许响应时间(毫秒)，仅慢调用策略有效
	Threshold                    float64 `mapstructure:"threshold"`                        // 熔断阈值，根据策略不同含义不同
	ProbeNum                     uint64  `mapstructure:"probe_num"`                        // 半开状态探测请求数量
	Description                  string  `mapstructure:"description"`
}

// RedisClusterConfig Redis集群配置
type RedisClusterConfig struct {
	Addrs      []string `mapstructure:"addrs"`       // 集群节点地址
	Password   string   `mapstructure:"password"`    // 密码
	MaxRetries int      `mapstructure:"max_retries"` // 最大重试次数

	// 连接池配置
	PoolSize           int `mapstructure:"pool_size"`            // 连接池大小
	MinIdleConns       int `mapstructure:"min_idle_conns"`       // 最小空闲连接数
	MaxConnAge         int `mapstructure:"max_conn_age"`         // 最大连接时间(秒)
	PoolTimeout        int `mapstructure:"pool_timeout"`         // 连接池超时(秒)
	IdleTimeout        int `mapstructure:"idle_timeout"`         // 空闲超时(秒)
	IdleCheckFrequency int `mapstructure:"idle_check_frequency"` // 空闲检查频率(秒)

	// 集群配置
	MaxRedirects   int  `mapstructure:"max_redirects"`    // 最大重定向次数
	ReadOnly       bool `mapstructure:"read_only"`        // 只读模式
	RouteByLatency bool `mapstructure:"route_by_latency"` // 按延迟路由
	RouteRandomly  bool `mapstructure:"route_randomly"`   // 随机路由

	// 超时配置
	DialTimeout  int `mapstructure:"dial_timeout"`  // 连接超时(秒)
	ReadTimeout  int `mapstructure:"read_timeout"`  // 读取超时(秒)
	WriteTimeout int `mapstructure:"write_timeout"` // 写入超时(秒)
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled         bool                    `mapstructure:"enabled"`           // 是否启用缓存
	DefaultType     string                  `mapstructure:"default_type"`      // 默认缓存类型 (memory, redis, hybrid)
	UseFramework    bool                    `mapstructure:"use_framework"`     // 是否使用框架Redis
	GlobalKeyPrefix string                  `mapstructure:"global_key_prefix"` // 全局键前缀
	DefaultTTL      string                  `mapstructure:"default_ttl"`       // 默认过期时间
	Caches          map[string]CacheInstance `mapstructure:"caches"`           // 预定义缓存实例
}

// CacheInstance 缓存实例配置
type CacheInstance struct {
	Type      string                 `mapstructure:"type"`       // 缓存类型
	KeyPrefix string                 `mapstructure:"key_prefix"` // 键前缀
	TTL       string                 `mapstructure:"ttl"`        // 过期时间
	Settings  map[string]interface{} `mapstructure:"settings"`   // 自定义设置
}

// CacheMemoryConfig 内存缓存配置
type CacheMemoryConfig struct {
	MaxSize         int    `mapstructure:"max_size"`         // 最大条目数
	CleanupInterval string `mapstructure:"cleanup_interval"` // 清理间隔
	EvictionPolicy  string `mapstructure:"eviction_policy"`  // 淘汰策略 (lru, lfu, fifo, random, ttl)
	EnableMetrics   bool   `mapstructure:"enable_metrics"`   // 启用指标收集
	EnableCallbacks bool   `mapstructure:"enable_callbacks"` // 启用回调函数
	ShardCount      int    `mapstructure:"shard_count"`      // 分片数量，提高并发性能
	PreAllocSize    int    `mapstructure:"pre_alloc_size"`   // 预分配大小
	TTLVariance     string `mapstructure:"ttl_variance"`     // TTL变化范围
	LFUDecayRate    string `mapstructure:"lfu_decay_rate"`   // LFU衰减率
	LFUMinFreq      int64  `mapstructure:"lfu_min_freq"`     // LFU最小频率
}

// CacheRedisConfig Redis缓存配置
type CacheRedisConfig struct {
	UseFramework bool   `mapstructure:"use_framework"` // 使用框架Redis
	KeyPrefix    string `mapstructure:"key_prefix"`    // 键前缀
	// 如果不使用框架Redis，可以单独配置
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// CacheHybridConfig 混合缓存配置
type CacheHybridConfig struct {
	L1Type        string                 `mapstructure:"l1_type"`        // L1缓存类型 (通常是memory)
	L1Settings    map[string]interface{} `mapstructure:"l1_settings"`    // L1缓存设置
	L2Type        string                 `mapstructure:"l2_type"`        // L2缓存类型 (通常是redis)
	L2Settings    map[string]interface{} `mapstructure:"l2_settings"`    // L2缓存设置
	SyncStrategy  string                 `mapstructure:"sync_strategy"`  // 同步策略
	L1TTL         string                 `mapstructure:"l1_ttl"`         // L1过期时间
	L2TTL         string                 `mapstructure:"l2_ttl"`         // L2过期时间
	WriteBack     CacheWriteBackConfig   `mapstructure:"write_back"`     // 写回配置
}

// CacheWriteBackConfig 写回配置
type CacheWriteBackConfig struct {
	Enabled   bool   `mapstructure:"enabled"`    // 是否启用写回
	Interval  string `mapstructure:"interval"`   // 写回间隔
	BatchSize int    `mapstructure:"batch_size"` // 批次大小
}

// IDGenConfig ID生成器配置
type IDGenConfig struct {
	Enabled         bool                    `mapstructure:"enabled"`           // 是否启用ID生成器
	Type            string                  `mapstructure:"type"`              // 生成器类型 (leaf, snowflake)
	UseFramework    bool                    `mapstructure:"use_framework"`     // 是否使用框架数据库配置
	DefaultStep     int32                   `mapstructure:"default_step"`      // 默认步长
	Database        IDGenDatabaseConfig     `mapstructure:"database"`          // 数据库配置（不使用框架时）
	Leaf            IDGenLeafConfig         `mapstructure:"leaf"`              // Leaf配置
	BizTags         map[string]IDGenBizTag  `mapstructure:"biz_tags"`          // 预定义业务标识
}

// IDGenDatabaseConfig ID生成器数据库配置
type IDGenDatabaseConfig struct {
	Driver          string `mapstructure:"driver"`            // 数据库驱动
	DSN             string `mapstructure:"dsn"`               // 数据源名称
	Host            string `mapstructure:"host"`              // 主机
	Port            int    `mapstructure:"port"`              // 端口
	Database        string `mapstructure:"database"`          // 数据库名
	Username        string `mapstructure:"username"`          // 用户名
	Password        string `mapstructure:"password"`          // 密码
	Charset         string `mapstructure:"charset"`           // 字符集
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int    `mapstructure:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"` // 连接最大生存时间
	LogLevel        string `mapstructure:"log_level"`         // 日志级别
}

// IDGenLeafConfig Leaf算法配置
type IDGenLeafConfig struct {
	DefaultStep      int32  `mapstructure:"default_step"`      // 默认步长
	PreloadThreshold string `mapstructure:"preload_threshold"` // 预加载阈值
	CleanupInterval  string `mapstructure:"cleanup_interval"`  // 清理间隔
	MaxStepSize      int32  `mapstructure:"max_step_size"`     // 最大步长
	MinStepSize      int32  `mapstructure:"min_step_size"`     // 最小步长
	StepAdjustRatio  string `mapstructure:"step_adjust_ratio"` // 步长调整比例
}

// IDGenBizTag 业务标识配置
type IDGenBizTag struct {
	Step        int32  `mapstructure:"step"`        // 步长
	Description string `mapstructure:"description"` // 描述
	AutoCreate  bool   `mapstructure:"auto_create"` // 是否自动创建
}
