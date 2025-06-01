# 资源命名规范

## 📚 概述

为了统一 Sentinel 保护机制中限流和熔断器的资源命名，我们采用统一的路径风格命名规范，并**支持通配符匹配**来简化配置。

## 🎯 命名规范

### 统一原则
- 所有 `resource` 字段都使用 **路径风格** (`/path/to/resource`)
- **支持通配符模式** (`*` 匹配任意字符)
- **支持多模式匹配** (使用逗号分隔多个模式)
- 层次结构清晰，便于管理和识别
- 限流规则和熔断器使用相同的命名规范

### 通配符规则
```yaml
# 单一通配符匹配
/api/v1/users/*                          # 匹配所有用户相关接口

# 多模式匹配 (逗号分隔)
/grpc/*/get*,/grpc/*/list*,/grpc/*/find* # 匹配所有gRPC读操作

# 具体路径匹配
/health                                   # 精确匹配健康检查接口

# 层级通配符
/api/*                                    # 匹配所有API接口 (兜底)
```

### 分类规范

#### 1. HTTP API 资源
```yaml
# 具体接口
/health                          # 健康检查

# 模块级别通配符
/api/v1/auth/*                   # 所有认证接口 (login, register, refresh等)
/api/v1/users/*                  # 所有用户接口 (CRUD操作)
/protection/*                    # 所有保护状态接口

# 通用兜底
/api/*                           # 所有API接口兜底限流
```

#### 2. gRPC 服务资源
```yaml
# 服务级别通配符
/grpc/user_service/*             # 所有用户服务方法

# 操作类型匹配
/grpc/*/get*,/grpc/*/list*,/grpc/*/find*     # 所有读操作
/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*  # 所有写操作

# 具体方法 (如需要特殊配置)
/grpc/user_service/get_user      # 特定方法
```

#### 3. 基础设施资源
```yaml
# 统一基础设施
/infrastructure/*                # 所有基础设施 (database, redis, mq等)

# 具体组件 (如需要差异化配置)
/infrastructure/database         # 数据库操作
/infrastructure/redis           # Redis操作
```

#### 4. 外部依赖资源
```yaml
# 统一外部依赖
/external/*                      # 所有外部服务调用

# 具体服务 (如需要差异化配置)
/external/payment_gateway        # 支付网关
/external/notification_service   # 通知服务
```

## 📋 配置示例

### 简化的限流规则配置
```yaml
rate_limit_rules:
  # 健康检查 - 精确匹配
  - name: "health_check_limiter"
    resource: "/health"
    threshold: 2
    stat_interval_ms: 1000
    
  # 认证接口 - 通配符匹配
  - name: "auth_api_limiter"
    resource: "/api/v1/auth/*"          # 匹配所有认证接口
    threshold: 10
    stat_interval_ms: 60000
    
  # 用户接口 - 通配符匹配
  - name: "users_api_limiter"
    resource: "/api/v1/users/*"         # 匹配所有用户接口
    threshold: 30
    stat_interval_ms: 60000
    
  # gRPC读操作 - 多模式匹配
  - name: "grpc_read_operations_limiter"
    resource: "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*"
    threshold: 80
    stat_interval_ms: 1000
    
  # API兜底限流 - 通用通配符
  - name: "api_general_limiter"
    resource: "/api/*"                  # 兜底保护所有API
    threshold: 100
    stat_interval_ms: 60000
```

### 简化的熔断器配置
```yaml
circuit_breakers:
  # 认证接口熔断 - 通配符匹配
  - name: "auth_api_circuit"
    resource: "/api/v1/auth/*"          # 保护所有认证接口
    strategy: "ErrorRatio"
    threshold: 0.5
    
  # gRPC用户服务熔断 - 服务级通配符
  - name: "grpc_user_service_circuit"
    resource: "/grpc/user_service/*"    # 保护整个用户服务
    strategy: "SlowRequestRatio"
    threshold: 0.3
    
  # 基础设施熔断 - 统一保护
  - name: "infrastructure_circuit"
    resource: "/infrastructure/*"       # 保护所有基础设施
    strategy: "ErrorRatio"
    threshold: 0.5
```

## ✅ 优势

### 配置简化
- **从 13 个限流规则减少到 8 个**
- **从 8 个熔断器减少到 7 个**
- 维护更容易，配置更清晰

### 通配符匹配优势
1. **灵活性**: 一个规则覆盖多个相似资源
2. **可维护性**: 新增接口无需修改配置
3. **层次化保护**: 支持细粒度到粗粒度的保护策略
4. **兜底机制**: 通用规则确保全覆盖

### 匹配优先级
```yaml
优先级从高到低:
1. 精确匹配: /health
2. 具体通配符: /api/v1/auth/*
3. 多模式匹配: /grpc/*/get*,/grpc/*/list*
4. 通用通配符: /api/*
```

## 🚀 迁移指南

### 配置对比

| 场景 | 旧配置 (详细) | 新配置 (简化) |
|------|---------------|---------------|
| 登录接口 | `/api/v1/auth/login` | `/api/v1/auth/*` |
| 注册接口 | `/api/v1/auth/register` | `/api/v1/auth/*` |
| 获取用户 | `/grpc/user_service/get_user` | `/grpc/*/get*` |
| 创建用户 | `/grpc/user_service/create_user` | `/grpc/*/create*` |
| 数据库操作 | `/infrastructure/database` | `/infrastructure/*` |
| Redis操作 | `/infrastructure/redis` | `/infrastructure/*` |

### 迁移步骤
1. **备份现有配置**
2. **应用新的简化配置**
3. **验证通配符匹配正常工作**
4. **监控保护效果**
5. **根据需要微调阈值**

## 🔧 实现建议

### 代码中的匹配逻辑
```go
// 示例: 资源匹配函数
func MatchResource(resource string, pattern string) bool {
    // 支持多模式匹配 (逗号分隔)
    patterns := strings.Split(pattern, ",")
    for _, p := range patterns {
        if matched, _ := filepath.Match(strings.TrimSpace(p), resource); matched {
            return true
        }
    }
    return false
}
```

### 监控和调试
- 记录匹配的规则和资源
- 提供规则匹配状态查询接口
- 支持动态调整阈值

## 📝 注意事项

1. **通配符性能**: 通配符匹配比精确匹配略慢，但差异很小
2. **规则冲突**: 确保通配符规则不会意外匹配到不相关的资源
3. **测试覆盖**: 充分测试各种路径匹配场景
4. **监控观察**: 上线后观察保护效果，及时调整

## 📖 相关文档

- [Sentinel保护机制测试文档](../test/README_Sentinel_Tests.md)
- [配置文件说明](../config/config.yaml)
- [保护机制设计文档](./PROTECTION_DESIGN.md) 