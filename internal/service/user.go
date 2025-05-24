package service

import (
	"context"
	"distributed-service/internal/model"
	"distributed-service/internal/repository"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/tracing"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserService defines the interface for user business logic
type UserService interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error

	// Register Authentication methods
	Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, user *model.User) error {
	ctx, span := tracing.StartSpan(ctx, "userService.Create")
	defer span.End()

	tracing.AddSpanAttributes(ctx,
		attribute.String("user.username", user.Username),
		attribute.String("user.email", user.Email),
	)

	existing, err := s.repo.GetByUsername(ctx, user.Username)
	if err == nil && existing != nil {
		tracing.RecordError(ctx, ErrUserExists)
		return ErrUserExists
	}

	logger.Info(ctx, "Creating new user",
		logger.String("username", user.Username),
		logger.String("email", user.Email),
	)

	if err := s.repo.Create(ctx, user); err != nil {
		tracing.RecordError(ctx, err)
		return err
	}

	return nil
}

func (s *userService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	ctx, span := tracing.StartSpan(ctx, "userService.GetByID")
	defer span.End()

	tracing.AddSpanAttributes(ctx, attribute.Int64("user.id", int64(id)))

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tracing.RecordError(ctx, ErrUserNotFound)
			return nil, ErrUserNotFound
		}
		tracing.RecordError(ctx, err)
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	ctx, span := tracing.StartSpan(ctx, "userService.GetByUsername")
	defer span.End()

	tracing.AddSpanAttributes(ctx, attribute.String("user.username", username))

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tracing.RecordError(ctx, ErrUserNotFound)
			return nil, ErrUserNotFound
		}
		tracing.RecordError(ctx, err)
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, user *model.User) error {
	ctx, span := tracing.StartSpan(ctx, "userService.Update")
	defer span.End()

	tracing.AddSpanAttributes(ctx,
		attribute.Int64("user.id", int64(user.ID)),
		attribute.String("user.username", user.Username),
	)

	if err := s.repo.Update(ctx, user); err != nil {
		tracing.RecordError(ctx, err)
		return err
	}
	return nil
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	ctx, span := tracing.StartSpan(ctx, "userService.Delete")
	defer span.End()

	tracing.AddSpanAttributes(ctx, attribute.Int64("user.id", int64(id)))

	if err := s.repo.Delete(ctx, id); err != nil {
		tracing.RecordError(ctx, err)
		return err
	}
	return nil
}

// Register creates a new user account
func (s *userService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	return tracing.WithSpanResult(ctx, "userService.Register", func(ctx context.Context) (*model.User, error) {
		tracing.AddSpanAttributes(ctx,
			attribute.String("user.username", req.Username),
			attribute.String("user.email", req.Email),
		)

		// Check if user already exists
		existing, err := s.repo.GetByUsername(ctx, req.Username)
		if err == nil && existing != nil {
			return nil, ErrUserExists
		}

		// Hash password
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			logger.Error(ctx, "Failed to hash password", logger.Error_(err))
			return nil, err
		}

		// Create user
		user := &model.User{
			Username: req.Username,
			Email:    req.Email,
			Password: hashedPassword,
			Status:   1, // Active
		}

		if err := s.repo.Create(ctx, user); err != nil {
			logger.Error(ctx, "Failed to create user", logger.Error_(err))
			return nil, err
		}

		logger.Info(ctx, "User registered successfully",
			logger.String("username", user.Username),
			logger.String("email", user.Email),
		)

		return user, nil
	})
}

// Login authenticates a user
func (s *userService) Login(ctx context.Context, req *model.LoginRequest) (*model.User, error) {
	return tracing.WithSpanResult(ctx, "userService.Login", func(ctx context.Context) (*model.User, error) {
		tracing.AddSpanAttributes(ctx, attribute.String("user.username", req.Username))

		// Get user by username
		user, err := s.repo.GetByUsername(ctx, req.Username)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrInvalidCredentials
			}
			return nil, err
		}

		// Check password
		if !auth.CheckPassword(req.Password, user.Password) {
			logger.Warn(ctx, "Invalid password attempt",
				logger.String("username", req.Username),
			)
			return nil, ErrInvalidCredentials
		}

		// Check if user is active
		if user.Status != 1 {
			logger.Warn(ctx, "Inactive user login attempt",
				logger.String("username", req.Username),
			)
			return nil, ErrInvalidCredentials
		}

		logger.Info(ctx, "User logged in successfully",
			logger.String("username", user.Username),
		)

		return user, nil
	})
}

// ChangePassword changes user's password
func (s *userService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	return tracing.WithSpan(ctx, "userService.ChangePassword", func(ctx context.Context) error {
		tracing.AddSpanAttributes(ctx, attribute.Int64("user.id", int64(userID)))

		// Get user
		user, err := s.repo.GetByID(ctx, userID)
		if err != nil {
			return ErrUserNotFound
		}

		// Check old password
		if !auth.CheckPassword(oldPassword, user.Password) {
			return ErrInvalidPassword
		}

		// Hash new password
		hashedPassword, err := auth.HashPassword(newPassword)
		if err != nil {
			logger.Error(ctx, "Failed to hash new password", logger.Error_(err))
			return err
		}

		// Update password
		user.Password = hashedPassword
		if err := s.repo.Update(ctx, user); err != nil {
			logger.Error(ctx, "Failed to update password", logger.Error_(err))
			return err
		}

		logger.Info(ctx, "Password changed successfully",
			logger.Int("user_id", int(userID)),
		)

		return nil
	})
}
