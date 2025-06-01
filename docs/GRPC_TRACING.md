# gRPC 分布式追踪集成

本文档描述了如何在分布式服务中实现 gRPC 的分布式追踪功能。

## 概述

分布式追踪允许我们跟踪请求在微服务架构中的完整路径，帮助识别性能瓶颈、调试问题和监控系统健康状况。我们的 gRPC 服务集成了 OpenTelemetry 来提供完整的追踪功能。

## 架构组件

### 1. 追踪中间件 (pkg/middleware/grpc.go)

#### 服务端拦截器

- **GRPCTracingInterceptor**: 一元 RPC 调用的追踪拦截器
- **GRPCStreamTracingInterceptor**: 流式 RPC 调用的追踪拦截器

```go
// 为每个 gRPC 调用创建 span
func GRPCTracingInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 开始 span
        spanName := "gRPC " + info.FullMethod
        ctx, span := tracing.StartSpan(ctx, spanName)
        defer span.End()

        // 添加 gRPC 特定属性
        span.SetAttributes(
            attribute.String("rpc.system", "grpc"),
            attribute.String("rpc.service", getServiceName(info.FullMethod)),
            attribute.String("rpc.method", getMethodName(info.FullMethod)),
        )

        // 执行处理器
        resp, err := handler(ctx, req)
        
        // 记录错误
        if err != nil {
            tracing.RecordError(ctx, err)
        }

        return resp, err
    }
}
```

#### 客户端拦截器

```go
// 客户端追踪拦截器
func grpcTracingInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        // 创建客户端 span
        tracer := otel.Tracer("grpc-client")
        ctx, span := tracer.Start(ctx, "gRPC Call: "+method)
        defer span.End()

        // 注入追踪上下文到 metadata
        md, ok := metadata.FromOutgoingContext(ctx)
        if !ok {
            md = metadata.New(nil)
        }
        
        propagator := otel.GetTextMapPropagator()
        propagator.Inject(ctx, &metadataCarrier{md})
        ctx = metadata.NewOutgoingContext(ctx, md)

        // 执行调用
        err := invoker(ctx, method, req, reply, cc, opts...)
        if err != nil {
            span.RecordError(err)
        }

        return err
    }
}
```

### 2. 追踪工具 (pkg/tracing/tracer.go)

提供了 gRPC 特定的追踪辅助函数：

```go
// 追踪 gRPC 调用
func TraceGRPC(ctx context.Context, method, service string, statusCode grpccodes.Code) {
    span := trace.SpanFromContext(ctx)
    span.SetAttributes(
        attribute.String("rpc.system", "grpc"),
        attribute.String("rpc.service", service),
        attribute.String("rpc.method", method),
        attribute.String("rpc.grpc.status_code", statusCode.String()),
    )
}

// 追踪 gRPC 客户端调用
func TraceGRPCClient(ctx context.Context, target, method string, statusCode grpccodes.Code) {
    span := trace.SpanFromContext(ctx)
    span.SetAttributes(
        attribute.String("rpc.system", "grpc"),
        attribute.String("rpc.target", target),
        attribute.String("rpc.method", method),
        attribute.String("rpc.grpc.status_code", statusCode.String()),
        attribute.String("component", "grpc-client"),
    )
}
```

### 3. 服务实现追踪 (internal/grpc/user_service.go)

在 gRPC 服务方法中添加详细的追踪信息：

```go
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 添加追踪属性
    tracing.AddSpanAttributes(ctx,
        attribute.String("grpc.service", "user.v1.UserService"),
        attribute.String("grpc.method", "GetUser"),
        attribute.Int64("user.id", int64(req.Id)),
    )

    // 业务逻辑...
    user, err := s.userService.GetByID(ctx, uint(req.Id))
    if err != nil {
        tracing.RecordError(ctx, err)
        tracing.AddSpanAttributes(ctx, attribute.String("error.type", "not_found"))
        return nil, status.Errorf(codes.NotFound, "user not found")
    }

    // 成功时添加属性
    tracing.AddSpanAttributes(ctx,
        attribute.String("user.username", user.Username),
        attribute.Bool("operation.success", true),
    )

    return &pb.GetUserResponse{User: convertUserToProto(user)}, nil
}
```

## 追踪属性标准

### gRPC 服务端属性

| 属性名 | 类型 | 描述 | 示例 |
|--------|------|------|------|
| `rpc.system` | string | RPC 系统类型 | "grpc" |
| `rpc.service` | string | 服务名称 | "user.v1.UserService" |
| `rpc.method` | string | 方法名称 | "GetUser" |
| `rpc.grpc.status_code` | string | gRPC 状态码 | "OK", "NOT_FOUND" |
| `request.id` | string | 请求 ID | "trace-id-123" |
| `user.id` | int64 | 用户 ID | 12345 |
| `error.type` | string | 错误类型 | "invalid_argument" |

### gRPC 客户端属性

| 属性名 | 类型 | 描述 | 示例 |
|--------|------|------|------|
| `rpc.system` | string | RPC 系统类型 | "grpc" |
| `rpc.target` | string | 目标服务地址 | "localhost:9090" |
| `rpc.method` | string | 调用的方法 | "/user.v1.UserService/GetUser" |
| `component` | string | 组件类型 | "grpc-client" |

## 配置和使用

### 1. 服务端配置

