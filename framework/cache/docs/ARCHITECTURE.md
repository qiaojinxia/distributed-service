# 缓存模块架构设计

## 🏗️ 整体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                           用户应用层                                 │
│                  User Application Layer                             │
├─────────────────────────────────────────────────────────────────────┤
│                         框架全局API                                  │
│                    Framework Global API                             │
│                                                                     │
│  core.GetUserCache()    core.GetSessionCache()    core.GetCache()  │
│  core.GetProductCache() core.GetConfigCache()     core.HasCache()  │
├─────────────────────────────────────────────────────────────────────┤
│                       框架缓存服务层                                  │
│                  Framework Cache Service Layer                      │
│                                                                     │
│                    FrameworkCacheService                            │
│  ┌─────────────────┬─────────────────┬─────────────────────────────┐ │
│  │   Named Cache   │   Named Cache   │      Named Cache        │ │
│  │   Management    │   Management    │      Management         │ │
│  │                 │                 │                         │ │
│  │  users(LRU)     │ sessions(TTL)   │   products(Simple)      │ │
│  │  configs(...)   │  temp_data(...) │   custom_cache(...)     │ │
│  └─────────────────┴─────────────────┴─────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────┤
│                         缓存接口层                                    │
│                      Cache Interface Layer                          │
│                                                                     │
│                        cache.Cache                                 │
│    ┌─────────────────────────────────────────────────────────────┐   │
│    │  Set() | Get() | Delete() | Exists() | Clear()            │   │
│    └─────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────┤
│                         缓存实现层                                    │
│                   Cache Implementation Layer                        │
│                                                                     │
│  ┌───────────────────┬───────────────────┬─────────────────────────┐ │
│  │   Memory Cache    │    Redis Cache    │    Future Extensions   │ │
│  │                   │                   │                         │ │
│  │  ┌─────────────┐  │  ┌─────────────┐  │  ┌─────────────────────┐ │ │
│  │  │ LRU Policy  │  │  │ Distributed │  │  │  • Memcached        │ │ │
│  │  │ TTL Policy  │  │  │ TTL Support │  │  │  • Multi-tier       │ │ │
│  │  │Simple Policy│  │  │ Pub/Sub     │  │  │  • Compressed       │ │ │
│  │  └─────────────┘  │  └─────────────┘  │  └─────────────────────┘ │ │
│  └───────────────────┴───────────────────┴─────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────┤
│                        存储引擎层                                     │
│                     Storage Engine Layer                            │
│                                                                     │
│  ┌───────────────────┬───────────────────┬─────────────────────────┐ │
│  │  Third-party Libs │   Redis Client    │    System Memory       │ │
│  │                   │                   │                         │ │
│  │ • hashicorp/lru   │ • go-redis/redis  │ • Built-in maps        │ │
│  │ • patrickmn/      │ • Connection Pool │ • GC Management        │ │
│  │   go-cache        │ • Failover        │ • Memory Limits        │ │
│  └───────────────────┴───────────────────┴─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

## 🔄 数据流设计

### 缓存写入流程
```
User App → Global API → Framework Service → Cache Interface → Implementation → Storage
```

### 缓存读取流程
```
User App → Global API → Framework Service → Cache Interface → Implementation → Storage
                ↓                                                               ↓
            Cache Hit/Miss ←─────────────────────────────────────────────────────┘
```

### 缓存失效流程
```
TTL Expiry ─┐
            ├→ Background Cleanup → Remove Entry → Update Index
LRU Evict ──┘
```

## 🧩 组件详细设计

### 1. 缓存接口设计 (cache.Cache)

```go
type Cache interface {
    // 基础操作
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Get(ctx context.Context, key string) (interface{}, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Clear(ctx context.Context) error
    
    // 扩展操作 (未来)
    // SetMulti(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
    // GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
    // Increment(ctx context.Context, key string, delta int64) (int64, error)
}
```

**设计原则:**
- 统一接口，屏蔽实现差异
- 支持上下文传递，便于超时控制
- 泛型友好，支持任意数据类型
- 错误处理明确，便于故障排查

### 2. 内存缓存实现 (MemoryCache)

```go
type MemoryCache struct {
    config      MemoryConfig
    lruCache    *lru.Cache          // LRU策略缓存
    goCache     *cache.Cache        // TTL策略缓存  
    simpleCache map[string]CacheItem // Simple策略缓存
    mutex       sync.RWMutex        // 并发安全
}

type MemoryConfig struct {
    MaxSize         int             // 最大条目数
    DefaultTTL      time.Duration   // 默认过期时间
    CleanupInterval time.Duration   // 清理间隔
    EvictionPolicy  EvictionPolicy  // 淘汰策略
}
```

