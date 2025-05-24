# 🛡️ API限流和熔断器使用指南

## 📋 功能概述

本框架集成了完整的API限流和熔断器功能，提供多层次的系统保护机制：

- **🚫 API限流**: 防止恶意请求和系统过载
- **🔥 熔断器**: 防止服务雪崩和级联故障
- **📊 实时监控**: 提供详细的限流和熔断状态监控

## 🚀 快速开始

### 1. 基本使用

限流和熔断器已自动集成到所有API端点中，无需额外配置即可使用。

### 2. 测试功能

```bash
# 运行完整的限流和熔断器测试
./scripts/test-ratelimit-circuitbreaker.sh

# 查看熔断器状态
curl http://localhost:8080/circuit-breaker/status

# 查看Hystrix监控流
curl http://localhost:8080/hystrix
```

## 📊 限流功能

### 限流类型

#### 1. IP限流
基于客户端IP地址的限流，适用于公开API端点。

```go
// 示例：每秒10次请求
r.GET("/health", rateLimiter.IPRateLimit("10-S"), handler)
```

#### 2. 用户限流
基于认证用户ID的限流，适用于需要登录的API。

```go
// 示例：每分钟50次请求
r.GET("/api/v1/users/me", rateLimiter.UserRateLimit("50-M"), handler)
```

#### 3. 自定义限流
基于自定义键值函数的限流。

```go
// 按端点限流
r.GET("/api", rateLimiter.CustomRateLimit(ratelimit.KeyFunctions.ByEndpoint, "100-M"), handler)
```

### 限流配置格式

限流配置使用 `数量-时间单位` 格式：

- `10-S`: 每秒10次
- `100-M`: 每分钟100次  
- `1000-H`: 每小时1000次
- `5000-D`: 每天5000次

### 当前限流配置

| 端点类型 | 限流规则 | 说明 |
|----------|----------|------|
| 健康检查 | 10次/秒 | 基于IP限流 |
| 认证端点 | 20次/分钟 | 基于IP限流 |
| 受保护认证 | 10次/分钟 | 基于用户限流 |
| 公开用户API | 30次/分钟 | 基于IP限流 |
| 受保护用户API | 50次/分钟 | 基于用户限流 |

### 限流响应

当触发限流时，API返回HTTP 429状态码：

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Limit: 10 per 1m0s",
  "retry_after": 1640995200
}
```

响应头包含限流信息：
- `X-RateLimit-Limit`: 限流阈值
- `X-RateLimit-Remaining`: 剩余请求数
- `X-RateLimit-Reset`: 重置时间戳

## 🔥 熔断器功能

### 熔断器配置

#### 默认配置

| 服务类型 | 超时时间 | 最大并发 | 请求阈值 | 错误率阈值 | 休眠窗口 |
|----------|----------|----------|----------|------------|----------|
| 数据库 | 5秒 | 100 | 20 | 50% | 10秒 |
| 外部API | 3秒 | 50 | 10 | 30% | 5秒 |
| 缓存 | 1秒 | 200 | 5 | 60% | 3秒 |

#### API专用配置

| API类型 | 超时时间 | 最大并发 | 请求阈值 | 错误率阈值 | 休眠窗口 |
|---------|----------|----------|----------|------------|----------|
| 用户注册 | 3秒 | 20 | 10 | 30% | 5秒 |
| 用户登录 | 3秒 | 30 | 15 | 25% | 5秒 |
| 用户查询 | 2秒 | 100 | 20 | 40% | 3秒 |

### 熔断器状态

#### 状态类型
- **CLOSED**: 正常状态，请求正常通过
- **OPEN**: 熔断状态，请求被拒绝并执行降级
- **HALF_OPEN**: 半开状态，允许少量请求测试服务恢复

#### 状态查询

```bash
# 查看所有熔断器状态
curl http://localhost:8080/circuit-breaker/status
```

响应示例：
```json
{
  "status": "healthy",
  "open_circuits": [],
  "total_circuits": 6,
  "states": {
    "auth_login": {
      "name": "auth_login",
      "is_open": false,
      "request_count": 0,
      "error_count": 0,
      "error_percentage": 0
    }
  }
}
```

### 熔断器降级

当熔断器打开时，API返回HTTP 503状态码：

```json
{
  "error": "Service Unavailable",
  "message": "The service is temporarily unavailable due to circuit breaker",
  "code": "CIRCUIT_BREAKER_OPEN"
}
```

## 📈 监控和指标

### Hystrix监控流

访问 `http://localhost:8080/hystrix` 获取实时监控数据流，可用于：

