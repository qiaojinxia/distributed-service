package ratelimit

import (
	"context"
	"distributed-service/pkg/config"
	"distributed-service/pkg/database"
	"distributed-service/pkg/logger"
	"go.uber.org/zap"
)

// NewRateLimiterFromConfig 根据配置创建适合的限流器
func NewRateLimiterFromConfig(cfg config.RateLimitConfig) (RateLimiter, error) {
	ctx := context.Background()
	if !cfg.Enabled {
		logger.Info(ctx, "Rate limiter is disabled")
		return NewRateLimiter(cfg)
	}

	switch cfg.StoreType {
	case "redis":
		if database.RedisClient == nil {
			logger.Warn(ctx, "Redis client not available, falling back to memory store")
			return NewRateLimiter(cfg)
		}
		logger.Info(ctx, "Creating Redis rate limiter",
			zap.String("prefix", cfg.RedisPrefix))
		return NewRedisRateLimiter(cfg, database.RedisClient)

	case "memory":
		logger.Info(ctx, "Creating memory rate limiter")
		return NewRateLimiter(cfg)

	default:
		logger.Warn(ctx, "Unknown store type, using memory",
			zap.String("store_type", cfg.StoreType))
		return NewRateLimiter(cfg)
	}
}
