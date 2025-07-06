# 分布式服务框架模块文档

## 📚 文档导航

欢迎使用分布式服务框架！本文档提供了框架各个模块的详细设计文档和使用指南。

### 🗂️ 模块列表

| 模块 | 功能描述 | 文档链接 | 状态 |
|------|----------|----------|------|
| **Core** | 框架核心和应用构建器 | [📖 查看文档](../core/) | ✅ |
| **Cache** | 多策略缓存系统 | [📖 查看文档](../cache/docs/) | ✅ |
| **Auth** | JWT认证和权限管理 | [📖 查看文档](../auth/docs/) | ✅ |
| **Logger** | 高性能结构化日志 | [📖 查看文档](../logger/docs/) | ✅ |
| **Database** | 数据库连接和ORM | [📖 查看文档](../database/docs/) | ✅ |
| **Middleware** | HTTP/gRPC中间件 | [📖 查看文档](../middleware/docs/) | ✅ |
| **Plugin** | 动态插件系统 | [📖 查看文档](../plugin/docs/) | ✅ |
| **Transport** | HTTP/gRPC传输层 | [📖 查看文档](../transport/docs/) | ✅ |
| **Metrics** | 监控指标收集 | [📖 查看文档](../metrics/) | 🚧 |
| **Tracing** | 分布式链路追踪 | [📖 查看文档](../tracing/) | 🚧 |
| **Config** | 配置管理 | [📖 查看文档](../config/) | 🚧 |

## 🚀 快速开始

### 1. 基础使用

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/core"
)

func main() {
    // 创建应用
    app := core.New().
        Name("my-service").
        Port(8080).
        WithCache().       // 启用缓存
        WithAuth().        // 启用认证
        WithMetrics().     // 启用监控
        WithTracing()      // 启用追踪
    
    // 注册路由
    app.GET("/users", getUsersHandler)
    app.POST("/users", createUserHandler)
    
    // 启动服务
    app.Run()
}
```

### 2. 高级配置

```yaml
# config/app.yaml
app:
  name: "distributed-service"
  port: 8080
  mode: "release"
  
cache:
  default_ttl: "1h"
  max_size: 1000
  
database:
  mysql:
    host: "localhost"
    port: 3306
    database: "myapp"
    
logger:
  level: "info"
  format: "json"
  
metrics:
  enabled: true
  port: 9090
```

## 📖 模块详细介绍

### 🎯 核心模块 (Core)

框架的核心模块，提供应用构建器和统一的API接口。

**主要功能：**
- 应用生命周期管理
- 模块注册和初始化
- 统一配置管理
- 优雅启动和关闭

**快速上手：**
```go
app := core.New().
    Name("my-service").
    Port(8080).
    OnlyHTTP().  // 只启用HTTP服务
    Run()
```

### 🗄️ 缓存模块 (Cache)

多策略缓存系统，支持LRU、TTL、Simple三种淘汰策略。

**主要功能：**
- 内存缓存和Redis缓存
- 多种淘汰策略
- 框架无缝集成
- 高性能设计

**快速上手：**
```go
// 使用框架缓存API
userCache := core.GetUserCache()
userCache.Set(ctx, "user:123", user, time.Hour)

// 直接创建缓存
cache, _ := cache.NewMemoryCache(cache.MemoryConfig{
    MaxSize: 1000,
    EvictionPolicy: cache.EvictionPolicyLRU,
})
```

### 🔐 认证模块 (Auth)

JWT令牌管理和权限控制系统。

**主要功能：**
- JWT令牌生成和验证
- 密码加密和校验
- 权限控制中间件
- 多种认证方式

**快速上手：**
```go
// 生成JWT令牌
token, _ := auth.GenerateToken(claims, "secret-key")

// 验证密码
isValid := auth.CheckPassword(password, hashedPassword)

// 中间件使用
r.Use(middleware.JWTAuth("secret-key"))
```

### 📝 日志模块 (Logger)

基于Zap的高性能结构化日志系统。

**主要功能：**
- 零分配高性能日志
- 结构化字段支持
- 分布式追踪集成
- 异步写入优化

**快速上手：**
```go
// 基础日志
logger.Info("用户登录", 
    logger.String("user_id", "123"),
    logger.String("ip", "192.168.1.1"))

// 带上下文的日志
logger.InfoContext(ctx, "处理订单",
    logger.String("order_id", orderID))
```

### 🗃️ 数据库模块 (Database)

MySQL和Redis的统一访问层。

**主要功能：**
- GORM集成的MySQL支持
- go-redis客户端集成
- 连接池管理
- 事务支持

**快速上手：**
```go
// MySQL操作
db.Create(&user)
db.First(&user, userID)
db.Model(&user).Update("name", newName)

