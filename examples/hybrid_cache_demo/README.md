# 混合缓存演示 (L1本地 + L2 Redis)

这个示例展示了框架的混合缓存功能，实现了多级缓存架构，结合本地内存缓存和Redis分布式缓存的优势。

## 🚀 核心特性

### 多级缓存架构
- **L1缓存**: 本地内存缓存，提供毫秒级访问速度
- **L2缓存**: Redis分布式缓存，提供持久化和共享能力
- **智能路由**: 优先从L1读取，未命中时从L2读取并回填L1

### 三种同步策略

#### 1. 📝 写穿透 (Write-Through)
```
写入请求 → 同时写L1和L2 → 返回结果
```
- **特点**: 数据一致性好，写延迟稍高
- **适用**: 数据一致性要求高的场景

#### 2. 🔄 写回 (Write-Back)
```
写入请求 → 写L1 → 定时批量写L2 → 返回结果
```
- **特点**: 写性能好，可能存在短暂不一致
- **适用**: 高并发写入场景

#### 3. 🎯 写绕过 (Write-Around)
```
写入请求 → 只写L2 → 返回结果
```
- **特点**: 节省L1空间，适合写多读少
- **适用**: 大数据量、偶尔访问的场景

## 🛠️ 使用方式

### 1. 快速开始

```go
// 创建缓存管理器
manager := framework.NewCacheManager()

// 创建混合缓存
err := manager.CreateCache(cache.Config{
    Type: cache.TypeHybrid,
    Name: "my_hybrid_cache",
    Settings: map[string]interface{}{
        "sync_strategy": "write_through",
        "l1_config": map[string]interface{}{
            "type": "memory",
            "settings": map[string]interface{}{
                "max_size": 1000,
                "default_ttl": "1h",
            },
        },
        "l2_config": map[string]interface{}{
            "type": "redis",
            "settings": map[string]interface{}{
                "addr": "localhost:6379",
                "db": 0,
            },
        },
    },
})

// 使用缓存
hybridCache, _ := manager.GetCache("my_hybrid_cache")
hybridCache.Set(ctx, "key", "value", time.Hour)
value, _ := hybridCache.Get(ctx, "key")
```

### 2. 使用配置预设

```go
// 默认配置 - 平衡性能和内存使用
defaultConfig := cache.Presets.GetDefaultHybridConfig()

// 高性能配置 - 大内存，写回模式
highPerfConfig := cache.Presets.GetHighPerformanceHybridConfig()

// 低内存配置 - 节省内存，写绕过模式
lowMemConfig := cache.Presets.GetLowMemoryHybridConfig()
```

### 3. 自定义配置

```go
customConfig := cache.NewCustomHybridConfig().
    WithL1Memory(5000, time.Minute*45).
    WithL2Redis("localhost:6379", "", 0, time.Hour*6).
    WithSyncStrategy(cache.SyncStrategyWriteBack).
    WithWriteBack(true, time.Minute*3, 50).
    Build()

hybridCache, err := cache.NewHybridCache(customConfig)
```

## 📊 监控和统计

混合缓存提供详细的统计信息：

```go
if hybridCache, ok := cache.(*cache.HybridCache); ok {
    stats := hybridCache.GetStats()
    fmt.Printf("L1命中率: %.2f%%\n", 
        float64(stats.L1Hits)/(float64(stats.L1Hits+stats.L1Misses))*100)
    fmt.Printf("L2命中率: %.2f%%\n", 
        float64(stats.L2Hits)/(float64(stats.L2Hits+stats.L2Misses))*100)
    fmt.Printf("写回次数: %d\n", stats.Writebacks)
}
```

## 🎯 最佳实践

### 1. 根据业务场景选择策略

| 场景 | 推荐策略 | 原因 |
|------|----------|------|
| 用户会话 | Write-Back | 高并发读写，允许短暂不一致 |
| 商品信息 | Write-Through | 数据一致性重要 |
| 日志统计 | Write-Around | 写多读少，节省内存 |

### 2. 合理设置TTL

```go
// L1缓存：短TTL，快速失效，节省内存
l1TTL := time.Minute * 30

// L2缓存：长TTL，减少数据库压力
l2TTL := time.Hour * 24
```

### 3. 监控关键指标

- **L1命中率**: 应该 > 80%
- **总命中率**: 应该 > 95%
- **写回延迟**: 应该 < 100ms
- **错误率**: 应该 < 1%

## 🔧 配置参数说明

### 基础配置
- `sync_strategy`: 同步策略 (`write_through`, `write_back`, `write_around`)
- `l1_ttl`: L1缓存TTL
- `l2_ttl`: L2缓存TTL

### 写回配置
- `write_back_enabled`: 是否启用写回
- `write_back_interval`: 写回间隔
- `write_back_batch_size`: 批量写回大小

### L1内存缓存配置
- `max_size`: 最大条目数
- `default_ttl`: 默认TTL
- `cleanup_interval`: 清理间隔

### L2 Redis配置
- `addr`: Redis地址
- `password`: 密码
- `db`: 数据库编号
- `pool_size`: 连接池大小

## 🚨 注意事项

1. **Redis依赖**: 确保Redis服务可用
2. **内存管理**: 合理设置L1缓存大小，避免OOM
3. **网络延迟**: L2缓存访问会有网络延迟
4. **数据一致性**: 写回模式可能存在短暂不一致
5. **故障转移**: 设计好Redis故障时的降级策略

## 🎬 运行演示

```bash
cd examples/hybrid_cache_demo
go run main.go
```

演示包含：
- 写穿透策略演示
- 写回策略演示  
- 自定义配置演示
- 配置预设展示
- 性能统计展示