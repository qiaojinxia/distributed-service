# 🚀 分布式服务框架 - 缓存模块

## 概述

缓存模块提供了完整的缓存解决方案，支持内存缓存、Redis缓存和混合缓存，完美集成框架的Redis连接管理。

## ✨ 主要特性

- 🔥 **依赖注入**: 只使用外部注入的Redis客户端，不在缓存模块中创建连接
- 🏗️ **框架集成**: 与分布式服务框架完美集成，支持选项式配置
- 🎛️ **多种缓存类型**: 内存缓存、Redis缓存、混合缓存
- 🧠 **智能淘汰策略**: LRU、TTL、Simple等多种淘汰算法
- 🔧 **简单配置**: 支持配置文件和编程式配置
- 📊 **性能监控**: 内置统计信息和性能监控
- 🔀 **批量操作**: 支持MSet、MGet等批量操作
- 📦 **对象缓存**: 支持JSON序列化的对象缓存
- 🏷️ **命名空间**: 键前缀支持，避免键冲突

## 🚀 快速开始

### 1. 最简单的使用方式

```go
package main

import "github.com/qiaojinxia/distributed-service/framework/core"

func main() {
    // 一行代码启动带缓存的服务
    core.New().
        WithRedis(&config.RedisConfig{
            Host: "localhost",
            Port: 6379,
        }).
        WithCacheDefaults().  // 使用默认缓存配置
        HTTP(func(r interface{}) {
            // 设置路由
        }).
        Run()
}
```

### 2. 选择缓存类型

```go
// 内存缓存 - 快速但容量有限
core.New().WithMemoryCache().Run()

// Redis缓存 - 持久化但有网络延迟  
core.New().
    WithRedis(redisConfig).
    WithRedisCache().
    Run()

// 混合缓存 - 最佳性能和容量平衡
core.New().
    WithRedis(redisConfig).
    WithHybridCache().
    Run()
```

### 3. 自定义配置

```go
cacheConfig := &config.CacheConfig{
    Enabled:         true,
    DefaultType:     "redis",
    UseFramework:    true,
    GlobalKeyPrefix: "myapp",
    DefaultTTL:      "2h",
    Caches: map[string]config.CacheInstance{
        "users": {
            Type:      "redis",
            KeyPrefix: "users",
            TTL:       "6h",
        },
        "sessions": {
            Type:      "memory",
            KeyPrefix: "sessions", 
            TTL:       "30m",
            Settings: map[string]interface{}{
                "max_size": 10000,
                "eviction_policy": "lru",
                "cleanup_interval": "5m",
            },
        },
        "products": {
            Type:      "hybrid",
            KeyPrefix: "products",
            TTL:       "2h",
        },
    },
}

core.New().
    WithRedis(redisConfig).
    WithCache(cacheConfig).
    Run()
```

## 📄 配置文件支持

### config.yaml

```yaml
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 50

cache:
  enabled: true
  default_type: "redis"
  use_framework: true
  global_key_prefix: "myapp"
  default_ttl: "2h"
  caches:
    users:
      type: "redis"
      key_prefix: "users"
      ttl: "6h"
    sessions:
      type: "memory"
      key_prefix: "sessions"
      ttl: "30m"
      settings:
        max_size: 10000
        eviction_policy: "lru"
        cleanup_interval: "5m"
    products:
      type: "hybrid"
      key_prefix: "products"
      ttl: "2h"
```

### 使用配置文件

```go
core.New().
    Config("config/config.yaml").  // 自动读取缓存配置
    Run()
```

## 💻 缓存操作

### 基本操作

```go
// 获取缓存服务
app, _ := core.New().WithCacheDefaults().Build()
cacheService := app.GetComponentManager().GetCacheService()

// 获取特定缓存
userCache, _ := cacheService.GetUserCache()

// 设置和获取
ctx := context.Background()
userCache.Set(ctx, "user:123", "John Doe", time.Hour)
value, _ := userCache.Get(ctx, "user:123")
fmt.Println(value) // "John Doe"
```

### 批量操作

```go
// 批量设置
userData := map[string]interface{}{
    "user:124": "Jane Smith",
    "user:125": "Bob Johnson",
}
userCache.MSet(ctx, userData, time.Hour)

// 批量获取
results, _ := userCache.MGet(ctx, []string{"user:124", "user:125"})
fmt.Println(results) // map[user:124:Jane Smith user:125:Bob Johnson]
```

### 对象缓存

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

user := User{ID: 123, Name: "John", Email: "john@example.com"}

// 设置对象
userCache.SetObject(ctx, "user:object:123", user, time.Hour)

// 获取对象
var retrievedUser User
userCache.GetObject(ctx, "user:object:123", &retrievedUser)
```

### 配置淘汰策略

```go
import "github.com/qiaojinxia/distributed-service/framework/cache"

// LRU策略缓存
lruConfig := cache.MemoryConfig{
    MaxSize:         1000,
    DefaultTTL:      time.Hour,
    CleanupInterval: time.Minute * 10,
    EvictionPolicy:  cache.EvictionPolicyLRU,
}
lruCache, _ := cache.NewMemoryCache(lruConfig)

// TTL策略缓存
ttlConfig := cache.MemoryConfig{
    MaxSize:         500,
    DefaultTTL:      time.Minute * 30,
    CleanupInterval: time.Minute * 5,
    EvictionPolicy:  cache.EvictionPolicyTTL,
}
ttlCache, _ := cache.NewMemoryCache(ttlConfig)

// 使用缓存
lruCache.Set(ctx, "user:123", user, time.Hour)
ttlCache.Set(ctx, "session:abc", session, time.Minute*30)
```

### 统计信息

```go
stats := userCache.GetStats()
fmt.Printf("命中率: %.2f%%\n", 
    float64(stats.Hits)/float64(stats.Hits+stats.Misses)*100)
