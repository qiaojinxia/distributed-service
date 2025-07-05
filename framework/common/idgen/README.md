# ID生成器 (IDGen) 模块

分布式唯一ID生成器，基于美团Leaf算法实现，完全集成到框架中，支持高性能、高可用的分布式ID生成。

## 🌟 特性

- **🚀 高性能**: 基于数据库号段模式，支持预加载和批量获取
- **📈 可扩展**: 动态步长调整，根据消耗速度自动优化
- **🔒 线程安全**: 支持高并发场景，内置缓存和锁机制
- **⚡ 零配置**: 框架自动管理，开箱即用
- **🎯 灵活配置**: 支持多种配置方式和业务标识管理
- **📊 监控友好**: 内置指标收集和性能监控
- **🛡️ 高可用**: 支持故障恢复和优雅降级

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Framework IDGen                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   App Builder   │  │ Component Mgr   │  │   Config    │  │
│  │                 │  │                 │  │             │  │
│  │ .WithIDGen()    │→ │ IDGenService    │→ │ IDGenConfig │  │
│  │ .WithIDGenAuto()│  │                 │  │             │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                Framework IDGen Service                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │    Leaf Core    │  │ Segment Buffer  │  │   Metrics   │  │
│  │                 │  │                 │  │             │  │
│  │ • ID Generation │  │ • L1/L2 Cache   │  │ • QPS Stats │  │
│  │ • Batch Support │  │ • Preloading    │  │ • Monitoring│  │
│  │ • Step Adjust   │  │ • Thread Safe   │  │ • Health    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                     Database Layer                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   GORM DAO      │  │  Framework DB   │  │   Custom DB │  │
│  │                 │  │                 │  │             │  │
│  │ • CRUD Ops      │  │ • Auto Config   │  │ • Manual    │  │
│  │ • Transactions  │  │ • Shared Pool   │  │ • Isolated  │  │
│  │ • Auto Migration│  │ • Zero Config   │  │ • Full Ctrl │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 1. 最简使用（推荐）

```go
package main

import "github.com/qiaojinxia/distributed-service/framework/core"

func main() {
    // 一行代码启动，自动检测配置并启用ID生成器
    core.New().AutoDetect().Run()
}
```

### 2. 显式配置

```go
func main() {
    core.New().
        WithDatabase(&config.MySQLConfig{
            Host:     "localhost",
            Port:     3306,
            Username: "root",
            Password: "password",
            Database: "distributed_service",
        }).
        WithIDGenDefaults(). // 使用默认ID生成器配置
        Run()
}
```

### 3. 配置文件方式

```yaml
# config.yaml
mysql:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "distributed_service"

idgen:
  enabled: true  # 仅此一行即可启用！
```

```go
func main() {
    core.New().
        Config("config.yaml").
        Run()
}
```

## 💻 基本用法

### 获取ID生成器服务

```go
// 方法1: 从组件管理器获取
componentManager := app.GetComponentManager()
idGenService := componentManager.GetIDGenService()

// 方法2: 直接创建（用于测试）
idGenService := idgen.NewFrameworkIDGenService()
err := idGenService.Initialize(context.Background())
```

### 生成ID

```go
ctx := context.Background()

// 生成单个ID
userID, err := idGenService.NextID(ctx, "user")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Generated User ID: %d\n", userID)

// 批量生成ID
orderIDs, err := idGenService.BatchNextID(ctx, "order", 10)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Generated Order IDs: %v\n", orderIDs)
```

### 业务标识管理

```go
// 创建新业务标识
err := idGenService.CreateBizTag(ctx, "payment", 3000, "支付ID生成器")

// 更新步长
err := idGenService.UpdateStep(ctx, "payment", 5000)

// 删除业务标识
err := idGenService.DeleteBizTag(ctx, "payment")
```

## ⚙️ 配置详解

### 基础配置

```yaml
idgen:
  enabled: true           # 是否启用ID生成器
  type: "leaf"           # 生成器类型 (当前支持: leaf)
  use_framework: true    # 是否使用框架数据库配置
  default_step: 1000     # 默认步长
```

