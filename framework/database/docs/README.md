# 数据库模块设计文档

## 📋 概述

数据库模块是分布式服务框架的数据持久化核心组件，提供MySQL和Redis的统一访问接口。基于GORM和go-redis构建，支持连接池管理、事务处理、读写分离和多数据源配置。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    应用数据层                            │
│              Application Data Layer                     │
│     Repository | Service | DAO Pattern                 │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  数据库抽象层                            │
│               Database Abstraction Layer               │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   MySQL管理器   │   Redis管理器   │   事务管理器    │ │
│  │ MySQL Manager   │ Redis Manager   │Transaction Mgr  │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   ORM & 客户端层                         │
│                ORM & Client Layer                       │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │     GORM        │   go-redis      │   连接池        │ │
│  │   ORM Engine    │  Redis Client   │Connection Pool  │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   数据库服务器                           │
│                Database Servers                        │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   MySQL主库     │   MySQL从库     │  Redis集群      │ │
│  │ MySQL Master    │ MySQL Slave     │ Redis Cluster   │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 🎯 核心特点

### 1. MySQL支持
- **GORM集成**: 基于GORM v2的完整ORM功能
- **连接池管理**: 自动连接池配置和监控
- **读写分离**: 支持主从数据库配置
- **事务支持**: 完整的事务管理和嵌套事务
- **迁移管理**: 自动数据库结构迁移

### 2. Redis支持
- **go-redis客户端**: 高性能Redis客户端
- **集群支持**: Redis集群和哨兵模式
- **管道操作**: 批量操作优化
- **分布式锁**: 基于Redis的分布式锁实现
- **缓存模式**: 多种缓存模式支持

### 3. 连接管理
- **连接池**: 智能连接池管理
- **健康检查**: 定期连接健康检查
- **故障恢复**: 自动故障检测和恢复
- **监控指标**: 连接池使用率监控

### 4. 事务处理
- **ACID支持**: 完整的事务ACID特性
- **嵌套事务**: 支持嵌套事务和保存点
- **分布式事务**: 跨服务事务协调(计划中)
- **回滚机制**: 自动和手动回滚支持

## 🚀 使用示例

### MySQL基础操作

```go
package main

import (
    "time"
    "github.com/qiaojinxia/distributed-service/framework/database"
    "gorm.io/gorm"
)

// 用户模型
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"uniqueIndex;size:100"`
    Age       int       
    CreatedAt time.Time
    UpdatedAt time.Time
}

func main() {
    // 初始化数据库
    config := database.MySQLConfig{
        Host:     "localhost",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
        
        // 连接池配置
        MaxOpenConns: 100,
        MaxIdleConns: 10,
        MaxLifetime:  time.Hour,
    }
    
    db, err := database.NewMySQL(config)
    if err != nil {
        panic(err)
    }
    
    // 自动迁移
    db.AutoMigrate(&User{})
    
    // 创建用户
    user := User{
        Name:  "张三",
        Email: "zhangsan@example.com",
        Age:   25,
    }
    
    result := db.Create(&user)
    if result.Error != nil {
        panic(result.Error)
    }
    
    // 查询用户
    var foundUser User
    db.First(&foundUser, user.ID)
    
    // 更新用户
    db.Model(&foundUser).Update("Age", 26)
    
    // 删除用户
    db.Delete(&foundUser)
}
```

### Redis基础操作

```go
package main

import (
    "context"
    "time"
    "github.com/qiaojinxia/distributed-service/framework/database"
)

