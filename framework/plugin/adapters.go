package plugin

import (
	"context"
	"fmt"

	"distributed-service/framework/config"
	"distributed-service/pkg/etcd"
	"distributed-service/pkg/kafka"
	"distributed-service/pkg/redis_cluster"
)

// RedisClusterPlugin Redis Cluster插件适配器
type RedisClusterPlugin struct {
	*BaseServicePlugin
	client *redis_cluster.Client
}

// NewRedisClusterPlugin 创建Redis集群插件
func NewRedisClusterPlugin() *RedisClusterPlugin {
	plugin := &RedisClusterPlugin{
		BaseServicePlugin: NewBaseServicePlugin(
			"redis-cluster",
			"v1.0.0",
			"Redis Cluster service plugin for distributed caching",
		),
	}

	// 设置初始化逻辑
	plugin.OnInitialize(plugin.initialize)
	plugin.OnStart(plugin.start)
	plugin.OnStop(plugin.stop)
	plugin.OnDestroy(plugin.destroy)

	// 添加服务端点
	plugin.AddEndpoint(Endpoint{
		Name:        "ping",
		Path:        "/redis-cluster/ping",
		Method:      "GET",
		Description: "Check Redis cluster health",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "set",
		Path:        "/redis-cluster/set",
		Method:      "POST",
		Description: "Set key-value pair",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "get",
		Path:        "/redis-cluster/get/{key}",
		Method:      "GET",
		Description: "Get value by key",
	})

	return plugin
}

func (p *RedisClusterPlugin) initialize(ctx context.Context, config Config) error {
	// 从配置中获取Redis集群配置
	clusterConfig := &redis_cluster.Config{
		Addrs:      []string{"localhost:7000", "localhost:7001", "localhost:7002"},
		Password:   config.GetString("password"),
		MaxRetries: config.GetInt("max_retries"),
		PoolSize:   config.GetInt("pool_size"),
	}

	// 如果配置中有自定义地址
	if addrs := config.Get("addrs"); addrs != nil {
		if addrList, ok := addrs.([]string); ok {
			clusterConfig.Addrs = addrList
		}
	}

	// 创建Redis集群客户端
	client, err := redis_cluster.NewClient(clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to create redis cluster client: %w", err)
	}

	p.client = client
	p.SetService(client)

	return nil
}

func (p *RedisClusterPlugin) start(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("redis cluster client not initialized")
	}

	// 测试连接
	if err := p.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping redis cluster: %w", err)
	}

	return nil
}

func (p *RedisClusterPlugin) stop(ctx context.Context) error {
	// Redis集群客户端停止逻辑（如果需要）
	return nil
}

func (p *RedisClusterPlugin) destroy(ctx context.Context) error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// GetClient 获取Redis集群客户端
func (p *RedisClusterPlugin) GetClient() *redis_cluster.Client {
	return p.client
}

// KafkaPlugin Kafka插件适配器
type KafkaPlugin struct {
	*BaseServicePlugin
	client *kafka.Client
}

// NewKafkaPlugin 创建Kafka插件
func NewKafkaPlugin() *KafkaPlugin {
	plugin := &KafkaPlugin{
		BaseServicePlugin: NewBaseServicePlugin(
			"kafka",
			"v1.0.0",
			"Apache Kafka service plugin for distributed messaging",
		),
	}

	// 设置依赖
	plugin.SetDependencies([]string{"logger"})

	// 设置生命周期逻辑
	plugin.OnInitialize(plugin.initialize)
	plugin.OnStart(plugin.start)
	plugin.OnStop(plugin.stop)
	plugin.OnDestroy(plugin.destroy)

	// 添加服务端点
	plugin.AddEndpoint(Endpoint{
		Name:        "send",
		Path:        "/kafka/send",
		Method:      "POST",
		Description: "Send message to topic",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "topics",
		Path:        "/kafka/topics",
		Method:      "GET",
		Description: "Get available topics",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "consume",
		Path:        "/kafka/consume",
		Method:      "POST",
		Description: "Start consuming messages",
	})

	return plugin
}

func (p *KafkaPlugin) initialize(ctx context.Context, config Config) error {
	// 从配置中获取Kafka配置
	kafkaConfig := &kafka.Config{
		Brokers:  []string{"localhost:9092"},
		ClientID: config.GetString("client_id"),
		Group:    config.GetString("group"),
		Version:  config.GetString("version"),
	}

	// 如果配置中有自定义brokers
	if brokers := config.Get("brokers"); brokers != nil {
		if brokerList, ok := brokers.([]string); ok {
			kafkaConfig.Brokers = brokerList
		}
	}

	// 设置默认值
	if kafkaConfig.ClientID == "" {
		kafkaConfig.ClientID = "plugin-kafka-client"
	}
	if kafkaConfig.Group == "" {
		kafkaConfig.Group = "plugin-kafka-group"
	}
	if kafkaConfig.Version == "" {
		kafkaConfig.Version = "2.8.0"
	}

	// 创建Kafka客户端
	client, err := kafka.NewClient(kafkaConfig)
	if err != nil {
		return fmt.Errorf("failed to create kafka client: %w", err)
	}

	p.client = client
	p.SetService(client)

	return nil
}

