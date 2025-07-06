# 缓存模块快速开始

## 🚀 5分钟上手指南

### 步骤1: 基础使用

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
        MaxSize:        100,
        DefaultTTL:     time.Hour,
        EvictionPolicy: cache.EvictionPolicyLRU,
    }
    
    myCache, _ := cache.NewMemoryCache(config)
    ctx := context.Background()
    
    // 存储和获取数据
    myCache.Set(ctx, "user:123", "张三", time.Minute*30)
    value, _ := myCache.Get(ctx, "user:123")
    fmt.Printf("用户: %v\n", value) // 输出: 用户: 张三
}
```

### 步骤2: 框架集成

```go
package main

import (
    "context"
    "time"
    
    "github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
    // 启动框架
    go core.New().Port(8080).Run()
    time.Sleep(time.Second * 2) // 等待初始化
    
    ctx := context.Background()
    
    // 直接使用全局缓存API
    userCache := core.GetUserCache()
    userCache.Set(ctx, "user:456", "李四", time.Hour)
    
    sessionCache := core.GetSessionCache()  
    sessionCache.Set(ctx, "session:abc", "会话数据", time.Minute*30)
    
    productCache := core.GetProductCache()
    productCache.Set(ctx, "hot_products", []string{"iPhone", "iPad"}, time.Hour*6)
}
```

## 📋 常用模式

### 模式1: 用户缓存服务

```go
type UserService struct{}

func (s *UserService) GetUser(ctx context.Context, userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    userCache := core.GetUserCache()
    
    // 尝试从缓存获取
    if data, err := userCache.Get(ctx, cacheKey); err == nil {
        if user, ok := data.(User); ok {
            return &user, nil
        }
    }
    
    // 缓存未命中，从数据库加载
    user := s.loadFromDB(userID)
    
    // 存入缓存
    userCache.Set(ctx, cacheKey, *user, time.Hour*2)
    return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // 更新数据库
    s.updateDB(user)
    
    // 删除缓存保证一致性
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    core.GetUserCache().Delete(ctx, cacheKey)
    return nil
}
```

### 模式2: 会话管理

```go
type SessionManager struct{}

func (sm *SessionManager) CreateSession(userID int) string {
    sessionID := generateSessionID()
    session := Session{
        ID:     sessionID,
        UserID: userID,
        CreatedAt: time.Now(),
    }
    
    ctx := context.Background()
    sessionCache := core.GetSessionCache()
    sessionCache.Set(ctx, sessionID, session, time.Minute*30)
    
    return sessionID
}

func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
    ctx := context.Background()
    sessionCache := core.GetSessionCache()
    
    data, err := sessionCache.Get(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("会话已过期")
    }
    
    session := data.(Session)
    return &session, nil
}
```

### 模式3: 商品缓存

```go
type ProductService struct{}

func (ps *ProductService) GetHotProducts() ([]Product, error) {
    ctx := context.Background()
    productCache := core.GetProductCache()
    
    // 检查缓存
    if data, err := productCache.Get(ctx, "hot_products"); err == nil {
        if products, ok := data.([]Product); ok {
            return products, nil
        }
    }
    
    // 计算热门商品
    products := ps.calculateHotProducts()
    
    // 缓存6小时
    productCache.Set(ctx, "hot_products", products, time.Hour*6)
    return products, nil
}
```

## ⚡ 性能优化技巧

### 技巧1: 选择合适的淘汰策略

```go
// 用户数据 - 使用LRU
userConfig := cache.MemoryConfig{
    EvictionPolicy: cache.EvictionPolicyLRU,  // 热点用户常驻内存
    MaxSize: 1000,
}

// 临时数据 - 使用TTL
tempConfig := cache.MemoryConfig{
    EvictionPolicy: cache.EvictionPolicyTTL,  // 自动过期清理
    DefaultTTL: time.Minute * 10,
}

// 配置数据 - 使用Simple
configConfig := cache.MemoryConfig{
    EvictionPolicy: cache.EvictionPolicySimple, // 长期存储
    DefaultTTL: time.Hour * 24,
}
```

### 技巧2: 合理设置TTL

```go
// 根据数据更新频率设置TTL
cache.Set(ctx, "user_profile", user, time.Hour*2)      // 用户资料，不常变
cache.Set(ctx, "stock_price", price, time.Minute*5)    // 股价，实时性高
cache.Set(ctx, "daily_report", report, time.Hour*12)   // 日报，每日更新
```

### 技巧3: 批量操作优化

```go
// 避免循环中的单个操作
for _, userID := range userIDs {
    // ❌ 不推荐：每次都访问缓存
    user, _ := userCache.Get(ctx, fmt.Sprintf("user:%d", userID))
}

