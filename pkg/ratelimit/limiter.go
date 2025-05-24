package ratelimit

import (
	"context"
	"distributed-service/pkg/config"
	"distributed-service/pkg/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	IPRateLimit(limit string) gin.HandlerFunc
	UserRateLimit(limit string) gin.HandlerFunc
	CustomRateLimit(keyFunc func(*gin.Context) string, limit string) gin.HandlerFunc
	EndpointRateLimit(endpoint string) gin.HandlerFunc
	GetConfiguredLimit(limitType string) string
}

// rateLimiter 限流器实现
type rateLimiter struct {
	store  limiter.Store
	config config.RateLimitConfig
}

// NewRateLimiter 创建新的限流器，支持配置文件
func NewRateLimiter(cfg config.RateLimitConfig) (RateLimiter, error) {
	if !cfg.Enabled {
		logger.Info(context.Background(), "Rate limiter is disabled")
		return &rateLimiter{config: cfg}, nil
	}

	// 使用内存存储，在后续版本中可以扩展支持Redis
	store := memory.NewStore()

	logger.Info(context.Background(), "Rate limiter initialized",
		zap.String("store_type", cfg.StoreType),
		zap.Bool("enabled", cfg.Enabled),
		zap.String("prefix", cfg.RedisPrefix))

	return &rateLimiter{
		store:  store,
		config: cfg,
	}, nil
}

// IPRateLimit 基于IP的限流中间件
func (rl *rateLimiter) IPRateLimit(limit string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate, err := limiter.NewRateFromFormatted(limit)
	if err != nil {
		logger.Error(context.Background(), "Invalid rate limit format",
			zap.String("limit", limit),
			zap.Error(err))
		return func(c *gin.Context) {
			c.Next()
		}
	}

	instance := limiter.New(rl.store, rate)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%sip:%s", rl.config.RedisPrefix, clientIP)

		rateLimitContext, err := instance.Get(c.Request.Context(), key)
		if err != nil {
			logger.Error(c.Request.Context(), "Rate limiter error",
				zap.String("key", key),
				zap.Error(err))
			c.Next()
			return
		}

		// 设置响应头信息
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitContext.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitContext.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitContext.Reset))

		if rateLimitContext.Reached {
			logger.Warn(c.Request.Context(), "Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.String("limit", limit))
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", rateLimitContext.Limit, rate.Period),
				"retry_after": rateLimitContext.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimit 基于用户的限流中间件
func (rl *rateLimiter) UserRateLimit(limit string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate, err := limiter.NewRateFromFormatted(limit)
	if err != nil {
		logger.Error(context.Background(), "Invalid rate limit format",
			zap.String("limit", limit),
			zap.Error(err))
		return func(c *gin.Context) {
			c.Next()
		}
	}

	instance := limiter.New(rl.store, rate)

	return func(c *gin.Context) {
		// 从JWT token中获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户ID，降级为IP限流
			clientIP := c.ClientIP()
			userID = fmt.Sprintf("anonymous:%s", clientIP)
		}

		key := fmt.Sprintf("%suser:%v", rl.config.RedisPrefix, userID)

		rateLimitContext, err := instance.Get(c.Request.Context(), key)
		if err != nil {
			logger.Error(c.Request.Context(), "Rate limiter error",
				zap.String("key", key),
				zap.Error(err))
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitContext.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitContext.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitContext.Reset))

		if rateLimitContext.Reached {
			logger.Warn(c.Request.Context(), "User rate limit exceeded",
				zap.Any("user_id", userID),
				zap.String("path", c.Request.URL.Path),
				zap.String("limit", limit))
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", rateLimitContext.Limit, rate.Period),
				"retry_after": rateLimitContext.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomRateLimit 自定义键值函数的限流中间件
func (rl *rateLimiter) CustomRateLimit(keyFunc func(*gin.Context) string, limit string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate, err := limiter.NewRateFromFormatted(limit)
	if err != nil {
		logger.Error(context.Background(), "Invalid rate limit format",
			zap.String("limit", limit),
			zap.Error(err))
		return func(c *gin.Context) {
			c.Next()
		}
	}

	instance := limiter.New(rl.store, rate)

	return func(c *gin.Context) {
		keyPrefix := keyFunc(c)
		if keyPrefix == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("%s%s", rl.config.RedisPrefix, keyPrefix)

		rateLimitContext, err := instance.Get(c.Request.Context(), key)
		if err != nil {
			logger.Error(c.Request.Context(), "Rate limiter error",
				zap.String("key", key),
				zap.Error(err))
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitContext.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitContext.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitContext.Reset))

		if rateLimitContext.Reached {
			logger.Warn(c.Request.Context(), "Custom rate limit exceeded",
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
				zap.String("limit", limit))
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", rateLimitContext.Limit, rate.Period),
				"retry_after": rateLimitContext.Reset,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointRateLimit 基于端点配置的限流中间件
func (rl *rateLimiter) EndpointRateLimit(endpoint string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 从配置中获取端点限流配置
	limit, exists := rl.config.Endpoints[endpoint]
	if !exists {
		// 如果没有特定配置，使用默认配置
		limit = rl.config.DefaultConfig.UserPublic
		logger.Debug(context.Background(), "Using default rate limit for endpoint",
			zap.String("endpoint", endpoint),
			zap.String("limit", limit))
	}

	return rl.IPRateLimit(limit)
}

// GetConfiguredLimit 根据限流类型获取配置的限制
func (rl *rateLimiter) GetConfiguredLimit(limitType string) string {
	switch limitType {
	case "health_check":
		return rl.config.DefaultConfig.HealthCheck
	case "auth_public":
		return rl.config.DefaultConfig.AuthPublic
	case "auth_protected":
		return rl.config.DefaultConfig.AuthProtected
	case "user_public":
		return rl.config.DefaultConfig.UserPublic
	case "user_protected":
		return rl.config.DefaultConfig.UserProtected
	default:
		return rl.config.DefaultConfig.UserPublic
	}
}

// KeyFunctions 提供常用的键值生成函数
var KeyFunctions = struct {
	// ByEndpoint 按端点限流
	ByEndpoint func(*gin.Context) string
	// ByUserAndEndpoint 按用户和端点组合限流
	ByUserAndEndpoint func(*gin.Context) string
}{
	ByEndpoint: func(c *gin.Context) string {
		return fmt.Sprintf("endpoint:%s:%s", c.Request.Method, c.Request.URL.Path)
	},
	ByUserAndEndpoint: func(c *gin.Context) string {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = fmt.Sprintf("anonymous:%s", c.ClientIP())
		}
		return fmt.Sprintf("user_endpoint:%v:%s:%s", userID, c.Request.Method, c.Request.URL.Path)
	},
}
