# 日志模块设计文档

## 📋 概述

日志模块是分布式服务框架的可观测性核心组件，基于Zap构建，提供高性能、结构化日志记录功能。支持多种输出格式、日志级别控制、字段扩展和分布式追踪集成。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    应用接口层                            │
│        Application Interface Layer                      │
│  logger.Info() | Error() | Debug() | WithFields()      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   日志处理层                             │
│                Log Processing Layer                     │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   字段处理      │    格式化       │    过滤器       │ │
│  │ Field Handler   │  Formatter      │   Filter        │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    Zap引擎层                             │
│                   Zap Engine Layer                      │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   JSON编码器    │   控制台编码    │   采样器        │ │
│  │ JSON Encoder    │Console Encoder  │   Sampler       │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   输出目标层                             │
│                 Output Target Layer                     │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │    文件输出     │    控制台       │   远程日志      │ │
│  │  File Output    │   Console       │ Remote Logging  │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 🎯 核心特点

### 1. 高性能日志
- **零分配**: 基于Zap的高性能日志引擎
- **异步写入**: 支持异步日志写入，不阻塞业务逻辑
- **批量处理**: 自动批量写入，减少I/O开销
- **内存优化**: 对象池复用，减少GC压力

### 2. 结构化日志
- **JSON格式**: 标准JSON输出，便于日志分析
- **字段扩展**: 支持任意字段添加和上下文传递
- **类型安全**: 强类型字段，避免序列化错误
- **嵌套支持**: 支持复杂对象和嵌套结构

### 3. 分布式追踪
- **TraceID集成**: 自动注入分布式追踪ID
- **上下文传递**: 跨服务的日志关联
- **链路追踪**: 与OpenTelemetry无缝集成
- **请求追踪**: HTTP/gRPC请求全链路日志

### 4. 灵活配置
- **多级别控制**: Debug、Info、Warn、Error、Fatal
- **动态调整**: 运行时动态调整日志级别
- **多输出**: 同时输出到文件、控制台、远程服务
- **日志轮转**: 自动日志文件轮转和清理

## 🚀 使用示例

### 基础日志记录

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/logger"
)

func main() {
    // 初始化日志器
    err := logger.Init(logger.Config{
        Level:      "info",
        Format:     "json",
        OutputPath: "./logs/app.log",
    })
    if err != nil {
        panic(err)
    }
    
    // 基础日志记录
    logger.Info("应用启动成功")
    logger.Error("发生错误", logger.String("error", "连接失败"))
    logger.Debug("调试信息", logger.Int("retry_count", 3))
    logger.Warn("警告信息", logger.String("component", "database"))
}
```

### 结构化字段

```go
package main

import (
    "time"
    "github.com/qiaojinxia/distributed-service/framework/logger"
)

func processOrder(orderID string, userID int) {
    // 创建带字段的日志器
    orderLogger := logger.WithFields(
        logger.String("order_id", orderID),
        logger.Int("user_id", userID),
        logger.String("operation", "process_order"),
    )
    
    orderLogger.Info("开始处理订单")
    
    // 模拟处理逻辑
    if err := validateOrder(orderID); err != nil {
        orderLogger.Error("订单验证失败",
            logger.String("error", err.Error()),
            logger.Duration("elapsed", time.Since(start)),
        )
        return
    }
    
    orderLogger.Info("订单处理完成",
        logger.String("status", "success"),
        logger.Time("completed_at", time.Now()),
    )
}
```

### HTTP请求日志

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/qiaojinxia/distributed-service/framework/logger"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
)

func main() {
    r := gin.Default()
    
    // 使用日志中间件
    r.Use(middleware.LoggerMiddleware())
    
    r.GET("/api/users/:id", func(c *gin.Context) {
        userID := c.Param("id")
        
        // 从上下文获取请求日志器
        reqLogger := logger.FromContext(c.Request.Context())
        
        reqLogger.Info("获取用户信息",
            logger.String("user_id", userID),
        )
        
        // 业务逻辑...
        user, err := getUserByID(userID)
        if err != nil {
            reqLogger.Error("获取用户失败",
                logger.String("user_id", userID),
                logger.String("error", err.Error()),
            )
            c.JSON(500, gin.H{"error": "内部错误"})
            return
        }
        
        reqLogger.Info("用户信息获取成功",
            logger.String("user_id", userID),
            logger.String("username", user.Name),
        )
        
        c.JSON(200, user)
    })
    
    r.Run(":8080")
}
```