func (p *KafkaPlugin) start(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	// 创建生产者
	if err := p.client.CreateProducer(); err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return nil
}

func (p *KafkaPlugin) stop(ctx context.Context) error {
	// Kafka客户端停止逻辑（如果需要）
	return nil
}

func (p *KafkaPlugin) destroy(ctx context.Context) error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// GetClient 获取Kafka客户端
func (p *KafkaPlugin) GetClient() *kafka.Client {
	return p.client
}

// EtcdPlugin Etcd插件适配器
type EtcdPlugin struct {
	*BaseServicePlugin
	client *etcd.Client
}

// NewEtcdPlugin 创建Etcd插件
func NewEtcdPlugin() *EtcdPlugin {
	plugin := &EtcdPlugin{
		BaseServicePlugin: NewBaseServicePlugin(
			"etcd",
			"v1.0.0",
			"Etcd service plugin for distributed configuration and coordination",
		),
	}

	// 设置依赖
	plugin.SetDependencies([]string{"logger"})

	// 设置生命周期逻辑
	plugin.OnInitialize(plugin.initialize)
	plugin.OnStart(plugin.start)
	plugin.OnStop(plugin.stop)
	plugin.OnDestroy(plugin.destroy)

	// 添加服务端点
	plugin.AddEndpoint(Endpoint{
		Name:        "put",
		Path:        "/etcd/put",
		Method:      "POST",
		Description: "Put key-value pair",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "get",
		Path:        "/etcd/get/{key}",
		Method:      "GET",
		Description: "Get value by key",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "delete",
		Path:        "/etcd/delete/{key}",
		Method:      "DELETE",
		Description: "Delete key",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "watch",
		Path:        "/etcd/watch/{key}",
		Method:      "GET",
		Description: "Watch key changes",
	})

	return plugin
}

func (p *EtcdPlugin) initialize(ctx context.Context, config Config) error {
	// 从配置中获取Etcd配置
	etcdConfig := &etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		Username:    config.GetString("username"),
		Password:    config.GetString("password"),
		DialTimeout: config.GetInt("dial_timeout"),
	}

	// 如果配置中有自定义端点
	if endpoints := config.Get("endpoints"); endpoints != nil {
		if endpointList, ok := endpoints.([]string); ok {
			etcdConfig.Endpoints = endpointList
		}
	}

	// 设置默认值
	if etcdConfig.DialTimeout == 0 {
		etcdConfig.DialTimeout = 5
	}

	// 创建Etcd客户端
	client, err := etcd.NewClient(etcdConfig)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}

	p.client = client
	p.SetService(client)

	return nil
}

func (p *EtcdPlugin) start(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("etcd client not initialized")
	}

	// 测试连接
	if err := p.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping etcd: %w", err)
	}

	return nil
}

func (p *EtcdPlugin) stop(ctx context.Context) error {
	// Etcd客户端停止逻辑（如果需要）
	return nil
}

func (p *EtcdPlugin) destroy(ctx context.Context) error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// GetClient 获取Etcd客户端
func (p *EtcdPlugin) GetClient() *etcd.Client {
	return p.client
}

// SchedulerPlugin 定时任务调度器插件适配器
type SchedulerPlugin struct {
	*BaseServicePlugin
	scheduler TaskScheduler
}

// NewSchedulerPlugin 创建定时任务调度器插件
func NewSchedulerPlugin() *SchedulerPlugin {
	plugin := &SchedulerPlugin{
		BaseServicePlugin: NewBaseServicePlugin(
			"scheduler",
			"v1.0.0",
			"Task scheduler plugin for distributed task management",
		),
	}

	// 设置依赖
	plugin.SetDependencies([]string{"logger"})

	// 设置生命周期逻辑
	plugin.OnInitialize(plugin.initialize)
	plugin.OnStart(plugin.start)
	plugin.OnStop(plugin.stop)
	plugin.OnDestroy(plugin.destroy)

	// 添加服务端点
	plugin.AddEndpoint(Endpoint{
		Name:        "schedule",
		Path:        "/scheduler/schedule",
		Method:      "POST",
		Description: "Schedule a new task",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "tasks",
		Path:        "/scheduler/tasks",
		Method:      "GET",
		Description: "Get all tasks",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "task",
		Path:        "/scheduler/tasks/{id}",
		Method:      "GET",
		Description: "Get task by ID",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "cancel",
		Path:        "/scheduler/tasks/{id}/cancel",
		Method:      "POST",
		Description: "Cancel a task",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "pause",
		Path:        "/scheduler/tasks/{id}/pause",
		Method:      "POST",
		Description: "Pause a task",
	})
	plugin.AddEndpoint(Endpoint{
		Name:        "resume",
		Path:        "/scheduler/tasks/{id}/resume",
		Method:      "POST",
		Description: "Resume a task",
	})

	return plugin
}

