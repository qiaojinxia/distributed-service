# 分布式微服务 Docker 部署指南

这是一个完整的分布式微服务应用的 Docker 容器化部署方案，包含了 MySQL、Redis、RabbitMQ、Consul、Prometheus、Grafana 等完整的服务栈。

## 🏗️ 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    分布式微服务架构                          │
├─────────────────────────────────────────────────────────────┤
│  📱 应用服务 (Go + Gin)                                     │
│  ├── API 层 (RESTful + Swagger)                            │
│  ├── 服务层 (Business Logic)                               │
│  ├── 仓库层 (Data Access)                                  │
│  └── 模型层 (Data Models)                                  │
├─────────────────────────────────────────────────────────────┤
│  🗄️ 数据存储                                               │
│  ├── MySQL (主数据库)                                       │
│  ├── Redis (缓存)                                          │
│  └── RabbitMQ (消息队列)                                    │
├─────────────────────────────────────────────────────────────┤
│  🔧 基础设施                                                │
│  ├── Consul (服务注册与发现)                                │
│  ├── Prometheus (监控指标收集)                              │
│  └── Grafana (可视化面板)                                   │
└─────────────────────────────────────────────────────────────┘
```

## 📋 前置要求

- Docker >= 20.10
- Docker Compose >= 1.29
- 至少 4GB 可用内存
- 至少 10GB 可用磁盘空间

## 🚀 一键部署

### 方式一：使用部署脚本（推荐）

```bash
./deploy.sh
```

### 方式二：手动部署

```bash
# 构建并启动所有服务
docker-compose up --build -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

## 📊 服务访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 主应用 | http://localhost:8080 | API 服务 |
| API 文档 | http://localhost:8080/swagger/index.html | Swagger UI |
| 健康检查 | http://localhost:8080/health | 服务健康状态 |
| 指标监控 | http://localhost:9090/metrics | Prometheus 指标 |
| 服务注册中心 | http://localhost:8500 | Consul UI |
| 消息队列管理 | http://localhost:15672 | RabbitMQ Management (guest/guest) |
| 监控系统 | http://localhost:9091 | Prometheus UI |
| 可视化面板 | http://localhost:3000 | Grafana (admin/admin123) |

## 🧪 API 测试

### 创建用户
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```

### 获取用户
```bash
curl http://localhost:8080/api/v1/users/1
```

### 删除用户
```bash
curl -X DELETE http://localhost:8080/api/v1/users/1
```

## 🔧 配置说明

### 环境配置文件

- `config/config.yaml` - 本地开发配置
- `config/config-docker.yaml` - Docker 环境配置
- `config/redis.conf` - Redis 配置
- `config/prometheus.yml` - Prometheus 配置

### 数据库配置

MySQL 默认配置：
- 用户名: `root`
- 密码: `root`
- 数据库: `distributed_service`
- 端口: `3306`

### 消息队列配置

RabbitMQ 默认配置：
- 用户名: `guest`
- 密码: `guest`
- 端口: `5672` (AMQP), `15672` (Management)

## 📁 目录结构

```
distributed-service/
├── Dockerfile                    # 应用构建文件
├── docker-compose.yaml          # 服务编排文件
├── deploy.sh                    # 一键部署脚本
├── .dockerignore                # Docker 忽略文件
├── config/                      # 配置文件目录
│   ├── config.yaml             # 本地配置
│   ├── config-docker.yaml      # Docker 配置
│   ├── redis.conf              # Redis 配置
│   └── prometheus.yml          # Prometheus 配置
├── scripts/                     # 初始化脚本
│   └── mysql-init.sql          # MySQL 初始化脚本
├── internal/                    # 应用源码
│   ├── api/                    # API 层
│   ├── service/                # 服务层
│   ├── repository/             # 仓库层
│   └── model/                  # 模型层
├── pkg/                        # 公共包
│   ├── config/                 # 配置管理
│   ├── database/               # 数据库连接
│   ├── logger/                 # 日志管理
│   ├── middleware/             # 中间件
│   ├── metrics/                # 指标收集
│   ├── mq/                     # 消息队列
│   └── registry/               # 服务注册
└── docs/                       # Swagger 文档
```

## 🛠️ 常用命令

### 查看服务状态
```bash
docker-compose ps
```

### 查看应用日志
```bash
docker-compose logs -f app
```

### 查看所有服务日志
```bash
docker-compose logs -f
```

### 重启服务
```bash
docker-compose restart app
```

### 停止所有服务
```bash
docker-compose down
```

### 重新构建应用
```bash
docker-compose up --build -d app
```

### 清理所有数据
```bash
docker-compose down -v --remove-orphans
```

## 🔍 故障排查

### 应用启动失败

1. 检查配置文件是否正确
2. 确保数据库连接正常
3. 查看应用日志：`docker-compose logs app`

### 数据库连接失败

1. 检查 MySQL 容器状态：`docker-compose ps mysql`
2. 查看 MySQL 日志：`docker-compose logs mysql`
3. 确保配置文件中的数据库连接信息正确

### 服务注册失败

1. 检查 Consul 容器状态：`docker-compose ps consul`
2. 访问 Consul UI：http://localhost:8500
3. 确保网络连接正常

## 📈 监控和指标

### Prometheus 指标
- HTTP 请求数量和响应时间
- 数据库查询性能
- 缓存命中率
- 系统资源使用情况

### Grafana 面板
默认账号：`admin` / `admin123`

可以导入预定义的面板来监控服务性能。

## 🔐 安全注意事项

1. **生产环境部署时，请修改所有默认密码**
2. **配置防火墙规则，限制端口访问**
3. **使用 HTTPS 和 TLS 加密通信**
4. **定期更新镜像和依赖**
5. **配置日志轮转和监控告警**

## 📝 开发指南

### 添加新的 API 端点

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中实现数据访问
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/api/` 中实现 HTTP 处理器
5. 在 `internal/api/router.go` 中注册路由
6. 重新生成 Swagger 文档：`swag init`

### 修改配置

1. 更新 `config/config-docker.yaml`
2. 重启应用：`docker-compose restart app`

## 🆘 支持

如有问题，请查看：
1. 应用日志
2. 各服务的健康检查状态
3. Prometheus 监控指标
4. 本文档的故障排查部分 