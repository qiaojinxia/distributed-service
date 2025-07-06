# 中间件模块设计文档

## 📋 概述

中间件模块是分布式服务框架的请求处理核心组件，提供HTTP和gRPC的统一中间件支持。包含认证、日志、监控、限流、链路追踪等功能，支持灵活的中间件组合和自定义扩展。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    请求入口层                            │
│               Request Entry Layer                       │
│        HTTP Request | gRPC Request                     │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  中间件管道层                            │
│               Middleware Pipeline Layer                 │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │  HTTP中间件     │   gRPC拦截器    │   通用中间件    │ │
│  │HTTP Middleware  │gRPC Interceptor │Common Middleware│ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   功能中间件层                           │
│               Functional Middleware Layer               │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────────┐ │
│  │  认证   │  日志   │  监控   │  限流   │   追踪      │ │
│  │  Auth   │ Logger  │Monitor  │ RateLimit│  Tracing   │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   业务处理层                             │
│                Business Logic Layer                     │
│              Handler | Service                          │
└─────────────────────────────────────────────────────────┘
```

## 🎯 核心特点

### 1. 双协议支持
- **HTTP中间件**: 基于Gin框架的HTTP中间件
- **gRPC拦截器**: 一元和流式RPC拦截器
- **统一接口**: HTTP和gRPC的统一中间件接口
- **协议转换**: 自动处理协议差异

### 2. 丰富的内置中间件
- **认证中间件**: JWT、OAuth2、API Key认证
- **日志中间件**: 请求/响应日志记录
- **监控中间件**: 指标收集和健康检查
- **限流中间件**: 基于令牌桶和滑动窗口
- **追踪中间件**: 分布式链路追踪
- **CORS中间件**: 跨域资源共享
- **安全中间件**: 安全头设置和防护

### 3. 高性能设计
- **零分配**: 避免不必要的内存分配
- **异步处理**: 支持异步日志和监控
- **连接复用**: 长连接和连接池优化
- **缓存机制**: 认证结果和配置缓存

### 4. 灵活配置
- **链式调用**: 支持中间件链式组合
- **条件启用**: 基于路径、方法的条件中间件
- **动态配置**: 运行时动态调整中间件参数
- **优先级控制**: 中间件执行顺序控制

## 🚀 使用示例

### HTTP中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
)

func main() {
    r := gin.Default()
    
    // 全局中间件
    r.Use(middleware.CORS())                    // 跨域处理
    r.Use(middleware.RequestID())               // 请求ID
    r.Use(middleware.Logger())                  // 访问日志
    r.Use(middleware.Recovery())                // 错误恢复
    r.Use(middleware.RateLimit(100, 60))        // 限流：100req/min
    r.Use(middleware.Metrics())                 // 监控指标
    
    // 认证中间件
    auth := r.Group("/api/v1")
    auth.Use(middleware.JWTAuth("your-secret-key"))
    {
        auth.GET("/users", getUsersHandler)
        auth.POST("/users", createUserHandler)
        
        // 角色权限中间件
        admin := auth.Group("/admin")
        admin.Use(middleware.RequireRole("admin"))
        {
            admin.GET("/stats", getStatsHandler)
            admin.DELETE("/users/:id", deleteUserHandler)
        }
    }
    
    // 公开API
    public := r.Group("/public")
    public.Use(middleware.RateLimit(20, 60))    // 更严格的限流
    {
        public.POST("/register", registerHandler)
        public.POST("/login", loginHandler)
    }
    
    r.Run(":8080")
}
```

### gRPC拦截器

```go
package main

import (
    "google.golang.org/grpc"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
)

func main() {
    // 一元RPC拦截器
    unaryInterceptors := []grpc.UnaryServerInterceptor{
        middleware.UnaryRequestID(),              // 请求ID
        middleware.UnaryLogger(),                 // 日志记录
        middleware.UnaryMetrics(),                // 指标收集
        middleware.UnaryAuth("your-secret-key"),  // 认证
        middleware.UnaryRateLimit(100, 60),       // 限流
        middleware.UnaryTracing(),                // 链路追踪
        middleware.UnaryRecovery(),               // 错误恢复
    }
    
    // 流式RPC拦截器
    streamInterceptors := []grpc.StreamServerInterceptor{
        middleware.StreamRequestID(),
        middleware.StreamLogger(),
        middleware.StreamMetrics(),
        middleware.StreamAuth("your-secret-key"),
        middleware.StreamRateLimit(50, 60),
        middleware.StreamTracing(),
        middleware.StreamRecovery(),
    }
    
    // 创建gRPC服务器
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(unaryInterceptors...),
        grpc.ChainStreamInterceptor(streamInterceptors...),
    )
    
    // 注册服务
    pb.RegisterUserServiceServer(server, &userService{})
    
    // 启动服务器
    lis, _ := net.Listen("tcp", ":50051")
    server.Serve(lis)
}
```

