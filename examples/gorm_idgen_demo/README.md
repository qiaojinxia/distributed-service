# 基于GORM的美团Leaf分布式ID生成器演示

这个示例展示了使用GORM重构后的美团Leaf分布式ID生成器，提供了更现代化的数据库操作和更好的性能。

## 🚀 核心特性

### 📊 基于GORM的现代化实现
- **多数据库支持**: MySQL、PostgreSQL、SQLite
- **连接池管理**: 自动管理数据库连接
- **事务支持**: 原子性操作保证数据一致性
- **日志控制**: 可配置的日志级别

### ⚡ 高性能双缓冲机制
- **双Buffer设计**: 预分配ID段，避免频繁数据库访问
- **异步预加载**: 90%使用率时自动预加载下一个号段
- **动态步长调整**: 根据消耗速度自动调整步长
- **并发安全**: 使用sync.Map和原子操作确保线程安全

### 📈 完整的监控体系
- **实时指标**: QPS、成功率、错误率统计
- **Buffer状态**: 号段使用情况、剩余ID数量
- **性能分析**: 号段加载次数、缓冲区切换次数
- **业务隔离**: 每个业务标识独立的指标统计

## 🛠️ 使用方式

### 1. 基本用法

```go
// 使用SQLite（适合开发和测试）
config := idgen.SQLiteConfig("./demo.db")

// 创建ID生成器
idGen, err := framework.NewIDGenerator(config)
gormGen := idGen.(*idgen.GormLeafIDGenerator)

// 创建表和业务标识
ctx := context.Background()
gormGen.CreateTable(ctx)
gormGen.CreateBizTag(ctx, "user", 1000, "用户ID")

// 生成ID
userID, err := gormGen.NextID(ctx, "user")
```

### 2. 配置构建器

```go
config := idgen.NewConfigBuilder().
    WithMySQL("localhost", 3306, "test_db", "root", "password").
    WithLeafConfig(&idgen.LeafConfig{
        DefaultStep:      1000,
        PreloadThreshold: 0.9,
        MaxStepSize:      100000,
        MinStepSize:      100,
    }).
    WithConnectionPool(20, 100, time.Hour).
    WithLogLevel("info").
    Build()
```

### 3. 预设配置

```go
// MySQL配置
config := idgen.MySQLConfig("localhost", 3306, "db", "user", "pass")

// PostgreSQL配置  
config := idgen.PostgreSQLConfig("localhost", 5432, "db", "user", "pass")

// SQLite配置
config := idgen.SQLiteConfig("./app.db")
```

## 📊 性能特点

### QPS表现
- **单机QPS**: 50,000+ (取决于硬件配置)
- **并发安全**: 支持多协程并发访问
- **低延迟**: 平均延迟 < 1ms

### 内存使用
- **双Buffer**: 每个业务标识约2个号段的内存占用
- **自动清理**: 定期清理不活跃的缓存
- **指标存储**: 轻量级的统计信息

### 数据库访问
- **批量获取**: 一次数据库访问获取多个ID
- **减少频次**: 90%减少数据库访问次数
- **事务保护**: 确保ID分配的原子性

## 🔧 配置参数

### LeafConfig 参数
- `DefaultStep`: 默认步长 (建议1000-10000)
- `PreloadThreshold`: 预加载阈值 (建议0.8-0.95)
- `CleanupInterval`: 清理间隔 (建议30分钟-2小时)
- `MaxStepSize`: 最大步长 (避免内存过度占用)
- `MinStepSize`: 最小步长 (保证基本性能)
- `StepAdjustRatio`: 步长调整比例 (建议1.5-3.0)

### DatabaseConfig 参数
- `MaxIdleConns`: 最大空闲连接数
- `MaxOpenConns`: 最大打开连接数
- `ConnMaxLifetime`: 连接最大生存时间
- `LogLevel`: 日志级别 (silent/error/warn/info)

## 📈 监控指标

### 核心指标
```go
metrics := gormGen.GetMetrics("user")
fmt.Printf("QPS: %.2f\n", metrics.AverageQPS)
fmt.Printf("成功率: %.2f%%\n", metrics.SuccessRate()*100)
fmt.Printf("号段加载次数: %d\n", metrics.SegmentLoads)
```

### Buffer状态
```go
status := gormGen.GetBufferStatus("user")
fmt.Printf("当前使用率: %.2f%%\n", 
    status["current_segment"].(map[string]interface{})["usage_ratio"])
```

## 🎯 最佳实践

### 1. 步长设置
- **高频业务**: 步长5000-20000
- **中频业务**: 步长1000-5000  
- **低频业务**: 步长100-1000

### 2. 预加载阈值
- **稳定负载**: 0.9 (90%时预加载)
- **突发负载**: 0.8 (80%时预加载)
- **低延迟要求**: 0.7 (70%时预加载)

### 3. 数据库配置
- **连接池**: 根据并发数设置，一般10-50
- **日志级别**: 生产环境使用error或warn
- **事务隔离**: 使用默认的READ_COMMITTED

### 4. 监控告警
- **成功率**: < 99% 时告警
- **QPS下降**: 比历史平均值低50%时告警
- **号段加载频率**: 异常高时检查步长配置

## 🚨 注意事项

1. **数据库时钟**: 确保数据库服务器时钟同步
2. **步长设置**: 避免设置过大的步长导致ID跳跃过大
3. **业务隔离**: 不同业务使用不同的biz_tag
4. **资源清理**: 程序退出时调用Close()方法
5. **并发控制**: 高并发下注意数据库连接数限制

## 🎬 运行演示

```bash
cd examples/gorm_idgen_demo
go run main.go
```

演示包含：
- 基本用法演示
- 配置构建器演示
- 性能测试演示
- 监控和指标演示

## 🔄 升级指南

从原始版本升级到GORM版本：

1. **依赖变更**: 添加GORM相关依赖
2. **配置更新**: 使用新的配置结构
3. **API兼容**: 保持相同的IDGenerator接口
4. **数据迁移**: 表结构基本兼容，可平滑迁移