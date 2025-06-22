# HTTP + gRPC 集成测试示例

这是一个展示 HTTP 和 gRPC 服务集成的完整示例，包含了 protobuf 定义、gRPC 服务实现以及 HTTP 接口调用 gRPC 服务的演示。

## 🚀 功能特性

### 🔌 gRPC 服务
- **UserService**: 完整的用户 CRUD 操作
- **健康检查**: gRPC 健康检查服务
- **服务反射**: 支持 gRPC 服务发现

### 🌐 HTTP 接口
- **RESTful API**: 标准的 REST 接口设计
- **gRPC 集成**: HTTP 接口内部调用 gRPC 服务
- **错误处理**: 完善的错误处理和状态码映射

### 📋 支持的操作
- ✅ 获取用户列表（支持分页和搜索）
- ✅ 获取单个用户信息
- ✅ 创建新用户
- ✅ 更新用户信息
- ✅ 删除用户
- ✅ 健康检查（HTTP 和 gRPC）

## 📁 项目结构

```
http_grpc_test/
├── proto/                      # Protocol Buffers 定义
│   ├── user.proto             # 用户服务 proto 文件
│   └── user/                  # 生成的 Go 代码
│       ├── user.pb.go         # 消息定义
│       └── user_grpc.pb.go    # gRPC 服务定义
├── service/                   # gRPC 服务实现
│   └── user_service.go        # 用户服务实现
├── client/                    # gRPC 客户端
│   └── grpc_client.go         # gRPC 客户端封装
├── config/                    # 配置文件
│   └── config.yaml           # 服务配置
├── main.go                   # 主程序入口
├── generate.sh               # protobuf 生成脚本
├── test_api.sh              # API 测试脚本
├── test_start.sh            # 服务启动脚本
└── README.md                # 项目说明
```

## 🛠️ 快速开始

### 1. 生成 Protobuf 文件

```bash
# 给脚本执行权限
chmod +x generate.sh

# 生成 protobuf 文件
./generate.sh
```

### 2. 启动服务

```bash
# 启动 HTTP + gRPC 服务
go run main.go

# 或使用启动脚本
chmod +x test_start.sh
./test_start.sh
```

### 3. 测试服务

```bash
# 给测试脚本执行权限
chmod +x test_api.sh

# 运行 API 测试
./test_api.sh
```

## 📡 服务端点

### HTTP 接口 (端口 8080)

#### 基础接口
- `GET /health` - HTTP 健康检查
- `GET /ping` - Ping 测试
- `GET /grpc/health` - gRPC 健康检查 (通过 HTTP 调用)

#### 用户 API
- `GET /api/users` - 列出用户
  - 查询参数: `page`, `page_size`, `search`
- `GET /api/users/:id` - 获取用户详情
- `POST /api/users` - 创建用户
- `PUT /api/users/:id` - 更新用户
- `DELETE /api/users/:id` - 删除用户

### gRPC 接口 (端口 9093)

#### UserService
- `GetUser(GetUserRequest) returns (GetUserResponse)`
- `ListUsers(ListUsersRequest) returns (ListUsersResponse)`
- `CreateUser(CreateUserRequest) returns (CreateUserResponse)`
- `UpdateUser(UpdateUserRequest) returns (UpdateUserResponse)`
- `DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)`
- `HealthCheck(HealthCheckRequest) returns (HealthCheckResponse)`

## 🧪 API 测试示例

### 1. 列出所有用户

```bash
curl -X GET "http://localhost:8080/api/users"
```

### 2. 创建新用户

```bash
curl -X POST "http://localhost:8080/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1-555-0123"
  }'
```

### 3. 获取用户详情

```bash
curl -X GET "http://localhost:8080/api/users/1"
```

### 4. 更新用户

```bash
curl -X PUT "http://localhost:8080/api/users/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith",
    "email": "john.smith@example.com",
    "phone": "+1-555-0124"
  }'
```

### 5. 删除用户

```bash
curl -X DELETE "http://localhost:8080/api/users/1"
```

### 6. 分页查询

```bash
curl -X GET "http://localhost:8080/api/users?page=1&page_size=5"
```

### 7. 搜索用户

```bash
curl -X GET "http://localhost:8080/api/users?search=Alice"
```

### 8. gRPC 健康检查

```bash
curl -X GET "http://localhost:8080/grpc/health"
```

## 🔧 技术特点

### Protocol Buffers
- 使用 proto3 语法
- 定义了完整的用户服务接口
- 支持消息验证和错误处理

### gRPC 服务实现
- 内存存储（生产环境建议使用数据库）
- 完整的 CRUD 操作
- 业务逻辑验证（如邮箱唯一性）
- 并发安全（使用读写锁）

### HTTP 到 gRPC 集成
- HTTP 接口作为 gRPC 服务的网关
- 自动错误码转换
- 超时控制
- 请求参数验证

### 错误处理
- gRPC 状态码到 HTTP 状态码的映射
- 详细的错误信息返回
- 统一的错误响应格式

## 📝 响应格式

### 成功响应
```json
{
  "user": {
    "id": "1",
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "phone": "+1-555-0101",
    "created_at": 1703947200,
    "updated_at": 1703947200
  },
  "message": "User retrieved successfully",
  "source": "gRPC UserService"
}
```

### 错误响应
```json
{
  "error": "Failed to get user",
  "message": "user not found",
  "code": "NotFound"
}
```

## 🔍 日志监控

服务运行时会输出详细的日志信息，包括：
- HTTP 请求日志
- gRPC 调用日志
- 业务操作日志
- 错误日志

通过日志可以观察到 HTTP 接口是如何调用后端 gRPC 服务的。

## 🎯 学习目标

通过这个示例，你可以学习到：

1. **Protobuf 定义**: 如何设计 gRPC 服务接口
2. **gRPC 服务实现**: 如何实现 gRPC 服务端
3. **HTTP 网关模式**: 如何通过 HTTP 接口调用 gRPC 服务
4. **错误处理**: gRPC 和 HTTP 之间的错误映射
5. **并发控制**: 多线程环境下的数据安全
6. **接口设计**: RESTful API 设计最佳实践

## 📚 相关文档

- [Protocol Buffers Documentation](https://developers.google.com/protocol-buffers)
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Gin Web Framework](https://gin-gonic.com/)

## 🚨 注意事项

1. 当前使用内存存储，重启服务会丢失数据
2. 生产环境建议使用数据库替代内存存储
3. gRPC 客户端连接需要在服务启动后建立
4. 确保相关依赖已正确安装 (protoc, protoc-gen-go, protoc-gen-go-grpc) 