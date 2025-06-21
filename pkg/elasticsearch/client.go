package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"distributed-service/framework/config"
	"distributed-service/framework/logger"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client Elasticsearch客户端
type Client struct {
	client *elasticsearch.Client
	config *Config
	logger logger.Logger
}

// Config Elasticsearch配置
type Config struct {
	Addresses []string `yaml:"addresses" json:"addresses"`
	Username  string   `yaml:"username" json:"username"`
	Password  string   `yaml:"password" json:"password"`
	CACert    string   `yaml:"ca_cert" json:"ca_cert"`
	Timeout   int      `yaml:"timeout" json:"timeout"` // 秒
}

// NewClient 创建Elasticsearch客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("elasticsearch config is required")
	}

	// 创建ES配置
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	if cfg.Timeout > 0 {
		esCfg.Transport = &http.Transport{
			ResponseHeaderTimeout: time.Duration(cfg.Timeout) * time.Second,
		}
	}

	// 创建ES客户端
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
		logger: logger.GetLogger(),
	}, nil
}

// Ping 检查连接
func (c *Client) Ping(ctx context.Context) error {
	req := esapi.PingRequest{}
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", res.Status())
	}

	return nil
}

// Index 索引文档
func (c *Client) Index(ctx context.Context, index, docID string, doc interface{}) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.Status())
	}

	return nil
}

// Search 搜索文档
func (c *Client) Search(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(data)),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// Delete 删除文档
func (c *Client) Delete(ctx context.Context, index, docID string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("failed to delete document: %s", res.Status())
	}

	return nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	var body string
	if mapping != nil {
		data, err := json.Marshal(mapping)
		if err != nil {
			return fmt.Errorf("failed to marshal mapping: %w", err)
		}
		body = string(data)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(body),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	return nil
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("failed to delete index: %s", res.Status())
	}

	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	// Elasticsearch客户端不需要显式关闭
	c.logger.Info("Elasticsearch client closed")
	return nil
}

// ConvertConfig 转换配置格式
func ConvertConfig(cfg *config.ElasticsearchConfig) (*Config, error) {
	return &Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		CACert:    cfg.CACert,
		Timeout:   cfg.Timeout,
	}, nil
}

// 全局Elasticsearch客户端实例
var globalClient *Client

// InitElasticsearch 初始化Elasticsearch客户端
func InitElasticsearch(ctx context.Context, cfg *Config) error {
	client, err := NewClient(cfg)
	if err != nil {
		return err
	}

	// 测试连接
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("elasticsearch connection test failed: %w", err)
	}

	globalClient = client
	logger.GetLogger().Info("Elasticsearch client initialized successfully")
	return nil
}

// GetClient 获取全局Elasticsearch客户端
func GetClient() *Client {
	return globalClient
}
