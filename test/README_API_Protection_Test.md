# HTTP API保护测试文档

## 📋 测试概述

这是一个comprehensive的HTTP API保护测试，基于`config.yaml`中的真实配置，验证Sentinel保护机制的限流和熔断功能。

## 🎯 测试目标

验证以下功能的正确性：
- ✅ **HTTP API限流**: 根据配置的QPS/阈值进行流量控制
- ✅ **HTTP API熔断**: 根据错误率自动熔断保护
- ✅ **通配符匹配**: 支持路径通配符模式匹配
- ✅ **优先级排序**: 精确匹配优先于通配符匹配
- ✅ **并发安全**: 多客户端同时访问的正确处理
- ✅ **动态规则**: 运行时动态创建和匹配规则

## 🚀 配置映射

### 限流规则 (Rate Limiting Rules)

| 规则名称 | 资源模式 | 阈值 | 窗口 | 描述 |
|---------|---------|------|------|------|
| `health_check_limiter` | `/health` | 2 | 1s | 健康检查接口 - 2 QPS |
| `auth_api_limiter` | `/api/v1/auth/*` | 10 | 60s | 认证接口 - 每分钟10次 |
| `users_api_limiter` | `/api/v1/users/*` | 30 | 60s | 用户接口 - 每分钟30次 |
| `api_general_limiter` | `/api/*` | 100 | 60s | 通用API - 每分钟100次 |
| `protection_status_limiter` | `/protection/*` | 20 | 1s | 保护状态接口 - 20 QPS |

### 熔断器规则 (Circuit Breaker Rules)

| 规则名称 | 资源模式 | 策略 | 错误阈值 | 最小请求数 | 描述 |
|---------|---------|------|----------|------------|------|
| `auth_api_circuit` | `/api/v1/auth/*` | ErrorRatio | 50% | 10 | 认证接口熔断器 |
| `users_api_circuit` | `/api/v1/users/*` | ErrorRatio | 60% | 8 | 用户接口熔断器 |
| `api_general_circuit` | `/api/*` | ErrorRatio | 80% | 20 | 通用API熔断器 |

## 🧪 测试用例详解

### 1. 限流测试 (Rate Limit Tests)

#### 1.1 健康检查限流测试
- **目标**: `/health` 接口
- **配置**: 2 QPS
- **测试方法**: 快速发送5个请求
- **预期结果**: 2个成功，3个被限流

#### 1.2 认证API限流测试
- **目标**: `/api/v1/auth/login` 接口
- **配置**: 10 requests/60s
- **测试方法**: 发送15个请求
- **预期结果**: 在60秒窗口内限制超出请求

#### 1.3 用户API限流测试
- **目标**: `/api/v1/users/profile` 接口
- **配置**: 30 requests/60s
- **测试方法**: 发送40个请求
- **预期结果**: 30个成功，10个被限流

#### 1.4 保护状态限流测试
- **目标**: `/protection/status` 接口
- **配置**: 20 QPS
- **测试方法**: 快速发送30个请求
- **预期结果**: 约20个成功，约10个被限流

#### 1.5 通用API限流测试
- **目标**: `/api/v1/products/list` 接口
- **配置**: 100 requests/60s
- **测试方法**: 发送120个请求
- **预期结果**: 100个成功，20个被限流

### 2. 熔断器测试 (Circuit Breaker Tests)

#### 2.1 认证API熔断测试
- **目标**: `/api/v1/auth/register` 接口
- **配置**: 50%错误率触发，最小10个请求
- **测试方法**: 发送高错误率请求
- **预期结果**: 达到阈值后熔断开启

#### 2.2 用户API熔断测试
- **目标**: `/api/v1/users/delete` 接口
- **配置**: 60%错误率触发，最小8个请求
- **测试方法**: 模拟服务器错误
- **预期结果**: 错误率超过60%时熔断

#### 2.3 通用API熔断测试
- **目标**: `/api/v1/orders/create` 接口
- **配置**: 80%错误率触发，最小20个请求
- **测试方法**: 高错误率模拟
- **预期结果**: 错误率超过80%时熔断

### 3. 功能测试 (Functional Tests)

#### 3.1 优先级匹配测试
- **测试目标**: 验证规则优先级排序逻辑
- **测试案例**:
  - `/api/v1/auth/login` → 匹配认证规则（优先级更高）
  - `/api/v1/users/profile` → 匹配用户规则（优先级更高）
  - `/api/v1/orders/list` → 匹配通用API规则（兜底规则）

#### 3.2 并发请求测试
- **测试目标**: 验证多客户端并发访问的正确处理
- **测试方法**: 20个goroutine同时发送请求
- **验证点**: 限流统计准确，无竞态条件

#### 3.3 通配符匹配测试
- **测试目标**: 验证路径通配符模式匹配
- **测试案例**:
  - `/api/v1/auth/login` → 匹配 `/api/v1/auth/*`
  - `/api/v1/users/123` → 匹配 `/api/v1/users/*`
  - `/protection/rules` → 匹配 `/protection/*`

## 🛠️ 技术实现要点

