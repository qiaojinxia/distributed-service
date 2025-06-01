# 分布式服务保护测试套件总览

## 📋 测试套件概述

本项目包含comprehensive的保护机制测试，完全基于`config.yaml`的真实配置，验证Sentinel保护功能在HTTP API和gRPC服务中的正确性。

## 🎯 测试覆盖范围

### 1. HTTP API保护测试 🌐
- **文件**: `test/api_protection_test.go`
- **脚本**: `test/run_api_test.sh`
- **文档**: `test/README_API_Protection_Test.md`

#### 测试功能
- ✅ **API限流**: 健康检查(2 QPS)、认证(10 req/min)、用户(30 req/min)、通用API(100 req/min)
- ✅ **API熔断**: 认证(50%错误率)、用户(60%错误率)、通用(80%错误率)
- ✅ **通配符匹配**: `/api/v1/auth/*`, `/api/v1/users/*`, `/api/*`
- ✅ **优先级排序**: 精确匹配 > 具体通配符 > 通用通配符
- ✅ **并发处理**: 多客户端同时访问的安全性验证

### 2. gRPC服务保护测试 📡
- **文件**: `test/grpc_rate_limit_test.go`
- **文档**: `test/README_GRPC_Rate_Limit_Test.md`

#### 测试功能
- ✅ **gRPC限流**: 用户服务(5 QPS)、读操作(10 QPS)、写操作(2 QPS)
- ✅ **方法映射**: `/user.UserService/GetUser` → `/grpc/user_service/get_user`
- ✅ **多模式匹配**: `/grpc/*/get*,/grpc/*/list*,/grpc/*/find*`
- ✅ **实际客户端**: 真实gRPC客户端/服务器环境测试
- ✅ **并发gRPC**: 多客户端并发调用验证

## 🚀 快速开始

### 演示测试（推荐）
```bash
# 快速演示主要功能
chmod +x test/demo_api_test.sh
./test/demo_api_test.sh
```

### 完整API测试
```bash
# 运行所有HTTP API保护测试
chmod +x test/run_api_test.sh
./test/run_api_test.sh all

# 运行完整集成测试
./test/run_api_test.sh comprehensive
```

### gRPC测试
```bash
# 运行gRPC保护测试
go test ./test -run TestGRPCRateLimitWithClient -v -timeout 60s
```

### 单独测试
```bash
# 健康检查限流
go test ./test -run TestAPIProtectionWithRealConfig/TestHealthCheckRateLimit -v

# gRPC用户服务限流
go test ./test -run TestGRPCRateLimitWithClient/TestUserServiceRateLimit -v
```

## 📊 配置映射表

### HTTP API配置 (基于config.yaml)

| 资源模式 | 类型 | 阈值 | 窗口 | 优先级 | 测试用例 |
|----------|------|------|------|--------|----------|
| `/health` | 限流 | 2 | 1s | 1 | 健康检查 |
| `/api/v1/auth/*` | 限流 | 10 | 60s | 2 | 认证接口 |
| `/api/v1/users/*` | 限流 | 30 | 60s | 2 | 用户接口 |
| `/api/*` | 限流 | 100 | 60s | 3 | 通用API |
| `/protection/*` | 限流 | 20 | 1s | 3 | 保护状态 |
| `/api/v1/auth/*` | 熔断 | 50% | - | 2 | 认证熔断 |
| `/api/v1/users/*` | 熔断 | 60% | - | 2 | 用户熔断 |
| `/api/*` | 熔断 | 80% | - | 3 | 通用熔断 |

### gRPC服务配置

| 资源模式 | 类型 | 阈值 | 窗口 | 优先级 | 测试用例 |
|----------|------|------|------|--------|----------|
| `/grpc/user_service/*` | 限流 | 5 | 1s | 2 | 用户服务 |
| `/grpc/*/get*,/grpc/*/list*,/grpc/*/find*` | 限流 | 10 | 1s | 4 | 读操作 |
| `/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*` | 限流 | 2 | 1s | 4 | 写操作 |

## 🧪 测试架构

### HTTP API测试架构
```
HTTP Client → Gin Router → Sentinel Middleware → Mock Handler
                ↓
           [ 限流/熔断检查 ]
                ↓
        [ 成功/限流/熔断响应 ]
```

