# 认证模块设计文档

## 📋 概述

认证模块是分布式服务框架的安全核心组件，提供JWT令牌生成与验证、密码加密与校验等功能。支持多种加密算法，具有完整的用户认证和授权机制。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    应用层接口                            │
│           Authentication Middleware                     │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  认证服务层                              │
│                Auth Service Layer                       │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │  JWT 处理器     │   密码管理器    │   权限验证器    │ │
│  │ JWT Handler     │ Password Mgr    │ Permission Mgr  │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  核心功能层                              │
│                Core Function Layer                      │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   JWT 编解码    │   密码加密      │   令牌验证      │ │
│  │ Token Codec     │ Hash Function   │ Token Verify    │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 🎯 核心特点

### 1. JWT令牌管理
- **令牌生成**: 支持自定义Claims和过期时间
- **令牌验证**: 完整的签名验证和有效期检查
- **刷新机制**: 支持令牌自动刷新和续期
- **多算法支持**: HS256、RS256等多种签名算法

### 2. 密码安全
- **加密存储**: 使用bcrypt算法安全加密
- **强度验证**: 密码复杂度检查
- **盐值处理**: 自动生成和验证盐值
- **防暴力破解**: 支持失败次数限制

### 3. 权限控制
- **角色管理**: 基于角色的访问控制(RBAC)
- **权限验证**: 细粒度权限检查
- **中间件集成**: 与HTTP/gRPC中间件无缝集成

## 🚀 使用示例

### JWT令牌操作

```go
package main

import (
    "time"
    "github.com/qiaojinxia/distributed-service/framework/auth"
)

func main() {
    // 生成JWT令牌
    claims := auth.Claims{
        UserID:   123,
        Username: "john_doe",
        Role:     "admin",
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
            Issuer:    "distributed-service",
        },
    }
    
    token, err := auth.GenerateToken(claims, "your-secret-key")
    if err != nil {
        panic(err)
    }
    
    // 验证JWT令牌
    parsedClaims, err := auth.ParseToken(token, "your-secret-key")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("用户ID: %d, 用户名: %s\n", parsedClaims.UserID, parsedClaims.Username)
}
```

### 密码管理

```go
package main

import (
    "fmt"
    "github.com/qiaojinxia/distributed-service/framework/auth"
)

func main() {
    password := "user123456"
    
    // 加密密码
    hashedPassword, err := auth.HashPassword(password)
    if err != nil {
        panic(err)
    }
    
    // 验证密码
    isValid := auth.CheckPassword(password, hashedPassword)
    if isValid {
        fmt.Println("密码验证成功")
    } else {
        fmt.Println("密码验证失败")
    }
}
```

### 中间件集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/qiaojinxia/distributed-service/framework/auth"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
)

func main() {
    r := gin.Default()
    
    // 使用JWT认证中间件
    r.Use(middleware.JWTAuth("your-secret-key"))
    
    // 受保护的路由
    r.GET("/user/profile", func(c *gin.Context) {
        // 从上下文获取用户信息
        claims, exists := c.Get("claims")
        if !exists {
            c.JSON(401, gin.H{"error": "未授权"})
            return
        }
        
        userClaims := claims.(*auth.Claims)
        c.JSON(200, gin.H{
            "user_id": userClaims.UserID,
            "username": userClaims.Username,
            "role": userClaims.Role,
        })
    })
    
    r.Run(":8080")
}
```

## 🔧 配置选项

### JWT配置

```go
type JWTConfig struct {
    SecretKey       string        // 签名密钥
    Algorithm       string        // 签名算法 (HS256, RS256)
    ExpireDuration  time.Duration // 令牌有效期
    RefreshDuration time.Duration // 刷新令牌有效期
    Issuer          string        // 发行者
    Subject         string        // 主题
}
```

### 密码策略配置

```go
type PasswordConfig struct {
    MinLength    int  // 最小长度
    RequireUpper bool // 需要大写字母
    RequireLower bool // 需要小写字母
    RequireDigit bool // 需要数字
    RequireSymbol bool // 需要特殊符号
    MaxAttempts  int  // 最大尝试次数
}
```

## 🛡️ 安全最佳实践

### 1. 令牌安全
```go
// ✅ 推荐：使用强密钥
jwtSecret := generateRandomString(32) // 32字节随机字符串

// ✅ 推荐：设置合理的过期时间
expireDuration := time.Hour * 2 // 2小时过期

// ❌ 避免：硬编码密钥
// jwtSecret := "123456" // 不安全
```

### 2. 密码处理
```go
// ✅ 推荐：使用bcrypt加密
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// ✅ 推荐：密码强度验证
if len(password) < 8 {
    return errors.New("密码长度不足8位")
}

// ❌ 避免：明文存储密码
// user.Password = password // 不安全
```

### 3. 中间件使用
```go
// ✅ 推荐：分层权限控制
r.Use(middleware.JWTAuth(secret))
r.Use(middleware.RoleAuth("admin", "user"))

// ✅ 推荐：API路径保护
admin := r.Group("/admin")
admin.Use(middleware.RequireRole("admin"))
```

## 📊 性能指标

### 令牌操作性能
```
JWT生成: ~100µs
JWT验证: ~50µs
密码加密: ~100ms (bcrypt cost=10)
密码验证: ~100ms (bcrypt验证)
```

### 内存使用
```
JWT Claims: ~200字节
加密密码: ~60字节 (bcrypt hash)
中间件开销: ~1KB/请求
```

## 🚨 常见问题

### Q1: JWT令牌过期如何处理？

```go
func RefreshTokenHandler(c *gin.Context) {
    oldToken := c.GetHeader("Authorization")
    
    // 解析过期令牌（忽略过期错误）
    claims, err := auth.ParseTokenIgnoreExpiry(oldToken, secretKey)
    if err != nil {
        c.JSON(401, gin.H{"error": "无效令牌"})
        return
    }
    
    // 生成新令牌
    newClaims := auth.Claims{
        UserID:   claims.UserID,
        Username: claims.Username,
        Role:     claims.Role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
        },
    }
    
    newToken, _ := auth.GenerateToken(newClaims, secretKey)
    c.JSON(200, gin.H{"token": newToken})
}
```

### Q2: 如何实现用户登出？

```go
// 使用令牌黑名单
var tokenBlacklist = make(map[string]bool)

func LogoutHandler(c *gin.Context) {
    token := c.GetHeader("Authorization")
    
    // 将令牌加入黑名单
    tokenBlacklist[token] = true
    
    c.JSON(200, gin.H{"message": "登出成功"})
}

func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        
        // 检查黑名单
        if tokenBlacklist[token] {
            c.JSON(401, gin.H{"error": "令牌已失效"})
            c.Abort()
            return
        }
        
        // 正常验证流程...
        c.Next()
    }
}
```

## 🔮 扩展功能

### 多因子认证(MFA)
```go
type MFAConfig struct {
    Enabled    bool
    Methods    []string // "sms", "email", "totp"
    ExpireSecs int
}
```

### OAuth2集成
```go
type OAuth2Config struct {
    Providers map[string]OAuthProvider // "google", "github"
    RedirectURL string
    Scopes    []string
}
```

### 审计日志
```go
type AuthAuditLog struct {
    UserID    int       `json:"user_id"`
    Action    string    `json:"action"` // "login", "logout", "refresh"
    IP        string    `json:"ip"`
    UserAgent string    `json:"user_agent"`
    Timestamp time.Time `json:"timestamp"`
    Success   bool      `json:"success"`
}
```

---

> 认证模块为框架提供了完整的安全保障，支持现代Web应用的各种认证需求。