### 分布式追踪集成

```go
package main

import (
    "context"
    "github.com/qiaojinxia/distributed-service/framework/logger"
    "github.com/qiaojinxia/distributed-service/framework/tracing"
)

func handleUserRequest(ctx context.Context, userID int) {
    // 启动span
    span, ctx := tracing.StartSpan(ctx, "handle_user_request")
    defer span.Finish()
    
    // 带追踪信息的日志
    logger.InfoContext(ctx, "开始处理用户请求",
        logger.Int("user_id", userID),
    )
    
    // 调用其他服务
    profile, err := getUserProfile(ctx, userID)
    if err != nil {
        logger.ErrorContext(ctx, "获取用户资料失败",
            logger.Int("user_id", userID),
            logger.String("error", err.Error()),
        )
        span.SetTag("error", true)
        return
    }
    
    logger.InfoContext(ctx, "用户请求处理完成",
        logger.Int("user_id", userID),
        logger.String("profile_status", profile.Status),
    )
}

func getUserProfile(ctx context.Context, userID int) (*UserProfile, error) {
    // 子span
    span, ctx := tracing.StartSpan(ctx, "get_user_profile")
    defer span.Finish()
    
    logger.DebugContext(ctx, "查询用户资料",
        logger.Int("user_id", userID),
    )
    
    // 数据库查询...
    return fetchFromDB(userID)
}
```

## 🔧 配置选项

### 完整配置

```go
type Config struct {
    // 基础配置
    Level      string `yaml:"level"`       // 日志级别: debug,info,warn,error,fatal
    Format     string `yaml:"format"`      // 输出格式: json,console
    OutputPath string `yaml:"output_path"` // 输出路径
    
    // 高级配置
    MaxSize     int  `yaml:"max_size"`     // 单个日志文件最大大小(MB)
    MaxAge      int  `yaml:"max_age"`      // 日志文件保留天数
    MaxBackups  int  `yaml:"max_backups"`  // 最大备份文件数
    Compress    bool `yaml:"compress"`     // 是否压缩备份文件
    
    // 性能配置
    AsyncWrite     bool          `yaml:"async_write"`     // 异步写入
    BufferSize     int           `yaml:"buffer_size"`     // 缓冲区大小
    FlushInterval  time.Duration `yaml:"flush_interval"`  // 刷新间隔
    SamplingEnable bool          `yaml:"sampling_enable"` // 启用采样
    SamplingRate   int           `yaml:"sampling_rate"`   // 采样率
    
    // 追踪配置
    EnableTracing bool `yaml:"enable_tracing"` // 启用分布式追踪
    TraceIDField  string `yaml:"trace_id_field"` // TraceID字段名
    SpanIDField   string `yaml:"span_id_field"`  // SpanID字段名
}
```

### 配置示例

```yaml
# config/logger.yaml
logger:
  level: "info"
  format: "json"
  output_path: "./logs/app.log"
  
  # 日志轮转
  max_size: 100      # 100MB
  max_age: 30        # 30天
  max_backups: 5     # 5个备份
  compress: true     # 压缩备份
  
  # 性能优化
  async_write: true
  buffer_size: 1024
  flush_interval: "5s"
  sampling_enable: true
  sampling_rate: 100  # 每100条采样1条
  
  # 分布式追踪
  enable_tracing: true
  trace_id_field: "trace_id"
  span_id_field: "span_id"
```

## 📊 字段类型支持

### 基础类型

```go
// 字符串
logger.String("key", "value")

// 数字类型
logger.Int("count", 42)
logger.Int64("id", 123456789)
logger.Float64("price", 99.99)

// 布尔值
logger.Bool("success", true)

// 时间
logger.Time("created_at", time.Now())
logger.Duration("elapsed", time.Second*5)
```

