package idgen

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/database"
	"gorm.io/gorm"
)

// FrameworkIDGenService 框架ID生成器服务
type FrameworkIDGenService struct {
	generator IDGenerator
	config    *config.IDGenConfig
	db        *gorm.DB
}

// NewFrameworkIDGenService 创建框架ID生成器服务
func NewFrameworkIDGenService() *FrameworkIDGenService {
	return &FrameworkIDGenService{}
}

// Initialize 初始化服务
func (s *FrameworkIDGenService) Initialize(ctx context.Context) error {
	// 检查配置是否启用
	if config.GlobalConfig.IDGen.Enabled == false {
		return fmt.Errorf("IDGen service is not enabled in configuration")
	}

	s.config = &config.GlobalConfig.IDGen

	// 获取数据库连接
	if s.config.UseFramework {
		// 使用框架数据库
		if database.DB == nil {
			return fmt.Errorf("framework database not initialized")
		}
		s.db = database.DB
	} else {
		// 使用自定义数据库配置
		db, err := s.createCustomDB()
		if err != nil {
			return fmt.Errorf("failed to create custom database connection: %w", err)
		}
		s.db = db
	}

	// 创建ID生成器
	generator, err := s.createGenerator()
	if err != nil {
		return fmt.Errorf("failed to create ID generator: %w", err)
	}

	s.generator = generator

	// 预创建业务标识
	if err := s.createPredefinedBizTags(ctx); err != nil {
		return fmt.Errorf("failed to create predefined biz tags: %w", err)
	}

	return nil
}

// Start 启动服务
func (s *FrameworkIDGenService) Start(ctx context.Context) error {
	if s.generator == nil {
		return fmt.Errorf("ID generator not initialized")
	}
	return nil
}

