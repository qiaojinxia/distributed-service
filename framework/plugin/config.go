package plugin

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// SimpleConfig 简单的插件配置实现
type SimpleConfig struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSimpleConfig 创建简单配置
func NewSimpleConfig(data map[string]interface{}) *SimpleConfig {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &SimpleConfig{
		data: data,
	}
}

// Get 获取配置值
func (c *SimpleConfig) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// GetString 获取字符串配置
func (c *SimpleConfig) GetString(key string) string {
	if value := c.Get(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// GetInt 获取整数配置
func (c *SimpleConfig) GetInt(key string) int {
	if value := c.Get(key); value != nil {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// GetBool 获取布尔配置
func (c *SimpleConfig) GetBool(key string) bool {
	if value := c.Get(key); value != nil {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		case int:
			return v != 0
		}
	}
	return false
}

// GetDuration 获取时间间隔配置
func (c *SimpleConfig) GetDuration(key string) time.Duration {
	if value := c.Get(key); value != nil {
		switch v := value.(type) {
		case time.Duration:
			return v
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return d
			}
		case int:
			return time.Duration(v) * time.Second
		case int64:
			return time.Duration(v) * time.Second
		case float64:
			return time.Duration(v) * time.Second
		}
	}
	return 0
}

// Set 设置配置值
func (c *SimpleConfig) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// All 获取所有配置
func (c *SimpleConfig) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// ConfigProvider 配置提供者实现
type DefaultConfigProvider struct {
	configs map[string]Config
	mu      sync.RWMutex
}

// NewDefaultConfigProvider 创建默认配置提供者
func NewDefaultConfigProvider() *DefaultConfigProvider {
	return &DefaultConfigProvider{
		configs: make(map[string]Config),
	}
}

// GetPluginConfig 获取插件配置
func (p *DefaultConfigProvider) GetPluginConfig(pluginName string) Config {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if config, exists := p.configs[pluginName]; exists {
		return config
	}

	// 返回空配置
	return NewSimpleConfig(nil)
}

// SetPluginConfig 设置插件配置
func (p *DefaultConfigProvider) SetPluginConfig(pluginName string, config Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.configs[pluginName] = config
	return nil
}

// LoadConfig 加载配置文件
func (p *DefaultConfigProvider) LoadConfig(path string) error {
	// TODO: 实现从文件加载配置
	return nil
}

// SaveConfig 保存配置到文件
func (p *DefaultConfigProvider) SaveConfig(path string) error {
	// TODO: 实现保存配置到文件
	return nil
}

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	data map[string]interface{}
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		data: make(map[string]interface{}),
	}
}

// Set 设置配置项
func (b *ConfigBuilder) Set(key string, value interface{}) *ConfigBuilder {
	b.data[key] = value
	return b
}

// SetString 设置字符串配置
func (b *ConfigBuilder) SetString(key, value string) *ConfigBuilder {
	return b.Set(key, value)
}

// SetInt 设置整数配置
func (b *ConfigBuilder) SetInt(key string, value int) *ConfigBuilder {
	return b.Set(key, value)
}

// SetBool 设置布尔配置
func (b *ConfigBuilder) SetBool(key string, value bool) *ConfigBuilder {
	return b.Set(key, value)
}

// SetDuration 设置时间间隔配置
func (b *ConfigBuilder) SetDuration(key string, value time.Duration) *ConfigBuilder {
	return b.Set(key, value)
}

// Build 构建配置
func (b *ConfigBuilder) Build() Config {
	return NewSimpleConfig(b.data)
}