### 自定义中间件

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/qiaojinxia/distributed-service/framework/logger"
)

// HTTP自定义中间件
func CustomHTTPMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 请求前处理
        logger.Info("请求开始",
            logger.String("method", c.Request.Method),
            logger.String("path", c.Request.URL.Path),
            logger.String("ip", c.ClientIP()),
        )
        
        // 处理请求
        c.Next()
        
        // 请求后处理
        duration := time.Since(start)
        logger.Info("请求完成",
            logger.Int("status", c.Writer.Status()),
            logger.Duration("duration", duration),
        )
    }
}

// gRPC自定义拦截器
func CustomUnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        // 请求前处理
        logger.InfoContext(ctx, "gRPC请求开始",
            logger.String("method", info.FullMethod),
        )
        
        // 处理请求
        resp, err := handler(ctx, req)
        
        // 请求后处理
        duration := time.Since(start)
        logger.InfoContext(ctx, "gRPC请求完成",
            logger.Duration("duration", duration),
            logger.Bool("success", err == nil),
        )
        
        return resp, err
    }
}
```

### 条件中间件

```go
package main

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
)

func main() {
    r := gin.Default()
    
    // 条件认证中间件
    r.Use(middleware.ConditionalAuth(func(c *gin.Context) bool {
        // 只对API路径启用认证
        return strings.HasPrefix(c.Request.URL.Path, "/api/")
    }, "your-secret-key"))
    
    // 条件限流中间件
    r.Use(middleware.ConditionalRateLimit(func(c *gin.Context) bool {
        // 对公开API启用限流
        return strings.HasPrefix(c.Request.URL.Path, "/public/")
    }, 20, 60))
    
    // 条件CORS中间件
    r.Use(middleware.ConditionalCORS(func(c *gin.Context) bool {
        // 只对前端API启用CORS
        return strings.HasPrefix(c.Request.URL.Path, "/frontend/")
    }))
    
    r.Run(":8080")
}
```

## 🔧 配置选项

### 认证中间件配置

```go
type AuthConfig struct {
    // JWT配置
    JWTSecret       string        `yaml:"jwt_secret"`
    JWTExpiration   time.Duration `yaml:"jwt_expiration"`
    
    // API Key配置
    APIKeyHeader    string        `yaml:"api_key_header"`
    APIKeyQuery     string        `yaml:"api_key_query"`
    
    // OAuth2配置
    OAuth2Endpoint  string        `yaml:"oauth2_endpoint"`
    OAuth2ClientID  string        `yaml:"oauth2_client_id"`
    
    // 白名单
    WhitelistPaths  []string      `yaml:"whitelist_paths"`
    WhitelistIPs    []string      `yaml:"whitelist_ips"`
    
    // 缓存配置
    CacheEnabled    bool          `yaml:"cache_enabled"`
    CacheTTL        time.Duration `yaml:"cache_ttl"`
}
```

### 限流中间件配置

```go
type RateLimitConfig struct {
    // 基础配置
    Rate     int           `yaml:"rate"`      // 每分钟请求数
    Burst    int           `yaml:"burst"`     // 突发请求数
    Window   time.Duration `yaml:"window"`    // 时间窗口
    
    // 存储配置
    Storage  string        `yaml:"storage"`   // memory, redis
    KeyFunc  string        `yaml:"key_func"`  // ip, user, custom
    
    // Redis配置
    RedisAddr     string `yaml:"redis_addr"`
    RedisPassword string `yaml:"redis_password"`
    RedisDB       int    `yaml:"redis_db"`
    
    // 响应配置
    ErrorMessage  string `yaml:"error_message"`
    RetryAfter    bool   `yaml:"retry_after"`
    
    // 白名单
    WhitelistIPs  []string `yaml:"whitelist_ips"`
    WhitelistKeys []string `yaml:"whitelist_keys"`
}
```

### 日志中间件配置

```go
type LoggerConfig struct {
    // 日志级别
    Level         string        `yaml:"level"`
    
    // 日志格式
    Format        string        `yaml:"format"`        // json, text
    TimeFormat    string        `yaml:"time_format"`   // 时间格式
    
    // 日志字段
    RequestIDKey  string        `yaml:"request_id_key"`
    UserIDKey     string        `yaml:"user_id_key"`
    TraceIDKey    string        `yaml:"trace_id_key"`
    
    // 过滤配置
    SkipPaths     []string      `yaml:"skip_paths"`    // 跳过的路径
    SkipMethods   []string      `yaml:"skip_methods"`  // 跳过的方法
    
    // 性能配置
    AsyncWrite    bool          `yaml:"async_write"`   // 异步写入
    BufferSize    int           `yaml:"buffer_size"`   // 缓冲区大小
    
    // 敏感信息处理
    MaskFields    []string      `yaml:"mask_fields"`   // 需要脱敏的字段
    MaskHeaders   []string      `yaml:"mask_headers"`  // 需要脱敏的请求头
}
```

### 配置文件示例

```yaml
# config/middleware.yaml
middleware:
  auth:
    jwt_secret: "your-jwt-secret-key"
    jwt_expiration: "24h"
    api_key_header: "X-API-Key"
    whitelist_paths:
      - "/health"
      - "/metrics"
      - "/public/*"
    cache_enabled: true
    cache_ttl: "5m"
    
  rate_limit:
    rate: 100              # 100 requests per minute
    burst: 20              # 20 burst requests
    window: "1m"
    storage: "redis"
    key_func: "ip"
    redis_addr: "localhost:6379"
    error_message: "请求频率过高，请稍后重试"
    retry_after: true
    
  logger:
    level: "info"
    format: "json"
    time_format: "2006-01-02T15:04:05.000Z07:00"
    request_id_key: "request_id"
    trace_id_key: "trace_id"
    skip_paths:
      - "/health"
      - "/metrics"
    async_write: true
    buffer_size: 1024
    mask_fields:
      - "password"
      - "credit_card"
    mask_headers:
      - "Authorization"
      - "X-API-Key"
      
  cors:
    allowed_origins:
      - "https://example.com"
      - "https://app.example.com"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
    allowed_headers:
      - "Content-Type"
      - "Authorization"
    max_age: 86400
    
  metrics:
    enabled: true
    namespace: "myapp"
    subsystem: "http"
    buckets: [0.1, 0.3, 1.2, 5.0]
