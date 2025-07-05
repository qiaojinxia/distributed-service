# 缓存淘汰策略文档

## 概述

框架的内存缓存组件现在支持多种淘汰策略，使用现成的成熟库来实现，避免了自己编写复杂的淘汰逻辑。

## 支持的淘汰策略

### 1. LRU (Least Recently Used) - 最近最少使用
- **库**: `github.com/hashicorp/golang-lru/v2`
- **策略**: 当缓存满时，淘汰最久未被访问的项
- **使用场景**: 适用于大多数通用缓存需求
- **特点**: 
  - 高效的O(1)操作
  - 自动淘汰机制
  - 线程安全

### 2. TTL (Time To Live) - 基于过期时间
- **库**: `github.com/hashicorp/golang-lru/v2/expirable`
- **策略**: 基于过期时间自动清理，结合LRU策略
- **使用场景**: 适用于有明确过期需求的缓存
- **特点**:
  - 自动过期清理
  - 结合LRU策略
  - 支持不同项的不同TTL

### 3. Simple - 简单策略
- **库**: `github.com/patrickmn/go-cache`
- **策略**: 基于TTL的简单缓存，支持定期清理
- **使用场景**: 适用于轻量级缓存需求
- **特点**:
  - 轻量级实现
  - 支持TTL
  - 定期清理机制

## 配置示例

### 基本配置

```yaml
cache:
  enabled: true
  default_type: "memory"
  caches:
    # LRU策略缓存
    user_cache:
      type: "memory"
      key_prefix: "users"
      ttl: "1h"
      settings:
        max_size: 1000
        eviction_policy: "lru"
        cleanup_interval: "10m"
    
    # TTL策略缓存  
    session_cache:
      type: "memory"
      key_prefix: "sessions"
      ttl: "30m"
      settings:
        max_size: 500
        eviction_policy: "ttl"
        cleanup_interval: "5m"
    
    # 简单策略缓存
    config_cache:
      type: "memory"
      key_prefix: "config"
      ttl: "10m"
      settings:
        max_size: 100
        eviction_policy: "simple"
        cleanup_interval: "1m"
```

### 代码配置

```go
import (
    "github.com/qiaojinxia/distributed-service/framework/cache"
    "time"
)

// LRU缓存配置
lruConfig := cache.MemoryConfig{
    MaxSize:         1000,
    DefaultTTL:      time.Hour,
    CleanupInterval: time.Minute * 10,
    EvictionPolicy:  cache.EvictionPolicyLRU,
}

// TTL缓存配置
ttlConfig := cache.MemoryConfig{
    MaxSize:         500,
    DefaultTTL:      time.Minute * 30,
    CleanupInterval: time.Minute * 5,
    EvictionPolicy:  cache.EvictionPolicyTTL,
}

// 简单缓存配置
simpleConfig := cache.MemoryConfig{
    MaxSize:         100,
    DefaultTTL:      time.Minute * 10,
    CleanupInterval: time.Minute,
    EvictionPolicy:  cache.EvictionPolicySimple,
}

// 创建缓存
lruCache, err := cache.NewMemoryCache(lruConfig)
if err != nil {
    log.Fatal(err)
}
```

## 配置参数说明

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `max_size` | int | 否 | 1000 | 最大缓存项数量 |
| `default_ttl` | duration | 否 | "1h" | 默认过期时间 |
| `cleanup_interval` | duration | 否 | "10m" | 清理间隔时间 |
| `eviction_policy` | string | 否 | "lru" | 淘汰策略: "lru", "ttl", "simple" |

## 性能特征

### LRU策略
- **时间复杂度**: Get/Set/Delete都是O(1)
- **空间复杂度**: O(n)，n为最大缓存大小
- **内存使用**: 中等，需要维护访问顺序链表
- **适用场景**: 通用缓存，访问模式有局部性

