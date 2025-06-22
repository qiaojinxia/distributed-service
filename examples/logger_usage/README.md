# 规范化日志系统设计

这个示例展示了重构后的规范化日志系统设计，解决了之前设计中的问题。

## 🎯 设计目标

- ✅ **统一且简洁的API**：单一接口，避免冗余
- ✅ **自动 Trace ID 注入**：所有日志自动包含链路追踪信息
- ✅ **职责分离**：日志逻辑与字段创建分离
- ✅ **符合 Go 习惯**：简洁、直观的包级函数
- ✅ **类型安全**：强类型字段，避免运行时错误
- ✅ **高性能**：基于 Zap，零分配设计

## 📐 架构设计

### 核心组件

```
logger/
├── logger.go      # 核心日志接口和实现
├── fields.go      # 字段创建工具函数
└── README.md      # 使用文档
```

### 接口设计

```go
// 统一的日志接口 - 简洁且功能完整
type Logger interface {
    // 结构化日志（推荐）
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, fields ...Field)
    Fatal(ctx context.Context, msg string, fields ...Field)

    // 格式化日志
    Debugf(ctx context.Context, template string, args ...interface{})
    Infof(ctx context.Context, template string, args ...interface{})
    Warnf(ctx context.Context, template string, args ...interface{})
    Errorf(ctx context.Context, template string, args ...interface{})
    Fatalf(ctx context.Context, template string, args ...interface{})

    // 子日志器
    With(fields ...Field) Logger
    WithContext(ctx context.Context, fields ...Field) Logger
}
```

## 🚀 使用方式

### 1. 基本使用

```go
import "github.com/qiaojinxia/distributed-service/framework/logger"

// 包级函数 - 最常用
logger.Info(ctx, "用户登录成功", 
    logger.UserID("123"),
    logger.Duration("latency", time.Millisecond*200),
)

// 格式化日志
logger.Infof(ctx, "处理请求 %s，耗时 %v", path, duration)
```

### 2. 子日志器

```go
// 创建专用日志器
userLogger := logger.Default().With(
    logger.Service("user-service"),
    logger.Version("v2.0.0"),
)

userLogger.Info(ctx, "服务启动", logger.Port(8080))
```

### 3. 业务字段

```go
// 内置业务字段
logger.Info(ctx, "API请求",
    logger.Method("POST"),           // HTTP方法
    logger.Path("/api/users"),       // 请求路径  
    logger.StatusCode(201),          // 状态码
    logger.ClientIP("192.168.1.1"),  // 客户端IP
    logger.ResponseTime(time.Millisecond*150), // 响应时间
)

// 数据库操作
logger.Debug(ctx, "执行查询",
    logger.Database("orders"),       // 数据库名
    logger.Table("order_items"),     // 表名
    logger.SQL("SELECT * FROM ..."), // SQL语句
)

// 消息队列
logger.Info(ctx, "消息发送",
    logger.Queue("notifications"),   // 队列名
    logger.Topic("user.created"),    // 主题
)
```

### 4. 链式字段构建

```go
fields := logger.NewFields().
    String("module", "payment").
    Int("amount", 9999).
    Bool("is_test", false).
    Duration("processing_time", time.Millisecond*500).
    Build()

logger.Info(ctx, "支付完成", fields...)
```

## 🔍 Trace ID 自动注入

所有日志方法都会自动从 `context.Context` 中提取 OpenTelemetry 的 trace 信息：

```json
{
  "level": "info",
  "ts": "2025-01-07T14:30:15.123Z",
  "msg": "用户登录成功",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "user_id": "123",
  "latency": "200ms"
}
```

## 📊 与旧设计对比

| 特性 | 旧设计 | 新设计 |
|------|--------|---------|
| 接口数量 | 2个（Logger + ContextLogger） | 1个（Logger） |
| API风格 | 混合（包函数+方法） | 统一（包函数为主） |
| 方法命名 | 冗余（InfoCtx, DebugfCtx） | 简洁（Info, Debugf） |
| Trace ID | 手动添加 | 自动注入 |
| 字段创建 | 混在一起 | 独立文件 |
| 代码重复 | 多处重复 | DRY原则 |

## 🛠️ 最佳实践

### 1. 优先使用结构化日志

```go
// ✅ 推荐 - 结构化日志
logger.Info(ctx, "订单创建成功",
    logger.String("order_id", orderID),
    logger.Int("amount", amount),
)

// ❌ 避免 - 格式化日志（除非必要）
logger.Infof(ctx, "订单 %s 创建成功，金额 %d", orderID, amount)
```

### 2. 使用业务语义字段

```go
// ✅ 推荐 - 语义化字段
logger.UserID("123")
logger.RequestID("req_456") 
logger.Latency(duration)

// ❌ 避免 - 通用字段
logger.String("user_id", "123")
logger.String("request_id", "req_456")
logger.Duration("latency", duration)
```

### 3. 合理使用子日志器

```go
// ✅ 对于模块化服务
serviceLogger := logger.Default().With(
    logger.Service("payment-service"),
    logger.Version("v1.2.0"),
)

// ✅ 对于特定上下文  
requestLogger := logger.Default().WithContext(ctx,
    logger.RequestID(reqID),
    logger.UserID(userID),
)
```

## 🔧 运行示例

```bash
cd examples/logger_usage
go run main.go
```

## 📝 输出示例

```json
{"level":"info","ts":"2025-01-07T14:30:15.123Z","caller":"main.go:45","msg":"服务启动成功","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","span_id":"00f067aa0ba902b7","service":"logger-demo","version":"v1.0.0","port":8080}

{"level":"info","ts":"2025-01-07T14:30:15.124Z","caller":"main.go:50","msg":"用户 alice 登录成功，耗时 250ms","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","span_id":"00f067aa0ba902b7"}

{"level":"info","ts":"2025-01-07T14:30:15.125Z","caller":"main.go:53","msg":"处理用户请求","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","span_id":"00f067aa0ba902b7","user_id":"user123","request_id":"req456","method":"POST","path":"/api/users","status_code":201,"latency":"150ms"}
```

## ✅ 改进总结

1. **🎯 统一接口**：合并重复接口，API更简洁
2. **🚀 自动追踪**：无需手动添加trace_id
3. **📂 职责分离**：logger.go专注日志，fields.go负责字段
4. **🔧 易于使用**：包级函数为主，符合Go习惯
5. **⚡ 高性能**：基于Zap零分配设计
6. **🛡️ 类型安全**：强类型字段，减少错误
7. **📖 良好文档**：清晰的使用指南和示例 