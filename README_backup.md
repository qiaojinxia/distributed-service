# 🚀 企业级分布式微服务框架

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

一个基于 Go 的生产就绪分布式微服务框架，集成了完整的认证、API保护、监控、服务发现和容器化部署能力。

## ✨ 核心特性

### 🏗️ 微服务架构
- **分层架构设计** - Handler → Service → Repository → Model
- **双协议支持** - HTTP REST API + gRPC 服务并行运行
- **依赖注入** - 接口抽象和依赖解耦
- **上下文传递** - 完整的请求链路追踪
- **优雅关闭** - 支持平滑重启和资源清理

### 🚀 gRPC 服务
- **高性能通信** - 基于 HTTP/2 的二进制协议
- **类型安全** - Protocol Buffers 强类型接口定义
- **完整 API** - 用户管理的完整 CRUD 和认证功能
- **服务反射** - 开发环境支持服务发现和调试
- **中间件支持** - 日志、指标、错误恢复等中间件
- **健康检查** - 内置 gRPC 健康检查协议

### 🔐 安全认证
- **JWT 认证** - 基于 Token 的无状态认证
- **密码加密** - Bcrypt 安全哈希算法
- **权限控制** - 分级 API 访问权限
- **Token 刷新** - 自动续期机制

### 🛡️ API 保护机制
- **Sentinel 集成** - 基于阿里巴巴Sentinel的流量控制
- **HTTP/gRPC 双协议保护** - 统一的限流和熔断机制
- **多种限流策略** - QPS限流、并发限流、系统保护
- **智能熔断** - 基于错误率和响应时间的熔断策略
- **通配符匹配** - 支持路径模式匹配的保护规则
- **实时监控** - 详细的保护状态和指标统计

### 🗄️ 数据存储
- **MySQL** - 主数据库，支持事务和连接池
- **Redis** - 高性能缓存，支持集群
- **RabbitMQ** - 可靠消息队列，支持重连

### 🔧 基础设施
- **Consul** - 服务注册与发现
- **Prometheus** - 指标收集和监控
- **Grafana** - 可视化监控面板
- **Jaeger** - 分布式链路追踪
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
- **gRPC 指标** - gRPC方法调用统计和性能监控
- **数据库指标** - 查询时间、操作类型、表级别统计
- **缓存指标** - 命中率、响应时间、内存使用
- **分布式追踪** - OpenTelemetry + Jaeger 完整请求链路追踪

## 🚀 快速开始

### 本地开发

```bash
# 克隆项目
git clone https://github.com/yourusername/distributed-service
cd distributed-service

# 安装依赖
go mod tidy

# 启动本地开发
go run main.go
```

### Docker 部署（推荐）

```bash
# 一键部署
./deploy.sh

# 或手动部署
docker-compose up --build -d
```

📖 **详细部署指南**: [Docker 部署文档](docs/README-Docker.md)

## 📊 服务访问地址

| 服务 | 地址 | 用途 | 认证 |
|------|------|------|------|
| 🏠 主应用 | http://localhost:8080 | HTTP REST API 服务 | JWT |
| 🚀 gRPC 服务 | grpc://localhost:9090 | gRPC API 服务 | JWT |
| 📖 API 文档 | http://localhost:8080/swagger/index.html | Swagger UI | - |
| 🏥 健康检查 | http://localhost:8080/health | HTTP 服务状态 | - |
| 🏥 gRPC 健康检查 | grpc://localhost:9090/grpc.health.v1.Health/Check | gRPC 服务状态 | - |
| 📊 指标监控 | http://localhost:9090/metrics | Prometheus 指标 | - |
| 🔍 链路追踪 | http://localhost:16686 | Jaeger UI | - |
| 🗂️ 服务注册 | http://localhost:8500 | Consul UI | - |
| 🐰 消息队列 | http://localhost:15672 | RabbitMQ 管理 | guest/guest |
| 📈 监控系统 | http://localhost:9091 | Prometheus | - |
| 📊 可视化 | http://localhost:3000 | Grafana | admin/admin123 |

## 🔐 API 使用示例

### HTTP REST API 测试
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

### gRPC API 测试

