# 链路追踪日志示例

这个示例展示了如何在分布式微服务框架中使用带 trace id 的日志功能。

## 🎯 功能特性

- ✅ **自动 Trace ID 注入**: 所有日志自动包含 trace_id 和 span_id
- ✅ **HTTP 请求日志**: 自动记录 HTTP 请求和响应信息
- ✅ **跨服务追踪**: 支持跨服务调用的 trace id 传播
- ✅ **JSON 格式日志**: 便于日志聚合和分析
- ✅ **多级别日志**: 支持 Debug、Info、Warn、Error、Fatal 等级别

## 🚀 快速开始

### 运行示例

```bash
cd examples/trace_logging
go run main.go
```

### 测试 API

```bash
# 获取用户信息
curl http://localhost:8080/api/v1/users/123

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# 获取订单信息
curl http://localhost:8080/api/v1/orders/order123

# 健康检查
curl http://localhost:8080/api/v1/health
```

## 📝 日志输出示例

### 带 Trace ID 的 JSON 日志

```json
{
  "level": "info",
  "ts": "2025-01-07T14:30:15.123Z",
  "caller": "main.go:85",
  "msg": "Getting user information",
  "trace_id": "a1b2c3d4e5f67890abcdef1234567890",
  "span_id": "1234567890abcdef",
  "user_id": "123",
  "operation": "get_user"
}
```

### HTTP 请求日志

```json
{
  "level": "info",
  "ts": "2025-01-07T14:30:15.100Z",
  "caller": "tracing.go:95",
  "msg": "HTTP request started",
  "trace_id": "a1b2c3d4e5f67890abcdef1234567890",
  "span_id": "1234567890abcdef",
  "method": "GET",
  "path": "/api/v1/users/123",
  "client_ip": "127.0.0.1",
  "user_agent": "curl/7.68.0"
}
```

## 🔧 使用方法

### 1. 基础日志记录

```go
func yourHandler(c *gin.Context) {
    ctx := c.MustGet("ctx").(context.Context)
    
    // 自动包含 trace_id 和 span_id
    logger.Info(ctx, "Processing request",
        logger.String("user_id", "123"),
        logger.String("action", "get_profile"),
    )
}
```

### 2. 错误日志记录

```go
if err != nil {
    logger.Error(ctx, "Database operation failed",
        logger.String("operation", "select"),
        logger.String("table", "users"),
        logger.Error_(err),
    )
}
```

### 3. 结构化日志字段

```go
logger.Info(ctx, "User operation completed",
    logger.String("user_id", userID),
    logger.String("operation", "create"),
    logger.Duration("duration", time.Since(start)),
    logger.Int("affected_rows", 1),
    logger.Bool("success", true),
)
```

### 4. 使用 ContextLogger 接口

```go
// 获取支持上下文的日志器
ctxLogger := logger.GetContextLogger()

// 直接传入 context，自动添加 trace id
ctxLogger.InfoCtx(ctx, "Operation completed")
ctxLogger.ErrorCtx(ctx, "Operation failed", logger.Error_(err))
```

## 🏗️ 架构说明

### 中间件链

```
HTTP Request
    ↓
TraceContextMiddleware   # 确保 trace context
    ↓
LoggingMiddleware       # 记录请求日志
    ↓
Business Logic          # 业务逻辑
    ↓
LoggingMiddleware       # 记录响应日志
    ↓
HTTP Response (含 X-Trace-ID header)
```

### 核心组件

1. **`logger` 包**: 提供带 trace id 的日志功能
2. **`tracing` 包**: 提供链路追踪工具函数
3. **`middleware` 包**: 提供各种中间件
4. **`transport/http` 包**: HTTP 响应处理

### 关键接口

```go
// 基础日志接口
func Info(ctx context.Context, msg string, fields ...zapcore.Field)
func Error(ctx context.Context, msg string, fields ...zapcore.Field)

// 上下文日志器接口
type ContextLogger interface {
    InfoCtx(ctx context.Context, msg string, fields ...zapcore.Field)
    ErrorCtx(ctx context.Context, msg string, fields ...zapcore.Field)
    // ...
}

// Trace ID 工具函数
func GetTraceID(ctx context.Context) string
func GetSpanID(ctx context.Context) string
```

## 📊 日志分析

### 查询特定请求的所有日志

```bash
# 假设使用 jq 分析日志
cat application.log | jq 'select(.trace_id == "a1b2c3d4e5f67890abcdef1234567890")'
```

### 统计错误日志

```bash
cat application.log | jq 'select(.level == "error") | .trace_id' | sort | uniq -c
```

## 🎯 最佳实践

### 1. 始终传递 Context

```go
// ✅ 正确
func businessLogic(ctx context.Context, userID string) error {
    logger.Info(ctx, "Starting business logic", logger.String("user_id", userID))
    return nil
}

// ❌ 错误 - 缺少 context
func businessLogic(userID string) error {
    log.Println("Starting business logic") // 无法关联到请求
    return nil
}
```

### 2. 使用结构化日志字段

```go
// ✅ 正确 - 结构化字段
logger.Info(ctx, "User operation",
    logger.String("operation", "create"),
    logger.String("user_id", userID),
    logger.Duration("duration", duration),
)

// ❌ 错误 - 字符串拼接
logger.Info(ctx, fmt.Sprintf("User %s created in %v", userID, duration))
```

### 3. 合理的日志级别

```go
// Debug: 调试信息
logger.Debug(ctx, "Database query", logger.String("sql", query))

// Info: 重要业务操作
logger.Info(ctx, "User login", logger.String("user_id", userID))

// Warn: 潜在问题
logger.Warn(ctx, "Rate limit approaching", logger.Int("current", current))

// Error: 错误情况
logger.Error(ctx, "Database connection failed", logger.Error_(err))
```

## 🔍 故障排查

当出现问题时，你可以：

1. **通过 Trace ID 查询**: 找到特定请求的所有相关日志
2. **跨服务追踪**: 同一个 trace id 可以跨多个服务
3. **性能分析**: 通过 span 信息分析各环节耗时
4. **错误定位**: 快速定位错误发生的具体位置

## 📚 相关文档

- [OpenTelemetry 官方文档](https://opentelemetry.io/)
- [Zap 日志库文档](https://github.com/uber-go/zap)
- [Gin 框架文档](https://gin-gonic.com/) 