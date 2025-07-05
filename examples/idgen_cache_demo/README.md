# 分布式ID生成器和缓存管理器演示

这个示例展示了如何使用框架的分布式ID生成器和缓存管理器功能。

## 功能特性

### 🆔 分布式ID生成器 (美团Leaf算法)
- **双Buffer机制**: 预分配ID段，避免频繁数据库访问
- **高性能**: 内存中生成ID，TPS可达数万
- **高可用**: 支持多实例部署，数据库宕机时仍可短时间工作
- **业务隔离**: 支持多业务标识，互不影响

### 💾 缓存管理器
- **多缓存类型**: 支持内存缓存、Redis缓存等
- **动态注册**: 可注册不同功能的缓存实例
- **统计监控**: 提供命中率、错误数等统计信息
- **批量操作**: 支持批量读写操作

## 使用方法

### 1. 运行演示
```bash
cd examples/idgen_cache_demo
go run main.go
```

### 2. 缓存管理器使用

```go
// 创建缓存管理器
manager := framework.NewCacheManager()

// 创建内存缓存
err := manager.CreateCache(cache.Config{
    Type: cache.TypeMemory,
    Name: "user_cache",
    Settings: map[string]interface{}{
        "max_size":         1000,
        "default_ttl":      "1h",
        "cleanup_interval": "10m",
    },
})

// 获取缓存实例
userCache, err := manager.GetCache("user_cache")

// 使用缓存
ctx := context.Background()
err = userCache.Set(ctx, "key", "value", time.Hour)
value, err := userCache.Get(ctx, "key")
```

### 3. 分布式ID生成器使用

```go
// 创建ID生成器
config := idgen.Config{
    Type:      "leaf",
    TableName: "leaf_alloc",
    Database: &idgen.DatabaseConfig{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "test_db",
        Username: "root",
        Password: "password",
        Charset:  "utf8mb4",
    },
}

idGen, err := framework.NewIDGenerator(config)

// 生成ID
ctx := context.Background()
userID, err := idGen.NextID(ctx, "user")

// 批量生成ID
orderIDs, err := idGen.BatchNextID(ctx, "order", 100)
```

## 数据库表结构

分布式ID生成器需要以下MySQL表结构：

```sql
CREATE TABLE leaf_alloc (
    biz_tag VARCHAR(128) NOT NULL PRIMARY KEY,
    max_id BIGINT NOT NULL DEFAULT 0,
    step INT NOT NULL DEFAULT 1000,
    description VARCHAR(256),
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 初始化业务数据
INSERT INTO leaf_alloc (biz_tag, max_id, step, description) VALUES
('user', 0, 1000, '用户ID'),
('order', 0, 2000, '订单ID'),
('product', 0, 1000, '商品ID');
```

## 配置说明

### 缓存配置

#### 内存缓存配置
- `max_size`: 最大缓存条目数
- `default_ttl`: 默认过期时间
- `cleanup_interval`: 清理过期数据间隔
- `eviction_policy`: 淘汰策略（LRU等）

#### Redis缓存配置
- `addr`: Redis地址
- `password`: Redis密码
- `db`: 数据库编号
- `pool_size`: 连接池大小

### ID生成器配置
- `type`: 生成器类型（目前支持"leaf"）
- `table_name`: 数据库表名
- `database`: 数据库连接配置

## 注意事项

1. **生产环境**: 确保数据库高可用，建议使用主从复制
2. **性能调优**: 根据业务需求调整step大小
3. **监控告警**: 监控ID生成器和缓存的健康状态
4. **容量规划**: 根据业务量规划缓存容量和数据库性能