func main() {
    // 初始化Redis
    config := database.RedisConfig{
        Host:     "localhost",
        Port:     6379,
        Password: "",
        DB:       0,
        
        // 连接池配置
        PoolSize:     10,
        MinIdleConns: 5,
        MaxRetries:   3,
    }
    
    rdb, err := database.NewRedis(config)
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // 字符串操作
    err = rdb.Set(ctx, "user:123", "张三", time.Hour).Err()
    if err != nil {
        panic(err)
    }
    
    val, err := rdb.Get(ctx, "user:123").Result()
    if err != nil {
        panic(err)
    }
    fmt.Println("用户名:", val)
    
    // 哈希操作
    err = rdb.HSet(ctx, "user:123:profile", map[string]interface{}{
        "name": "张三",
        "age":  25,
        "city": "北京",
    }).Err()
    
    profile := rdb.HGetAll(ctx, "user:123:profile").Val()
    fmt.Printf("用户资料: %+v\n", profile)
    
    // 列表操作
    rdb.LPush(ctx, "user:123:orders", "order1", "order2", "order3")
    orders := rdb.LRange(ctx, "user:123:orders", 0, -1).Val()
    fmt.Printf("用户订单: %+v\n", orders)
}
```

### 事务处理

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/database"
    "gorm.io/gorm"
)

type Account struct {
    ID      uint
    UserID  uint
    Balance float64
}

type Transaction struct {
    ID       uint
    FromID   uint
    ToID     uint
    Amount   float64
    Status   string
}

func transferMoney(db *gorm.DB, fromUserID, toUserID uint, amount float64) error {
    // 开始事务
    return db.Transaction(func(tx *gorm.DB) error {
        // 检查发送方余额
        var fromAccount Account
        if err := tx.Where("user_id = ?", fromUserID).First(&fromAccount).Error; err != nil {
            return err
        }
        
        if fromAccount.Balance < amount {
            return errors.New("余额不足")
        }
        
        // 扣款
        if err := tx.Model(&fromAccount).Update("balance", fromAccount.Balance-amount).Error; err != nil {
            return err
        }
        
        // 加款
        if err := tx.Model(&Account{}).Where("user_id = ?", toUserID).
            Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }
        
        // 记录交易
        transaction := Transaction{
            FromID: fromUserID,
            ToID:   toUserID,
            Amount: amount,
            Status: "completed",
        }
        
        if err := tx.Create(&transaction).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### 读写分离配置

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/database"
    "gorm.io/gorm"
)

func setupMasterSlave() {
    // 主库配置
    masterConfig := database.MySQLConfig{
        Host:     "mysql-master.example.com",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
        Role:     "master", // 标记为主库
    }
    
    // 从库配置
    slaveConfig := database.MySQLConfig{
        Host:     "mysql-slave.example.com", 
        Port:     3306,
        Database: "myapp",
        Username: "readonly",
        Password: "password",
        Role:     "slave", // 标记为从库
    }
    
    // 初始化主从数据库
    dbManager, err := database.NewMasterSlaveDB(masterConfig, slaveConfig)
    if err != nil {
        panic(err)
    }
    
    // 写操作自动路由到主库
    user := User{Name: "张三", Email: "zhangsan@example.com"}
    dbManager.Create(&user)
    
    // 读操作自动路由到从库
    var users []User
    dbManager.Find(&users)
}
```

### Redis集群配置

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/database"
    "github.com/go-redis/redis/v8"
)