### Leaf算法配置

```yaml
idgen:
  leaf:
    default_step: 1000              # 默认步长
    preload_threshold: "0.9"        # 预加载阈值 (90%时开始预加载)
    cleanup_interval: "1h"          # 清理不活跃buffer的间隔
    max_step_size: 100000           # 最大步长
    min_step_size: 100              # 最小步长
    step_adjust_ratio: "2.0"        # 步长调整比例
```

### 业务标识预配置

```yaml
idgen:
  biz_tags:
    user:
      step: 5000                    # 用户ID步长
      description: "用户ID生成器"
      auto_create: true             # 自动创建
    order:
      step: 10000                   # 订单ID步长
      description: "订单ID生成器"
      auto_create: true
    product:
      step: 1000                    # 产品ID步长
      description: "产品ID生成器"
      auto_create: true
```

### 自定义数据库配置

```yaml
idgen:
  enabled: true
  use_framework: false     # 不使用框架数据库
  database:               # 自定义数据库配置
    driver: "mysql"
    host: "localhost"
    port: 3306
    database: "custom_idgen_db"
    username: "idgen_user"
    password: "idgen_password"
    max_idle_conns: 5
    max_open_conns: 50
    conn_max_lifetime: "1h"
```

## 🔧 Builder API

### 基础配置方法

```go
// 自定义配置
builder.WithIDGen(&config.IDGenConfig{...})

// 使用默认配置
builder.WithIDGenDefaults()

// 从配置文件读取
builder.WithIDGenFromConfig()

// 智能自动配置
builder.WithIDGenAuto()
```

### 组合使用示例

```go
app := core.New().
    Name("MyApp").
    WithDatabase(&config.MySQLConfig{...}).
    WithIDGenDefaults().
    WithCache(&config.CacheConfig{...}).
    HTTP(func(r interface{}) {
        // 设置路由
    }).
    Build()
```

## 📊 监控和指标

### 获取指标

```go
// 获取单个业务标识指标
metrics := idGenService.GetMetrics("user")
fmt.Printf("Metrics: %+v\n", metrics)

// 获取所有指标
allMetrics := idGenService.GetAllMetrics()
for bizTag, metrics := range allMetrics {
    fmt.Printf("%s: %+v\n", bizTag, metrics)
}

// 获取Buffer状态
status := idGenService.GetBufferStatus("user")
fmt.Printf("Buffer Status: %+v\n", status)
```

### 指标说明

```go
type LeafMetrics struct {
    TotalRequests   int64   // 总请求数
    SuccessRequests int64   // 成功请求数  
    FailedRequests  int64   // 失败请求数
    SegmentLoads    int64   // 号段加载次数
    BufferSwitches  int64   // Buffer切换次数
    QPS             float64 // 每秒查询数
    LastUpdateTime  time.Time // 最后更新时间
}
```

## 🏗️ 数据库表结构

框架会自动创建所需的数据库表：

```sql
CREATE TABLE IF NOT EXISTS leaf_alloc (
    biz_tag VARCHAR(128) NOT NULL COMMENT '业务标识',
    max_id BIGINT NOT NULL DEFAULT 1 COMMENT '当前最大ID',
    step INT NOT NULL DEFAULT 1000 COMMENT '步长',
    description VARCHAR(256) COMMENT '描述',
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (biz_tag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Leaf分布式ID生成器';
```

## 📈 性能优化

### 步长设置建议

| 业务类型 | 推荐步长 | 说明 |
|---------|---------|------|
| 高频业务 (用户、订单) | 5000-10000 | 减少数据库访问 |
| 中频业务 (产品、支付) | 1000-3000 | 平衡性能和浪费 |
| 低频业务 (配置、日志) | 100-1000 | 减少ID浪费 |

### 预加载策略