### TTL策略  
- **时间复杂度**: Get/Set都是O(1)，过期清理是O(n)
- **空间复杂度**: O(n)
- **内存使用**: 中等，需要存储过期时间
- **适用场景**: 有明确过期需求的缓存

### Simple策略
- **时间复杂度**: Get/Set/Delete都是O(1)
- **空间复杂度**: O(n)  
- **内存使用**: 较低，简单的map实现
- **适用场景**: 轻量级缓存需求

## 使用建议

### 选择策略的指导原则

1. **LRU策略** - 推荐用于:
   - 通用业务缓存
   - 用户数据缓存
   - 计算结果缓存
   - 访问模式有局部性的场景

2. **TTL策略** - 推荐用于:
   - 会话数据缓存
   - 临时令牌缓存
   - 有明确过期需求的数据
   - 需要精确控制数据生命周期

3. **Simple策略** - 推荐用于:
   - 配置数据缓存
   - 轻量级临时缓存
   - 开发和测试环境
   - 对性能要求不高的场景

### 配置优化建议

1. **合理设置max_size**:
   - 根据内存限制和数据大小估算
   - 监控命中率，适当调整大小
   - 避免设置过大导致内存占用过多

2. **合理设置TTL**:
   - 根据数据的时效性设置
   - 过短的TTL会增加缓存miss
   - 过长的TTL会占用过多内存

3. **合理设置cleanup_interval**:
   - TTL和Simple策略需要定期清理
   - 清理频率影响内存使用和CPU消耗
   - 一般设置为TTL的1/6到1/3

## 监控和调试

### 统计信息

所有缓存都提供统计信息:

```go
stats := cache.GetStats()
fmt.Printf("命中率: %.2f%%\n", float64(stats.Hits)/float64(stats.Hits+stats.Misses)*100)
fmt.Printf("总操作数: %d\n", stats.Sets+stats.Gets+stats.Deletes)
fmt.Printf("淘汰数: %d\n", stats.Evictions)
```

### 性能监控

```go
// 设置淘汰回调，监控淘汰行为
cache.SetEvictionCallback(func(key string, value interface{}) {
    log.Printf("Cache evicted: key=%s", key)
})
```

## 迁移指南

### 从旧版本迁移

如果你之前使用的是框架的旧版内存缓存，迁移步骤如下：

1. **更新配置文件**，添加`eviction_policy`参数
2. **选择合适的策略**，默认使用LRU
3. **测试验证**缓存行为是否符合预期

### 配置兼容性

- 旧配置文件仍然兼容，会自动使用LRU策略
- 新的配置参数都有合理的默认值
- 可以逐步迁移，不需要一次性更改所有缓存配置

## 最佳实践

1. **开发环境**使用Simple策略，减少资源消耗
2. **生产环境**根据业务需求选择LRU或TTL策略  
3. **监控缓存命中率**，优化缓存配置
4. **定期评估缓存大小**，避免内存泄漏
5. **使用适当的TTL**，平衡数据新鲜度和性能
6. **测试不同策略**在你的使用场景下的表现

## 故障排除

### 常见问题

1. **内存使用过高**
   - 检查max_size设置是否合理
   - 检查cleanup_interval是否太长
   - 监控淘汰行为是否正常

2. **命中率过低**
   - 检查TTL设置是否太短
   - 检查缓存大小是否太小
   - 分析访问模式是否适合当前策略

3. **性能问题**
   - LRU策略在高并发下性能最好
   - TTL策略的清理操作可能影响性能
   - Simple策略适合低并发场景

### 调试技巧

```go
// 启用详细日志
cache.SetEvictionCallback(func(key string, value interface{}) {
    log.Printf("Evicted: %s (reason: %s)", key, "policy")
})

// 定期输出统计信息
go func() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        stats := cache.GetStats()
        log.Printf("Cache stats: hits=%d, misses=%d, evictions=%d", 
            stats.Hits, stats.Misses, stats.Evictions)
    }
}()
```