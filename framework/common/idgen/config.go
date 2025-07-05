package idgen

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Config struct {
	Type     string                 `json:"type" yaml:"type"`
	Database *DatabaseConfig        `json:"database" yaml:"database"`
	Leaf     *LeafConfig            `json:"leaf" yaml:"leaf"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

type DatabaseConfig struct {
	Driver          string        `json:"driver" yaml:"driver"`
	DSN             string        `json:"dsn" yaml:"dsn"`
	Host            string        `json:"host" yaml:"host"`
	Port            int           `json:"port" yaml:"port"`
	Database        string        `json:"database" yaml:"database"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	Charset         string        `json:"charset" yaml:"charset"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	LogLevel        string        `json:"log_level" yaml:"log_level"`
}

func (c *DatabaseConfig) BuildDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	switch c.Driver {
	case "mysql":
		charset := c.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database, charset)
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.Username, c.Password, c.Database)
	case "sqlite", "sqlite3":
		return c.Database
	default:
		return ""
	}
}

func (c *DatabaseConfig) GetLogLevel() logger.LogLevel {
	switch c.LogLevel {
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	case "silent":
		return logger.Silent
	default:
		return logger.Warn
	}
}

func NewIDGeneratorFromConfig(config Config) (IDGenerator, error) {
	if config.Database == nil {
		return nil, fmt.Errorf("database config is required")
	}

	// 创建GORM数据库连接
	db, err := createGormDB(config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if config.Database.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	}
	if config.Database.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	}
	if config.Database.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.Database.ConnMaxLifetime)
	}

	// 创建Leaf配置
	leafConfig := config.Leaf
	if leafConfig == nil {
		leafConfig = DefaultLeafConfig()
	}

	// 从settings中解析配置
	if config.Settings != nil {
		if data, err := json.Marshal(config.Settings); err == nil {
			_ = json.Unmarshal(data, leafConfig)
		}
	}

	switch config.Type {
	case "leaf", "gorm-leaf", "":
		return NewGormLeafIDGenerator(db, leafConfig), nil
	default:
		return nil, fmt.Errorf("unsupported id generator type: %s", config.Type)
	}
}

func createGormDB(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := config.BuildDSN()
	if dsn == "" {
		return nil, fmt.Errorf("failed to build DSN for driver: %s", config.Driver)
	}

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.GetLogLevel()),
	}

	// 根据驱动类型创建数据库连接
	switch config.Driver {
	case "mysql":
		return gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres", "postgresql":
		return gorm.Open(postgres.Open(dsn), gormConfig)
	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Open(dsn), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// DefaultDatabaseConfig 默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:          "mysql",
		Host:            "localhost",
		Port:            3306,
		Database:        "distributed_service",
		Username:        "root",
		Password:        "",
		Charset:         "utf8mb4",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn",
	}
}

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	config Config
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: Config{
			Type:     "gorm-leaf",
			Database: DefaultDatabaseConfig(),
			Leaf:     DefaultLeafConfig(),
		},
	}
}

// WithType 设置生成器类型
func (b *ConfigBuilder) WithType(idType string) *ConfigBuilder {
	b.config.Type = idType
	return b
}

// WithMySQL 设置MySQL数据库
func (b *ConfigBuilder) WithMySQL(host string, port int, database, username, password string) *ConfigBuilder {
	b.config.Database = &DatabaseConfig{
		Driver:          "mysql",
		Host:            host,
		Port:            port,
		Database:        database,
		Username:        username,
		Password:        password,
		Charset:         "utf8mb4",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn",
	}
	return b
}

// WithPostgreSQL 设置PostgreSQL数据库
func (b *ConfigBuilder) WithPostgreSQL(host string, port int, database, username, password string) *ConfigBuilder {
	b.config.Database = &DatabaseConfig{
		Driver:          "postgres",
		Host:            host,
		Port:            port,
		Database:        database,
		Username:        username,
		Password:        password,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn",
	}
	return b
}

// WithSQLite 设置SQLite数据库
func (b *ConfigBuilder) WithSQLite(dbPath string) *ConfigBuilder {
	b.config.Database = &DatabaseConfig{
		Driver:          "sqlite",
		Database:        dbPath,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn",
	}
	return b
}

// WithDSN 直接设置DSN
func (b *ConfigBuilder) WithDSN(driver, dsn string) *ConfigBuilder {
	b.config.Database = &DatabaseConfig{
		Driver:          driver,
		DSN:             dsn,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn",
	}
	return b
}

// WithLeafConfig 设置Leaf配置
func (b *ConfigBuilder) WithLeafConfig(leafConfig *LeafConfig) *ConfigBuilder {
	b.config.Leaf = leafConfig
	return b
}

// WithConnectionPool 设置连接池参数
func (b *ConfigBuilder) WithConnectionPool(maxIdle, maxOpen int, maxLifetime time.Duration) *ConfigBuilder {
	b.config.Database.MaxIdleConns = maxIdle
	b.config.Database.MaxOpenConns = maxOpen
	b.config.Database.ConnMaxLifetime = maxLifetime
	return b
}

// WithLogLevel 设置日志级别
func (b *ConfigBuilder) WithLogLevel(level string) *ConfigBuilder {
	b.config.Database.LogLevel = level
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() Config {
	return b.config
}

// 预设配置

// MySQLConfig MySQL预设配置
func MySQLConfig(host string, port int, database, username, password string) Config {
	return NewConfigBuilder().
		WithMySQL(host, port, database, username, password).
		Build()
}

// PostgreSQLConfig PostgreSQL预设配置
func PostgreSQLConfig(host string, port int, database, username, password string) Config {
	return NewConfigBuilder().
		WithPostgreSQL(host, port, database, username, password).
		Build()
}

// SQLiteConfig SQLite预设配置
func SQLiteConfig(dbPath string) Config {
	return NewConfigBuilder().
		WithSQLite(dbPath).
		Build()
}
