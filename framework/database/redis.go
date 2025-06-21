package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"time"
)

var RedisClient *redis.Client

func InitRedis(ctx context.Context, cfg *config.RedisConfig) error {
	logger.Info(ctx, "Initializing Redis connection",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.Int("db", cfg.DB),
	)

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error(ctx, "Failed to connect to Redis",
			logger.Error_(err),
			logger.String("host", cfg.Host),
			logger.Int("port", cfg.Port),
		)
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info(ctx, "Successfully connected to Redis",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.Int("db", cfg.DB),
	)

	RedisClient = client
	return nil
}