func setupRedisCluster() {
    // 集群配置
    config := database.RedisClusterConfig{
        Addrs: []string{
            "redis-1.example.com:7000",
            "redis-2.example.com:7001", 
            "redis-3.example.com:7002",
        },
        Password: "cluster-password",
        
        // 集群选项
        MaxRedirects:   8,
        ReadOnly:       false,
        RouteByLatency: true,
        RouteRandomly:  true,
    }
    
    cluster, err := database.NewRedisCluster(config)
    if err != nil {
        panic(err)
    }
    
    // 使用集群客户端
    ctx := context.Background()
    cluster.Set(ctx, "key", "value", 0)
    val := cluster.Get(ctx, "key").Val()
}
```

## 🔧 配置选项

### MySQL配置

```go
type MySQLConfig struct {
    // 连接信息
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Charset  string `yaml:"charset"`
    
    // 连接池配置
    MaxOpenConns int           `yaml:"max_open_conns"` // 最大连接数
    MaxIdleConns int           `yaml:"max_idle_conns"` // 最大空闲连接数
    MaxLifetime  time.Duration `yaml:"max_lifetime"`   // 连接最大生存时间
    
    // 性能配置
    SlowThreshold time.Duration `yaml:"slow_threshold"` // 慢查询阈值
    LogLevel      string        `yaml:"log_level"`      // 日志级别
    
    // SSL配置
    SSLMode string `yaml:"ssl_mode"` // SSL模式
    SSLCert string `yaml:"ssl_cert"` // SSL证书
    SSLKey  string `yaml:"ssl_key"`  // SSL密钥
    
    // 其他选项
    Timezone string `yaml:"timezone"` // 时区
    ParseTime bool  `yaml:"parse_time"` // 解析时间
}
```

### Redis配置

```go
type RedisConfig struct {
    // 连接信息
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
    
    // 连接池配置
    PoolSize     int           `yaml:"pool_size"`      // 连接池大小
    MinIdleConns int           `yaml:"min_idle_conns"` // 最小空闲连接
    MaxConnAge   time.Duration `yaml:"max_conn_age"`   // 连接最大年龄
    PoolTimeout  time.Duration `yaml:"pool_timeout"`   // 获取连接超时
    IdleTimeout  time.Duration `yaml:"idle_timeout"`   // 空闲连接超时
    
    // 重试配置
    MaxRetries      int           `yaml:"max_retries"`       // 最大重试次数
    MinRetryBackoff time.Duration `yaml:"min_retry_backoff"` // 最小重试间隔
    MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"` // 最大重试间隔
    
    // 超时配置
    DialTimeout  time.Duration `yaml:"dial_timeout"`  // 连接超时
    ReadTimeout  time.Duration `yaml:"read_timeout"`  // 读取超时
    WriteTimeout time.Duration `yaml:"write_timeout"` // 写入超时
}
```

### 配置文件示例

```yaml
# config/database.yaml
database:
  mysql:
    host: "localhost"
    port: 3306
    database: "myapp"
    username: "root"
    password: "password"
    charset: "utf8mb4"
    
    # 连接池
    max_open_conns: 100
    max_idle_conns: 10
    max_lifetime: "1h"
    
    # 性能监控
    slow_threshold: "200ms"
    log_level: "warn"
    
    # 时区设置
    timezone: "Asia/Shanghai"
    parse_time: true
    
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    
    # 连接池
    pool_size: 10
    min_idle_conns: 5
    max_conn_age: "30m"
    pool_timeout: "4s"
    idle_timeout: "5m"
    
    # 重试配置
    max_retries: 3
    min_retry_backoff: "8ms"
    max_retry_backoff: "512ms"
    
    # 超时配置
    dial_timeout: "5s"
    read_timeout: "3s"
    write_timeout: "3s"
```

## 📊 监控与指标

### 连接池监控

```go
// MySQL连接池状态
func getDBStats(db *gorm.DB) map[string]interface{} {
    sqlDB, _ := db.DB()
    stats := sqlDB.Stats()
    
    return map[string]interface{}{
        "open_connections":     stats.OpenConnections,
        "in_use":              stats.InUse,
        "idle":                stats.Idle,
        "wait_count":          stats.WaitCount,
        "wait_duration":       stats.WaitDuration,
        "max_idle_closed":     stats.MaxIdleClosed,
        "max_lifetime_closed": stats.MaxLifetimeClosed,
    }
}

// Redis连接池状态
func getRedisStats(rdb *redis.Client) map[string]interface{} {
    stats := rdb.PoolStats()
    
    return map[string]interface{}{
        "hits":         stats.Hits,
        "misses":       stats.Misses,
        "timeouts":     stats.Timeouts,
        "total_conns":  stats.TotalConns,
        "idle_conns":   stats.IdleConns,
        "stale_conns":  stats.StaleConns,
    }
}
```

### 性能指标

```go
// 查询性能统计
type QueryStats struct {
    SlowQueries    int64         `json:"slow_queries"`
    TotalQueries   int64         `json:"total_queries"`
    AvgQueryTime   time.Duration `json:"avg_query_time"`
    MaxQueryTime   time.Duration `json:"max_query_time"`
    ErrorCount     int64         `json:"error_count"`
}