```

## 📊 监控与指标

### HTTP指标

```go
// HTTP请求指标
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.1, 0.3, 1.2, 5.0},
        },
        []string{"method", "path"},
    )
    
    httpRequestSize = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_size_bytes",
            Help:    "HTTP request size in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "path"},
    )
)
```

### gRPC指标

```go
// gRPC请求指标
var (
    grpcRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"method", "status"},
    )
    
    grpcRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "gRPC request duration in seconds",
            Buckets: []float64{0.01, 0.1, 0.3, 1.2, 5.0},
        },
        []string{"method"},
    )
)
```

### 中间件性能指标

```go
// 中间件性能指标
type MiddlewareStats struct {
    AuthRequests      int64         `json:"auth_requests"`
    AuthSuccesses     int64         `json:"auth_successes"`
    AuthFailures      int64         `json:"auth_failures"`
    AuthLatency       time.Duration `json:"auth_latency"`
    
    RateLimitRequests int64         `json:"rate_limit_requests"`
    RateLimitBlocked  int64         `json:"rate_limit_blocked"`
    RateLimitLatency  time.Duration `json:"rate_limit_latency"`
    
    LoggerRequests    int64         `json:"logger_requests"`
    LoggerErrors      int64         `json:"logger_errors"`
    LoggerLatency     time.Duration `json:"logger_latency"`
}
```

## 🔍 最佳实践

### 1. 中间件顺序

```go
// ✅ 推荐的中间件顺序
r.Use(middleware.Recovery())        // 1. 错误恢复（最外层）
r.Use(middleware.CORS())            // 2. 跨域处理
r.Use(middleware.RequestID())       // 3. 请求ID生成
r.Use(middleware.Logger())          // 4. 日志记录
r.Use(middleware.Metrics())         // 5. 指标收集
r.Use(middleware.RateLimit())       // 6. 限流控制
r.Use(middleware.Auth())            // 7. 认证授权
r.Use(middleware.Tracing())         // 8. 链路追踪（最内层）
```

### 2. 性能优化

```go
// ✅ 推荐：缓存认证结果
authMiddleware := middleware.JWTAuthWithCache(
    "secret-key",
    middleware.CacheConfig{
        TTL:  5 * time.Minute,
        Size: 1000,
    },
)

