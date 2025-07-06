# 缓存模块设计文档

## 📋 概述

缓存模块是分布式服务框架的核心组件之一，提供了统一的缓存接口和多种实现策略。支持内存缓存和Redis缓存，具有丰富的淘汰策略和灵活的配置选项。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    框架全局API                           │
│  core.GetUserCache() | GetSessionCache() | GetCache()   │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                框架缓存服务                              │
│            FrameworkCacheService                        │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   用户缓存      │    会话缓存     │   产品缓存      │ │
│  │   (LRU策略)     │   (TTL策略)     │  (Simple策略)   │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  缓存接口层                              │
│                 cache.Cache                             │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   内存缓存      │    Redis缓存    │   未来扩展      │ │
│  │ MemoryCache     │  RedisCache     │      ...        │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 核心组件

#### 1. 缓存接口 (Cache Interface)
```go
type Cache interface {
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Get(ctx context.Context, key string) (interface{}, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Clear(ctx context.Context) error
}
```

#### 2. 内存缓存实现 (MemoryCache)
- 支持三种淘汰策略：LRU、TTL、Simple
- 基于成熟的第三方库实现
- 支持自定义TTL和默认TTL

#### 3. Redis缓存实现 (RedisCache)
- 基于Redis的分布式缓存
- 支持TTL过期机制
- 适用于多实例共享数据

#### 4. 框架缓存服务 (FrameworkCacheService)
- 统一管理多个缓存实例
- 自动初始化默认缓存
- 提供缓存生命周期管理

## 🎯 核心特点

### 1. 多策略支持
- **LRU策略**: 最近最少使用淘汰，适用于有限内存环境
- **TTL策略**: 基于时间过期，适用于临时数据存储
- **Simple策略**: 简单存储，适用于配置缓存

### 2. 灵活配置
```go
type MemoryConfig struct {
    MaxSize         int           // 最大缓存条目数
    DefaultTTL      time.Duration // 默认过期时间
    CleanupInterval time.Duration // 清理间隔
    EvictionPolicy  EvictionPolicy // 淘汰策略
}
```

### 3. 框架集成
- 全局API访问：`core.GetUserCache()`
- 自动初始化：框架启动时自动创建缓存实例
- 优雅降级：Redis不可用时自动使用内存缓存

### 4. 类型安全
- 泛型支持（未来版本）
- 接口统一
- 错误处理完善

## 🚀 使用示例

### 基础使用

#### 1. 直接创建缓存实例
```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/qiaojinxia/distributed-service/framework/cache"
)

func main() {
    // 创建LRU缓存
    config := cache.MemoryConfig{
        MaxSize:        1000,
        DefaultTTL:     time.Hour,
        EvictionPolicy: cache.EvictionPolicyLRU,
    }
    
    lruCache, err := cache.NewMemoryCache(config)
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // 存储数据
    err = lruCache.Set(ctx, "user:123", map[string]interface{}{
        "name": "张三",
        "age":  25,
    }, time.Minute*30)
    
    // 获取数据
    data, err := lruCache.Get(ctx, "user:123")
    if err != nil {
        fmt.Printf("获取失败: %v\n", err)
        return
    }
    
    fmt.Printf("用户信息: %+v\n", data)
}
```

#### 2. 框架集成使用
```go
package main

import (
    "context"
    "time"
    
    "github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
    // 启动框架
    go func() {
        core.New().
            Port(8080).
            Name("cache-example").
            Run()
    }()
    
    // 等待框架初始化
    time.Sleep(time.Second * 2)
    
    ctx := context.Background()
    
    // 使用用户缓存 (LRU策略)
    userCache := core.GetUserCache()
    userCache.Set(ctx, "user:123", UserInfo{
        ID:   123,
        Name: "张三",
        Role: "admin",
    }, time.Hour)
    
    // 使用会话缓存 (TTL策略)
    sessionCache := core.GetSessionCache()
    sessionCache.Set(ctx, "session:abc123", SessionData{
        UserID:    123,
        LoginTime: time.Now(),
        IP:        "192.168.1.100",
    }, time.Minute*30)
    
    // 使用产品缓存 (Simple策略)
    productCache := core.GetProductCache()
    productCache.Set(ctx, "product:hot", []string{
        "iPhone 15", "MacBook Pro", "iPad Air",
    }, time.Hour*6)
}

type UserInfo struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Role string `json:"role"`
}

type SessionData struct {
    UserID    int       `json:"user_id"`
    LoginTime time.Time `json:"login_time"`
    IP        string    `json:"ip"`
}
```

