# 🔍 分布式链路追踪 (Distributed Tracing)

本文档介绍如何使用和配置分布式链路追踪功能。

## 📋 概述

分布式链路追踪使用 **OpenTelemetry + Jaeger** 实现，提供以下功能：

- 🔍 **请求链路可视化**: 完整追踪 HTTP 请求在微服务中的调用路径
- 📊 **性能分析**: 识别性能瓶颈和延迟问题
- 🐛 **错误追踪**: 快速定位分布式系统中的错误
- 📈 **服务依赖分析**: 了解服务间的调用关系

## 🏗️ 架构组件

```
HTTP Request → Gin Middleware → Service Layer → Repository Layer → Database
      ↓              ↓              ↓              ↓              ↓
   Trace ID      HTTP Span     Service Span   Repository Span  DB Span
```

### 核心组件

1. **OpenTelemetry SDK**: 追踪数据收集和导出
2. **Jaeger**: 追踪数据存储和可视化
3. **Gin 中间件**: 自动为 HTTP 请求创建 span
4. **Service/Repository 追踪**: 业务逻辑和数据访问层追踪

## ⚙️ 配置说明

### 配置文件 (`config/config.yaml`)

```yaml
tracing:
  service_name: distributed-service      # 服务名称
  service_version: v1.0.0               # 服务版本
  environment: development              # 环境标识
  enabled: true                         # 是否启用追踪
  exporter_type: stdout                 # 导出器类型: "otlp", "stdout"
  endpoint: http://localhost:4318/v1/traces  # OTLP endpoint
  sample_ratio: 1.0                     # 采样率 (0.0-1.0)
```

### 导出器类型

- **stdout**: 输出到控制台，适合开发调试
- **otlp**: 发送到 Jaeger，适合生产环境

### 采样率建议

- **开发环境**: `1.0` (100% 采样)
- **测试环境**: `0.5` (50% 采样)
- **生产环境**: `0.1` (10% 采样)

## 🚀 快速开始

### 1. 启动服务

```bash
# 启动所有服务 (包括 Jaeger)
docker-compose up -d

# 或者仅启动应用
go run main.go
```

### 2. 运行测试

```bash
# 执行追踪测试脚本
./scripts/test-tracing.sh
```

### 3. 查看追踪数据

访问 Jaeger UI: http://localhost:16686

1. 在 Service 下拉框中选择 `distributed-service`
2. 点击 "Find Traces" 查看追踪数据
3. 点击具体的 trace 查看详细信息

## 📊 追踪数据结构

### Span 层次结构

```
HTTP Request Span
├── userService.Register
│   ├── userRepository.GetByUsername
│   └── userRepository.Create
├── userService.Login
│   └── userRepository.GetByUsername
└── userService.ChangePassword
    ├── userRepository.GetByID
    └── userRepository.Update
```

### Span 属性

#### HTTP Span 属性
- `http.method`: HTTP 方法
- `http.route`: 路由路径
- `http.status_code`: 响应状态码
- `http.user_agent`: 用户代理
- `request.id`: 请求 ID

#### Service Span 属性
- `user.username`: 用户名
- `user.email`: 邮箱
- `user.id`: 用户 ID

#### Database Span 属性
- `db.operation`: 数据库操作 (SELECT, INSERT, UPDATE, DELETE)
- `db.table`: 表名
- `db.system`: 数据库系统 (mysql)
- `db.rows_affected`: 影响行数

## 🛠️ 开发指南

### 在代码中添加追踪

#### 1. 基本 Span 创建

```go
import "distributed-service/pkg/tracing"

func MyFunction(ctx context.Context) error {
    ctx, span := tracing.StartSpan(ctx, "MyFunction")
    defer span.End()
    
    // 添加属性
    tracing.AddSpanAttributes(ctx, 
        attribute.String("key", "value"),
        attribute.Int("count", 42),
    )
    
    // 业务逻辑...
    
    return nil
}
```

#### 2. 使用 WithSpan 辅助函数

```go
func MyFunction(ctx context.Context) error {
    return tracing.WithSpan(ctx, "MyFunction", func(ctx context.Context) error {
        // 业务逻辑...
        return nil
    })
}
```

#### 3. 带返回值的 Span

```go
func MyFunction(ctx context.Context) (*Result, error) {
    return tracing.WithSpanResult(ctx, "MyFunction", func(ctx context.Context) (*Result, error) {
        // 业务逻辑...
        return &Result{}, nil
    })
}
```

#### 4. 错误处理

```go
func MyFunction(ctx context.Context) error {
    ctx, span := tracing.StartSpan(ctx, "MyFunction")
    defer span.End()
    
    if err := someOperation(); err != nil {
        tracing.RecordError(ctx, err)
        return err
    }
    
    return nil
}
```

### 专用追踪函数

```go
// 数据库操作追踪
tracing.TraceDatabase(ctx, "SELECT", "users", 1)

// 缓存操作追踪
tracing.TraceCache(ctx, "GET", "user:123", true)

// 消息队列追踪
tracing.TraceMessageQueue(ctx, "PUBLISH", "user.events", 1)

// HTTP 请求追踪
tracing.TraceHTTPRequest(ctx, "POST", "/api/users", 201)
```

## 🔧 故障排除

### 常见问题

#### 1. 看不到追踪数据

**检查项目:**
- 确认 `tracing.enabled: true`
- 检查 Jaeger 服务是否正常运行
- 验证 endpoint 配置是否正确
- 检查采样率是否过低

#### 2. Span 数据不完整

**可能原因:**
- Context 没有正确传递
- Span 没有正确结束 (缺少 `defer span.End()`)
- 属性设置在 span 结束之后

#### 3. 性能影响

**优化建议:**
- 降低生产环境采样率
- 避免在高频函数中创建过多 span
- 合理设置 span 属性数量

### 调试命令

```bash
# 检查 Jaeger 服务状态
docker-compose ps jaeger

# 查看 Jaeger 日志
docker-compose logs jaeger

# 测试 OTLP endpoint
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{}'
```

## 📈 最佳实践

### 1. Span 命名规范

- **HTTP Span**: `HTTP {method} {route}`
- **Service Span**: `{serviceName}.{methodName}`
- **Repository Span**: `{repositoryName}.{methodName}`
- **Database Span**: `db.{operation}.{table}`

### 2. 属性设置

- 使用语义化的属性名
- 避免包含敏感信息 (密码、token)
- 合理控制属性数量和大小

### 3. 错误处理

- 始终记录错误到 span
- 设置适当的 span 状态
- 包含足够的上下文信息

### 4. 性能考虑

- 生产环境使用合适的采样率
- 避免在循环中创建大量 span
- 定期清理 Jaeger 存储数据

## 🔗 相关链接

- [OpenTelemetry 官方文档](https://opentelemetry.io/docs/)
- [Jaeger 官方文档](https://www.jaegertracing.io/docs/)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [Gin OpenTelemetry 中间件](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/gin-gonic/gin/otelgin) 