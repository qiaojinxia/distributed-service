package service

import (
	"context"
	"fmt"
	user "github.com/qiaojinxia/distributed-service/examples/http-grpc-test/proto"
	"strconv"
	"sync"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	user.UnimplementedUserServiceServer

	// 内存存储 (生产环境应该使用数据库)
	users  map[string]*user.User
	mutex  sync.RWMutex
	nextID int64
}

// NewUserService 创建用户服务实例
func NewUserService() *UserServiceImpl {
	svc := &UserServiceImpl{
		users:  make(map[string]*user.User),
		nextID: 1,
	}

	// 初始化一些测试数据
	svc.initTestData()

	return svc
}

// initTestData 初始化测试数据
func (s *UserServiceImpl) initTestData() {
	now := time.Now().Unix()
	testUsers := []*user.User{
		{
			Id:        "1",
			Name:      "Alice Johnson",
			Email:     "alice@example.com",
			Phone:     "+1-555-0101",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Id:        "2",
			Name:      "Bob Smith",
			Email:     "bob@example.com",
			Phone:     "+1-555-0102",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Id:        "3",
			Name:      "Charlie Brown",
			Email:     "charlie@example.com",
			Phone:     "+1-555-0103",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, u := range testUsers {
		s.users[u.Id] = u
	}
	s.nextID = 4
}

// GetUser 获取用户信息
func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	logger.Info(ctx, "GetUser called", logger.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	u, exists := s.users[req.UserId]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &user.GetUserResponse{
		User:    u,
		Message: "User retrieved successfully",
	}, nil
}

// ListUsers 列出用户
func (s *UserServiceImpl) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	logger.Info(ctx, "ListUsers called",
		logger.Int32("page", req.Page),
		logger.Int32("page_size", req.PageSize),
		logger.String("search", req.Search))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 收集所有用户
	var allUsers []*user.User
	for _, u := range s.users {
		// 如果有搜索条件，进行简单的名称匹配
		if req.Search != "" {
			if !contains(u.Name, req.Search) && !contains(u.Email, req.Search) {
				continue
			}
		}
		allUsers = append(allUsers, u)
	}

	// 分页处理
	total := int32(len(allUsers))
	page := req.Page
	pageSize := req.PageSize

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return &user.ListUsersResponse{
			Users:   []*user.User{},
			Total:   total,
			Message: "No users found for this page",
		}, nil
	}

	if end > total {
		end = total
	}

	pageUsers := allUsers[start:end]

	return &user.ListUsersResponse{
		Users:   pageUsers,
		Total:   total,
		Message: fmt.Sprintf("Retrieved %d users (page %d)", len(pageUsers), page),
	}, nil
}

// CreateUser 创建用户
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	logger.Info(ctx, "CreateUser called",
		logger.String("name", req.Name),
		logger.String("email", req.Email))

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查邮箱是否已存在
	for _, u := range s.users {
		if u.Email == req.Email {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
	}

	// 创建新用户
	now := time.Now().Unix()
	newUser := &user.User{
		Id:        strconv.FormatInt(s.nextID, 10),
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.users[newUser.Id] = newUser
	s.nextID++

	return &user.CreateUserResponse{
		User:    newUser,
		Message: "User created successfully",
	}, nil
}

// UpdateUser 更新用户
func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	logger.Info(ctx, "UpdateUser called", logger.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	u, exists := s.users[req.UserId]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// 检查邮箱是否被其他用户使用
	if req.Email != "" && req.Email != u.Email {
		for id, existingUser := range s.users {
			if id != req.UserId && existingUser.Email == req.Email {
				return nil, status.Error(codes.AlreadyExists, "email already exists")
			}
		}
	}

	// 更新字段
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	u.UpdatedAt = time.Now().Unix()

	return &user.UpdateUserResponse{
		User:    u,
		Message: "User updated successfully",
	}, nil
}

// DeleteUser 删除用户
func (s *UserServiceImpl) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	logger.Info(ctx, "DeleteUser called", logger.String("user_id", req.UserId))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.users[req.UserId]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	delete(s.users, req.UserId)

	return &user.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// HealthCheck 健康检查
func (s *UserServiceImpl) HealthCheck(ctx context.Context, req *user.HealthCheckRequest) (*user.HealthCheckResponse, error) {
	logger.Info(ctx, "HealthCheck called", logger.String("service", req.Service))

	return &user.HealthCheckResponse{
		Status:    "healthy",
		Message:   "User service is running normally",
		Timestamp: time.Now().Unix(),
	}, nil
}

// contains 简单的字符串包含检查 (不区分大小写)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			(len(s) > 0 && (s[0:len(substr)] == substr ||
				(len(s) > len(substr) && contains(s[1:], substr)))))
}
