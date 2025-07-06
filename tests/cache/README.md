# 缓存框架测试

## 🎯 测试目标

验证缓存框架的核心修复：
1. **框架初始化修复** - 缓存API不再返回nil
2. **TTL过期修复** - 自定义TTL正确工作
3. **策略功能验证** - LRU、TTL、Simple策略正常

## 📋 测试文件

### 🔴 关键测试
- `cache_integration_test.go` - 框架集成测试
- `ttl_behavior_test.go` - TTL行为测试

### 🟡 功能测试  
- `cache_policies_test.go` - 缓存策略测试

## 🚀 运行方式

### 运行所有测试（推荐）
```bash
# 从项目根目录运行
cd /Users/caomaoboy/GolandProjects/distributed-service
go test ./tests/ -v

# 或者进入测试目录运行
cd tests
go test -v
```

### 单独运行特定测试
```bash
# 运行框架集成测试
go test ./tests/ -v -run TestCacheIntegration

# 运行TTL行为测试
go test ./tests/ -v -run TestTTLBehavior

# 运行缓存策略测试
go test ./tests/ -v -run TestCachePolicies
```

## ✅ 预期结果

如果修复成功，应该看到：

**框架集成测试**：
- ✅ GetUserCache() 返回非nil实例
- ✅ 基本CRUD操作正常
- ✅ 缓存隔离正确

**TTL行为测试**：
- ✅ 自定义TTL（500ms）正确过期
- ✅ 默认TTL数据正常保留
- ✅ 多种TTL值并存

**策略测试**：
- ✅ LRU淘汰机制正确
- ✅ TTL策略支持自定义过期
- ✅ Simple策略基础功能正常

## 🔍 重点验证

1. **之前的问题**：`GetUserCache(): false` → `GetUserCache(): true`
2. **TTL修复**：500ms TTL数据在800ms后正确过期
3. **编译正常**：所有测试正常编译运行