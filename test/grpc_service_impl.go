package test

import (
	"context"
	"fmt"
	"time"

	pb "distributed-service/test/distributed-service/test/proto/user"
)

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	pb.UnimplementedUserServiceServer
}

// GetUser 获取用户信息
func (s *UserServiceImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// 模拟处理时间
	time.Sleep(10 * time.Millisecond)

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:        req.UserId,
			Name:      fmt.Sprintf("User-%s", req.UserId),
			Email:     fmt.Sprintf("user%s@example.com", req.UserId),
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}

// ListUsers 列出用户
func (s *UserServiceImpl) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// 模拟处理时间
	time.Sleep(20 * time.Millisecond)

	users := make([]*pb.User, req.PageSize)
	for i := int32(0); i < req.PageSize; i++ {
		users[i] = &pb.User{
			Id:        fmt.Sprintf("%d", req.Page*req.PageSize+i+1),
			Name:      fmt.Sprintf("User-%d", req.Page*req.PageSize+i+1),
			Email:     fmt.Sprintf("user%d@example.com", req.Page*req.PageSize+i+1),
			CreatedAt: time.Now().Unix(),
		}
	}

	return &pb.ListUsersResponse{
		Users: users,
		Total: 100, // 模拟总数
	}, nil
}

// CreateUser 创建用户
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 模拟处理时间（写操作通常耗时更长）
	time.Sleep(50 * time.Millisecond)

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:        fmt.Sprintf("new-%d", time.Now().Unix()),
			Name:      req.Name,
			Email:     req.Email,
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}

// UpdateUser 更新用户
func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// 模拟处理时间
	time.Sleep(40 * time.Millisecond)

	return &pb.UpdateUserResponse{
		User: &pb.User{
			Id:        req.UserId,
			Name:      req.Name,
			Email:     req.Email,
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}

// DeleteUser 删除用户
func (s *UserServiceImpl) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// 模拟处理时间
	time.Sleep(30 * time.Millisecond)

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}