在 `pkg/grpc/server.go` 中，追踪拦截器已经集成到拦截器链中：

```go
grpc.ChainUnaryInterceptor(
    middleware.GRPCTracingInterceptor(), // 分布式追踪
    middleware.GRPCLoggingInterceptor(),  // 日志记录
    middleware.GRPCRecoveryInterceptor(), // 恐慌恢复
    middleware.GRPCMetricsInterceptor(),  // 指标收集
),
```

### 2. 客户端配置

在客户端代码中添加追踪拦截器：

```go
conn, err := grpc.NewClient("localhost:9090",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithUnaryInterceptor(grpcTracingInterceptor()),
)
```

### 3. 追踪初始化

```go
func initTracing() func() {
    exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
    if err != nil {
        log.Fatalf("Failed to create trace exporter: %v", err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("grpc-service"),
            semconv.ServiceVersionKey.String("1.0.0"),
        )),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return func() {
        tp.Shutdown(context.Background())
    }
}
```

## 测试和验证

### 1. 运行测试脚本

```bash
./scripts/test-grpc-tracing.sh
```

### 2. 手动测试

```bash
# 启动服务器
go run ./cmd/server

# 运行客户端示例
go run ./examples/grpc-client

# 使用 grpcurl 测试
grpcurl -plaintext -d '{"username":"test","email":"test@example.com","password":"pass"}' \
    localhost:9090 user.v1.UserService/CreateUser
```

### 3. 验证追踪输出

追踪输出应该包含：
- Trace ID 和 Span ID
- gRPC 方法信息
- 请求/响应属性
- 错误信息（如果有）
- 执行时间

示例输出：
```json
{
  "Name": "gRPC /user.v1.UserService/GetUser",
  "SpanContext": {
    "TraceID": "4bf92f3577b34da6a3ce929d0e0e4736",
    "SpanID": "00f067aa0ba902b7"
  },
  "Attributes": [
    {"Key": "rpc.system", "Value": {"Type": "STRING", "Value": "grpc"}},
    {"Key": "rpc.service", "Value": {"Type": "STRING", "Value": "user.v1.UserService"}},
    {"Key": "rpc.method", "Value": {"Type": "STRING", "Value": "GetUser"}},
    {"Key": "user.id", "Value": {"Type": "INT64", "Value": 1}}
  ]
}
```

## 生产环境配置

### 1. Jaeger 集成

```go
import (
    "go.opentelemetry.io/otel/exporters/jaeger"
)

func initJaegerTracing() func() {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
    ))
    if err != nil {
        log.Fatal(err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("grpc-service"),
        )),
    )

    otel.SetTracerProvider(tp)
    return func() { tp.Shutdown(context.Background()) }
}
```

### 2. OTLP 集成

```go
import (
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func initOTLPTracing() func() {
    exp, err := otlptracehttp.New(context.Background(),
        otlptracehttp.WithEndpoint("http://otel-collector:4318"),
    )
    if err != nil {
        log.Fatal(err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String("grpc-service"),
        )),
    )

    otel.SetTracerProvider(tp)
    return func() { tp.Shutdown(context.Background()) }
}
```

### 3. 环境变量配置

```bash
# OpenTelemetry 配置
export OTEL_SERVICE_NAME="grpc-service"
export OTEL_SERVICE_VERSION="1.0.0"
export OTEL_EXPORTER_OTLP_ENDPOINT="http://otel-collector:4317"
export OTEL_EXPORTER_OTLP_PROTOCOL="grpc"

# Jaeger 配置
export JAEGER_AGENT_HOST="jaeger"
export JAEGER_AGENT_PORT="6831"
export JAEGER_SAMPLER_TYPE="const"
export JAEGER_SAMPLER_PARAM="1"
```

## 最佳实践

### 1. Span 命名

- 使用描述性的 span 名称：`gRPC /service.v1.Service/Method`
- 包含服务和方法信息
- 保持一致的命名约定

### 2. 属性添加

- 添加有意义的业务属性
- 避免添加敏感信息（密码、令牌等）
- 使用标准的属性名称

### 3. 错误处理

- 总是记录错误到 span
- 添加错误类型属性
- 设置适当的 span 状态

### 4. 性能考虑

- 使用采样来控制追踪开销
- 批量导出 span 数据
- 监控追踪系统的性能影响

### 5. 安全性

- 不要在 span 中记录敏感数据
- 使用安全的传输协议
- 实施适当的访问控制

## 故障排除

### 常见问题

1. **追踪数据未显示**
   - 检查 exporter 配置
   - 验证网络连接
   - 确认采样配置

2. **Span 链断裂**
   - 检查上下文传播
   - 验证 metadata 注入
   - 确认拦截器顺序

3. **性能影响**
   - 调整采样率
   - 优化批量大小
   - 监控内存使用

### 调试命令

```bash
# 检查 gRPC 服务
grpcurl -plaintext localhost:9090 list

# 测试健康检查
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check

# 查看追踪输出
export OTEL_LOG_LEVEL=debug
go run ./examples/grpc-client
```

## 相关文档

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [gRPC Go 文档](https://grpc.io/docs/languages/go/)
- [分布式追踪最佳实践](https://opentelemetry.io/docs/concepts/observability-primer/)
- [Jaeger 部署指南](https://www.jaegertracing.io/docs/deployment/) 