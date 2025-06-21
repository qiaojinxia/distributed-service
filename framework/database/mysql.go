package database

import (
	"context"
	"distributed-service/framework/config"
	"distributed-service/framework/logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitMySQL(ctx context.Context, cfg *config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	logger.Info(ctx, "Initializing MySQL connection",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Database),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error(ctx, "Failed to connect to MySQL",
			logger.Error_(err),
			logger.String("host", cfg.Host),
			logger.Int("port", cfg.Port),
		)
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(ctx, "Failed to get underlying *sql.DB", logger.Error_(err))
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		logger.Error(ctx, "Failed to ping MySQL", logger.Error_(err))
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	logger.Info(ctx, "Successfully connected to MySQL",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Database),
	)

	DB = db
	return nil
}
