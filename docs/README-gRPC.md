# gRPC 服务集成指南

本项目已成功集成 gRPC 服务，提供高性能的 RPC 通信能力。

## 功能特性

### gRPC 服务器
- **高性能**: 基于 HTTP/2 协议，支持多路复用和流式传输
- **类型安全**: 使用 Protocol Buffers 定义强类型接口
- **中间件支持**: 集成日志记录、错误恢复、指标收集等中间件
- **健康检查**: 内置健康检查服务
- **服务反射**: 开发环境支持服务反射，便于调试
- **优雅关闭**: 支持优雅关闭，确保请求完整处理

### 用户服务 gRPC API
- `GetUser`: 根据 ID 获取用户信息
- `CreateUser`: 创建新用户
- `UpdateUser`: 更新用户信息
- `DeleteUser`: 删除用户
- `ListUsers`: 分页查询用户列表
- `Login`: 用户登录认证
- `Check`: 健康检查

## 项目结构

```
distributed-service/
├── proto/                          # Protocol Buffers 定义文件
│   └── user/
│       └── user.proto              # 用户服务 proto 定义
├── api/proto/                      # 生成的 Go 代码
│   └── user/
│       ├── user.pb.go              # 消息定义
│       └── user_grpc.pb.go         # gRPC 服务定义
├── pkg/grpc/                       # gRPC 服务器包
│   ├── server.go                   # gRPC 服务器实现
│   └── config.go                   # 配置转换工具
├── pkg/middleware/                 # 中间件
│   └── grpc.go                     # gRPC 中间件
├── internal/grpc/                  # gRPC 服务实现
│   └── user_service.go             # 用户服务 gRPC 实现
└── examples/grpc-client/           # 客户端示例
    └── main.go                     # gRPC 客户端示例
```

## 配置说明

在 `config/config.yaml` 中添加了 gRPC 服务器配置：

```yaml
grpc:
  port: 9090                        # gRPC 服务端口
  max_recv_msg_size: 4194304        # 最大接收消息大小 (4MB)
  max_send_msg_size: 4194304        # 最大发送消息大小 (4MB)
  connection_timeout: "5s"          # 连接超时
  max_connection_idle: "15s"        # 最大连接空闲时间
  max_connection_age: "30s"         # 最大连接存活时间
  max_connection_age_grace: "5s"    # 连接优雅关闭时间
  time: "5s"                        # Keep-alive 时间
  timeout: "1s"                     # Keep-alive 超时
  enable_reflection: true           # 启用服务反射 (开发环境)
  enable_health_check: true         # 启用健康检查
```

## 快速开始

### 1. 生成 Protocol Buffers 代码

```bash
# 使用 Makefile
make proto

# 或手动执行
protoc --proto_path=proto \
  --go_out=api/proto --go_opt=paths=source_relative \
  --go-grpc_out=api/proto --go-grpc_opt=paths=source_relative \
  proto/user/user.proto
```

### 2. 启动服务

```bash
# 启动服务 (同时启动 HTTP 和 gRPC 服务器)
make run

# 或
go run main.go
```

服务启动后：
- HTTP 服务器运行在 `:8080`
- gRPC 服务器运行在 `:9090`

### 3. 测试 gRPC 服务

#### 使用示例客户端

```bash
cd examples/grpc-client
go run main.go
```

#### 使用 grpcurl 工具

```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 查看可用服务
grpcurl -plaintext localhost:9090 list

# 查看服务方法
grpcurl -plaintext localhost:9090 list user.v1.UserService

# 健康检查
grpcurl -plaintext -d '{"service": "user.v1.UserService"}' \
  localhost:9090 user.v1.UserService/Check

# 创建用户
grpcurl -plaintext -d '{
  "username": "testuser",
  "email": "test@example.com", 
  "password": "password123"
}' localhost:9090 user.v1.UserService/CreateUser

# 用户登录
grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "password123"
}' localhost:9090 user.v1.UserService/Login
```

## 中间件功能

### 日志记录
- 记录所有 gRPC 请求和响应
- 包含请求方法、持续时间、错误信息等

### 错误恢复
- 自动捕获和恢复 panic
- 记录堆栈信息
- 返回标准 gRPC 错误响应

### 指标收集
- 请求计数器：`grpc_requests_total`
- 请求持续时间：`grpc_request_duration_seconds`
- 按方法和状态码分组

### 健康检查
- 标准 gRPC 健康检查协议
- 支持服务级别的健康状态

## 性能优化

### 连接管理
- Keep-alive 配置优化
- 连接池管理
- 优雅关闭

### 消息大小限制
- 可配置的消息大小限制
- 防止内存溢出

### 并发处理
- 基于 goroutine 的并发处理
- 上下文超时控制

## 安全考虑

### 认证授权
- JWT 令牌验证
- 可扩展的认证中间件

### 传输安全
- 支持 TLS 加密 (生产环境推荐)
- 证书管理

## 监控和观测

### 指标监控
- Prometheus 指标集成
- Grafana 仪表板支持

### 分布式追踪
- OpenTelemetry 集成
- 请求链路追踪

### 日志记录
- 结构化日志
- 上下文信息传递

## 开发工具

### Makefile 命令
```bash
make proto        # 生成 protobuf 代码
make build        # 构建应用
make run          # 运行应用
make test         # 运行测试
make clean        # 清理构建文件
```

### 代码生成
- 自动生成 gRPC 客户端和服务端代码
- 类型安全的消息定义

## 部署说明

### Docker 部署
```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

### Kubernetes 部署
- 支持 Kubernetes 部署
- 服务发现和负载均衡
- 健康检查集成

## 故障排除

### 常见问题

1. **端口冲突**
    - 检查 9090 端口是否被占用
    - 修改配置文件中的端口设置

2. **protobuf 代码生成失败**
    - 确保安装了 protoc 编译器
    - 检查 Go 插件是否正确安装

3. **连接失败**
    - 检查服务器是否正常启动
    - 验证网络连接和防火墙设置

### 调试技巧
- 启用 gRPC 日志记录
- 使用服务反射查看可用方法
- 检查健康检查状态

## 扩展开发

### 添加新服务
1. 在 `proto/` 目录下定义新的 .proto 文件
2. 生成 Go 代码
3. 实现服务接口
4. 注册到 gRPC 服务器

### 自定义中间件
- 实现 `grpc.UnaryServerInterceptor` 接口
- 在服务器配置中添加中间件

## 最佳实践

1. **错误处理**: 使用标准 gRPC 状态码
2. **超时设置**: 为所有请求设置合理的超时
3. **资源管理**: 正确关闭连接和释放资源
4. **版本管理**: 使用语义化版本管理 API
5. **文档维护**: 保持 proto 文件和文档同步

## 参考资料

- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers 指南](https://developers.google.com/protocol-buffers)
- [Go gRPC 教程](https://grpc.io/docs/languages/go/)