package ratelimit

import (
	"context"
	"distributed-service/pkg/config"
	"distributed-service/pkg/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

// RedisRateLimiter Redis 限流器实现
type RedisRateLimiter struct {
	client *redis.Client
	config config.RateLimitConfig
}

// NewRedisRateLimiter 创建Redis限流器
func NewRedisRateLimiter(cfg config.RateLimitConfig, redisClient *redis.Client) (RateLimiter, error) {
	if !cfg.Enabled {
		logger.Info(context.Background(), "Redis rate limiter is disabled")
		return &RedisRateLimiter{config: cfg}, nil
	}

	if redisClient == nil {
		return nil, fmt.Errorf("redis client is required for Redis rate limiter")
	}

	logger.Info(context.Background(), "Redis rate limiter initialized",
		zap.String("store_type", cfg.StoreType),
		zap.Bool("enabled", cfg.Enabled),
		zap.String("prefix", cfg.RedisPrefix))

	return &RedisRateLimiter{
		client: redisClient,
		config: cfg,
	}, nil
}

// IPRateLimit 基于IP的Redis限流
func (rl *RedisRateLimiter) IPRateLimit(limit string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		return fmt.Sprintf("%sip:%s", rl.config.RedisPrefix, c.ClientIP())
	}, limit)
}

// UserRateLimit 基于用户的Redis限流
func (rl *RedisRateLimiter) UserRateLimit(limit string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		userID, exists := c.Get("user_id")
		if !exists {
			userID = fmt.Sprintf("anonymous:%s", c.ClientIP())
		}
		return fmt.Sprintf("%suser:%v", rl.config.RedisPrefix, userID)
	}, limit)
}

// CustomRateLimit 自定义Redis限流
func (rl *RedisRateLimiter) CustomRateLimit(keyFunc func(*gin.Context) string, limit string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		keyPrefix := keyFunc(c)
		if keyPrefix == "" {
			return ""
		}
		return fmt.Sprintf("%s%s", rl.config.RedisPrefix, keyPrefix)
	}, limit)
}

// EndpointRateLimit 基于端点配置的Redis限流
func (rl *RedisRateLimiter) EndpointRateLimit(endpoint string) gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limit, exists := rl.config.Endpoints[endpoint]
	if !exists {
		limit = rl.config.DefaultConfig.UserPublic
		logger.Debug(context.Background(), "Using default rate limit for endpoint",
			zap.String("endpoint", endpoint),
			zap.String("limit", limit))
	}

	return rl.IPRateLimit(limit)
}

// GetConfiguredLimit 获取配置的限制
func (rl *RedisRateLimiter) GetConfiguredLimit(limitType string) string {
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

// rateLimitMiddleware Redis限流中间件核心实现
func (rl *RedisRateLimiter) rateLimitMiddleware(keyFunc func(*gin.Context) string, limit string) gin.HandlerFunc {
	// 解析限流配置
	requests, window, err := parseRateLimit(limit)
	if err != nil {
		logger.Error(context.Background(), "Invalid rate limit format",
			zap.String("limit", limit),
			zap.Error(err))
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		// 使用滑动窗口算法进行限流
		now := time.Now().Unix()
		windowStart := now - int64(window.Seconds())

		// 清除过期的请求记录
		rl.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// 获取当前窗口内的请求数量
		currentCount, err := rl.client.ZCard(ctx, key).Result()
		if err != nil {
			logger.Error(ctx, "Redis rate limiter error",
				zap.String("key", key),
				zap.Error(err))
			c.Next()
			return
		}

		// 设置响应头
		remaining := int64(requests) - currentCount
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now+int64(window.Seconds())))

		if currentCount >= int64(requests) {
			logger.Warn(ctx, "Redis rate limit exceeded",
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
				zap.Int64("current_count", currentCount),
				zap.String("limit", limit))

			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", requests, window),
				"retry_after": now + int64(window.Seconds()),
			})
			c.Abort()
			return
		}

		// 记录当前请求
		rl.client.ZAdd(ctx, key, &redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d-%d", now, time.Now().Nanosecond()),
		})

		// 设置过期时间
		rl.client.Expire(ctx, key, window)

		c.Next()
	}
}

// parseRateLimit 解析限流配置字符串
func parseRateLimit(limit string) (int, time.Duration, error) {
	if len(limit) < 3 {
		return 0, 0, fmt.Errorf("invalid rate limit format: %s", limit)
	}

	// 分离数量和时间单位
	parts := []rune(limit)
	var numPart, unitPart string

	for i, r := range parts {
		if r == '-' && i > 0 {
			numPart = string(parts[:i])
			unitPart = string(parts[i+1:])
			break
		}
	}

	if numPart == "" || unitPart == "" {
		return 0, 0, fmt.Errorf("invalid rate limit format: %s", limit)
	}

	// 解析数量
	var requests int
	if _, err := fmt.Sscanf(numPart, "%d", &requests); err != nil {
		return 0, 0, fmt.Errorf("invalid request count: %s", numPart)
	}

	// 解析时间单位
	var window time.Duration
	switch unitPart {
	case "S":
		window = time.Second
	case "M":
		window = time.Minute
	case "H":
		window = time.Hour
	case "D":
		window = 24 * time.Hour
	default:
		return 0, 0, fmt.Errorf("invalid time unit: %s", unitPart)
	}

	return requests, window, nil
}