### Mock Handler设计
```go
type MockHandler struct {
    delayMs    int      // 模拟处理延迟
    errorRate  float64  // 错误率 (0.0-1.0)
    callCount  int64    // 调用计数
    shouldFail bool     // 强制失败模式
}
```

### 测试服务器架构
- **Gin Router**: 轻量级HTTP框架
- **Sentinel中间件**: 自动应用限流和熔断规则
- **Mock Handlers**: 模拟不同的服务响应模式
- **TestServer**: httptest.Server提供真实HTTP环境

### 资源名映射规则
- **精确匹配**: `/health` → `/health`
- **通配符匹配**: `/api/v1/auth/login` → 匹配 `/api/v1/auth/*`
- **优先级**: 精确 > 具体通配符 > 通用通配符

## 🚦 运行测试

### 快速开始
```bash
# 运行所有测试
chmod +x test/run_api_test.sh
./test/run_api_test.sh all

# 运行完整集成测试
./test/run_api_test.sh comprehensive

# 生成测试报告
./test/run_api_test.sh report
```

### 单独运行测试
```bash
# 运行特定测试用例
go test ./test -run TestAPIProtectionWithRealConfig/TestHealthCheckRateLimit -v

# 运行所有API保护测试
go test ./test -run TestAPIProtectionWithRealConfig -v -timeout 300s

# 运行并发测试
go test ./test -run TestAPIProtectionWithRealConfig/TestConcurrentRequests -v
```

### 测试参数调优
```bash
# 增加超时时间（长窗口限流测试）
go test ./test -run TestAPIProtectionWithRealConfig -v -timeout 600s

# 启用详细日志
go test ./test -run TestAPIProtectionWithRealConfig -v -args -debug

# 并行运行（注意资源共享）
go test ./test -run TestAPIProtectionWithRealConfig -v -parallel 4
```

## 📊 预期结果

### 成功指标
- ✅ 所有限流测试通过：正确控制请求速率
- ✅ 所有熔断测试通过：错误率达到阈值时正确熔断
- ✅ 优先级匹配正确：规则按预期优先级应用
- ✅ 并发安全：无竞态条件，统计准确
- ✅ 通配符匹配：路径模式正确匹配

### 输出示例
```
🧪 开始健康检查限流测试: /health (阈值=2, 窗口=1s)
✅ 请求 1/5 成功
✅ 请求 2/5 成功
🚫 请求 3/5 被限流
🚫 请求 4/5 被限流
🚫 请求 5/5 被限流
📊 健康检查测试结果: 成功=2, 限流=3, 耗时=52ms
```

## 🔧 故障排除

### 常见问题

#### 1. 限流未生效
- **现象**: 请求数超过阈值但未被限流
- **原因**: 时间窗口配置问题或规则未正确加载
- **解决**: 检查`StatIntervalMs`配置，确认规则已加载

#### 2. 熔断器未触发
- **现象**: 错误率达到阈值但未熔断
- **原因**: 请求数未达到`MinRequestAmount`最小值
- **解决**: 确保发送足够的请求数

#### 3. 通配符匹配失败
- **现象**: 请求未匹配到预期规则
- **原因**: 资源名映射错误或优先级问题
- **解决**: 检查中间件资源名提取逻辑

#### 4. 并发测试不稳定
- **现象**: 并发测试结果不一致
- **原因**: 竞态条件或时间精度问题
- **解决**: 增加测试间隔，使用更稳定的断言

### 调试技巧

#### 启用详细日志
```go
logger.InitLogger(&logger.Config{
    Level:      "debug",  // 改为debug
    Encoding:   "console",
    OutputPath: "stdout",
})
```

#### 检查规则状态
```go
// 添加规则检查代码
rules := flow.GetRules()
for _, rule := range rules {
    t.Logf("规则: %s -> %s (阈值=%v)", rule.Resource, rule.TokenCalculateStrategy, rule.Threshold)
}
```

#### 监控统计信息
```go
// 获取统计信息
node := base.GetResourceNode("resource_name")
if node != nil {
    t.Logf("资源统计: 通过=%d, 阻塞=%d", node.PassQps(), node.BlockQps())
}
```

## 📈 性能考虑

### 测试效率
- **并发度**: 适度并发避免资源竞争
- **请求间隔**: 短窗口限流测试需要合适间隔
- **超时设置**: 长窗口测试需要足够的超时时间

### 资源使用
- **内存**: Sentinel规则存储在内存中
- **CPU**: 高并发测试时CPU使用率较高
- **网络**: 本地测试，网络开销最小

## 🎉 总结

这套API保护测试完整验证了基于`config.yaml`配置的HTTP API保护功能：

1. **配置驱动**: 完全基于真实配置文件
2. **功能完整**: 覆盖限流、熔断、通配符、优先级等所有功能
3. **真实环境**: 使用真实HTTP客户端和服务器
4. **自动化**: 完整的测试自动化脚本
5. **文档齐全**: 详细的使用和故障排除文档

通过这套测试，可以确信API保护功能在生产环境中能够正确工作，为系统提供可靠的流量控制和故障保护。 