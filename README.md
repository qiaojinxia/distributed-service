# 🚀 企业级分布式微服务框架

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

一个基于 Go 的生产就绪分布式微服务框架，集成了完整的认证、监控、服务发现和容器化部署能力。

## ✨ 核心特性

### 🏗️ 微服务架构
- **分层架构设计** - Handler → Service → Repository → Model
- **依赖注入** - 接口抽象和依赖解耦
- **上下文传递** - 完整的请求链路追踪
- **优雅关闭** - 支持平滑重启和资源清理

### 🔐 安全认证
- **JWT 认证** - 基于 Token 的无状态认证
- **密码加密** - Bcrypt 安全哈希算法
- **权限控制** - 分级 API 访问权限
- **Token 刷新** - 自动续期机制

### 🛡️ 系统保护
- **API 限流** - 多层次限流保护（IP、用户、自定义）
  - 配置文件驱动的限流规则
  - 支持内存和Redis双存储后端
  - 端点级别的精细化限流控制
  - 标准HTTP响应头和429状态码
- **熔断器** - 防止服务雪崩和级联故障
- **降级处理** - 服务不可用时的优雅降级
- **实时监控** - 限流和熔断状态监控

### 🗄️ 数据存储
- **MySQL** - 主数据库，支持事务和连接池
- **Redis** - 高性能缓存，支持集群
- **RabbitMQ** - 可靠消息队列，支持重连

### 🔧 基础设施
- **Consul** - 服务注册与发现
- **Prometheus** - 指标收集和监控
- **Grafana** - 可视化监控面板
- **健康检查** - 自动故障检测

### 📚 API 文档
- **Swagger/OpenAPI** - 自动生成 API 文档
- **交互式测试** - 在线 API 调试
- **认证支持** - Bearer Token 集成

### 🐳 容器化部署
- **Docker** - 多阶段构建优化
- **Docker Compose** - 一键部署全栈
- **健康检查** - 容器自动恢复
- **数据持久化** - 卷管理和备份

### 📊 监控日志
- **结构化日志** - 基于 Zap 的高性能日志
- **HTTP 指标** - 请求数量、响应时间、状态码分布
- **数据库指标** - 查询时间、操作类型、表级别统计
- **缓存指标** - 命中率、响应时间、内存使用
- **分布式追踪** - OpenTelemetry + Jaeger 完整请求链路追踪
- **告警支持** - 集成 Prometheus AlertManager

## 🚀 快速开始

### 本地开发

```bash
# 克隆项目
git clone <repository-url>
cd distributed-service

# 安装依赖
go mod tidy

# 启动本地开发
go run main.go
```

### Docker 部署

```bash
# 一键部署（推荐）
./deploy.sh

# 或手动部署
docker-compose up --build -d
```

📖 **详细部署指南**: [Docker 部署文档](README-Docker.md)

## 📊 服务访问地址

| 服务 | 地址 | 用途 | 认证 |
|------|------|------|------|
| 🏠 主应用 | http://localhost:8080 | API 服务 | JWT |
| 📖 API 文档 | http://localhost:8080/swagger/index.html | Swagger UI | - |
| 🏥 健康检查 | http://localhost:8080/health | 服务状态 | - |
| 📊 指标监控 | http://localhost:9090/metrics | Prometheus 指标 | - |
| 🛡️ 熔断器状态 | http://localhost:8080/circuit-breaker/status | 熔断器监控 | - |
| 📡 Hystrix 流 | http://localhost:8080/hystrix | 实时监控流 | - |
| 🔍 链路追踪 | http://localhost:16686 | Jaeger UI | - |
| 🗂️ 服务注册 | http://localhost:8500 | Consul UI | - |
| 🐰 消息队列 | http://localhost:15672 | RabbitMQ 管理 | guest/guest |
| 📈 监控系统 | http://localhost:9091 | Prometheus | - |
| 📊 可视化 | http://localhost:3000 | Grafana | admin/admin123 |

