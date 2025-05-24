package repository

import (
	"context"
	"distributed-service/internal/model"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	ctx, span := tracing.StartSpan(ctx, "userRepository.Create")
	defer span.End()

	tracing.AddSpanAttributes(ctx,
		attribute.String("user.username", user.Username),
		attribute.String("user.email", user.Email),
	)

	logger.Info(ctx, "Creating user", logger.String("username", user.Username))

	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		tracing.RecordError(ctx, err)
		return err
	}

	tracing.TraceDatabase(ctx, "INSERT", "users", 1)
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	ctx, span := tracing.StartSpan(ctx, "userRepository.GetByID")
	defer span.End()

	tracing.AddSpanAttributes(ctx, attribute.Int64("user.id", int64(id)))

	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		tracing.RecordError(ctx, err)
		return nil, err
	}

	tracing.TraceDatabase(ctx, "SELECT", "users", 1)
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	ctx, span := tracing.StartSpan(ctx, "userRepository.GetByUsername")
	defer span.End()

	tracing.AddSpanAttributes(ctx, attribute.String("user.username", username))

	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		tracing.RecordError(ctx, err)
		return nil, err
	}

	tracing.TraceDatabase(ctx, "SELECT", "users", 1)
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	ctx, span := tracing.StartSpan(ctx, "userRepository.Update")
	defer span.End()

	tracing.AddSpanAttributes(ctx,
		attribute.Int64("user.id", int64(user.ID)),
		attribute.String("user.username", user.Username),
	)

	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		tracing.RecordError(ctx, err)
		return err
	}

	tracing.TraceDatabase(ctx, "UPDATE", "users", 1)
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	ctx, span := tracing.StartSpan(ctx, "userRepository.Delete")
	defer span.End()

	tracing.AddSpanAttributes(ctx, attribute.Int64("user.id", int64(id)))

	err := r.db.WithContext(ctx).Delete(&model.User{}, id).Error
	if err != nil {
		tracing.RecordError(ctx, err)
		return err
	}

	tracing.TraceDatabase(ctx, "DELETE", "users", 1)
	return nil
}
