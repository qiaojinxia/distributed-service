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
│  ├── Grafana (可视化面板)                                   │
│  └── Jaeger (分布式链路追踪)                                │
├─────────────────────────────────────────────────────────────┤
│  🔍 观测性 (Observability)                                 │
│  ├── OpenTelemetry (统一遥测数据)                          │
│  ├── 结构化日志 (Zap)                                       │
│  ├── 指标监控 (Prometheus)                                 │
│  └── 链路追踪 (Jaeger)                                      │
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
| 链路追踪 | http://localhost:16686 | Jaeger UI |
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

## 🔍 分布式链路追踪测试

### 自动化测试

运行完整的追踪功能测试：
```bash
# 完整追踪测试
./scripts/test-tracing.sh

# 快速验证
./scripts/verify-tracing.sh
```

### 手动测试

#### 1. JWT 认证追踪测试
```bash
# 用户注册（生成完整追踪链路）
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: trace-register-$(date +%s)" \
  -d '{
    "username": "traceuser",
    "email": "trace@example.com",
    "password": "password123"
  }'

# 用户登录（生成认证追踪）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: trace-login-$(date +%s)" \
  -d '{
    "username": "traceuser",
    "password": "password123"
  }'
```

#### 2. 获取当前用户信息（需要JWT Token）
```bash
# 先登录获取Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"traceuser","password":"password123"}' | \
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 获取用户信息
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: trace-userinfo-$(date +%s)"
```

#### 3. 修改密码追踪
```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: trace-changepass-$(date +%s)" \
  -d '{
    "old_password": "password123",
    "new_password": "newpassword123"
  }'
```

### 查看追踪数据

1. **访问 Jaeger UI**: http://localhost:16686
2. **选择服务**: 在 Service 下拉框选择 `distributed-service`
3. **设置时间范围**: 选择最近 15 分钟
4. **查找追踪**: 点击 "Find Traces" 按钮
5. **分析链路**: 点击具体的 trace 查看详细调用链

### 追踪验证要点

在 Jaeger UI 中应该能看到以下追踪层次：

```
HTTP Request Span
├── userService.Register/Login/ChangePassword
│   ├── userRepository.GetByUsername
│   ├── userRepository.Create/Update
│   └── 业务逻辑处理
├── 数据库操作追踪
│   ├── SQL 查询时间
│   ├── 影响行数
│   └── 错误信息（如有）
└── 中间件追踪
    ├── JWT 认证处理
    ├── 请求ID传播
    └── 响应时间统计
```

**关键指标检查**：
- ✅ 每个 span 都有正确的名称和操作类型
- ✅ HTTP span 包含方法、路径、状态码、响应时间
- ✅ Service span 包含用户名、邮箱等业务属性  
- ✅ Repository span 包含数据库操作类型和表名
- ✅ 错误 span 包含异常信息和错误堆栈
- ✅ 整个调用链路完整且时间合理
- ✅ 请求ID在整个链路中正确传播

## 🔧 配置说明

### 环境配置文件

- `config/config.yaml` - 本地开发配置
- `config/config-docker.yaml` - Docker 环境配置
- `config/config-local.yaml` - 本地测试配置（无外部依赖）
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

### 分布式链路追踪配置

Jaeger 配置：
- UI 端口: `16686`
- OTLP HTTP 端口: `4318` (应用数据上报)
- OTLP gRPC 端口: `4317`
- 数据收集端口: `14268`

OpenTelemetry 配置（`config-docker.yaml`）：
```yaml
tracing:
  service_name: "distributed-service"
  service_version: "1.0.0"
  environment: "docker"
  enabled: true
  exporter_type: "otlp"  # 发送到 Jaeger
  endpoint: "http://jaeger:4318/v1/traces"
  sample_ratio: 0.1  # 10% 采样率
```

本地开发配置（`config-local.yaml`）：
```yaml
tracing:
  service_name: "distributed-service"
  environment: "local"
  enabled: true
  exporter_type: "stdout"  # 输出到控制台
  sample_ratio: 1.0  # 100% 采样率
```

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
│   ├── config-local.yaml       # 本地测试配置
│   ├── redis.conf              # Redis 配置
│   └── prometheus.yml          # Prometheus 配置
├── scripts/                     # 初始化脚本
│   ├── mysql-init.sql          # MySQL 初始化脚本
│   ├── test-tracing.sh         # 分布式追踪测试脚本
│   └── verify-tracing.sh       # 追踪功能快速验证脚本
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
│   ├── registry/               # 服务注册
│   └── tracing/                # 分布式链路追踪
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

### 查看 Jaeger 日志
```bash
docker-compose logs -f jaeger
```

### 重启服务
```bash
docker-compose restart app
```

### 重启追踪服务
```bash
docker-compose restart jaeger
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

### 追踪功能测试
```bash
# 运行完整追踪测试
./scripts/test-tracing.sh

# 快速验证追踪功能
./scripts/verify-tracing.sh
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