// 缓存命中率
type CacheStats struct {
    HitCount   int64   `json:"hit_count"`
    MissCount  int64   `json:"miss_count"`
    HitRate    float64 `json:"hit_rate"`
    SetCount   int64   `json:"set_count"`
    DelCount   int64   `json:"del_count"`
}
```

## 🔍 最佳实践

### 1. 连接池优化

```go
// ✅ 推荐配置
config := database.MySQLConfig{
    MaxOpenConns: 100,                    // 根据并发量调整
    MaxIdleConns: 10,                     // 通常为MaxOpenConns的10%
    MaxLifetime:  time.Hour,              // 避免长连接问题
}

// ✅ 监控连接池使用率
go func() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        stats := getDBStats(db)
        if stats["in_use"].(int) > stats["max_open_conns"].(int)*0.8 {
            logger.Warn("数据库连接池使用率过高", logger.Any("stats", stats))
        }
    }
}()
```

### 2. 事务最佳实践

```go
// ✅ 推荐：使用闭包事务
func transferMoney(db *gorm.DB, fromID, toID uint, amount float64) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 事务逻辑
        return nil
    })
}

// ✅ 推荐：手动事务控制
func complexOperation(db *gorm.DB) error {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    if err := tx.Error; err != nil {
        return err
    }
    
    // 业务操作...
    if err := businessLogic(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}
```

### 3. 查询优化

```go
// ✅ 推荐：使用索引
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex"`           // 唯一索引
    Name  string `gorm:"index"`                 // 普通索引
    Age   int    `gorm:"index:idx_age_city"`    // 组合索引
    City  string `gorm:"index:idx_age_city"`    // 组合索引
}

// ✅ 推荐：预加载关联
var users []User
db.Preload("Orders").Preload("Profile").Find(&users)

// ✅ 推荐：分页查询
var users []User
db.Limit(20).Offset(100).Find(&users)

// ❌ 避免：N+1查询问题
for _, user := range users {
    // 这会产生N+1查询
    var orders []Order
    db.Where("user_id = ?", user.ID).Find(&orders)
}
```

### 4. Redis使用模式

```go
// ✅ 推荐：管道操作
pipe := rdb.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Set(ctx, "key3", "value3", 0)
_, err := pipe.Exec(ctx)

// ✅ 推荐：使用连接池
// 避免频繁创建Redis连接
var redisClient *redis.Client

func init() {
    redisClient = database.NewRedis(config)
}

// ✅ 推荐：设置合理的过期时间
rdb.Set(ctx, "session:"+sessionID, sessionData, time.Hour*24)

// ❌ 避免：大Key
// 避免存储过大的值（> 1MB）
```

## 🚨 故障排查

### 常见问题

**Q1: 连接池耗尽**
```go
// 检查连接是否正确关闭
rows, err := db.Raw("SELECT * FROM users").Rows()
if err != nil {
    return err
}
defer rows.Close() // 必须关闭

// 检查长事务
db.Set("gorm:query_option", "SET innodb_lock_wait_timeout = 5")
```

**Q2: 慢查询问题**
```go
// 启用慢查询日志
db.Logger = logger.Default.LogMode(logger.Info)

// 分析执行计划
db.Raw("EXPLAIN SELECT * FROM users WHERE email = ?", email).Scan(&result)
```

**Q3: Redis连接问题**
```go
// 检查Redis连接
pong, err := rdb.Ping(ctx).Result()
if err != nil {
    logger.Error("Redis连接失败", logger.Err(err))
}

// 连接池监控
stats := rdb.PoolStats()
if stats.Timeouts > 0 {
    logger.Warn("Redis连接池超时", logger.Any("stats", stats))
}
```

## 🔮 高级功能

### 分库分表支持

```go
// 分表路由
func getTableName(userID uint) string {
    return fmt.Sprintf("users_%d", userID%10)
}

// 动态表名
type User struct {
    ID   uint
    Name string
}

func (User) TableName() string {
    // 根据上下文动态决定表名
    return "users_" + getCurrentShardSuffix()
}
```

### 读写分离中间件

```go
// 读写分离中间件
func ReadWriteSplitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" {
            // 读操作使用从库
            c.Set("db", slaveDB)
        } else {
            // 写操作使用主库
            c.Set("db", masterDB)
        }
        c.Next()
    }
}
```

---

> 数据库模块为框架提供了可靠的数据持久化能力，支持企业级应用的各种数据访问需求。