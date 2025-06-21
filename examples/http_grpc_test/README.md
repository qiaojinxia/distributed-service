# HTTP + gRPC 集成测试示例

这个示例演示如何使用分布式服务框架同时启动HTTP和gRPC服务。

## 功能特性

- ✅ HTTP REST API 服务 (端口: 8080)
- ✅ gRPC 服务 (端口: 9000) 
- ✅ 健康检查端点
- ✅ 服务发现支持
- ✅ 中间件集成
- ✅ 生命周期管理

## 快速开始

### 1. 启动服务

```bash
cd examples/http_grpc_test
go run main.go
```

### 2. 测试HTTP服务

#### 健康检查
```bash
curl http://localhost:8080/health
```

#### API版本信息
```bash
curl http://localhost:8080/api/version
```

#### 用户管理
```bash
# 获取用户列表
curl http://localhost:8080/api/users

# 获取特定用户
curl http://localhost:8080/api/users/123

# 创建用户
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'
```

#### 订单管理
```bash
# 获取订单列表
curl http://localhost:8080/api/orders

# 获取特定订单
curl http://localhost:8080/api/orders/123
```

#### gRPC测试端点
```bash
curl http://localhost:8080/api/test/grpc
```

### 3. 测试gRPC服务

#### 使用grpcurl测试健康检查
```bash
# 安装grpcurl (如果尚未安装)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出可用服务
grpcurl -plaintext localhost:9000 list

# 健康检查
grpcurl -plaintext localhost:9000 grpc.health.v1.Health/Check
```

#### 使用evans CLI测试
```bash
# 安装evans (如果尚未安装)
go install github.com/ktr0731/evans@latest

# 连接到gRPC服务
evans --host localhost --port 9000 -r repl
```

## API端点

### HTTP REST API

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/version` | 版本信息 |
| GET | `/api/users` | 获取用户列表 |
| GET | `/api/users/:id` | 获取特定用户 |
| POST | `/api/users` | 创建用户 |
| GET | `/api/orders` | 获取订单列表 |
| GET | `/api/orders/:id` | 获取特定订单 |
| GET | `/api/test/grpc` | gRPC服务状态 |

### gRPC服务

| 服务 | 描述 |
|------|------|
| `grpc.health.v1.Health` | 健康检查服务 |
| `grpc.reflection.v1alpha.ServerReflection` | 服务反射 |
| `UserService` | 用户服务 (示例) |
| `OrderService` | 订单服务 (示例) |

## 日志输出示例

```
🚀 启动HTTP + gRPC集成测试服务...
🔧 初始化服务依赖...
🌐 注册HTTP路由:
  ✅ GET /health
  ✅ GET /api/version
  ✅ GET /api/users
  ✅ GET /api/users/:id
  ✅ POST /api/users
  ✅ GET /api/orders
  ✅ GET /api/orders/:id
  ✅ GET /api/test/grpc
🔌 注册gRPC服务:
  ✅ UserService 已注册
  ✅ OrderService 已注册
  ✅ HealthService 已自动注册
✅ 服务启动完成!
🌐 HTTP服务监听: http://localhost:8080
🔌 gRPC服务监听: localhost:9000
```

## 配置说明

- **HTTP端口**: 8080
- **gRPC端口**: 9000 (框架默认)
- **运行模式**: debug (开发模式)
- **日志级别**: info
- **健康检查**: 启用
- **服务反射**: 启用 (gRPC)

## 扩展功能

这个示例可以扩展以下功能：

1. **添加实际的protobuf定义**
2. **集成数据库连接**
3. **添加认证中间件**
4. **集成服务发现**
5. **添加监控指标**
6. **集成分布式追踪**

## 注意事项

- 确保端口8080和9000未被占用
- 在生产环境中，建议使用配置文件而非硬编码
- gRPC服务示例中的UserService和OrderService需要实际的protobuf定义 