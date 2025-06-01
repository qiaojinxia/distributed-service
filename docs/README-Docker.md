# 🚀 Docker 容器化部署指南

本文档详细介绍如何使用 Docker 和 Docker Compose 部署分布式微服务项目，包括完整的基础设施堆栈和服务配置。

## 📋 目录

- [快速部署](#快速部署)
- [部署模式](#部署模式)
- [服务架构](#服务架构)
- [详细配置](#详细配置)
- [API测试验证](#API测试验证)
- [监控和追踪](#监控和追踪)
- [故障排查](#故障排查)
- [生产环境建议](#生产环境建议)

## 🚀 快速部署

### 一键部署（推荐）

```bash
# 克隆项目
git clone https://github.com/yourusername/distributed-service
cd distributed-service

# 执行一键部署脚本
./deploy.sh
```

### 手动部署

```bash
# 停止现有服务
docker-compose down --remove-orphans

# 构建并启动所有服务
docker-compose up --build -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

## 🎯 部署模式

部署脚本支持两种模式，满足不同的使用场景：

### 1️⃣ **仅基础设施模式** (本地调试)
```bash
./deploy.sh
# 选择 1 - 仅基础设施
```

**启动的服务：**
- 🗃️ MySQL 数据库
- 🚀 Redis 缓存  
- 🐰 RabbitMQ 消息队列
- 🗂️ Consul 服务发现
- 📊 Prometheus 监控
- 📈 Grafana 可视化
- 🔍 Jaeger 链路追踪

**用途：**
- 本地开发调试
- 热重载开发 (`go run main.go`)
- IDE 调试支持

### 2️⃣ **完整部署模式** (生产环境)
```bash
./deploy.sh  
# 选择 2 - 完整部署 (默认)
```

**启动的服务：**
- 🏠 应用程序容器 + 所有基础设施服务

**用途：**
- 生产环境部署
- 容器化测试
- CI/CD 流水线

## 🏗️ 服务架构

### 核心服务栈

```
┌─────────────────────────────────────────────────────────────┐
│                   分布式微服务架构                          │
├─────────────────────────────────────────────────────────────┤
│  🌐 API网关层                                               │
│  ├── HTTP API Gateway (Port: 8080)                         │
│  └── gRPC API Gateway (Port: 9090)                         │
├─────────────────────────────────────────────────────────────┤
│  🛡️ API 保护层                                              │
│  ├── Sentinel 限流控制                                      │
│  ├── 熔断器保护                                             │
│  ├── JWT 认证和授权                                         │
│  └── 通配符路径匹配                                         │
├─────────────────────────────────────────────────────────────┤
│  💼 应用服务层                                              │
│  ├── 分布式服务主应用 (Go)                                  │
│  ├── HTTP REST API                                         │
│  ├── gRPC API                                              │
│  └── 业务逻辑处理                                           │
├─────────────────────────────────────────────────────────────┤
│  🗄️ 数据存储层                                              │
│  ├── MySQL 8.0 (数据持久化)                                │
│  ├── Redis 7.0 (缓存和会话)                                │
│  └── RabbitMQ 3.12 (消息队列)                              │
├─────────────────────────────────────────────────────────────┤
│  🔧 基础设施层                                              │
│  ├── Consul (服务发现)                                     │
│  ├── Prometheus (指标收集)                                 │
│  ├── Grafana (监控面板)                                    │
│  └── Jaeger (分布式追踪)                                   │
└─────────────────────────────────────────────────────────────┘
```

### 容器服务详情

| 服务 | 镜像 | 端口 | 状态检查 | 用途 |
|------|------|------|----------|------|
| 🏠 app | distributed-service:latest | 8080, 9090 | /health | 主应用 (HTTP + gRPC) |
| 🗃️ mysql | mysql:8.0 | 3306 | mysqladmin ping | 数据库 |
| 🚀 redis | redis:7.0-alpine | 6379 | redis-cli ping | 缓存 |
| 🐰 rabbitmq | rabbitmq:3.12-management | 5672, 15672 | rabbitmq-diagnostics ping | 消息队列 |
| 🗂️ consul | consul:1.16 | 8500, 8600 | /v1/status/leader | 服务发现 |
| 📊 prometheus | prom/prometheus:latest | 9091 | /-/healthy | 指标监控 |
| 📈 grafana | grafana/grafana:latest | 3000 | /api/health | 可视化 |
| 🔍 jaeger | jaegertracing/all-in-one:latest | 16686, 14268 | / | 分布式追踪 |

### 网络配置

- **自定义网络**: `distributed-network`
- **服务发现**: 通过服务名进行容器间通信
- **端口映射**: 仅必要端口对外暴露
- **健康检查**: 所有服务配置健康检查机制

## 📊 服务访问地址

| 🎯 服务类型 | 📍 访问地址 | 🔐 认证 | 📝 说明 |
|------------|-------------|---------|---------|
| **🏠 核心服务** | | | |
| HTTP REST API | http://localhost:8080 | JWT | 主要业务 API |
| gRPC API | grpc://localhost:9090 | JWT | 高性能 gRPC 接口 |
| **📚 文档和监控** | | | |
| API 文档 | http://localhost:8080/swagger/index.html | - | Swagger UI |
| 健康检查 | http://localhost:8080/health | - | 服务状态 |
| gRPC 健康检查 | grpc://localhost:9090/grpc.health.v1.Health/Check | - | gRPC 服务状态 |
| Prometheus 指标 | http://localhost:9090/metrics | - | 指标导出 |
| **🔍 监控基础设施** | | | |
| 链路追踪 | http://localhost:16686 | - | Jaeger UI |
| 服务注册中心 | http://localhost:8500 | - | Consul UI |
| 消息队列管理 | http://localhost:15672 | guest/guest | RabbitMQ 管理界面 |
| 监控系统 | http://localhost:9091 | - | Prometheus |
| 可视化面板 | http://localhost:3000 | admin/admin123 | Grafana |

## 🧪 API测试验证

### HTTP REST API

#### 认证流程测试
```bash
# 1. 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"dockeruser","email":"docker@example.com","password":"password123"}'

# 2. 用户登录
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"dockeruser","password":"password123"}' | jq -r '.data.token')

# 3. 访问受保护 API
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Authorization: Bearer '$TOKEN \
  -H 'Content-Type: application/json' \
  -d '{"username":"protecteduser","email":"protected@example.com","password":"password123"}'

# 4. 获取用户信息
curl http://localhost:8080/api/v1/users/1
```

#### API保护机制测试
```bash
# 快速限流测试 - 健康检查端点 (2 QPS限制)
echo "🧪 测试健康检查限流 (预期前2个成功，后3个被限流):"
for i in {1..5}; do
  echo -n "请求 $i: "
  curl -w "HTTP_%{http_code}\n" -s -o /dev/null http://localhost:8080/health
  sleep 0.1
done

# 认证接口限流测试 (10次/分钟)
echo -e "\n🧪 测试认证接口限流:"
for i in {1..15}; do
  echo -n "注册请求 $i: "
  curl -X POST http://localhost:8080/api/v1/auth/register \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"testuser$i\",\"email\":\"test$i@example.com\",\"password\":\"password123\"}" \
    -w "HTTP_%{http_code}\n" -s -o /dev/null
  sleep 3
done
```

### gRPC API

#### gRPC 基本功能测试
```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 1. gRPC 健康检查
grpcurl -plaintext -d '{"service": "user.v1.UserService"}' \
  localhost:9090 grpc.health.v1.Health/Check

# 2. 创建用户
grpcurl -plaintext -d '{
  "username": "grpcuser",
  "email": "grpc@example.com", 
  "password": "password123"
}' localhost:9090 user.v1.UserService/CreateUser

# 3. 用户登录
grpcurl -plaintext -d '{
  "username": "grpcuser",
  "password": "password123"
}' localhost:9090 user.v1.UserService/Login

# 4. 查看可用服务
grpcurl -plaintext localhost:9090 list

# 5. 查看服务方法
grpcurl -plaintext localhost:9090 list user.v1.UserService
```

#### gRPC 保护机制测试
```bash
# gRPC 限流测试
echo "🧪 测试 gRPC 限流保护:"
for i in {1..30}; do
  echo -n "gRPC请求 $i: "
  grpcurl -plaintext -d "{\"username\":\"test$i\",\"email\":\"test$i@example.com\",\"password\":\"password123\"}" \
    localhost:9090 user.v1.UserService/CreateUser 2>&1 | \
    grep -o "Code: [A-Z_]*" || echo "SUCCESS"
  sleep 0.1
done

# gRPC 熔断器测试 - 访问不存在的用户触发错误
echo -e "\n🧪 测试 gRPC 熔断器:"
for i in {1..20}; do
  echo -n "熔断测试 $i: "
  grpcurl -plaintext -d '{"id": 999999}' \
    localhost:9090 user.v1.UserService/GetUser 2>&1 | \
    grep -o "Code: [A-Z_]*" || echo "SUCCESS"
  sleep 0.1
done
```

## 🛡️ API保护测试套件

### 容器内测试
```bash
# 进入应用容器
docker-compose exec app /bin/sh

# 运行API保护测试
cd test && go test -v -run TestAPIProtectionWithRealConfig
```

### 主机测试（推荐）
```bash
# 在主机上直接运行（需要安装Go）
cd test

# 运行完整测试套件
go test -v -run TestAPIProtectionWithRealConfig

# 运行特定测试
go test -v -run TestAPIProtectionWithRealConfig/TestAuthAPICircuitBreaker
go test -v -run TestAPIProtectionWithRealConfig/TestHealthCheckRateLimit

# 使用测试脚本
./run_api_test.sh      # 完整测试
./demo_api_test.sh     # 快速演示
```

### 测试覆盖范围

- ✅ **HTTP限流测试** - 验证健康检查、认证、用户等端点的限流
- ✅ **熔断器测试** - 验证基于错误率的智能熔断
- ✅ **通配符匹配** - 验证路径模式匹配规则
- ✅ **并发安全** - 验证多线程环境下的保护机制
- ✅ **优先级匹配** - 验证具体路径优先于通配符路径

## 🔍 监控和追踪

### 自动化追踪测试

```bash
# 使用现有的追踪测试脚本
./scripts/test-tracing.sh       # 完整追踪测试
./scripts/verify-tracing.sh     # 快速验证
./scripts/test-metrics.sh       # 数据库指标测试
```

### 手动分布式追踪测试

```bash
# HTTP 追踪测试
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -H 'X-Request-ID: docker-http-trace' \
  -d '{"username":"traceuser","email":"trace@example.com","password":"password123"}'

# gRPC 追踪测试
grpcurl -plaintext \
  -H 'x-request-id: docker-grpc-trace' \
  -d '{"username":"grpc-trace","email":"grpc@example.com","password":"password123"}' \
  localhost:9090 user.v1.UserService/CreateUser

# 查看追踪结果
echo "📊 访问 Jaeger UI 查看追踪: http://localhost:16686"
```

### 监控指标验证

```bash
# 查看应用指标
curl http://localhost:9090/metrics | grep -E "(http_requests_total|grpc_requests_total|database_query_duration)"

# 查看Prometheus目标状态
curl http://localhost:9091/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# 验证 Grafana 连接
curl -u admin:admin123 http://localhost:3000/api/health
```

### 监控面板访问

- **📊 Prometheus**: http://localhost:9091 - 指标收集和查询
- **📈 Grafana**: http://localhost:3000 (admin/admin123) - 可视化监控面板
- **🔍 Jaeger**: http://localhost:16686 - 分布式追踪分析
- **🗂️ Consul**: http://localhost:8500 - 服务注册和发现
- **🐰 RabbitMQ**: http://localhost:15672 (guest/guest) - 消息队列管理

## 🔍 故障排查

### 常见问题排查

#### 1. 服务启动问题
```bash
# 查看所有服务状态
docker-compose ps

# 查看特定服务日志
docker-compose logs app
docker-compose logs mysql
docker-compose logs redis

# 查看服务健康状态
docker-compose exec app wget -qO- http://localhost:8080/health
```

#### 2. 数据库连接问题
```bash
# 检查MySQL连接
docker-compose exec mysql mysql -u testuser -ptestpass -e "SHOW DATABASES;"

# 查看MySQL日志
docker-compose logs mysql | tail -20

# 重启数据库
docker-compose restart mysql
```

#### 3. Redis连接问题
```bash
# 检查Redis连接
docker-compose exec redis redis-cli ping

# 查看Redis配置
docker-compose exec redis redis-cli config get "*"
```

#### 4. gRPC服务问题
```bash
# 测试gRPC健康检查
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check

# 查看gRPC服务列表
grpcurl -plaintext localhost:9090 list

# 如果grpcurl未安装
docker-compose exec app nc -zv localhost 9090
```

#### 5. API保护问题
```bash
# 检查Sentinel配置加载
docker-compose logs app | grep -i sentinel

# 查看限流统计
curl http://localhost:9090/metrics | grep sentinel

# 测试API保护功能
docker-compose exec app sh -c "cd test && go test -v -run TestAPIProtectionWithRealConfig/TestHealthCheckRateLimit"
```

### 性能问题排查

```bash
# 查看容器资源使用
docker stats

# 查看应用内存使用
docker-compose exec app ps aux

# 查看数据库性能
docker-compose exec mysql mysqladmin -u root -prootpass processlist

# 查看Redis内存使用
docker-compose exec redis redis-cli info memory
```

### 网络问题排查

```bash
# 检查容器网络连接
docker network ls
docker network inspect distributed-service_distributed-network

# 测试服务间连接
docker-compose exec app ping mysql
docker-compose exec app ping redis
docker-compose exec app ping consul
```

### 日志聚合查看

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看最近日志
docker-compose logs --tail=50

# 按时间查看日志
docker-compose logs --since="2024-01-01T10:00:00"

# 过滤特定日志
docker-compose logs app | grep -i error
```

### 服务恢复

```bash
# 重启单个服务
docker-compose restart app

# 重新构建并启动
docker-compose up --build -d app

# 完全重新部署
docker-compose down -v --remove-orphans
docker-compose up --build -d

# 清理未使用资源
docker system prune -f
docker volume prune -f
```

## ⚙️ 详细配置

### 环境变量配置

```bash
# 数据库配置
MYSQL_ROOT_PASSWORD=rootpass
MYSQL_DATABASE=distributed_service
MYSQL_USER=testuser
MYSQL_PASSWORD=testpass

# Grafana配置
GF_SECURITY_ADMIN_PASSWORD=admin123

# Consul配置
CONSUL_BIND_INTERFACE=eth0
```

### 数据持久化

```yaml
volumes:
  mysql_data:          # MySQL数据持久化
  redis_data:          # Redis数据持久化
  consul_data:         # Consul配置持久化
  prometheus_data:     # Prometheus指标持久化
  grafana_data:        # Grafana面板持久化
```

### 配置文件挂载

```yaml
volumes:
  - ./config/config-docker.yaml:/app/config/config.yaml:ro
  - ./config/prometheus.yml:/etc/prometheus/prometheus.yml:ro
  - ./scripts/mysql-init.sql:/docker-entrypoint-initdb.d/init.sql:ro
```

## 🚀 生产环境建议

### 安全配置

1. **修改默认密码**
```bash
# 修改数据库密码
MYSQL_ROOT_PASSWORD=your_secure_password
MYSQL_PASSWORD=your_secure_password

# 修改Grafana密码
GF_SECURITY_ADMIN_PASSWORD=your_secure_password

# 修改JWT密钥
JWT_SECRET_KEY=your_jwt_secret_key
```

2. **网络安全**
```yaml
# 仅暴露必要端口
ports:
  - "8080:8080"  # HTTP API
  - "9090:9090"  # gRPC API
  # 其他服务端口仅内网访问
```

3. **资源限制**
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '1'
      memory: 1G
```

### 高可用配置

1. **数据库集群**
```yaml
# 使用MySQL主从或集群
# 配置Redis Sentinel或集群模式
# 设置定期数据备份
```

2. **负载均衡**
```yaml
# 添加nginx或traefik负载均衡
# 配置健康检查
# 设置故障转移
```

3. **监控告警**
```yaml
# 配置Prometheus告警规则
# 设置Grafana告警通知
# 监控关键指标阈值
```

### 备份和恢复

```bash
# 数据库备份
docker-compose exec mysql mysqldump -u root -prootpass distributed_service > backup.sql

# Redis备份
docker-compose exec redis redis-cli BGSAVE

# 配置备份
tar -czf config-backup.tar.gz config/

# 恢复数据
docker-compose exec -T mysql mysql -u root -prootpass distributed_service < backup.sql
```

## 📊 部署验证清单

部署完成后，请按以下清单验证各项功能：

### ✅ 基础服务检查
- [ ] 所有容器正常运行 (`docker-compose ps`)
- [ ] HTTP API响应正常 (`curl http://localhost:8080/health`)
- [ ] gRPC服务响应正常 (`grpcurl -plaintext localhost:9090 list`)
- [ ] 数据库连接正常
- [ ] Redis缓存连接正常

### ✅ 认证功能检查
- [ ] 用户注册功能正常
- [ ] 用户登录获取JWT token
- [ ] JWT认证保护API正常工作

### ✅ API保护功能检查  
- [ ] 限流功能正常 (健康检查接口限流测试)
- [ ] 熔断器功能正常 (错误率熔断测试)
- [ ] API保护测试套件通过 (`cd test && ./run_api_test.sh`)

### ✅ 监控追踪检查
- [ ] Prometheus指标正常收集 (`./scripts/test-metrics.sh`)
- [ ] Grafana面板显示正常  
- [ ] Jaeger追踪数据正常显示 (`./scripts/verify-tracing.sh`)
- [ ] 分布式追踪链路完整

### ✅ 基础设施检查
- [ ] Consul服务注册正常
- [ ] RabbitMQ消息队列正常
- [ ] 各服务健康检查通过

完成以上检查后，您的分布式微服务系统就可以正常提供服务了！

## 📚 相关文档

- [项目总览](../README.md) - 项目介绍和本地开发指南
- [gRPC使用指南](README-gRPC.md) - gRPC服务详细文档
- [分布式追踪](TRACING.md) - 链路追踪使用指南
- [API保护测试](../test/README_API_Protection_Test.md) - 保护机制测试文档