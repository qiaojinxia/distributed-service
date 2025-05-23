# 🚀 企业级分布式微服务框架

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
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
- **指标收集** - HTTP、数据库、缓存指标
- **分布式追踪** - 完整请求链路
- **告警支持** - 集成 Prometheus AlertManager

## 🚀 快速开始

### 开发环境

```bash
# 克隆项目
git clone <repository-url>
cd distributed-service

# 安装依赖
go mod tidy

# 启动本地开发
go run main.go
```

### 一键部署（推荐）

```bash
# 使用 Docker Compose 一键部署
./deploy.sh
```

### 手动部署

```bash
# 构建并启动所有服务
docker-compose up --build -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

## 📊 服务访问地址

| 服务 | 地址 | 用途 | 认证 |
|------|------|------|------|
| 🏠 主应用 | http://localhost:8080 | API 服务 | JWT |
| 📖 API 文档 | http://localhost:8080/swagger/index.html | Swagger UI | - |
| 🏥 健康检查 | http://localhost:8080/health | 服务状态 | - |
| 📊 指标监控 | http://localhost:9090/metrics | Prometheus 指标 | - |
| 🗂️ 服务注册 | http://localhost:8500 | Consul UI | - |
| 🐰 消息队列 | http://localhost:15672 | RabbitMQ 管理 | guest/guest |
| 📈 监控系统 | http://localhost:9091 | Prometheus | - |
| 📊 可视化 | http://localhost:3000 | Grafana | admin/admin123 |

## 🔐 认证 API

### 用户注册
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"newuser","email":"user@example.com","password":"password123"}'
```

### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"newuser","password":"password123"}'
```

### 访问受保护 API
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -d '{"username":"protected","email":"protected@example.com","password":"password123"}'
```

### 修改密码
```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -d '{"old_password":"password123","new_password":"newpassword456"}'
```

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
├── 🗄️ 数据库脚本
│   └── scripts/
│       └── mysql-init.sql        # 数据库初始化
├── 📚 API 文档
│   └── docs/                     # Swagger 生成文档
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
│       └── registry/             # 服务注册
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
- 监控关键指标：响应时间、错误率、吞吐量

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