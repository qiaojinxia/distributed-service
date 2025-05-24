# 限流功能文档

## 概述

本项目实现了基于配置文件的API限流功能，支持多种限流策略和存储后端。

## 功能特性

- **多种限流策略**: IP限流、用户限流、自定义限流
- **配置化管理**: 通过YAML配置文件管理所有限流规则
- **多存储后端**: 支持内存存储和Redis存储
- **端点级配置**: 为不同API端点设置不同的限流规则
- **响应头信息**: 提供详细的限流状态信息
- **优雅降级**: 限流器故障时不影响正常业务

## 配置说明

### 基本配置

```yaml
ratelimit:
  enabled: true                    # 是否启用限流
  store_type: redis               # 存储类型: memory, redis
  redis_prefix: "ratelimit:"      # Redis键前缀
  default_config:                 # 默认限流配置
    health_check: "10-S"          # 健康检查：每秒10次
    auth_public: "20-M"           # 认证公开端点：每分钟20次
    auth_protected: "10-M"        # 认证保护端点：每分钟10次
    user_public: "30-M"           # 用户公开API：每分钟30次
    user_protected: "50-M"        # 用户保护API：每分钟50次
  endpoints:                      # 端点特定配置
    "/health": "10-S"
    "/api/v1/auth/register": "20-M"
    "/api/v1/auth/login": "20-M"
    "/api/v1/auth/refresh": "20-M"
    "/api/v1/auth/change-password": "10-M"
    "/api/v1/users/me": "50-M"
    "/api/v1/users": "30-M"
```

### 限流格式说明

限流配置使用 `数量-时间单位` 格式：

- `10-S`: 每秒10次
- `20-M`: 每分钟20次
- `100-H`: 每小时100次
- `1000-D`: 每天1000次

支持的时间单位：
- `S`: 秒 (Second)
- `M`: 分钟 (Minute)
- `H`: 小时 (Hour)
- `D`: 天 (Day)

## 限流策略

### 1. IP限流

基于客户端IP地址进行限流，适用于公开API端点。

```go
// 使用配置的限流规则
rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("auth_public"))

// 使用自定义限流规则
rateLimiter.IPRateLimit("20-M")
```

### 2. 用户限流

基于JWT token中的用户ID进行限流，适用于需要认证的API。

```go
// 使用配置的限流规则
rateLimiter.UserRateLimit(rateLimiter.GetConfiguredLimit("user_protected"))

// 使用自定义限流规则
rateLimiter.UserRateLimit("50-M")
```

### 3. 自定义限流

使用自定义键生成函数进行限流。

```go
// 按端点限流
rateLimiter.CustomRateLimit(ratelimit.KeyFunctions.ByEndpoint, "100-M")

// 按用户和端点组合限流
rateLimiter.CustomRateLimit(ratelimit.KeyFunctions.ByUserAndEndpoint, "10-M")
```

### 4. 端点限流

根据配置文件中的端点配置自动选择限流规则。

```go
rateLimiter.EndpointRateLimit("/api/v1/auth/register")
```

## 存储后端

### 内存存储

适用于单实例部署，重启后限流状态会丢失。

```yaml
ratelimit:
  store_type: memory
```

### Redis存储

适用于多实例部署，支持分布式限流。

```yaml
ratelimit:
  store_type: redis
  redis_prefix: "ratelimit:"
```

## 响应头信息

限流中间件会在响应中添加以下头信息：

- `X-RateLimit-Limit`: 限流上限
- `X-RateLimit-Remaining`: 剩余请求数
- `X-RateLimit-Reset`: 重置时间戳

## 限流响应

当触发限流时，API会返回HTTP 429状态码和以下响应：

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Limit: 10 per 1m0s",
  "retry_after": 1640995200
}
```

## 使用示例

### 在路由中使用

```go
// 健康检查端点
r.GET("/health", 
    rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("health_check")), 
    healthHandler)

// 认证端点组
authGroup := r.Group("/auth")
authGroup.Use(rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("auth_public")))
{
    authGroup.POST("/register", registerHandler)
    authGroup.POST("/login", loginHandler)
}

// 保护的用户端点
userGroup := r.Group("/users")
userGroup.Use(middleware.JWTAuth(jwtManager))
userGroup.Use(rateLimiter.UserRateLimit(rateLimiter.GetConfiguredLimit("user_protected")))
{
    userGroup.GET("/me", getMeHandler)
    userGroup.POST("", createUserHandler)
}
```

### 创建限流器

```go
// 使用配置文件创建
rateLimiter, err := ratelimit.NewRateLimiterFromConfig(config.GlobalConfig.RateLimit)

// 手动创建内存限流器
rateLimiter, err := ratelimit.NewRateLimiter(config.RateLimitConfig{
    Enabled: true,
    StoreType: "memory",
    RedisPrefix: "ratelimit:",
    DefaultConfig: config.RateLimitDefaultConfig{
        HealthCheck: "10-S",
        AuthPublic: "20-M",
    },
})

// 手动创建Redis限流器
rateLimiter, err := ratelimit.NewRedisRateLimiter(cfg, redisClient)
```

## 测试

使用提供的测试脚本验证限流功能：

```bash
# 运行限流测试
./scripts/test-ratelimit.sh
```

测试脚本会：
1. 测试健康检查端点的限流 (10次/秒)
2. 测试认证注册端点的限流 (20次/分钟)
3. 检查响应头信息

## 监控和调试

### 日志信息

限流器会记录以下日志：

```
INFO  Rate limiter initialized  store_type=redis enabled=true prefix=ratelimit:
WARN  Rate limit exceeded       ip=192.168.1.100 path=/health limit=10-S
```

### 配置验证

启动时会验证限流配置格式：

```
ERROR Invalid rate limit format  limit=invalid-format error=...
```

## 最佳实践

1. **生产环境使用Redis**: 确保多实例部署时限流状态一致
2. **合理设置限流值**: 根据业务需求和服务器性能设置
3. **监控限流指标**: 关注429响应的频率和分布
4. **优雅降级**: 限流器故障时不应影响正常业务
5. **定期调整**: 根据实际使用情况调整限流配置

## 故障排除

### 常见问题

1. **限流不生效**
   - 检查 `enabled` 配置是否为 `true`
   - 验证限流格式是否正确
   - 查看日志中的错误信息

2. **Redis连接失败**
   - 检查Redis服务是否正常
   - 验证Redis连接配置
   - 查看是否自动降级到内存存储

3. **限流过于严格**
   - 调整配置文件中的限流值
   - 考虑使用用户限流替代IP限流
   - 为不同端点设置不同的限流规则

### 调试命令

```bash
# 检查Redis中的限流键
redis-cli --scan --pattern "ratelimit:*"

# 查看特定键的值
redis-cli ZRANGE "ratelimit:ip:192.168.1.100" 0 -1 WITHSCORES

# 清除所有限流数据
redis-cli --scan --pattern "ratelimit:*" | xargs redis-cli DEL
``` 