```yaml
idgen:
  leaf:
    preload_threshold: "0.9"  # 90%时预加载，推荐0.8-0.95
    cleanup_interval: "1h"    # 清理间隔，推荐1h-4h
```

### 并发优化

```go
// 批量获取ID以提高性能
ids, err := idGenService.BatchNextID(ctx, "order", 100)

// 异步预创建业务标识
go func() {
    idGenService.CreateBizTag(ctx, "new_business", 1000, "新业务")
}()
```

## 🛡️ 最佳实践

### 1. 业务标识命名规范

```
✅ 推荐命名:
- user          (用户)
- order         (订单)  
- product       (产品)
- payment       (支付)
- message       (消息)

❌ 避免命名:
- id            (太通用)
- data          (不明确)
- temp          (临时性)
```

### 2. 错误处理

```go
func generateUserID(ctx context.Context, idGen IDGenService) (int64, error) {
    // 带重试的ID生成
    for i := 0; i < 3; i++ {
        id, err := idGen.NextID(ctx, "user")
        if err == nil {
            return id, nil
        }
        
        // 记录错误并重试
        log.Printf("ID generation failed (attempt %d): %v", i+1, err)
        time.Sleep(time.Millisecond * 100)
    }
    
    return 0, fmt.Errorf("failed to generate ID after 3 attempts")
}
```

### 3. 高并发场景

```go
// 使用缓冲通道预生成ID
type IDPool struct {
    idChan chan int64
    idGen  IDGenService
    bizTag string
}

func NewIDPool(idGen IDGenService, bizTag string, bufferSize int) *IDPool {
    pool := &IDPool{
        idChan: make(chan int64, bufferSize),
        idGen:  idGen,
        bizTag: bizTag,
    }
    
    // 启动补充协程
    go pool.refill()
    return pool
}

func (p *IDPool) GetID() int64 {
    return <-p.idChan
}

func (p *IDPool) refill() {
    for {
        // 当缓冲区不足时批量补充
        if len(p.idChan) < cap(p.idChan)/2 {
            ids, _ := p.idGen.BatchNextID(context.Background(), p.bizTag, 50)
            for _, id := range ids {
                p.idChan <- id
            }
        }
        time.Sleep(time.Second)
    }
}
```

### 4. 监控告警

```go
// 定期检查ID生成器健康状态
func checkIDGenHealth(idGen IDGenService) {
    metrics := idGen.GetAllMetrics()
    
    for bizTag, metric := range metrics {
        // 检查失败率
        if metric.FailedRequests > 0 {
            failureRate := float64(metric.FailedRequests) / float64(metric.TotalRequests)
            if failureRate > 0.01 { // 失败率超过1%
                log.Printf("WARNING: High failure rate for %s: %.2f%%", bizTag, failureRate*100)
            }
        }
        
        // 检查QPS
        if metric.QPS > 10000 { // QPS超过1万
            log.Printf("INFO: High QPS for %s: %.2f", bizTag, metric.QPS)
        }
    }
}
```

## 🔍 故障排查

### 常见问题

1. **ID生成失败**
   ```
   原因: 数据库连接问题或表不存在
   解决: 检查数据库配置和连接状态
   ```

2. **性能下降**
   ```
   原因: 步长设置过小或并发过高
   解决: 增加步长或使用批量获取
   ```

3. **ID重复**
   ```
   原因: 多实例使用相同业务标识且数据库配置错误
   解决: 确保所有实例连接同一数据库
   ```

### 调试模式

```go
// 启用详细日志
idGenConfig := &config.IDGenConfig{
    Database: config.IDGenDatabaseConfig{
        LogLevel: "info", // error/warn/info/silent
    },
}
```

## 📚 示例代码

完整示例请参考 [framework_examples.go](./framework_examples.go) 文件，包含：

- 基础使用示例
- 自定义配置示例  
- 高级功能示例
- 性能测试示例
- 最佳实践示例

## 🤝 贡献

欢迎提交Issue和Pull Request来改进ID生成器模块！

## 📄 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](../../../LICENSE) 文件。