// ✅ 推荐：异步日志记录
loggerMiddleware := middleware.AsyncLogger(
    middleware.LoggerConfig{
        BufferSize:    1000,
        FlushInterval: 5 * time.Second,
    },
)

// ✅ 推荐：条件中间件
r.Use(middleware.ConditionalMetrics(func(c *gin.Context) bool {
    return !strings.HasPrefix(c.Request.URL.Path, "/health")
}))
```

### 3. 错误处理

```go
// ✅ 推荐：统一错误处理
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            switch e := err.Err.(type) {
            case *middleware.AuthError:
                c.JSON(401, gin.H{"error": "认证失败", "code": "AUTH_FAILED"})
            case *middleware.RateLimitError:
                c.JSON(429, gin.H{"error": "请求过于频繁", "code": "RATE_LIMIT_EXCEEDED"})
            default:
                c.JSON(500, gin.H{"error": "内部服务器错误", "code": "INTERNAL_ERROR"})
            }
        }
    }
}
```

### 4. 安全最佳实践

```go
// ✅ 推荐：安全头设置
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    }
}

// ✅ 推荐：输入验证
func InputValidation() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查请求大小
        if c.Request.ContentLength > 10*1024*1024 { // 10MB
            c.AbortWithStatusJSON(413, gin.H{"error": "请求体过大"})
            return
        }
        
        // 检查Content-Type
        contentType := c.GetHeader("Content-Type")
        if !isValidContentType(contentType) {
            c.AbortWithStatusJSON(400, gin.H{"error": "不支持的Content-Type"})
            return
        }
        
        c.Next()
    }
}
```

## 🚨 故障排查

### 常见问题

**Q1: 认证中间件性能问题**
```go
// 启用认证缓存
config := middleware.AuthConfig{
    CacheEnabled: true,
    CacheTTL:     5 * time.Minute,
}

// 监控认证延迟
auth.Use(middleware.MonitorAuth())
```

**Q2: 限流误判问题**
```go
// 调整限流算法
config := middleware.RateLimitConfig{
    Algorithm: "sliding_window", // 滑动窗口更精确
    KeyFunc:   "user_id",        // 基于用户而非IP
}

// 添加白名单
config.WhitelistIPs = []string{"127.0.0.1", "10.0.0.0/8"}
```

**Q3: 日志性能影响**
```go
// 启用异步日志
config := middleware.LoggerConfig{
    AsyncWrite:    true,
    BufferSize:    2048,
    FlushInterval: 3 * time.Second,
}

// 跳过健康检查路径
config.SkipPaths = []string{"/health", "/metrics", "/ping"}
```

## 🔮 高级功能

### 自适应限流

```go
type AdaptiveRateLimit struct {
    baseLine    int
    maxRate     int
    errorRate   float64
    adjustInterval time.Duration
}

func (a *AdaptiveRateLimit) AdjustRate() {
    if a.errorRate > 0.1 { // 错误率超过10%
        a.baseLine = int(float64(a.baseLine) * 0.8) // 降低20%
    } else if a.errorRate < 0.01 { // 错误率低于1%
        a.baseLine = int(float64(a.baseLine) * 1.1) // 提高10%
    }
}
```

### 断路器模式

```go
func CircuitBreaker() gin.HandlerFunc {
    cb := circuit.NewCircuitBreaker(circuit.Config{
        MaxRequests: 3,
        Interval:    time.Minute,
        Timeout:     30 * time.Second,
    })
    
    return func(c *gin.Context) {
        result, err := cb.Execute(func() (interface{}, error) {
            c.Next()
            if c.Writer.Status() >= 500 {
                return nil, errors.New("server error")
            }
            return nil, nil
        })
        
        if err != nil {
            c.JSON(503, gin.H{"error": "服务暂时不可用"})
            c.Abort()
        }
    }
}
```

---

> 中间件模块为框架提供了完整的请求处理能力，支持HTTP和gRPC的统一中间件管理和丰富的功能扩展。