func (p *SchedulerPlugin) initialize(ctx context.Context, config Config) error {
	// 创建任务调度器
	scheduler := NewDefaultTaskScheduler()

	// 设置日志记录器
	if logger := NewSimplePluginLogger("scheduler"); logger != nil {
		scheduler.SetLogger(logger)
	}

	// 设置事件处理器
	scheduler.SetEventHandler(func(event *TaskEvent) {
		// 可以通过插件的事件总线转发调度器事件
		if p.GetContext() != nil && p.GetContext().EventBus != nil {
			pluginEvent := &Event{
				Type:      fmt.Sprintf("scheduler.task.%s", event.Type.String()),
				Source:    "scheduler",
				Data:      event,
				Timestamp: event.Timestamp,
			}
			p.GetContext().EventBus.Publish(pluginEvent)
		}
	})

	p.scheduler = scheduler
	p.SetService(scheduler)

	return nil
}

func (p *SchedulerPlugin) start(ctx context.Context) error {
	if p.scheduler == nil {
		return fmt.Errorf("scheduler not initialized")
	}

	// 启动调度器
	if err := p.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	return nil
}

func (p *SchedulerPlugin) stop(ctx context.Context) error {
	if p.scheduler != nil {
		return p.scheduler.Stop(ctx)
	}
	return nil
}

func (p *SchedulerPlugin) destroy(ctx context.Context) error {
	// 停止调度器
	if p.scheduler != nil {
		return p.scheduler.Stop(ctx)
	}
	return nil
}

// GetScheduler 获取任务调度器
func (p *SchedulerPlugin) GetScheduler() TaskScheduler {
	return p.scheduler
}

// ScheduleTask 调度任务的便捷方法
func (p *SchedulerPlugin) ScheduleTask(task *Task) error {
	if p.scheduler == nil {
		return fmt.Errorf("scheduler not initialized")
	}
	return p.scheduler.ScheduleTask(task)
}

// LoggerPlugin Logger插件适配器
type LoggerPlugin struct {
	*BaseServicePlugin
}

// NewLoggerPlugin 创建日志插件
func NewLoggerPlugin() *LoggerPlugin {
	plugin := &LoggerPlugin{
		BaseServicePlugin: NewBaseServicePlugin(
			"logger",
			"v1.0.0",
			"Logger service plugin for application logging",
		),
	}

	// 设置生命周期逻辑
	plugin.OnInitialize(plugin.initialize)

	return plugin
}

func (p *LoggerPlugin) initialize(ctx context.Context, config Config) error {
	// 日志插件初始化逻辑
	// 这里可以配置日志级别、输出格式等
	return nil
}

// ConfigPlugin 配置插件适配器
type ConfigPlugin struct {
	*BaseServicePlugin
	config *config.Config
}

// NewConfigPlugin 创建配置插件
func NewConfigPlugin() *ConfigPlugin {
	plugin := &ConfigPlugin{
		BaseServicePlugin: NewBaseServicePlugin(
			"config",
			"v1.0.0",
			"Configuration service plugin for application settings",
		),
	}

	// 设置生命周期逻辑
	plugin.OnInitialize(plugin.initialize)

	return plugin
}

func (p *ConfigPlugin) initialize(ctx context.Context, config Config) error {
	// 配置插件初始化逻辑
	// 这里可以加载配置文件等
	return nil
}

// DefaultPluginFactory PluginFactory 插件工厂实现
type DefaultPluginFactory struct{}

// NewDefaultPluginFactory 创建默认插件工厂
func NewDefaultPluginFactory() *DefaultPluginFactory {
	return &DefaultPluginFactory{}
}

// CreatePlugin 创建插件
func (f *DefaultPluginFactory) CreatePlugin(pluginType string, config Config) (Plugin, error) {
	switch pluginType {
	case "redis-cluster":
		return NewRedisClusterPlugin(), nil
	case "kafka":
		return NewKafkaPlugin(), nil
	case "etcd":
		return NewEtcdPlugin(), nil
	case "logger":
		return NewLoggerPlugin(), nil
	case "config":
		return NewConfigPlugin(), nil
	case "scheduler":
		return NewSchedulerPlugin(), nil
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", pluginType)
	}
}

// GetSupportedTypes 获取支持的插件类型
func (f *DefaultPluginFactory) GetSupportedTypes() []string {
	return []string{
		"redis-cluster",
		"kafka",
		"etcd",
		"logger",
		"config",
		"scheduler",
	}
}

// SimplePluginLogger 简单插件日志实现
type SimplePluginLogger struct {
	name string
}

// NewSimplePluginLogger 创建简单插件日志
func NewSimplePluginLogger(name string) *SimplePluginLogger {
	return &SimplePluginLogger{name: name}
}

func (l *SimplePluginLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG][%s] %s %v\n", l.name, msg, fields)
}

func (l *SimplePluginLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO][%s] %s %v\n", l.name, msg, fields)
}

func (l *SimplePluginLogger) Warn(msg string, fields ...interface{}) {
	fmt.Printf("[WARN][%s] %s %v\n", l.name, msg, fields)
}

func (l *SimplePluginLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR][%s] %s %v\n", l.name, msg, fields)
}

func (l *SimplePluginLogger) Fatal(msg string, fields ...interface{}) {
	fmt.Printf("[FATAL][%s] %s %v\n", l.name, msg, fields)
}