### 高级使用场景

#### 1. 电商用户缓存系统
```go
package examples

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/qiaojinxia/distributed-service/framework/core"
)

// 用户服务
type UserService struct{}

func (s *UserService) GetUser(ctx context.Context, userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    
    // 先从缓存获取
    userCache := core.GetUserCache()
    if data, err := userCache.Get(ctx, cacheKey); err == nil {
        var user User
        if jsonStr, ok := data.(string); ok {
            json.Unmarshal([]byte(jsonStr), &user)
            return &user, nil
        }
    }
    
    // 缓存未命中，从数据库获取
    user := s.getUserFromDB(userID)
    
    // 存入缓存
    userJSON, _ := json.Marshal(user)
    userCache.Set(ctx, cacheKey, string(userJSON), time.Hour*2)
    
    return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // 更新数据库
    err := s.updateUserToDB(user)
    if err != nil {
        return err
    }
    
    // 删除缓存，确保数据一致性
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    userCache := core.GetUserCache()
    userCache.Delete(ctx, cacheKey)
    
    return nil
}

type User struct {
    ID       int       `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    CreateAt time.Time `json:"create_at"`
}

func (s *UserService) getUserFromDB(userID int) *User {
    // 模拟数据库查询
    return &User{
        ID:       userID,
        Name:     "用户" + fmt.Sprint(userID),
        Email:    fmt.Sprintf("user%d@example.com", userID),
        CreateAt: time.Now(),
    }
}

func (s *UserService) updateUserToDB(user *User) error {
    // 模拟数据库更新
    return nil
}
```

#### 2. 会话管理系统
```go
package examples

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "time"
    
    "github.com/qiaojinxia/distributed-service/framework/core"
)

// 会话管理器
type SessionManager struct{}

func (sm *SessionManager) CreateSession(ctx context.Context, userID int) (string, error) {
    // 生成会话ID
    sessionID := sm.generateSessionID()
    
    // 创建会话数据
    session := Session{
        ID:        sessionID,
        UserID:    userID,
        CreatedAt: time.Now(),
        LastAccess: time.Now(),
    }
    
    // 存储到会话缓存 (TTL策略，30分钟过期)
    sessionCache := core.GetSessionCache()
    err := sessionCache.Set(ctx, sessionID, session, time.Minute*30)
    
    return sessionID, err
}

func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
    sessionCache := core.GetSessionCache()
    
    data, err := sessionCache.Get(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    session, ok := data.(Session)
    if !ok {
        return nil, fmt.Errorf("会话数据格式错误")
    }
    
    // 更新最后访问时间
    session.LastAccess = time.Now()
    sessionCache.Set(ctx, sessionID, session, time.Minute*30)
    
    return &session, nil
}

func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
    sessionCache := core.GetSessionCache()
    return sessionCache.Delete(ctx, sessionID)
}

func (sm *SessionManager) generateSessionID() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

type Session struct {
    ID         string    `json:"id"`
    UserID     int       `json:"user_id"`
    CreatedAt  time.Time `json:"created_at"`
    LastAccess time.Time `json:"last_access"`
}
```

#### 3. 商品推荐缓存
```go
package examples

import (
    "context"
    "fmt"
    "time"
    
    "github.com/qiaojinxia/distributed-service/framework/core"
)

// 推荐服务
type RecommendService struct{}

func (rs *RecommendService) GetHotProducts(ctx context.Context, category string) ([]Product, error) {
    cacheKey := fmt.Sprintf("hot_products:%s", category)
    
    // 从产品缓存获取
    productCache := core.GetProductCache()
    if data, err := productCache.Get(ctx, cacheKey); err == nil {
        if products, ok := data.([]Product); ok {
            return products, nil
        }
    }
    
    // 计算热门商品
    products := rs.calculateHotProducts(category)
    
    // 缓存6小时
    productCache.Set(ctx, cacheKey, products, time.Hour*6)
    
    return products, nil
}