- Hystrix Dashboard
- 自定义监控系统
- 实时告警

### Prometheus指标

熔断器会自动记录以下指标：

```promql
# 熔断器请求计数
circuit_breaker_requests_total{command="auth_login",status="success"}

# 熔断器错误计数  
circuit_breaker_requests_total{command="auth_login",status="error"}
```

## 🔧 高级配置

### 自定义熔断器

```go
// 配置自定义熔断器
circuitbreaker.ConfigureCommand("my_service", circuitbreaker.Config{
    Timeout:                2000, // 2秒超时
    MaxConcurrentRequests:  50,   // 最大50并发
    RequestVolumeThreshold: 10,   // 10个请求后开始统计
    SleepWindow:            5000, // 5秒休眠窗口
    ErrorPercentThreshold:  30,   // 30%错误率阈值
})
```

### 自定义降级处理

```go
// 自定义降级函数
fallbackHandler := func(c *gin.Context) {
    c.JSON(503, gin.H{
        "error": "Service temporarily unavailable",
        "fallback": true,
    })
}

// 使用自定义降级
r.GET("/api", circuitBreaker.Middleware("my_command", fallbackHandler), handler)
```

### 编程式使用

```go
// 直接使用熔断器保护代码
result, err := circuitBreaker.ExecuteCommand("my_operation", 
    func() (interface{}, error) {
        // 业务逻辑
        return doSomething()
    }, 
    func(err error) (interface{}, error) {
        // 降级逻辑
        return getDefaultValue(), nil
    })
```

## 🧪 测试和验证

### 限流测试

```bash
# 快速发送多个请求测试限流
for i in {1..20}; do
    curl -w "HTTP_%{http_code}\n" http://localhost:8080/health
    sleep 0.1
done
```

### 熔断器测试

```bash
# 发送大量错误请求触发熔断器
for i in {1..30}; do
    curl -w "HTTP_%{http_code}\n" http://localhost:8080/api/v1/users/999
    sleep 0.1
done
```

### 自动化测试

```bash
# 运行完整测试套件
./scripts/test-ratelimit-circuitbreaker.sh
```

## 🚨 故障排查

### 常见问题

#### 1. 限流不生效
- 检查限流配置格式是否正确
- 确认中间件是否正确应用到路由
- 查看应用日志中的错误信息

#### 2. 熔断器未触发
- 确认请求量是否达到阈值
- 检查错误率是否超过配置值
- 验证熔断器配置是否正确

#### 3. 监控数据缺失
- 确认Hystrix流端点可访问
- 检查Prometheus指标是否正常暴露
- 验证网络连接和防火墙设置

### 调试命令

```bash
# 查看限流状态
curl -I http://localhost:8080/health

# 查看熔断器状态
curl http://localhost:8080/circuit-breaker/status

# 查看应用日志
docker-compose logs -f app

# 查看Prometheus指标
curl http://localhost:9090/metrics | grep circuit_breaker
```

## 📚 最佳实践

### 限流配置建议

1. **分层限流**: 不同类型的API使用不同的限流策略
2. **合理阈值**: 根据系统容量和业务需求设置限流阈值
3. **用户友好**: 提供清晰的限流错误信息和重试建议
4. **监控告警**: 设置限流触发的监控告警

### 熔断器配置建议

1. **快速失败**: 设置合理的超时时间，避免长时间等待
2. **渐进恢复**: 使用半开状态逐步恢复服务
3. **降级策略**: 为每个关键服务准备降级方案
4. **监控观察**: 持续监控熔断器状态和触发频率

### 生产环境建议

1. **性能测试**: 在生产环境部署前进行充分的性能测试
2. **容量规划**: 根据业务增长预期调整限流和熔断配置
3. **告警设置**: 配置限流和熔断的实时告警
4. **定期回顾**: 定期回顾和优化配置参数

## 🔗 相关文档

- [分布式链路追踪指南](TRACING.md)
- [监控指标说明](../README-Docker.md#监控和指标)
- [部署文档](../README-Docker.md)
- [项目路线图](ROADMAP.md) 