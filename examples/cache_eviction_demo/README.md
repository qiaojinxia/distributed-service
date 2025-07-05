# 缓存淘汰策略演示

本示例演示了框架缓存模块支持的三种淘汰策略：

## 淘汰策略

### 1. LRU (Least Recently Used)
- 当缓存满时，淘汰最久未被访问的项
- 适用于有访问局部性的数据
- 使用 `github.com/hashicorp/golang-lru/v2` 库实现

### 2. TTL (Time To Live)  
- 基于过期时间自动清理项目
- 结合LRU策略处理容量限制
- 使用 `github.com/hashicorp/golang-lru/v2/expirable` 库实现

### 3. Simple
- 轻量级实现，支持TTL和定期清理
- 适用于简单的缓存需求
- 使用 `github.com/patrickmn/go-cache` 库实现

## 运行示例

```bash
cd examples/cache_eviction_demo
go run main.go
```

## 示例输出

示例会演示：
1. LRU策略的淘汰行为
2. TTL策略的过期机制
3. Simple策略的基本功能
4. 各种策略的统计信息

## 配置参数

- `MaxSize`: 最大缓存项数量
- `DefaultTTL`: 默认过期时间
- `CleanupInterval`: 清理间隔
- `EvictionPolicy`: 淘汰策略 ("lru", "ttl", "simple")

## 性能建议

- **LRU**: 适合大多数通用场景
- **TTL**: 适合有明确过期需求的场景
- **Simple**: 适合轻量级、低并发场景