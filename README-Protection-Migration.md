# 🛡️ 保护机制配置迁移指南

## 📋 概述

本文档说明了如何从传统的 `ratelimit` 配置迁移到统一的 `protection` 保护机制配置。

## 🔄 迁移内容

### 1. 配置文件更改

**删除的配置：**
- ❌ `ratelimit` 传统限流配置
- ❌ `RateLimitConfig` 结构体
- ❌ `RateLimitDefaultConfig` 结构体

**保留的配置：**
- ✅ `protection` 统一保护机制配置
- ✅ 所有 `ProtectionConfig` 相关结构体

### 2. 测试脚本更改

**删除的脚本：**
- ❌ `scripts/test-ratelimit.sh`
- ❌ `scripts/test-ratelimit-circuitbreaker.sh`

**新增的脚本：**
- ✅ `scripts/test-protection.sh` - 统一保护机制测试

### 3. 部署脚本更新

**deploy.sh 更改：**
- 所有对旧测试脚本的引用都已更新为 `test-protection.sh`
- 保护监控链接更新为新的 `/admin` 接口
- 测试选项说明已更新

## 📚 配置对照表

### 传统配置 vs 新配置

| 传统 ratelimit | 新 protection | 说明 |
|----------------|---------------|------|
| `ratelimit.enabled` | `protection.enabled` | 启用保护机制 |
| `ratelimit.store_type` | `protection.storage.type` | 存储类型 |
| `ratelimit.redis_prefix` | `protection.storage.prefix` | 键前缀 |
| `ratelimit.endpoints` | `protection.rate_limit_rules` | 限流规则 |
| 无 | `protection.circuit_breakers` | 熔断器规则 |
| 无 | `protection.web_admin` | Web管理界面 |

### 限流规则迁移示例

**旧配置：**
```yaml
ratelimit:
  enabled: true
  store_type: memory
  endpoints:
    "/api/v1/auth/login": "5-M"
    "/health": "60-S"
```

**新配置：**
```yaml
protection:
  enabled: true
  storage:
    type: memory
  rate_limit_rules:
    - key: "api:login"
      limit: 5
      window: "1m"
      enabled: true
      description: "登录接口限流"
    - key: "api:health"
      limit: 60
      window: "1s"
      enabled: true
      description: "健康检查限流"
```

## 🔧 迁移步骤

### 1. 更新配置文件

选择使用新的配置文件：

```bash
# 开发环境
cp config/config-clean.yaml config/config.yaml

# 生产环境  
cp config/config-docker-clean.yaml config/config-docker.yaml
```

### 2. 更新代码引用

如果你的代码中引用了 `GlobalConfig.RateLimit`，需要更新为 `GlobalConfig.Protection`：

```go
// 旧代码
if config.GlobalConfig.RateLimit.Enabled {
    // ...
}

// 新代码
if config.GlobalConfig.Protection.Enabled {
    // ...
}
```

### 3. 运行新的测试

```bash
# 运行新的保护机制测试
./scripts/test-protection.sh

# 或使用部署脚本的测试选项
./deploy.sh
# 选择 3 (保护机制测试) 或 4 (综合测试)
```

## 🎯 新功能优势

### 1. 统一管理界面

访问 `http://localhost:8080/admin` 可以：
- 📊 查看限流和熔断器状态
- ⚙️ 动态更新保护规则
- 📈 监控保护机制效果

### 2. 更丰富的存储选项

```yaml
protection:
  storage:
    type: consul  # memory, redis, consul
    # 支持分布式配置管理
```

### 3. 熔断器支持

```yaml
protection:
  circuit_breakers:
    - name: "external_api"
      timeout: "5s"
      error_percent_threshold: 50
      # 完整的熔断器配置
```

### 4. 更强的测试覆盖

`test-protection.sh` 包含：
- ✅ 限流功能测试
- ✅ 熔断器功能测试
- ✅ Web管理界面测试
- ✅ 动态配置更新测试
- ✅ 性能测试

## 🚀 使用建议

### 开发环境

```yaml
protection:
  storage:
    type: memory
  web_admin:
    auth_enabled: false  # 方便开发调试
```

### 生产环境

```yaml
protection:
  storage:
    type: consul  # 分布式存储
  web_admin:
    auth_enabled: true   # 启用认证
    username: "${ADMIN_USER}"
    password: "${ADMIN_PASSWORD}"
```

## 📊 监控和调试

### 关键监控指标

- **限流状态**: `/admin/api/rate-limits`
- **熔断器状态**: `/admin/api/circuit-breakers`
- **系统指标**: `/metrics`

### 调试命令

```bash
# 查看限流状态
curl http://localhost:8080/admin/api/rate-limits | jq .

# 查看熔断器状态
curl http://localhost:8080/admin/api/circuit-breakers | jq .

# 测试限流
for i in {1..10}; do 
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}'
  sleep 0.1
done
```

## ❓ 常见问题

### Q: 如何保持向后兼容？

A: 虽然配置格式已更改，但可以在应用层添加配置转换逻辑来支持旧配置。

### Q: 性能有影响吗？

A: 新的保护机制经过优化，性能更好且功能更强。

### Q: 如何批量迁移规则？

A: 可以使用脚本将旧的 endpoints 配置转换为新的 rate_limit_rules 格式。

## 📝 相关文档

- [项目主文档](README.md)
- [保护机制详细说明](pkg/protection/README.md)
- [Web管理界面使用指南](examples/web-admin/README.md)

---

> 💡 **提示**: 建议在生产环境部署前，先在测试环境验证新配置的功能和性能。 