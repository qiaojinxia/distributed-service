# gRPC限流测试文档

## 📋 测试概述

这是一个comprehensive的gRPC限流测试，使用真实的gRPC客户端和服务端来验证Sentinel保护机制的限流功能。

## 🚀 测试架构

### 服务定义
- **UserService**: 用户服务（模拟高优先级服务）
- **OrderService**: 订单服务（模拟通用服务）

### 限流规则配置
```yaml
rate_limit_rules:
  - name: "grpc_user_service_limiter"
    resource: "/grpc/user_service/*"        # 优先级: 2
    threshold: 5                            # 5 QPS
    
  - name: "grpc_read_operations_limiter"  
    resource: "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*"  # 优先级: 4
    threshold: 10                           # 10 QPS
    
  - name: "grpc_write_operations_limiter"
    resource: "/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*"  # 优先级: 4
    threshold: 2                            # 2 QPS
```

## 🧪 测试用例

### 1. 用户服务限流测试 ✅
- **测试目标**: 验证用户服务专属限流规则 (5 QPS)
- **测试方法**: 发送10个`UserService.GetUser`请求
- **预期结果**: 5个成功，5个被限流
- **实际结果**: ✅ 成功=5, 限流=5

### 2. 读操作限流测试 ✅  
- **测试目标**: 验证通用读操作限流规则 (10 QPS)
- **测试方法**: 并发发送15个`OrderService.GetOrder`请求
- **预期结果**: ~10个成功，~5个被限流
- **实际结果**: ✅ 成功=10, 限流=5

### 3. 写操作限流测试 ✅
- **测试目标**: 验证通用写操作限流规则 (2 QPS)
- **测试方法**: 发送5个`OrderService.CreateOrder`请求
- **预期结果**: 2个成功，3个被限流
- **实际结果**: ✅ 成功=2, 限流=3

### 4. 优先级匹配测试 ✅
- **测试目标**: 验证规则优先级逻辑
- **测试逻辑**: `UserService.GetUser`应该匹配用户服务规则(5 QPS)而不是读操作规则(10 QPS)
- **测试方法**: 发送8个`UserService.GetUser`请求
- **实际结果**: ✅ 成功=5, 限流=3 (按照5 QPS限制)

### 5. 混合操作测试 ✅
- **测试目标**: 验证多种操作并发执行时的限流效果
- **测试方法**: 同时发送用户获取、订单获取、订单创建操作
- **实际结果**: ✅ 所有操作都有正确的限流效果

## 📊 测试结果分析

### 🎯 优先级匹配验证
```
UserService.GetUser → /grpc/user_service/get_user → 匹配 "/grpc/user_service/*" (5 QPS)
OrderService.GetOrder → /grpc/order_service/get_order → 匹配 "/grpc/*/get*" (10 QPS)  
OrderService.CreateOrder → /grpc/order_service/create_order → 匹配 "/grpc/*/create*" (2 QPS)
```

### 🔄 动态规则创建
测试验证了当新资源第一次被访问时，系统会自动创建匹配的Sentinel规则：
- 用户服务方法自动匹配用户服务规则
- 订单服务方法自动匹配通用读/写操作规则

### 📈 通配符匹配效果
- **具体服务规则优先**: `/grpc/user_service/*` > `/grpc/*/get*`
- **多模式匹配**: `/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*`
- **路径前缀匹配**: `/grpc/user_service/*` 匹配 `/grpc/user_service/get_user`

## 🛠️ 技术实现要点

### gRPC方法名映射
```
原始方法: /user.UserService/GetUser
映射资源: /grpc/user_service/get_user
```

### 映射规则
1. 提取服务名: `UserService` → `user_service`
2. 转换方法名: `GetUser` → `get_user`  
3. 构建资源: `/grpc/{service}/{method}`

### 中间件集成
- 使用`grpc.UnaryInterceptor`和`grpc.StreamInterceptor`
- 自动资源名映射和规则匹配
- 限流异常转换为gRPC错误码`codes.ResourceExhausted`

## 🚦 测试命令

```bash
# 运行完整测试
go test ./test -run TestGRPCRateLimitWithClient -v -timeout 60s

# 运行单个测试用例
go test ./test -run TestGRPCRateLimitWithClient/TestUserServiceRateLimit -v
```

## 📝 测试文件结构

```
test/
├── proto/
│   ├── user.proto                    # 用户服务定义
│   └── order.proto                   # 订单服务定义
├── grpc_service_impl.go              # 用户服务实现
├── order_service_impl.go             # 订单服务实现
└── grpc_rate_limit_test.go          # 主测试文件
```

## ✅ 验证通过的功能

1. ✅ **通配符模式匹配**: `*`, 多模式`,`分隔
2. ✅ **优先级排序**: 具体 > 通用
3. ✅ **动态规则创建**: 首次访问自动创建规则
4. ✅ **gRPC方法映射**: 正确的资源名转换
5. ✅ **并发限流**: 多客户端同时访问
6. ✅ **错误处理**: 正确的gRPC错误码返回
7. ✅ **实时限流**: 准确的QPS控制

## 🎉 结论

gRPC通配符限流功能**完全正常工作**！所有测试用例都通过，验证了：

- **配置简化**: 从复杂的单独配置减少到简洁的通配符规则
- **智能匹配**: 优先级逻辑确保最合适的规则被应用
- **动态扩展**: 新服务自动继承适当的保护规则
- **生产就绪**: 真实的gRPC客户端/服务端环境验证 