// ✅ 推荐：预先检查缓存，批量处理
var cachedUsers []User
var missedIDs []int

for _, userID := range userIDs {
    key := fmt.Sprintf("user:%d", userID)
    if data, err := userCache.Get(ctx, key); err == nil {
        cachedUsers = append(cachedUsers, data.(User))
    } else {
        missedIDs = append(missedIDs, userID)
    }
}

// 批量加载未命中的数据
dbUsers := loadUsersFromDB(missedIDs)
for _, user := range dbUsers {
    userCache.Set(ctx, fmt.Sprintf("user:%d", user.ID), user, time.Hour*2)
}
```

## 🔍 调试与监控

### 检查缓存状态

```go
func checkCacheHealth() {
    caches := []struct{
        name string
        cache cache.Cache
    }{
        {"users", core.GetUserCache()},
        {"sessions", core.GetSessionCache()}, 
        {"products", core.GetProductCache()},
    }
    
    for _, c := range caches {
        if c.cache != nil {
            fmt.Printf("✅ %s 缓存可用\n", c.name)
        } else {
            fmt.Printf("❌ %s 缓存不可用\n", c.name)
        }
    }
}
```

### 缓存性能测试

```go
func benchmarkCache() {
    userCache := core.GetUserCache()
    ctx := context.Background()
    
    start := time.Now()
    
    // 写入测试
    for i := 0; i < 1000; i++ {
        key := fmt.Sprintf("test:%d", i)
        userCache.Set(ctx, key, fmt.Sprintf("value_%d", i), time.Hour)
    }
    writeTime := time.Since(start)
    
    start = time.Now()
    
    // 读取测试
    for i := 0; i < 1000; i++ {
        key := fmt.Sprintf("test:%d", i)
        userCache.Get(ctx, key)
    }
    readTime := time.Since(start)
    
    fmt.Printf("写入1000条: %v\n", writeTime)
    fmt.Printf("读取1000条: %v\n", readTime)
}
```

## 🚨 常见问题

### Q1: 缓存返回nil怎么办？

```go
// ❌ 错误处理
cache := core.GetUserCache()
cache.Set(ctx, "key", "value", time.Hour) // 可能panic

// ✅ 正确处理
cache := core.GetUserCache()
if cache == nil {
    fmt.Println("缓存服务未初始化，使用默认逻辑")
    return fallbackLogic()
}
cache.Set(ctx, "key", "value", time.Hour)
```

### Q2: 如何处理缓存失效？

```go
func getUserWithFallback(userID int) (*User, error) {
    ctx := context.Background()
    userCache := core.GetUserCache()
    
    // 尝试缓存
    if userCache != nil {
        if data, err := userCache.Get(ctx, fmt.Sprintf("user:%d", userID)); err == nil {
            if user, ok := data.(User); ok {
                return &user, nil
            }
        }
    }
    
    // 缓存失效，从数据库获取
    return loadUserFromDB(userID)
}
```

### Q3: 如何避免缓存穿透？

```go
func getPopularProduct(productID int) (*Product, error) {
    ctx := context.Background()
    productCache := core.GetProductCache()
    cacheKey := fmt.Sprintf("product:%d", productID)
    
    // 检查缓存
    if data, err := productCache.Get(ctx, cacheKey); err == nil {
        if product, ok := data.(Product); ok {
            return &product, nil
        }
        // 缓存了空值，避免穿透
        if data == nil {
            return nil, fmt.Errorf("商品不存在")
        }
    }
    
    // 从数据库查询
    product, err := loadProductFromDB(productID)
    if err != nil {
        // 缓存空值，防止频繁查询
        productCache.Set(ctx, cacheKey, nil, time.Minute*5)
        return nil, err
    }
    
    // 缓存正常数据
    productCache.Set(ctx, cacheKey, *product, time.Hour*2)
    return product, nil
}
```

## 🎯 最佳实践

1. **缓存键命名规范**
   ```go
   // 推荐格式: 模块:类型:ID
   "user:profile:123"
   "session:data:abc456"
   "product:detail:789"
   ```

2. **合理设置过期时间**
   ```go
   // 根据业务特性设置
   用户信息: 2小时
   会话数据: 30分钟  
   热门商品: 6小时
   系统配置: 24小时
   ```

3. **避免缓存大对象**
   ```go
   // ❌ 避免存储大对象
   cache.Set(ctx, "huge_data", hugeObject, time.Hour)
   
   // ✅ 存储引用或分片
   cache.Set(ctx, "data_ref", dataID, time.Hour)
   ```

---

🎉 **恭喜！** 你已经掌握了缓存模块的基本使用。更多高级特性请参考完整文档。