func (rs *RecommendService) GetUserRecommendations(ctx context.Context, userID int) ([]Product, error) {
    cacheKey := fmt.Sprintf("user_recommend:%d", userID)
    
    // 用户个性化推荐缓存2小时
    userCache := core.GetUserCache()
    if data, err := userCache.Get(ctx, cacheKey); err == nil {
        if products, ok := data.([]Product); ok {
            return products, nil
        }
    }
    
    // 生成个性化推荐
    products := rs.generateUserRecommendations(userID)
    
    userCache.Set(ctx, cacheKey, products, time.Hour*2)
    
    return products, nil
}

func (rs *RecommendService) calculateHotProducts(category string) []Product {
    // 模拟热门商品计算
    return []Product{
        {ID: 1, Name: "热门商品1", Category: category, Price: 199.99},
        {ID: 2, Name: "热门商品2", Category: category, Price: 299.99},
        {ID: 3, Name: "热门商品3", Category: category, Price: 399.99},
    }
}

func (rs *RecommendService) generateUserRecommendations(userID int) []Product {
    // 模拟个性化推荐算法
    return []Product{
        {ID: 100 + userID, Name: fmt.Sprintf("推荐商品%d", userID), Price: 159.99},
        {ID: 200 + userID, Name: fmt.Sprintf("定制商品%d", userID), Price: 259.99},
    }
}

type Product struct {
    ID       int     `json:"id"`
    Name     string  `json:"name"`
    Category string  `json:"category"`
    Price    float64 `json:"price"`
}
```

## 🔧 配置与优化

### 默认配置
```go
// 用户缓存 - LRU策略
userConfig := MemoryConfig{
    MaxSize:         1000,
    DefaultTTL:      time.Hour * 2,
    CleanupInterval: time.Minute * 10,
    EvictionPolicy:  EvictionPolicyLRU,
}

// 会话缓存 - TTL策略
sessionConfig := MemoryConfig{
    MaxSize:         5000,
    DefaultTTL:      time.Minute * 30,
    CleanupInterval: time.Minute * 5,
    EvictionPolicy:  EvictionPolicyTTL,
}

// 产品缓存 - Simple策略
productConfig := MemoryConfig{
    MaxSize:         500,
    DefaultTTL:      time.Hour * 6,
    CleanupInterval: time.Minute * 30,
    EvictionPolicy:  EvictionPolicySimple,
}
```

### 性能优化建议

1. **合理设置MaxSize**: 根据内存容量和数据大小调整
2. **选择合适的淘汰策略**: 
   - 用户数据使用LRU
   - 临时数据使用TTL
   - 配置数据使用Simple
3. **调整清理间隔**: 平衡内存使用和性能开销
4. **监控缓存命中率**: 及时调整缓存策略

## 🧪 测试与验证

### 运行测试
```bash
# 运行所有缓存测试
go test ./tests/ -v

# 运行特定测试
go test ./tests/ -v -run TestCacheIntegration
go test ./tests/ -v -run TestTTLBehavior
go test ./tests/ -v -run TestCachePolicies
```

### 测试覆盖
- ✅ 框架集成测试
- ✅ TTL过期行为测试
- ✅ 三种淘汰策略测试
- ✅ 并发安全测试
- ✅ 错误处理测试

## 🔮 未来规划

### 短期计划
- [ ] 添加缓存监控指标
- [ ] 实现缓存预热机制
- [ ] 支持缓存分层

### 长期计划
- [ ] 泛型支持
- [ ] 分布式缓存一致性
- [ ] 自适应淘汰策略
- [ ] 缓存压缩

## 📚 相关文档

- [缓存接口设计](./cache.go)
- [内存缓存实现](../cache_memory.go)
- [Redis缓存实现](../cache_redis.go)
- [框架服务集成](../framework_service.go)
- [测试用例](../tests/)

---

> 此文档持续更新中，如有问题请参考源码或提交Issue。