fmt.Printf("命中: %d, 未命中: %d, 设置: %d\n", 
    stats.Hits, stats.Misses, stats.Sets)
fmt.Printf("淘汰次数: %d\n", stats.Evictions)
```

## 🧠 内存缓存淘汰策略

### 支持的淘汰策略

| 策略 | 库 | 特点 | 适用场景 |
|------|-----|------|----------|
| **LRU** | hashicorp/golang-lru | 最近最少使用，O(1)操作 | 通用缓存，访问有局部性 |
| **TTL** | golang-lru/expirable | 基于过期时间，自动清理 | 有明确过期需求的数据 |
| **Simple** | patrickmn/go-cache | 轻量级，支持TTL | 轻量级缓存需求 |

### 配置示例

```yaml
cache:
  caches:
    # LRU策略 - 推荐用于用户数据
    users:
      type: "memory"
      settings:
        max_size: 1000
        eviction_policy: "lru"
        cleanup_interval: "10m"
    
    # TTL策略 - 推荐用于会话数据
    sessions:
      type: "memory"
      settings:
        max_size: 500
        eviction_policy: "ttl"
        cleanup_interval: "5m"
    
    # Simple策略 - 推荐用于配置数据
    configs:
      type: "memory"
      settings:
        max_size: 100
        eviction_policy: "simple"
        cleanup_interval: "1m"
```

📖 **详细文档**: [缓存淘汰策略文档](./CACHE_EVICTION_POLICIES.md)

## 🎯 缓存类型对比

| 类型 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| 内存缓存 | 极快速度，多种淘汰策略 | 容量限制、数据不持久 | 小量热点数据 |
| Redis缓存 | 大容量、持久化 | 网络延迟 | 大量数据、多实例共享 |
| 混合缓存 | 最佳性能和容量 | 复杂度较高 | 生产环境推荐 |

## 🛠️ 最佳实践

### 1. 键命名规范

```go
// 使用有意义的前缀
"users:123"        // 用户数据
"sessions:abc"     // 会话数据
"products:456"     // 产品数据
"config:settings"  // 配置数据
```

### 2. TTL设置建议

```go
// 不同类型数据的TTL建议
userCache.Set(ctx, "user:123", user, 2*time.Hour)      // 用户数据: 2-6小时
sessionCache.Set(ctx, "session:abc", session, 30*time.Minute) // 会话: 30分钟-2小时
productCache.Set(ctx, "product:456", product, 6*time.Hour)    // 产品: 1-24小时
configCache.Set(ctx, "config:key", config, 24*time.Hour)      // 配置: 24小时+
```

### 3. 错误处理

```go
func GetUser(ctx context.Context, userID string) (*User, error) {
    // 先尝试从缓存获取
    var user User
    err := userCache.GetObject(ctx, "user:"+userID, &user)
    if err == nil {
        return &user, nil
    }
    
    // 缓存未命中，从数据库获取
    user, err = database.GetUser(userID)
    if err != nil {
        return nil, err
    }
    
    // 存入缓存（忽略缓存错误）
    _ = userCache.SetObject(ctx, "user:"+userID, user, 2*time.Hour)
    
    return &user, nil
}
```

### 4. 监控缓存性能

```go
// 定期检查缓存统计
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        stats := userCache.GetStats()
        hitRate := float64(stats.Hits) / float64(stats.Hits + stats.Misses) * 100
        
        if hitRate < 80 {
            log.Warn("Cache hit rate is low", "rate", hitRate)
        }
        
        log.Info("Cache stats", 
            "hits", stats.Hits,
            "misses", stats.Misses, 
            "rate", hitRate)
    }
}()
```

## 🏗️ 架构设计

### 依赖注入模式

```
Framework Redis Client (database.RedisClient)
           ↓
Cache Manager (注入框架Redis客户端)
           ↓
Cache Instances (users, sessions, products...)
```

### 键前缀层次

```
GlobalKeyPrefix:CacheName:Key
     ↓           ↓        ↓
   myapp    :   users  : 123
```

## 🔧 组件管理器访问

```go
// 通过Builder获取组件管理器
app, _ := core.New().WithCacheDefaults().Build()
componentManager := app.GetComponentManager()

// 获取缓存服务
cacheService := componentManager.GetCacheService()

// 创建自定义缓存
cacheService.CreateRedisCache("custom", "custom:prefix")
customCache, _ := cacheService.GetCache("custom")
```

## 📊 性能监控

### 连接池监控

```go
import "github.com/qiaojinxia/distributed-service/framework/cache"

// 监控Redis连接池
optimizer := cache.NewPoolOptimizer(redisClient, cache.DefaultPoolOptimizerConfig())
stats, _ := optimizer.GetPoolStats(ctx)

fmt.Printf("总连接数: %d\n", stats.TotalConns)
fmt.Printf("空闲连接数: %d\n", stats.IdleConns) 
fmt.Printf("平均延迟: %v\n", stats.AvgLatency)
```

### 缓存健康检查

```go
healthChecker := cache.NewPoolHealthChecker(redisClient)
health, _ := healthChecker.CheckHealth(ctx)

fmt.Printf("健康状态: %t\n", health.Healthy)
fmt.Printf("健康评分: %d/100\n", health.Score)
```

## 🚫 注意事项

1. **不要在缓存模块中创建Redis连接** - 只使用注入的客户端
2. **合理设置TTL** - 避免内存泄漏和数据过期
3. **监控命中率** - 命中率低于80%需要优化策略
4. **错误处理** - 缓存失败不应影响业务逻辑
5. **键命名规范** - 使用有意义的前缀避免冲突



## 🤝 Contributing

欢迎提交Issue和Pull Request来改进缓存模块!