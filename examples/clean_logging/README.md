# 简洁日志输出示例

这个示例展示了如何配置**简洁而有用**的日志输出，避免冗长的 OpenTelemetry span 信息。

## 🚨 **问题分析**

之前你看到的冗长日志是因为：
```go
WithTracing(&config.TracingConfig{
    ExporterType: "stdout", // ❌ 这会输出完整的span JSON到控制台
})
```

## ✅ **优化解决方案**

### 1. **关键配置改动**

```go
WithTracing(&config.TracingConfig{
    ServiceName:    "clean-logging-demo",
    ServiceVersion: "v1.0.0", 
    Environment:    "development",
    Enabled:        true,
    ExporterType:   "none", // 🎯 关键：使用 "none" 避免span详情输出
    SampleRatio:    1.0,
})
```

### 2. **支持的导出器类型**

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| `"none"` | 🎯 **推荐** - 不输出span详情，只保留trace_id | 开发/生产环境，简洁日志 |
| `"noop"` | 完全禁用追踪输出 | 性能测试 |
| `"jaeger"` | 发送到Jaeger收集器 | 生产环境APM |
| `"stdout"` | ❌ 输出完整span JSON | 仅调试追踪系统时使用 |

## 🎯 **优化后的日志格式**

### HTTP访问日志
```json
{
  "level": "info",
  "ts": "2025-06-22T10:00:00.123+0800",
  "caller": "clean_logging/main.go:85",
  "msg": "HTTP request completed",
  "trace_id": "7f693971b0217d8476879db2f737a548",
  "span_id": "afa5a3713d20ad94", 
  "method": "GET",
  "path": "/api/v1/users/123",
  "status": 200,
  "latency": "52.3ms",
  "ip": "127.0.0.1"
}
```

### 业务逻辑日志
```json
{
  "level": "info",
  "ts": "2025-06-22T10:00:00.150+0800",
  "caller": "clean_logging/main.go:105",
  "msg": "Processing get user request",
  "trace_id": "7f693971b0217d8476879db2f737a548",
  "span_id": "afa5a3713d20ad94",
  "user_id": "123"
}
```

### Warning级别日志
```json
{
  "level": "warn",
  "ts": "2025-06-22T10:00:00.200+0800", 
  "caller": "clean_logging/main.go:115",
  "msg": "User not found",
  "trace_id": "7f693971b0217d8476879db2f737a548",
  "span_id": "afa5a3713d20ad94",
  "user_id": "404"
}
```

## 🚀 **运行测试**

```bash
cd examples/clean_logging
go run main.go
```

### 测试API

```bash
# 正常请求
curl http://localhost:8080/api/v1/users/123

# 触发404警告
curl http://localhost:8080/api/v1/users/404

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# 健康检查
curl http://localhost:8080/api/v1/health
```

## ✨ **关键优势**

1. **✅ 保留trace_id**: 仍然可以追踪请求链路
2. **✅ 简洁输出**: 没有冗长的span JSON
3. **✅ 包含caller**: Warning日志正确显示代码位置
4. **✅ 结构化**: JSON格式便于日志聚合分析
5. **✅ 性能**: 减少日志输出，提升性能

## 🔧 **生产环境建议**

### 开发环境
```go
WithTracing(&config.TracingConfig{
    ExporterType: "none",  // 简洁日志
    SampleRatio:  1.0,     // 100%采样用于开发
})
```

### 生产环境
```go
WithTracing(&config.TracingConfig{
    ExporterType: "jaeger",                    // 发送到APM系统
    Endpoint:     "http://jaeger:14268/api/traces",
    SampleRatio:  0.1,                         // 10%采样降低开销
})
```

### 性能测试
```go
WithTracing(&config.TracingConfig{
    ExporterType: "noop",  // 完全禁用
    SampleRatio:  0.0,     // 0%采样
})
```

## 📊 **对比效果**

| 配置 | 输出长度 | trace_id | caller | 性能 |
|------|----------|----------|--------|------|
| `ExporterType: "stdout"` | ❌ 极长(>2KB) | ✅ | ✅ | ❌ 慢 |
| `ExporterType: "none"` | ✅ 简洁(<200B) | ✅ | ✅ | ✅ 快 |
| `ExporterType: "noop"` | ✅ 最简(<100B) | ✅ | ✅ | ✅ 最快 |

## 🎯 **推荐方案**

对于你的使用场景，推荐：

```go
// 日常开发
ExporterType: "none"

// 生产部署  
ExporterType: "jaeger"
Endpoint: "your-jaeger-endpoint"
```

这样既保留了完整的追踪能力，又避免了冗长的控制台输出！ 