// Redis操作
rdb.Set(ctx, "key", "value", time.Hour)
value := rdb.Get(ctx, "key").Val()
```

### 🔗 中间件模块 (Middleware)

HTTP和gRPC的统一中间件系统。

**主要功能：**
- 认证、日志、监控中间件
- 限流和熔断支持
- HTTP和gRPC统一接口
- 灵活的中间件组合

**快速上手：**
```go
// HTTP中间件
r.Use(middleware.Logger())
r.Use(middleware.Auth("secret-key"))
r.Use(middleware.RateLimit(100, 60))

// gRPC拦截器
server.AddUnaryInterceptor(middleware.UnaryAuth("secret-key"))
server.AddStreamInterceptor(middleware.StreamLogger())
```

### 🔌 插件模块 (Plugin)

动态插件加载和管理系统。

**主要功能：**
- 热插拔插件支持
- 插件生命周期管理
- 事件驱动通信
- 依赖解析和隔离

**快速上手：**
```go
// 创建插件
type MyPlugin struct {
    plugin.BasePlugin
}

func (p *MyPlugin) Start(ctx context.Context) error {
    p.EventBus().Subscribe("user.login", p.handleLogin)
    return nil
}

// 注册插件
plugin.Register("my-plugin", NewMyPlugin)
```

### 🌐 传输模块 (Transport)

HTTP和gRPC的统一传输层。

**主要功能：**
- HTTP/gRPC服务器
- 负载均衡和服务发现
- 协议转换
- 网关代理

**快速上手：**
```go
// HTTP服务器
server := http.NewServer(config)
server.GET("/users", getUsersHandler)
server.Start()

// gRPC服务器
server := grpc.NewServer(config)
pb.RegisterUserServiceServer(server.Server(), &userService{})
server.Start()
```

## 🛠️ 开发指南

### 项目结构

```
framework/
├── core/           # 核心模块
├── cache/          # 缓存模块
│   └── docs/       # 模块文档
├── auth/           # 认证模块
│   └── docs/
├── logger/         # 日志模块
│   └── docs/
├── database/       # 数据库模块
│   └── docs/
├── middleware/     # 中间件模块
│   └── docs/
├── plugin/         # 插件模块
│   └── docs/
├── transport/      # 传输模块
│   └── docs/
├── metrics/        # 监控模块
├── tracing/        # 追踪模块
├── config/         # 配置模块
└── docs/           # 框架文档
```

### 文档规范

每个模块的文档应包含：

1. **README.md** - 主要设计文档
   - 模块概述和架构设计
   - 核心特点和功能说明
   - 详细使用示例
   - 配置选项和最佳实践

2. **ARCHITECTURE.md** - 架构设计文档（可选）
   - 详细的架构图和设计思路
   - 组件交互和数据流
   - 扩展性和性能考虑

3. **QUICKSTART.md** - 快速开始指南（可选）
   - 5分钟上手指南
   - 常见使用模式
   - 问题排查指南

### 代码规范

- 使用统一的错误处理模式
- 遵循Go语言最佳实践
- 添加完整的注释和文档
- 编写单元测试和集成测试
- 使用统一的日志格式

### 贡献指南

1. Fork项目
2. 创建功能分支
3. 编写代码和测试
4. 更新文档
5. 提交Pull Request

## 📊 性能基准

### 基础性能指标

| 组件 | 吞吐量 | 延迟 | 内存使用 |
|------|--------|------|----------|
| HTTP服务器 | 50K req/s | <1ms | 50MB |
| gRPC服务器 | 100K req/s | <0.5ms | 30MB |
| 内存缓存 | 2M op/s | <1μs | 可配置 |
| 日志系统 | 500K log/s | <2μs | 10MB |

### 扩展性指标

- **并发连接**: 支持10K+并发连接
- **插件数量**: 支持100+插件同时运行
- **集群规模**: 支持1000+节点集群
- **数据库连接**: 支持1000+连接池

## 🔗 相关链接

- [框架源码](https://github.com/qiaojinxia/distributed-service)
- [示例项目](../examples/)
- [API文档](./api/)
- [部署指南](./deployment/)
- [监控指南](./monitoring/)

## ❓ 常见问题

### Q: 如何选择合适的缓存策略？

**A:** 根据数据特点选择：
- **LRU**: 适用于热点数据缓存
- **TTL**: 适用于有时效性的数据
- **Simple**: 适用于配置类数据

### Q: 如何优化服务性能？

**A:** 参考以下建议：
- 使用连接池和对象池
- 启用异步日志和监控
- 合理配置缓存大小
- 使用批量操作

### Q: 如何进行故障排查？

**A:** 检查以下方面：
- 查看日志和监控指标
- 检查配置文件正确性
- 验证网络连接状态
- 分析性能瓶颈

---

> 📧 如有问题或建议，请联系框架维护团队或提交Issue。