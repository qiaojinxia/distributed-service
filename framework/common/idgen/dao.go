package idgen

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// LeafDAO 数据访问对象接口
type LeafDAO interface {
	// GetLeafAlloc 获取叶子分配记录
	GetLeafAlloc(ctx context.Context, bizTag string) (*LeafAlloc, error)

	// UpdateMaxID 原子性更新最大ID
	UpdateMaxID(ctx context.Context, bizTag string, step int32) (*LeafAlloc, error)

	// CreateLeafAlloc 创建新的业务标识
	CreateLeafAlloc(ctx context.Context, bizTag string, step int32, description string) error

	// GetAllBizTags 获取所有业务标识
	GetAllBizTags(ctx context.Context) ([]string, error)

	// UpdateStep 更新步长
	UpdateStep(ctx context.Context, bizTag string, newStep int32) error

	// DeleteLeafAlloc 删除业务标识
	DeleteLeafAlloc(ctx context.Context, bizTag string) error

	// GetLeafAllocWithLock 带锁获取叶子分配记录（用于更新）
	GetLeafAllocWithLock(ctx context.Context, bizTag string) (*LeafAlloc, error)

	// BatchGetLeafAllocs 批量获取叶子分配记录
	BatchGetLeafAllocs(ctx context.Context, bizTags []string) ([]*LeafAlloc, error)

	// CreateTable 创建表
	CreateTable(ctx context.Context) error
}

// GormLeafDAO GORM实现的数据访问对象
type GormLeafDAO struct {
	db *gorm.DB
}

// NewGormLeafDAO 创建新的GORM DAO
func NewGormLeafDAO(db *gorm.DB) LeafDAO {
	return &GormLeafDAO{db: db}
}

// GetLeafAlloc 获取叶子分配记录
func (dao *GormLeafDAO) GetLeafAlloc(ctx context.Context, bizTag string) (*LeafAlloc, error) {
	var leafAlloc LeafAlloc
	err := dao.db.WithContext(ctx).Where("biz_tag = ?", bizTag).First(&leafAlloc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBizTagNotFound
		}
		return nil, fmt.Errorf("failed to get leaf alloc: %w", err)
	}
	return &leafAlloc, nil
}

// UpdateMaxID 原子性更新最大ID并返回新值
func (dao *GormLeafDAO) UpdateMaxID(ctx context.Context, bizTag string, step int32) (*LeafAlloc, error) {
	var leafAlloc LeafAlloc

	// 使用事务确保原子性
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询当前记录（加排他锁）
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("biz_tag = ?", bizTag).First(&leafAlloc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBizTagNotFound
			}
			return fmt.Errorf("failed to lock leaf alloc: %w", err)
		}

		// 计算新的max_id
		newMaxID := leafAlloc.MaxID + int64(step)

		// 更新max_id和update_time
		result := tx.Model(&leafAlloc).Where("biz_tag = ?", bizTag).Updates(map[string]interface{}{
			"max_id":      newMaxID,
			"update_time": time.Now(),
		})

		if result.Error != nil {
			return fmt.Errorf("failed to update max_id: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("no rows affected when updating max_id for bizTag: %s", bizTag)
		}

		// 更新内存中的值
		leafAlloc.MaxID = newMaxID
		leafAlloc.UpdateTime = time.Now()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &leafAlloc, nil
}

// CreateLeafAlloc 创建新的业务标识
func (dao *GormLeafDAO) CreateLeafAlloc(ctx context.Context, bizTag string, step int32, description string) error {
	leafAlloc := &LeafAlloc{
		BizTag:      bizTag,
		MaxID:       0,
		Step:        step,
		Description: description,
		UpdateTime:  time.Now(),
		AutoClean:   0,
	}

	err := dao.db.WithContext(ctx).Create(leafAlloc).Error
	if err != nil {
		return fmt.Errorf("failed to create leaf alloc: %w", err)
	}

	return nil
}

// GetAllBizTags 获取所有业务标识
func (dao *GormLeafDAO) GetAllBizTags(ctx context.Context) ([]string, error) {
	var bizTags []string
	err := dao.db.WithContext(ctx).Model(&LeafAlloc{}).
		Select("biz_tag").Find(&bizTags).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all biz tags: %w", err)
	}
	return bizTags, nil
}

// UpdateStep 更新步长
func (dao *GormLeafDAO) UpdateStep(ctx context.Context, bizTag string, newStep int32) error {
	result := dao.db.WithContext(ctx).Model(&LeafAlloc{}).
		Where("biz_tag = ?", bizTag).
		Updates(map[string]interface{}{
			"step":        newStep,
			"update_time": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update step: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrBizTagNotFound
	}

	return nil
}

// DeleteLeafAlloc 删除业务标识
func (dao *GormLeafDAO) DeleteLeafAlloc(ctx context.Context, bizTag string) error {
	result := dao.db.WithContext(ctx).Where("biz_tag = ?", bizTag).Delete(&LeafAlloc{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete leaf alloc: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrBizTagNotFound
	}

	return nil
}

// GetLeafAllocWithLock 带锁获取叶子分配记录
func (dao *GormLeafDAO) GetLeafAllocWithLock(ctx context.Context, bizTag string) (*LeafAlloc, error) {
	var leafAlloc LeafAlloc
	err := dao.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("biz_tag = ?", bizTag).First(&leafAlloc).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBizTagNotFound
		}
		return nil, fmt.Errorf("failed to get leaf alloc with lock: %w", err)
	}

	return &leafAlloc, nil
}

// BatchGetLeafAllocs 批量获取叶子分配记录
func (dao *GormLeafDAO) BatchGetLeafAllocs(ctx context.Context, bizTags []string) ([]*LeafAlloc, error) {
	var leafAllocs []*LeafAlloc
	err := dao.db.WithContext(ctx).Where("biz_tag IN ?", bizTags).Find(&leafAllocs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to batch get leaf allocs: %w", err)
	}
	return leafAllocs, nil
}

// CreateTable 创建表（如果不存在）
func (dao *GormLeafDAO) CreateTable(ctx context.Context) error {
	// 检查表是否已存在
	if dao.db.WithContext(ctx).Migrator().HasTable(&LeafAlloc{}) {
		return nil
	}

	// 自动迁移表结构
	err := dao.db.WithContext(ctx).AutoMigrate(&LeafAlloc{})
	if err != nil {
		return fmt.Errorf("failed to create leaf_alloc table: %w", err)
	}

	// 创建索引（使用Migrator确保兼容性）
	if !dao.db.WithContext(ctx).Migrator().HasIndex(&LeafAlloc{}, "idx_leaf_alloc_update_time") {
		err = dao.db.WithContext(ctx).Exec(`
			CREATE INDEX IF NOT EXISTS idx_leaf_alloc_update_time 
			ON leaf_alloc(update_time)
		`).Error
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// GetNextSegment 获取下一个号段
func (dao *GormLeafDAO) GetNextSegment(ctx context.Context, bizTag string, step int32) (*LeafSegment, error) {
	leafAlloc, err := dao.UpdateMaxID(ctx, bizTag, step)
	if err != nil {
		return nil, err
	}

	// 计算号段范围
	minID := leafAlloc.MaxID - int64(step) + 1
	maxID := leafAlloc.MaxID

	segment := NewLeafSegment(minID, maxID, step)
	return segment, nil
}