### gRPC测试架构
```
gRPC Client → gRPC Server → Sentinel Interceptor → Service Implementation
                ↓
        [ gRPC方法名映射 ]
                ↓
           [ 限流/熔断检查 ]
                ↓
     [ 成功/ResourceExhausted响应 ]
```

## 📈 测试结果示例

### HTTP API限流测试结果
```
🧪 开始健康检查限流测试: /health (阈值=2, 窗口=1s)
✅ 请求 1/5 成功
✅ 请求 2/5 成功
🚫 请求 3/5 被限流
🚫 请求 4/5 被限流
🚫 请求 5/5 被限流
📊 健康检查测试结果: 成功=2, 限流=3, 耗时=69ms
```

### gRPC限流测试结果
```
📊 用户服务限流测试结果: 成功=5, 限流=5
📊 读操作限流测试结果: 成功=10, 限流=5
📊 写操作限流测试结果: 成功=2, 限流=3
```

## 🔧 技术特点

### 通配符匹配系统
- **单一通配符**: `/api/v1/auth/*`
- **多模式匹配**: `/grpc/*/get*,/grpc/*/list*`
- **优先级排序**: 精确匹配 > 具体通配符 > 通用通配符
- **动态规则创建**: 首次访问时自动创建匹配规则

### 资源名映射
- **HTTP**: 直接使用请求路径作为资源名
- **gRPC**: `/user.UserService/GetUser` → `/grpc/user_service/get_user`
- **统一前缀**: HTTP使用路径，gRPC使用`/grpc/`前缀

### 测试环境
- **真实服务器**: httptest.Server和真实gRPC服务器
- **真实客户端**: http.Client和gRPC客户端
- **并发安全**: 多goroutine并发测试
- **错误模拟**: 可控的错误率和延迟模拟

## 📚 文档结构

```
test/
├── README_Test_Suite_Overview.md          # 📖 测试套件总览 (本文档)
├── README_API_Protection_Test.md          # 🌐 HTTP API测试详细文档
├── README_GRPC_Rate_Limit_Test.md         # 📡 gRPC测试详细文档
├── api_protection_test.go                 # 🧪 HTTP API测试实现
├── grpc_rate_limit_test.go               # 🧪 gRPC测试实现
├── run_api_test.sh                       # 🚀 API测试执行脚本
├── demo_api_test.sh                      # 🎬 快速演示脚本
├── grpc_service_impl.go                  # 🔧 gRPC服务实现
├── order_service_impl.go                 # 🔧 订单服务实现
└── proto/                                # 📋 Protocol Buffers定义
    ├── user.proto
    └── order.proto
```

## ✅ 验证清单

使用以下清单验证所有功能正常工作：

### HTTP API保护
- [ ] 健康检查限流 (2 QPS)
- [ ] 认证接口限流 (10 req/min)
- [ ] 用户接口限流 (30 req/min)
- [ ] 通用API限流 (100 req/min)
- [ ] 认证接口熔断 (50%错误率)
- [ ] 用户接口熔断 (60%错误率)
- [ ] 通用API熔断 (80%错误率)
- [ ] 优先级匹配正确
- [ ] 并发请求安全

### gRPC服务保护
- [ ] 用户服务限流 (5 QPS)
- [ ] 读操作限流 (10 QPS)
- [ ] 写操作限流 (2 QPS)
- [ ] 方法名映射正确
- [ ] 多模式匹配正确
- [ ] 优先级排序正确

### 通用功能
- [ ] 通配符模式匹配
- [ ] 动态规则创建
- [ ] 规则优先级排序
- [ ] 并发安全
- [ ] 错误处理正确

## 🎉 总结

这套comprehensive的测试系统完整验证了分布式服务保护机制的所有关键功能：

1. **配置驱动**: 完全基于真实config.yaml配置
2. **功能全面**: 覆盖限流、熔断、通配符、优先级等所有功能
3. **环境真实**: 使用真实的HTTP/gRPC客户端和服务器
4. **自动化**: 完整的测试自动化脚本和CI/CD集成
5. **文档完善**: 详细的使用文档和故障排除指南

通过这套测试，可以确信保护机制在生产环境中能够正确工作，为分布式系统提供可靠的流量控制和故障保护。 