package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"time"
)

// Client Kafka客户端
type Client struct {
	config   *Config
	logger   logger.Logger
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	client   sarama.Client
}

// Config Kafka配置
type Config struct {
	Brokers       []string `yaml:"brokers" json:"brokers"`               // Broker地址列表
	ClientID      string   `yaml:"client_id" json:"client_id"`           // 客户端ID
	Group         string   `yaml:"group" json:"group"`                   // 消费者组
	Version       string   `yaml:"version" json:"version"`               // Kafka版本
	RetryBackoff  int      `yaml:"retry_backoff" json:"retry_backoff"`   // 重试退避时间(毫秒)
	RetryMax      int      `yaml:"retry_max" json:"retry_max"`           // 最大重试次数
	FlushMessages int      `yaml:"flush_messages" json:"flush_messages"` // 刷新消息数
	FlushBytes    int      `yaml:"flush_bytes" json:"flush_bytes"`       // 刷新字节数
	FlushTimeout  int      `yaml:"flush_timeout" json:"flush_timeout"`   // 刷新超时(毫秒)

	// SASL认证配置
	SASL struct {
		Enable    bool   `yaml:"enable" json:"enable"`
		Mechanism string `yaml:"mechanism" json:"mechanism"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
		Username  string `yaml:"username" json:"username"`
		Password  string `yaml:"password" json:"password"`
	} `yaml:"sasl" json:"sasl"`

	// TLS配置
	TLS struct {
		Enable   bool   `yaml:"enable" json:"enable"`
		CertFile string `yaml:"cert_file" json:"cert_file"`
		KeyFile  string `yaml:"key_file" json:"key_file"`
		CAFile   string `yaml:"ca_file" json:"ca_file"`
	} `yaml:"tls" json:"tls"`
}

// Message Kafka消息
type Message struct {
	Topic     string            // 主题
	Partition int32             // 分区
	Offset    int64             // 偏移量
	Key       []byte            // 键
	Value     []byte            // 值
	Headers   map[string][]byte // 头部信息
	Timestamp time.Time         // 时间戳
}

// MessageHandler 消息处理器
type MessageHandler func(ctx context.Context, message *Message) error

// NewClient 创建Kafka客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka newConfig is required")
	}

	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	// 创建Sarama配置
	newConfig := sarama.NewConfig()
	newConfig.ClientID = cfg.ClientID

	// 设置版本 - 使用较新但稳定的版本
	if cfg.Version != "" {
		version, err := sarama.ParseKafkaVersion(cfg.Version)
		if err == nil {
			newConfig.Version = version
		}
	}

	// 网络配置
	newConfig.Net.DialTimeout = 30 * time.Second
	newConfig.Net.ReadTimeout = 30 * time.Second
	newConfig.Net.WriteTimeout = 30 * time.Second

	// SASL配置
	if cfg.SASL.Enable {
		newConfig.Net.SASL.Enable = true
		newConfig.Net.SASL.User = cfg.SASL.Username
		newConfig.Net.SASL.Password = cfg.SASL.Password
		// 简化SASL配置，避免版本兼容问题
	}

	// TLS配置
	if cfg.TLS.Enable {
		newConfig.Net.TLS.Enable = true
	}

	// 生产者配置
	if cfg.RetryMax > 0 {
		newConfig.Producer.Retry.Max = cfg.RetryMax
	} else {
		newConfig.Producer.Retry.Max = 3
	}

	if cfg.RetryBackoff > 0 {
		newConfig.Producer.Retry.Backoff = time.Duration(cfg.RetryBackoff) * time.Millisecond
	}

	newConfig.Producer.Return.Successes = true
	newConfig.Producer.Return.Errors = true

	if cfg.FlushMessages > 0 {
		newConfig.Producer.Flush.Messages = cfg.FlushMessages
	}

	if cfg.FlushBytes > 0 {
		newConfig.Producer.Flush.Bytes = cfg.FlushBytes
	}

	if cfg.FlushTimeout > 0 {
		newConfig.Producer.Flush.Frequency = time.Duration(cfg.FlushTimeout) * time.Millisecond
	}

	// 消费者配置
	newConfig.Consumer.Return.Errors = true
	newConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 创建客户端
	client, err := sarama.NewClient(cfg.Brokers, newConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &Client{
		config: cfg,
		client: client,
		logger: logger.GetLogger(),
	}, nil
}

// CreateProducer 创建生产者
func (c *Client) CreateProducer() error {
	if c.producer != nil {
		return nil
	}

	producer, err := sarama.NewSyncProducerFromClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}

	c.producer = producer
	c.logger.Info("Kafka producer created")
	return nil
}

// CreateConsumer 创建消费者
func (c *Client) CreateConsumer() error {
	if c.consumer != nil {
		return nil
	}

	consumer, err := sarama.NewConsumerGroupFromClient(c.config.Group, c.client)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	c.consumer = consumer
	c.logger.Info("Kafka consumer created")
	return nil
}

// SendMessage 发送消息
func (c *Client) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	if c.producer == nil {
		if err := c.CreateProducer(); err != nil {
			return err
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := c.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Debugf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// SendMessageWithHeaders 发送带头部的消息
func (c *Client) SendMessageWithHeaders(ctx context.Context, topic string, key, value []byte, headers map[string][]byte) error {
	if c.producer == nil {
		if err := c.CreateProducer(); err != nil {
			return err
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	// 添加头部
	for k, v := range headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: v,
		})
	}

	partition, offset, err := c.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message with headers: %w", err)
	}

	c.logger.Debugf("Message with headers sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// ConsumeMessages 消费消息
func (c *Client) ConsumeMessages(ctx context.Context, topics []string, handler MessageHandler) error {
	if c.consumer == nil {
		if err := c.CreateConsumer(); err != nil {
			return err
		}
	}

	consumerHandler := &consumerGroupHandler{
		handler: handler,
		logger:  c.logger,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := c.consumer.Consume(ctx, topics, consumerHandler)
			if err != nil {
				c.logger.Errorf("Error from consumer: %v", err)
				return err
			}
		}
	}
}

// GetTopics 获取主题列表
func (c *Client) GetTopics(ctx context.Context) ([]string, error) {
	topics, err := c.client.Topics()
	if err != nil {
		return nil, fmt.Errorf("failed to get topics: %w", err)
	}
	return topics, nil
}

// GetPartitions 获取主题分区信息
func (c *Client) GetPartitions(ctx context.Context, topic string) ([]int32, error) {
	partitions, err := c.client.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
	}
	return partitions, nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	var errs []error

	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("producer close error: %w", err))
		}
	}

	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("consumer close error: %w", err))
		}
	}

	if c.client != nil {
		if err := c.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("client close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("kafka close errors: %v", errs)
	}

	c.logger.Info("Kafka client closed")
	return nil
}

// consumerGroupHandler 消费者组处理器
type consumerGroupHandler struct {
	handler MessageHandler
	logger  logger.Logger
}

// Setup 设置消费者组
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 清理消费者组
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		msg := &Message{
			Topic:     message.Topic,
			Partition: message.Partition,
			Offset:    message.Offset,
			Key:       message.Key,
			Value:     message.Value,
			Timestamp: message.Timestamp,
			Headers:   make(map[string][]byte),
		}

		// 转换头部信息
		for _, header := range message.Headers {
			msg.Headers[string(header.Key)] = header.Value
		}

		// 处理消息
		if err := h.handler(session.Context(), msg); err != nil {
			h.logger.Errorf("Message handler error: %v", err)
			continue
		}

		// 标记消息已处理
		session.MarkMessage(message, "")
	}

	return nil
}

// ConvertConfig 转换配置格式
func ConvertConfig(cfg *config.KafkaConfig) (*Config, error) {
	result := &Config{
		Brokers:       cfg.Brokers,
		ClientID:      cfg.ClientID,
		Group:         cfg.Group,
		Version:       cfg.Version,
		RetryBackoff:  cfg.RetryBackoff,
		RetryMax:      cfg.RetryMax,
		FlushMessages: cfg.FlushMessages,
		FlushBytes:    cfg.FlushBytes,
		FlushTimeout:  cfg.FlushTimeout,
	}

	result.SASL.Enable = cfg.SASL.Enable
	result.SASL.Mechanism = cfg.SASL.Mechanism
	result.SASL.Username = cfg.SASL.Username
	result.SASL.Password = cfg.SASL.Password

	result.TLS.Enable = cfg.TLS.Enable
	result.TLS.CertFile = cfg.TLS.CertFile
	result.TLS.KeyFile = cfg.TLS.KeyFile
	result.TLS.CAFile = cfg.TLS.CAFile

	return result, nil
}

// 全局Kafka客户端实例
var globalClient *Client

// InitKafka 初始化Kafka客户端
func InitKafka(_ context.Context, cfg *Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return err
	}

	globalClient = client
	logger.GetLogger().Info("Kafka client initialized successfully")
	return nil
}

// GetClient 获取全局Kafka客户端
func GetClient() *Client {
	return globalClient
}