### 链路追踪问题

1. **Jaeger UI 无法访问**
   - 检查 Jaeger 容器状态：`docker-compose ps jaeger`
   - 查看 Jaeger 日志：`docker-compose logs jaeger`
   - 确认端口 16686 没有被占用

2. **没有追踪数据**
   - 检查应用追踪配置是否启用（`tracing.enabled: true`）
   - 确认导出器类型设置正确（Docker环境使用 `otlp`）
   - 检查应用到 Jaeger 的网络连接
   - 验证采样率设置（开发环境建议设为 1.0）

3. **追踪数据不完整**
   - 检查中间件是否正确加载
   - 确认所有服务层都正确实现了追踪
   - 查看应用日志中的追踪错误信息

4. **性能问题**
   - 调整采样率（生产环境建议 0.1 或更低）
   - 检查 Jaeger 存储配置
   - 监控追踪数据量和存储空间

### 追踪验证步骤

```bash
# 1. 检查 Jaeger 服务状态
curl -f http://localhost:16686/api/services

# 2. 检查应用健康状态
curl -f http://localhost:8080/health

# 3. 生成测试追踪数据
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: debug-test-$(date +%s)" \
  -d '{"username":"debuguser","email":"debug@test.com","password":"test123"}'

# 4. 在 Jaeger UI 中查找 debug-test-* 请求ID
```

## 📈 监控和指标

### Prometheus 指标
- HTTP 请求数量和响应时间
- 数据库查询性能
- 缓存命中率
- 系统资源使用情况

### Grafana 面板
默认账号：`admin` / `admin123`

可以导入预定义的面板来监控服务性能。

### Jaeger 链路追踪
- **请求链路可视化**: 查看完整的请求调用链
- **性能分析**: 识别性能瓶颈和延迟热点
- **错误追踪**: 快速定位错误发生的具体位置
- **依赖关系**: 了解服务间的依赖和调用关系
- **采样控制**: 根据需要调整数据收集粒度

#### 关键追踪指标
- **延迟 (Latency)**: 各层服务的响应时间
- **错误率 (Error Rate)**: 各操作的失败比例  
- **吞吐量 (Throughput)**: 每秒处理的请求数
- **调用深度 (Call Depth)**: 服务调用的层次结构
- **并发度 (Concurrency)**: 同时处理的请求数量

#### 追踪数据分析
访问 Jaeger UI (http://localhost:16686) 可以：
1. 按服务、操作、标签筛选追踪数据
2. 查看请求的完整时间线
3. 分析错误和异常的根本原因
4. 比较不同时间段的性能表现
5. 导出追踪数据进行离线分析

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

### 添加分布式追踪

在新的服务方法中添加追踪：

```go
// Service 层示例
func (s *userService) CreateUser(ctx context.Context, user *model.User) error {
    return tracing.WithSpan(ctx, "userService.CreateUser", func(ctx context.Context) error {
        // 添加业务属性
        tracing.AddSpanAttributes(ctx, map[string]interface{}{
            "user.username": user.Username,
            "user.email":    user.Email,
        })
        
        // 调用 Repository 层
        return s.userRepo.Create(ctx, user)
    })
}

// Repository 层示例
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    return tracing.TraceDatabase(ctx, "userRepository.Create", "users", "create", func() error {
        return r.db.Create(user).Error
    })
}

// API 层使用请求ID
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 中间件已自动添加追踪，只需添加业务属性
    ctx := c.MustGet("ctx").(context.Context)
    tracing.AddSpanAttributes(ctx, map[string]interface{}{
        "api.endpoint": "/api/v1/users",
        "api.method":   "POST",
    })
    
    // 业务逻辑...
}
```

### 追踪最佳实践

1. **合理命名 Span**: 使用 `service.method` 格式
2. **添加有意义的属性**: 用户ID、操作类型、资源名称等
3. **记录错误信息**: 使用 `tracing.RecordError(ctx, err)`
4. **控制采样率**: 生产环境避免100%采样
5. **避免敏感信息**: 不要在 span 中记录密码等敏感数据

### 修改配置

1. 更新 `config/config-docker.yaml`
2. 重启应用：`docker-compose restart app`

## 🆘 支持

如有问题，请查看：
1. 应用日志
2. 各服务的健康检查状态
3. Prometheus 监控指标
4. Jaeger 链路追踪数据
5. 本文档的故障排查部分

### 相关文档

- **[分布式链路追踪详细文档](docs/TRACING.md)** - 追踪功能的完整说明
- **[项目总体介绍](README.md)** - 项目概览和快速开始
- **[API 文档](http://localhost:8080/swagger/index.html)** - 在线 API 文档

### 快速链接

- 🔍 **追踪 UI**: http://localhost:16686
- 📊 **监控面板**: http://localhost:3000  
- 📈 **指标收集**: http://localhost:9091
- 🗂️ **服务发现**: http://localhost:8500
- 📚 **API 文档**: http://localhost:8080/swagger/index.html 