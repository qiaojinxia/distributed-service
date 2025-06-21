package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled   bool     `json:"enabled"`
	SkipPaths []string `json:"skip_paths"`
	TokenPath string   `json:"token_path"`
	Validator func(token string) (bool, map[string]interface{}, error)
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Enabled: true,
		SkipPaths: []string{
			"/health",
			"/monitor",
			"/api/auth/login",
			"/api/auth/register",
		},
		TokenPath: "Authorization",
		Validator: func(token string) (bool, map[string]interface{}, error) {
			// 默认的简单验证器（仅做示例）
			if token == "valid-token" {
				return true, map[string]interface{}{
					"user_id": 1,
					"role":    "user",
				}, nil
			}
			return false, nil, nil
		},
	}
}

// authMiddleware 认证中间件
func authMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过认证
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				c.Next()
				return
			}
		}

		// 获取token
		token := extractToken(c, config.TokenPath)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid token",
			})
			c.Abort()
			return
		}

		// 验证token
		valid, claims, err := config.Validator(token)
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息设置到context中
		for key, value := range claims {
			c.Set(key, value)
		}

		c.Next()
	}
}

// extractToken 从请求中提取token
func extractToken(c *gin.Context, tokenPath string) string {
	// 从Header中获取
	auth := c.GetHeader(tokenPath)
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// 从Query参数中获取
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}

// AuthMiddleware 创建认证中间件
func AuthMiddleware(config ...AuthConfig) gin.HandlerFunc {
	cfg := DefaultAuthConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	return authMiddleware(cfg)
}

// BasicAuthRoutes 添加基础认证路由
func BasicAuthRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		// 登录端点
		auth.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid request format",
				})
				return
			}

			// 简单的演示验证（实际应用中应该使用真实的用户验证）
			if req.Username == "admin" && req.Password == "password" {
				c.JSON(http.StatusOK, gin.H{
					"token": "valid-token", // 实际应用中应该生成真实的JWT token
					"user": gin.H{
						"id":       1,
						"username": req.Username,
						"role":     "admin",
					},
				})
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
		})

		// 注册端点
		auth.POST("/register", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required,min=6"`
				Email    string `json:"email" binding:"required,email"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid request format",
				})
				return
			}

			// 简单的演示注册（实际应用中应该保存到数据库）
			c.JSON(http.StatusCreated, gin.H{
				"token": "valid-token",
				"user": gin.H{
					"id":       2,
					"username": req.Username,
					"email":    req.Email,
					"role":     "user",
				},
				"message": "user created successfully",
			})
		})

		// Token验证端点
		auth.GET("/verify", func(c *gin.Context) {
			token := extractToken(c, "Authorization")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "missing token",
				})
				return
			}

			// 简单验证
			if token == "valid-token" {
				c.JSON(http.StatusOK, gin.H{
					"valid": true,
					"user": gin.H{
						"id":       1,
						"username": "admin",
						"role":     "admin",
					},
				})
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"valid": false,
				"error": "invalid token",
			})
		})
	}
}
