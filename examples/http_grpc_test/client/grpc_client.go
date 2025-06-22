package client

import (
	"context"
	"fmt"
	"net"
	"time"

	user "github.com/qiaojinxia/distributed-service/examples/http-grpc-test/proto"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient gRPC客户端
type GRPCClient struct {
	conn       *grpc.ClientConn
	userClient user.UserServiceClient
}

// NewGRPCClient 创建gRPC客户端
func NewGRPCClient(address string) (*GRPCClient, error) {
	// 创建连接
	conn, err := grpc.NewClient(
		address,
		// 指定传输层凭证
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 设置连接超时
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "tcp", addr)
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// 创建客户端
	userClient := user.NewUserServiceClient(conn)

	return &GRPCClient{
		conn:       conn,
		userClient: userClient,
	}, nil
}

// Close 关闭连接
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetUser 获取用户
func (c *GRPCClient) GetUser(ctx context.Context, userID string) (*user.GetUserResponse, error) {
	logger.Info(ctx, "Calling gRPC GetUser", logger.String("user_id", userID))

	req := &user.GetUserRequest{
		UserId: userID,
	}

	resp, err := c.userClient.GetUser(ctx, req)
	if err != nil {
		logger.Error(ctx, "gRPC GetUser failed", logger.Err(err))
		return nil, err
	}

	logger.Info(ctx, "gRPC GetUser success", logger.String("user_name", resp.User.Name))
	return resp, nil
}

// ListUsers 列出用户
func (c *GRPCClient) ListUsers(ctx context.Context, page, pageSize int32, search string) (*user.ListUsersResponse, error) {
	logger.Info(ctx, "Calling gRPC ListUsers",
		logger.Int32("page", page),
		logger.Int32("page_size", pageSize),
		logger.String("search", search))

	req := &user.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	resp, err := c.userClient.ListUsers(ctx, req)
	if err != nil {
		logger.Error(ctx, "gRPC ListUsers failed", logger.Err(err))
		return nil, err
	}

	logger.Info(ctx, "gRPC ListUsers success",
		logger.Int32("total", resp.Total),
		logger.Int("returned", len(resp.Users)))
	return resp, nil
}

// CreateUser 创建用户
func (c *GRPCClient) CreateUser(ctx context.Context, name, email, phone string) (*user.CreateUserResponse, error) {
	logger.Info(ctx, "Calling gRPC CreateUser",
		logger.String("name", name),
		logger.String("email", email))

	req := &user.CreateUserRequest{
		Name:  name,
		Email: email,
		Phone: phone,
	}

	resp, err := c.userClient.CreateUser(ctx, req)
	if err != nil {
		logger.Error(ctx, "gRPC CreateUser failed", logger.Err(err))
		return nil, err
	}

	logger.Info(ctx, "gRPC CreateUser success", logger.String("user_id", resp.User.Id))
	return resp, nil
}

// UpdateUser 更新用户
func (c *GRPCClient) UpdateUser(ctx context.Context, userID, name, email, phone string) (*user.UpdateUserResponse, error) {
	logger.Info(ctx, "Calling gRPC UpdateUser", logger.String("user_id", userID))

	req := &user.UpdateUserRequest{
		UserId: userID,
		Name:   name,
		Email:  email,
		Phone:  phone,
	}

	resp, err := c.userClient.UpdateUser(ctx, req)
	if err != nil {
		logger.Error(ctx, "gRPC UpdateUser failed", logger.Err(err))
		return nil, err
	}

	logger.Info(ctx, "gRPC UpdateUser success")
	return resp, nil
}

// DeleteUser 删除用户
func (c *GRPCClient) DeleteUser(ctx context.Context, userID string) (*user.DeleteUserResponse, error) {
	logger.Info(ctx, "Calling gRPC DeleteUser", logger.String("user_id", userID))

	req := &user.DeleteUserRequest{
		UserId: userID,
	}

	resp, err := c.userClient.DeleteUser(ctx, req)
	if err != nil {
		logger.Error(ctx, "gRPC DeleteUser failed", logger.Err(err))
		return nil, err
	}

	logger.Info(ctx, "gRPC DeleteUser success")
	return resp, nil
}

// HealthCheck 健康检查
func (c *GRPCClient) HealthCheck(ctx context.Context) (*user.HealthCheckResponse, error) {
	logger.Info(ctx, "Calling gRPC HealthCheck")

	req := &user.HealthCheckRequest{
		Service: "user-service",
	}

	resp, err := c.userClient.HealthCheck(ctx, req)
	if err != nil {
		logger.Error(ctx, "gRPC HealthCheck failed", logger.Err(err))
		return nil, err
	}

	logger.Info(ctx, "gRPC HealthCheck success", logger.String("status", resp.Status))
	return resp, nil
}