#### 使用 grpcurl
```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 健康检查
grpcurl -plaintext -d '{"service": "user.v1.UserService"}' \
  localhost:9090 user.v1.UserService/Check

# 创建用户
grpcurl -plaintext -d '{
  "username": "grpcuser",
  "email": "grpc@example.com",
  "password": "password123"
}' localhost:9090 user.v1.UserService/CreateUser

# 用户登录
grpcurl -plaintext -d '{
  "username": "grpcuser",
  "password": "password123"
}' localhost:9090 user.v1.UserService/Login

# 查看可用服务
grpcurl -plaintext localhost:9090 list

# 查看服务方法
grpcurl -plaintext localhost:9090 list user.v1.UserService
```

📖 **完整gRPC文档**: [gRPC 使用指南](docs/README-gRPC.md)

## 🛡️ API 保护机制测试

项目集成了基于Sentinel的API保护机制，支持限流和熔断功能。

### 运行保护机制测试

```bash
# 进入测试目录
cd test

# 运行完整API保护测试套件
go test -v -run TestAPIProtectionWithRealConfig

# 运行特定测试
go test -v -run TestAPIProtectionWithRealConfig/TestAuthAPICircuitBreaker

# 使用测试脚本
./run_api_test.sh

# 快速演示
./demo_api_test.sh
```

### 测试覆盖功能

- **HTTP API 限流测试** - 验证不同端点的QPS限制
- **gRPC API 限流测试** - 验证gRPC方法的流量控制
- **熔断器测试** - 验证错误率触发的熔断机制
- **通配符匹配测试** - 验证路径模式匹配
- **并发测试** - 验证多线程环境下的保护机制

📖 **详细测试文档**: [API保护测试指南](test/README_API_Protection_Test.md)

## 🔍 分布式追踪

集成 OpenTelemetry 分布式追踪，支持 HTTP 和 gRPC 双协议追踪。

### 快速测试

```bash
# HTTP 追踪测试
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'X-Request-ID: http-trace-test' \
  -d '{"username":"traceuser","email":"trace@example.com","password":"password123"}'

# gRPC 追踪测试
grpcurl -plaintext -d '{"username":"grpc-trace-user","email":"grpc@example.com","password":"password123"}' \
  -H 'x-request-id: grpc-trace-test' \
  localhost:9090 user.v1.UserService/CreateUser
```

### 查看追踪链路
- 访问 [Jaeger UI](http://localhost:16686)
- 选择服务 `distributed-service`
- 查看完整的请求调用链

📖 **追踪详细文档**: [分布式追踪指南](docs/TRACING.md) | [gRPC追踪集成](docs/GRPC_TRACING.md)

## 📊 监控指标

### 指标类型
- **HTTP请求**: 请求数量、响应时间、状态码分布
- **gRPC请求**: gRPC方法调用数量、响应时间、状态码分布
- **数据库查询**: 查询时间、操作类型、表级别统计  
- **缓存性能**: 命中率、响应时间、内存使用
- **API保护**: 限流触发次数、熔断器状态变化

### 监控面板
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3000 (admin/admin123)
- **指标端点**: http://localhost:9090/metrics

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

4. **API保护配置**
   - 根据业务需求调整限流阈值
   - 监控熔断器触发频率
   - 为重要API设置更严格的保护规则

## 📚 文档导航

| 文档 | 内容 | 适用场景 |
|------|------|----------|
| [README.md](README.md) | 项目概览、核心功能、开发指南 | 了解项目、本地开发 |
| [README-Docker.md](docs/README-Docker.md) | Docker部署、运维、故障排查 | 容器化部署、生产运维 |
| [README-gRPC.md](docs/README-gRPC.md) | gRPC 服务使用指南 | gRPC 开发和调试 |
| [TRACING.md](docs/TRACING.md) | 分布式追踪详细说明 | 深入了解追踪功能 |
| [GRPC_TRACING.md](docs/GRPC_TRACING.md) | gRPC 分布式追踪集成 | gRPC 追踪专项指南 |
| [API保护测试指南](test/README_API_Protection_Test.md) | API保护机制测试详解 | 测试保护功能 |
| [测试套件概览](test/README_Test_Suite_Overview.md) | 完整测试套件说明 | 了解测试体系 |
| [Swagger UI](http://localhost:8080/swagger/index.html) | 在线API文档 | HTTP API接口调试 |

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
- [Sentinel](https://github.com/alibaba/sentinel-golang) - 流量控制和熔断降级 