### 复杂类型

```go
// 数组
logger.Strings("tags", []string{"api", "user", "v1"})
logger.Ints("ids", []int{1, 2, 3, 4, 5})

// 对象
logger.Any("user", User{
    ID:   123,
    Name: "John",
    Age:  30,
})

// 错误
logger.Error("处理失败", logger.Err(err))

// 自定义序列化
logger.Object("request", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
    enc.AddString("method", req.Method)
    enc.AddString("url", req.URL.String())
    return nil
}))
```

## 🚀 性能优化

### 性能基准

```
操作类型              吞吐量           延迟        内存分配
同步JSON日志         ~500K op/s      ~2µs        0 allocs
异步JSON日志         ~2M op/s        ~0.5µs      0 allocs  
结构化字段(5个)       ~300K op/s      ~3µs        0 allocs
复杂对象序列化        ~100K op/s      ~10µs       1 alloc
```

### 优化建议

```go
// ✅ 推荐：使用强类型字段
logger.Info("用户登录", 
    logger.String("username", user.Name),
    logger.Int("user_id", user.ID),
)

// ❌ 避免：使用Any类型
// logger.Info("用户登录", logger.Any("user", user)) // 慢

// ✅ 推荐：预定义日志器
var userLogger = logger.WithFields(
    logger.String("component", "user_service"),
    logger.String("version", "v1.0"),
)

// ✅ 推荐：条件日志
if logger.IsDebugEnabled() {
    logger.Debug("详细调试信息", logger.Any("details", expensiveOperation()))
}
```

## 🔍 监控与观测

### 日志统计

```go
// 获取日志统计信息
stats := logger.GetStats()
fmt.Printf("总日志数: %d\n", stats.TotalLogs)
fmt.Printf("错误日志数: %d\n", stats.ErrorLogs)
fmt.Printf("平均延迟: %v\n", stats.AvgLatency)
```

### 健康检查

```go
func healthCheck() map[string]interface{} {
    return map[string]interface{}{
        "logger": map[string]interface{}{
            "status":      logger.IsHealthy(),
            "buffer_size": logger.GetBufferUsage(),
            "last_flush":  logger.GetLastFlushTime(),
        },
    }
}
```

## 🚨 故障排查

### 常见问题

**Q1: 日志丢失问题**
```go
// 确保程序退出前刷新缓冲区
defer logger.Sync()

// 或使用同步模式
logger.SetAsync(false)
```

**Q2: 性能问题**
```go
// 启用采样减少日志量
logger.SetSamplingRate(100) // 每100条记录1条

// 调整缓冲区大小
logger.SetBufferSize(2048)
```

**Q3: 文件权限问题**
```go
// 检查文件写入权限
if err := logger.TestWrite(); err != nil {
    log.Fatalf("日志文件写入权限错误: %v", err)
}
```

## 🔮 高级功能

### 自定义编码器

```go
type CustomEncoder struct {
    zapcore.Encoder
}

func (enc *CustomEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
    // 自定义编码逻辑
    buf := &buffer.Buffer{}
    
    // 添加自定义前缀
    buf.AppendString("[MYAPP] ")
    
    // 调用原始编码器
    return enc.Encoder.EncodeEntry(entry, fields)
}
```

### 日志钩子

```go
// 注册日志钩子
logger.AddHook(func(entry *logger.Entry) {
    if entry.Level >= logger.ErrorLevel {
        // 发送告警通知
        sendAlert(entry.Message, entry.Fields)
    }
})
```

### 多环境配置

```go
func initLogger() {
    var config logger.Config
    
    switch os.Getenv("ENV") {
    case "production":
        config = logger.Config{
            Level:     "warn",
            Format:    "json",
            AsyncWrite: true,
        }
    case "development":
        config = logger.Config{
            Level:     "debug", 
            Format:    "console",
            AsyncWrite: false,
        }
    default:
        config = logger.DefaultConfig()
    }
    
    logger.Init(config)
}
```

---

> 日志模块为框架提供了企业级的日志记录能力，支持高并发、大规模分布式系统的可观测性需求。