// Stop 停止服务
func (s *FrameworkIDGenService) Stop(ctx context.Context) error {
	if s.generator != nil {
		if closer, ok := s.generator.(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	return nil
}

// NextID 获取下一个ID
func (s *FrameworkIDGenService) NextID(ctx context.Context, bizTag string) (int64, error) {
	if s.generator == nil {
		return 0, fmt.Errorf("ID generator not initialized")
	}
	return s.generator.NextID(ctx, bizTag)
}

// BatchNextID 批量获取ID
func (s *FrameworkIDGenService) BatchNextID(ctx context.Context, bizTag string, count int) ([]int64, error) {
	if s.generator == nil {
		return nil, fmt.Errorf("ID generator not initialized")
	}
	return s.generator.BatchNextID(ctx, bizTag, count)
}

// createCustomDB 创建自定义数据库连接
func (s *FrameworkIDGenService) createCustomDB() (*gorm.DB, error) {
	dbConfig := &DatabaseConfig{
		Driver:   s.config.Database.Driver,
		DSN:      s.config.Database.DSN,
		Host:     s.config.Database.Host,
		Port:     s.config.Database.Port,
		Database: s.config.Database.Database,
		Username: s.config.Database.Username,
		Password: s.config.Database.Password,
		Charset:  s.config.Database.Charset,
		LogLevel: s.config.Database.LogLevel,
	}

	if s.config.Database.MaxIdleConns > 0 {
		dbConfig.MaxIdleConns = s.config.Database.MaxIdleConns
	} else {
		dbConfig.MaxIdleConns = 10
	}

	if s.config.Database.MaxOpenConns > 0 {
		dbConfig.MaxOpenConns = s.config.Database.MaxOpenConns
	} else {
		dbConfig.MaxOpenConns = 100
	}

	if s.config.Database.ConnMaxLifetime != "" {
		if duration, err := time.ParseDuration(s.config.Database.ConnMaxLifetime); err == nil {
			dbConfig.ConnMaxLifetime = duration
		} else {
			dbConfig.ConnMaxLifetime = time.Hour
		}
	} else {
		dbConfig.ConnMaxLifetime = time.Hour
	}

	return createGormDB(dbConfig)
}

// createGenerator 创建ID生成器
func (s *FrameworkIDGenService) createGenerator() (IDGenerator, error) {
	// 创建Leaf配置
	leafConfig := s.createLeafConfig()

	// 根据类型创建生成器
	switch s.config.Type {
	case "leaf", "gorm-leaf", "":
		return NewGormLeafIDGenerator(s.db, leafConfig), nil
	default:
		return nil, fmt.Errorf("unsupported ID generator type: %s", s.config.Type)
	}
}

// createLeafConfig 创建Leaf配置
func (s *FrameworkIDGenService) createLeafConfig() *LeafConfig {
	leafConfig := DefaultLeafConfig()

	// 从框架配置中覆盖参数
	if s.config.Leaf.DefaultStep > 0 {
		leafConfig.DefaultStep = s.config.Leaf.DefaultStep
	} else if s.config.DefaultStep > 0 {
		leafConfig.DefaultStep = s.config.DefaultStep
	}

	if s.config.Leaf.MaxStepSize > 0 {
		leafConfig.MaxStepSize = s.config.Leaf.MaxStepSize
	}

	if s.config.Leaf.MinStepSize > 0 {
		leafConfig.MinStepSize = s.config.Leaf.MinStepSize
	}

	if s.config.Leaf.PreloadThreshold != "" {
		if threshold, err := strconv.ParseFloat(s.config.Leaf.PreloadThreshold, 64); err == nil {
			leafConfig.PreloadThreshold = threshold
		}
	}

	if s.config.Leaf.CleanupInterval != "" {
		if interval, err := time.ParseDuration(s.config.Leaf.CleanupInterval); err == nil {
			leafConfig.CleanupInterval = interval
		}
	}

	if s.config.Leaf.StepAdjustRatio != "" {
		if ratio, err := strconv.ParseFloat(s.config.Leaf.StepAdjustRatio, 64); err == nil {
			leafConfig.StepAdjustRatio = ratio
		}
	}

	return leafConfig
}

// createPredefinedBizTags 创建预定义的业务标识
func (s *FrameworkIDGenService) createPredefinedBizTags(ctx context.Context) error {
	if s.config.BizTags == nil || len(s.config.BizTags) == 0 {
		return nil
	}

	// 如果生成器支持管理方法，使用它来创建业务标识
	if manager, ok := s.generator.(interface {
		CreateBizTag(ctx context.Context, bizTag string, step int32, description string) error
	}); ok {
		for bizTag, bizConfig := range s.config.BizTags {
			if bizConfig.AutoCreate {
				step := bizConfig.Step
				if step <= 0 {
					step = s.config.DefaultStep
					if step <= 0 {
						step = 1000 // 默认步长
					}
				}

				description := bizConfig.Description
				if description == "" {
					description = fmt.Sprintf("Auto created biz tag: %s", bizTag)
				}

				// 忽略已存在的错误
				_ = manager.CreateBizTag(ctx, bizTag, step, description)
			}
		}
	}

	return nil
}

// GetGenerator 获取底层ID生成器（用于高级操作）
func (s *FrameworkIDGenService) GetGenerator() IDGenerator {
	return s.generator
}

// CreateBizTag 创建业务标识
func (s *FrameworkIDGenService) CreateBizTag(ctx context.Context, bizTag string, step int32, description string) error {
	if manager, ok := s.generator.(interface {
		CreateBizTag(ctx context.Context, bizTag string, step int32, description string) error
	}); ok {
		return manager.CreateBizTag(ctx, bizTag, step, description)
	}
	return fmt.Errorf("current generator does not support biz tag management")
}

// UpdateStep 更新步长
func (s *FrameworkIDGenService) UpdateStep(ctx context.Context, bizTag string, newStep int32) error {
	if manager, ok := s.generator.(interface {
		UpdateStep(ctx context.Context, bizTag string, newStep int32) error
	}); ok {
		return manager.UpdateStep(ctx, bizTag, newStep)
	}
	return fmt.Errorf("current generator does not support step management")
}

// DeleteBizTag 删除业务标识
func (s *FrameworkIDGenService) DeleteBizTag(ctx context.Context, bizTag string) error {
	if manager, ok := s.generator.(interface {
		DeleteBizTag(ctx context.Context, bizTag string) error
	}); ok {
		return manager.DeleteBizTag(ctx, bizTag)
	}
	return fmt.Errorf("current generator does not support biz tag management")
}

// GetMetrics 获取指标
func (s *FrameworkIDGenService) GetMetrics(bizTag string) interface{} {
	if metricsProvider, ok := s.generator.(interface {
		GetMetrics(bizTag string) interface{}
	}); ok {
		return metricsProvider.GetMetrics(bizTag)
	}
	return nil
}

// GetAllMetrics 获取所有指标
func (s *FrameworkIDGenService) GetAllMetrics() map[string]interface{} {
	if metricsProvider, ok := s.generator.(interface {
		GetAllMetrics() map[string]interface{}
	}); ok {
		return metricsProvider.GetAllMetrics()
	}
	return nil
}

// GetBufferStatus 获取buffer状态
func (s *FrameworkIDGenService) GetBufferStatus(bizTag string) map[string]interface{} {
	if statusProvider, ok := s.generator.(interface {
		GetBufferStatus(bizTag string) map[string]interface{}
	}); ok {
		return statusProvider.GetBufferStatus(bizTag)
	}
	return nil
}