## 🔐 API 使用示例

### 快速测试
```bash
# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"newuser","email":"user@example.com","password":"password123"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"newuser","password":"password123"}'

# 访问受保护 API
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -d '{"username":"protected","email":"protected@example.com","password":"password123"}'
```

📖 **完整API文档**: [Swagger UI](http://localhost:8080/swagger/index.html) | [部署文档](README-Docker.md)

## 🔍 分布式追踪

### 快速测试
```bash
# 运行追踪测试
./scripts/test-tracing.sh

# 生成追踪数据
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'X-Request-ID: trace-test' \
  -d '{"username":"traceuser","email":"trace@example.com","password":"password123"}'
```

### 查看追踪链路
- 访问 [Jaeger UI](http://localhost:16686)
- 选择服务 `distributed-service`
- 查看完整的请求调用链

### 追踪覆盖
- **HTTP层**: 请求路径、状态码、响应时间
- **Service层**: 业务逻辑执行时间
- **Repository层**: 数据库操作和SQL执行时间
- **错误追踪**: 异常信息和错误堆栈

📖 **详细文档**: [分布式追踪使用指南](docs/TRACING.md)

## 📊 监控指标

### 指标类型
- **HTTP请求**: 请求数量、响应时间、状态码分布
- **数据库查询**: 查询时间、操作类型、表级别统计  
- **缓存性能**: 命中率、响应时间、内存使用
- **分布式追踪**: 完整请求链路追踪

### 快速验证
```bash
# 测试数据库指标
./scripts/test-metrics.sh

# 查看指标数据
curl http://localhost:9090/metrics | grep database_query_duration
```

### 监控面板
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3000 (admin/admin123)
- **指标端点**: http://localhost:9090/metrics

## 🛡️ 系统保护

### 功能特性
- **多层限流**: IP限流、用户限流、自定义限流
- **智能熔断**: 防止服务雪崩和级联故障
- **优雅降级**: 服务不可用时的备用响应
- **实时监控**: 限流和熔断状态可视化

### 限流配置
项目支持通过配置文件管理所有限流规则，支持多种存储后端：

```yaml
ratelimit:
  enabled: true                    # 启用限流
  store_type: redis               # memory 或 redis
  redis_prefix: "ratelimit:"      # Redis键前缀
  default_config:                 # 默认配置
    health_check: "10-S"          # 健康检查：10次/秒
    auth_public: "20-M"           # 认证端点：20次/分钟
    auth_protected: "10-M"        # 保护端点：10次/分钟
    user_public: "30-M"           # 用户公开API：30次/分钟
    user_protected: "50-M"        # 用户保护API：50次/分钟
  endpoints:                      # 端点特定配置
    "/health": "10-S"
    "/api/v1/auth/register": "20-M"
    "/api/v1/auth/login": "20-M"
    # ... 更多端点配置
```

### 快速测试
```bash
# 1. 运行基础限流功能测试
./scripts/test-ratelimit.sh
# 测试内容: IP限流、端点限流、响应头验证、配置验证

# 2. 运行系统保护综合测试 
./scripts/test-ratelimit-circuitbreaker.sh
# 测试内容: 限流+熔断器、Redis存储、监控端点、错误处理

# 3. 手动快速验证限流
for i in {1..15}; do curl -w "HTTP_%{http_code}\n" http://localhost:8080/health; sleep 0.1; done
# 预期: 前10个返回200，后5个返回429

# 4. 验证配置文件正确性
./scripts/validate-config.sh
# 检查: 配置文件完整性、限流格式、编译验证

# 5. 查看熔断器实时状态
curl http://localhost:8080/circuit-breaker/status | jq '.'
# 显示: 熔断器状态、错误率、请求量等
```

### 测试脚本特性
- ✅ **服务健康检查**: 自动检测服务可用性，支持重试机制
- ✅ **彩色输出**: 直观的测试结果展示，成功/失败/警告状态区分
- ✅ **详细报告**: 测试通过率统计，问题诊断建议
- ✅ **错误处理**: 优雅的错误处理和回退机制
- ✅ **配置验证**: 自动验证配置文件完整性和格式正确性

### 限流策略
- **IP限流**: 基于客户端IP地址，适用于公开API
- **用户限流**: 基于JWT token中的用户ID，适用于认证API
- **自定义限流**: 支持自定义键生成函数的灵活限流
- **端点限流**: 根据配置文件自动应用不同规则

### 存储后端
- **内存存储**: 适用于开发环境和单实例部署
- **Redis存储**: 适用于生产环境和分布式部署，支持自动降级

### 配置概览
- **健康检查**: 10次/秒 IP限流
- **认证端点**: 20次/分钟 IP限流  
- **用户API**: 30-50次/分钟 用户限流
- **熔断保护**: 2-5秒超时，25-50%错误率阈值

### 响应头信息
限流触发时会返回标准的HTTP响应头：
- `X-RateLimit-Limit`: 限流上限
- `X-RateLimit-Remaining`: 剩余请求数
- `X-RateLimit-Reset`: 重置时间戳
- `HTTP 429`: Too Many Requests 状态码

### 监控端点
- **熔断器状态**: http://localhost:8080/circuit-breaker/status
- **Hystrix流**: http://localhost:8080/hystrix

📖 **详细文档**: [限流功能详细说明](docs/RATELIMIT.md) | [限流和熔断器使用指南](docs/RATELIMIT-CIRCUITBREAKER.md)

## 📁 项目结构

```
distributed-service/
├── 🐳 容器化部署
│   ├── Dockerfile                 # 多阶段构建配置
│   ├── docker-compose.yaml       # 服务编排配置
│   ├── deploy.sh                 # 一键部署脚本
│   └── .dockerignore             # Docker 忽略文件
├── ⚙️ 配置管理
│   └── config/
│       ├── config.yaml           # 开发环境配置
│       ├── config-docker.yaml    # 生产环境配置
│       ├── redis.conf            # Redis 配置
│       └── prometheus.yml        # 监控配置
├── 🗄️ 脚本工具
│   └── scripts/
│       ├── mysql-init.sql        # 数据库初始化
│       ├── test-tracing.sh       # 分布式追踪测试脚本
│       ├── test-metrics.sh       # 数据库指标测试脚本
│       ├── test-ratelimit.sh     # 限流功能测试脚本(改进版)
│       ├── test-ratelimit-circuitbreaker.sh  # 限流和熔断器综合测试脚本(改进版)
│       └── validate-config.sh    # 配置验证脚本
├── 📚 API 文档
│   └── docs/                     # Swagger 生成文档
│       ├── TRACING.md            # 分布式追踪使用指南
│       ├── RATELIMIT.md          # 限流功能详细文档
│       └── RATELIMIT-SUMMARY.md  # 限流功能实现总结
├── 🏗️ 应用代码
│   ├── internal/                 # 内部业务逻辑
│   │   ├── api/                  # HTTP 处理层
│   │   │   ├── auth.go           # 认证接口
│   │   │   ├── user.go           # 用户接口
│   │   │   └── router.go         # 路由配置
│   │   ├── service/              # 业务逻辑层
│   │   │   └── user.go           # 用户服务
│   │   ├── repository/           # 数据访问层
│   │   │   └── user.go           # 用户仓库
│   │   └── model/                # 数据模型层
│   │       ├── user.go           # 用户模型
│   │       └── auth.go           # 认证模型
│   └── pkg/                      # 公共组件包
│       ├── config/               # 配置管理
│       ├── database/             # 数据库连接
│       ├── logger/               # 日志管理
│       ├── middleware/           # 中间件
│       ├── metrics/              # 指标收集
│       ├── auth/                 # 认证组件
│       ├── mq/                   # 消息队列
│       ├── registry/             # 服务注册
│       ├── ratelimit/            # API限流
│       ├── circuitbreaker/       # 熔断器
│       └── tracing/              # 分布式链路追踪
├── go.mod                        # Go 模块依赖
├── go.sum                        # 依赖校验文件
├── main.go                       # 应用入口
├── README.md                     # 项目文档
└── README-Docker.md              # 部署文档
```

## 🛠️ 开发指南

### 添加新的 API 端点

1. **定义数据模型** (`internal/model/`)
```go
type YourModel struct {
    ID   uint   `json:"id" gorm:"primarykey"`
    Name string `json:"name"`
}
```

2. **实现数据访问** (`internal/repository/`)
```go
type YourRepository interface {
    Create(ctx context.Context, model *YourModel) error
}
```

3. **编写业务逻辑** (`internal/service/`)
```go
type YourService interface {
    Create(ctx context.Context, req *CreateRequest) error
}
```

4. **创建 HTTP 处理器** (`internal/api/`)
```go
// @Summary Create item
// @Router /api/v1/items [post]
func (h *YourHandler) Create(c *gin.Context) {
    // 实现逻辑
}
```

5. **注册路由** (`internal/api/router.go`)
```go
items := v1.Group("/items")
items.POST("", handler.Create)
```

6. **生成文档**
```bash
swag init
```

### 添加限流和熔断器保护

在新的API端点中添加限流和熔断器保护：

```go
// 在路由注册时添加限流中间件
authBase := v1.Group("/auth")
authBase.Use(rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("auth_public"))) // 使用配置的限流规则
{
    authBase.POST("/register", 
        circuitBreaker.Middleware("auth_register", nil),
        authHandler.Register)
}

// 不同类型的限流策略
// 1. 使用配置的限流规则
r.GET("/health", rateLimiter.IPRateLimit(rateLimiter.GetConfiguredLimit("health_check")), handler)

// 2. 使用自定义限流规则
r.POST("/api/v1/upload", rateLimiter.IPRateLimit("5-M"), handler) // 每分钟5次

// 3. 用户级别限流（需要JWT认证）
protectedGroup := v1.Group("/protected")
protectedGroup.Use(middleware.JWTAuth(jwtManager))
protectedGroup.Use(rateLimiter.UserRateLimit(rateLimiter.GetConfiguredLimit("user_protected")))

// 4. 端点特定限流（从配置文件读取）
r.POST("/special", rateLimiter.EndpointRateLimit("/special"), handler)

// 配置自定义熔断器
circuitbreaker.ConfigureCommand("my_service", circuitbreaker.Config{
    Timeout:                3000, // 3秒超时
    MaxConcurrentRequests:  50,   // 最大50并发
    RequestVolumeThreshold: 10,   // 10个请求后开始统计
    SleepWindow:            5000, // 5秒休眠窗口
    ErrorPercentThreshold:  30,   // 30%错误率阈值
})
```

### 限流配置管理

在配置文件中添加新的端点限流规则：

```yaml
# config/config.yaml 或 config/config-docker.yaml
ratelimit:
  enabled: true
  store_type: redis  # 生产环境使用redis，开发环境可以使用memory
  redis_prefix: "ratelimit:"
  default_config:
    health_check: "10-S"
    auth_public: "20-M"
    auth_protected: "10-M"
    user_public: "30-M"
    user_protected: "50-M"
    custom_api: "100-H"        # 新增：自定义API每小时100次
  endpoints:
    "/health": "10-S"
    "/api/v1/auth/register": "20-M"
    "/api/v1/auth/login": "20-M"
    "/api/v1/special": "5-M"   # 新增：特殊端点每分钟5次
    "/api/v1/upload": "10-H"   # 新增：上传端点每小时10次
```

**限流格式说明**：
- `10-S`: 每秒10次
- `20-M`: 每分钟20次  
- `100-H`: 每小时100次
- `1000-D`: 每天1000次

### 添加数据库指标监控

在新的 Repository 层添加数据库指标记录：

```go
import "distributed-service/pkg/metrics"

// Repository 层示例
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    // 使用 MeasureDatabaseQuery 包装数据库操作
    err := metrics.MeasureDatabaseQuery("CREATE", "users", func() error {
        return r.db.WithContext(ctx).Create(user).Error
    })
    return err
}

// 带返回值的查询操作
func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
    user, err := metrics.MeasureDatabaseQueryWithResult("SELECT", "users", func() (*model.User, error) {
        var user model.User
        err := r.db.WithContext(ctx).First(&user, id).Error
        return &user, err
    })
    return user, err
}
```

**指标标签规范**：
- `operation`: CREATE, SELECT, UPDATE, DELETE
- `table`: 实际的表名（如：users, orders 等）

### 配置管理

修改配置文件后重启应用：
```bash
# 开发环境
vim config/config.yaml
go run main.go

# Docker 环境
vim config/config-docker.yaml
docker-compose restart app
```

## 🔍 故障排查

### 查看服务状态
```bash
docker-compose ps
```

### 查看应用日志
```bash
docker-compose logs -f app
```

### 查看特定服务日志
```bash
docker-compose logs mysql
docker-compose logs redis
docker-compose logs rabbitmq
```

### 重启服务
```bash
docker-compose restart app
```

### 完全重新部署
```bash
docker-compose down -v --remove-orphans
docker-compose up --build -d
```

## 📈 性能优化

### 数据库优化
- 连接池配置：`max_idle_conns`, `max_open_conns`
- 索引优化：为常用查询字段添加索引
- 查询优化：使用 GORM 的预加载和选择字段

### 缓存策略
- Redis 缓存热点数据
- 设置合理的过期时间
- 使用缓存击穿保护

### 监控告警
- 设置 Prometheus 告警规则
- 配置 Grafana 监控面板
- 监控关键指标：响应时间、错误率、吞吐量、数据库性能

### 数据库性能监控
- 查询响应时间分布
- 各操作类型（CRUD）的执行频率
- 表级别的查询统计
- 慢查询识别和优化

## 🔒 安全建议

### 生产环境配置
1. **修改默认密码**
   - JWT 密钥：`config.jwt.secret_key`
   - 数据库密码：`config.mysql.password`
   - Grafana 密码：`GF_SECURITY_ADMIN_PASSWORD`

2. **网络安全**
   - 使用 HTTPS/TLS 加密通信
   - 配置防火墙规则
   - 限制端口访问

3. **认证安全**
   - 设置合理的 Token 过期时间
   - 实现 Token 黑名单机制
   - 添加 API 限流保护

4. **限流安全配置**
   - 根据业务需求调整限流阈值
   - 生产环境使用 Redis 存储确保分布式一致性
   - 监控429状态码的频率，及时调整限流策略
   - 为重要API设置更严格的限流规则

## 📚 文档导航

| 文档 | 内容 | 适用场景 |
|------|------|----------|
| [README.md](README.md) | 项目概览、核心功能、开发指南 | 了解项目、本地开发 |
| [README-Docker.md](README-Docker.md) | Docker部署、运维、故障排查 | 容器化部署、生产运维 |
| [TRACING.md](docs/TRACING.md) | 分布式追踪详细说明 | 深入了解追踪功能 |
| [RATELIMIT.md](docs/RATELIMIT.md) | 限流功能完整文档 | 限流配置和使用 |
| [RATELIMIT-SUMMARY.md](docs/RATELIMIT-SUMMARY.md) | 限流功能实现总结 | 技术实现详情 |
| [RATELIMIT-CIRCUITBREAKER.md](docs/RATELIMIT-CIRCUITBREAKER.md) | 限流熔断器详细配置 | 系统保护机制配置 |
| [Swagger UI](http://localhost:8080/swagger/index.html) | 在线API文档 | API接口调试 |

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

感谢以下开源项目：
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 高性能日志库
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT 实现
- [Consul](https://github.com/hashicorp/consul) - 服务发现
- [Prometheus](https://github.com/prometheus/prometheus) - 监控系统 