**设计特点:**
- **多策略支持**: 根据配置选择不同的淘汰策略
- **第三方库集成**: 使用成熟的LRU和TTL实现
- **并发安全**: 使用读写锁保护并发访问
- **内存可控**: 支持最大条目数限制

### 3. 框架缓存服务 (FrameworkCacheService)

```go
type FrameworkCacheService struct {
    caches      map[string]Cache    // 命名缓存实例
    redisClient *redis.Client       // Redis客户端
    config      ServiceConfig       // 服务配置
    mutex       sync.RWMutex        // 并发保护
}

type ServiceConfig struct {
    DefaultCaches map[string]CacheConfig // 默认缓存配置
    RedisConfig   *RedisConfig           // Redis配置
    EnableRedis   bool                   // 是否启用Redis
}
```

**设计特点:**
- **命名管理**: 支持多个命名缓存实例
- **自动初始化**: 框架启动时自动创建默认缓存
- **Redis集成**: 支持Redis分布式缓存
- **优雅降级**: Redis不可用时自动使用内存缓存

## 🎯 淘汰策略设计

### LRU (Least Recently Used)
```
访问顺序: A → B → C → D (缓存已满)
添加 E: 淘汰 A (最久未使用)
结果: B → C → D → E
```

**适用场景:**
- 用户信息缓存
- 热点数据缓存
- 有限内存环境

### TTL (Time To Live)
```
时间线: |--[A(2s)]--[B(5s)]--[C(1s)]--| → 3秒后: [B]
        0s         1s        2s       3s
```

**适用场景:**
- 会话数据
- 临时token
- 实时性要求高的数据

### Simple (简单存储)
```
无淘汰策略，直到手动清理或重启
Key-Value 直接映射存储
```

**适用场景:**
- 配置数据
- 静态资源
- 长期有效的数据

## 🔧 配置策略

### 默认配置矩阵

| 缓存类型 | 策略 | 大小 | 默认TTL | 清理间隔 | 用途 |
|---------|------|------|---------|----------|------|
| users | LRU | 1000 | 2小时 | 10分钟 | 用户信息 |
| sessions | TTL | 5000 | 30分钟 | 5分钟 | 会话数据 |
| products | Simple | 500 | 6小时 | 30分钟 | 商品信息 |
| configs | Simple | 100 | 24小时 | 1小时 | 配置数据 |

### 性能调优指南

1. **内存使用优化**
   ```go
   // 根据数据大小调整MaxSize
   MaxSize = MemoryLimit / AvgItemSize / SafetyFactor
   ```

2. **TTL设置建议**
   ```go
   // 根据数据更新频率设置
   HighFrequency: 5-30分钟
   MediumFrequency: 1-6小时  
   LowFrequency: 12-24小时
   ```

3. **清理间隔优化**
   ```go
   // 清理间隔 = TTL / 10 (建议值)
   CleanupInterval = DefaultTTL / 10
   ```

## 🚀 扩展设计

### 分层缓存架构
```
L1 Cache (Memory) → L2 Cache (Redis) → L3 Cache (Database)
```

### 缓存预热机制
```go
type Preheater interface {
    Preheat(ctx context.Context, keys []string) error
    PreheatPattern(ctx context.Context, pattern string) error
}
```

### 监控指标设计
```go
type Metrics struct {
    HitCount    int64 // 命中次数
    MissCount   int64 // 未命中次数
    SetCount    int64 // 设置次数
    DeleteCount int64 // 删除次数
    EvictCount  int64 // 淘汰次数
}
```

### 缓存一致性保证
```go
type Consistency interface {
    InvalidatePattern(pattern string) error
    BroadcastInvalidate(key string) error
    WatchChanges(callback func(key string)) error
}
```

## 📊 性能基准

### 内存使用估算
```
每个缓存条目 ≈ Key大小 + Value大小 + 元数据开销(~64字节)
1000个用户缓存 ≈ 1000 * (50 + 500 + 64) ≈ 600KB
```

### 操作性能预期
```
内存缓存:
- Set: ~1µs
- Get: ~0.5µs  
- Delete: ~0.5µs

Redis缓存:
- Set: ~1ms
- Get: ~0.5ms
- Delete: ~0.5ms
```

---

> 该架构设计支持高并发、低延迟的缓存访问，同时保持了良好的扩展性和可维护性。