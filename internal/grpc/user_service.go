package grpc

import (
	"context"

	pb "distributed-service/api/proto/user"
	"distributed-service/internal/model"
	"distributed-service/internal/service"
	"distributed-service/pkg/auth"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServiceServer implements the gRPC UserService
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
	jwtManager  *auth.JWTManager
}

// NewUserServiceServer creates a new gRPC user service server
func NewUserServiceServer(userService service.UserService, jwtManager *auth.JWTManager) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// GetUser retrieves a user by ID
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Add tracing attributes
	tracing.AddSpanAttributes(ctx,
		attribute.String("grpc.service", "user.v1.UserService"),
		attribute.String("grpc.method", "GetUser"),
		attribute.Int64("user.id", int64(req.Id)),
	)

	logger.Info(ctx, "gRPC GetUser called", logger.Int("user_id", int(req.Id)))

	if req.Id <= 0 {
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "invalid_argument"))
		return nil, status.Errorf(codes.InvalidArgument, "user ID must be positive")
	}

	user, err := s.userService.GetByID(ctx, uint(req.Id))
	if err != nil {
		logger.Error(ctx, "Failed to get user", logger.Error_(err))
		tracing.RecordError(ctx, err)
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "not_found"))
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	tracing.AddSpanAttributes(ctx,
		attribute.String("user.username", user.Username),
		attribute.String("user.email", user.Email),
		attribute.Bool("operation.success", true),
	)

	return &pb.GetUserResponse{
		User: convertUserToProto(user),
	}, nil
}

// CreateUser creates a new user
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Add tracing attributes
	tracing.AddSpanAttributes(ctx,
		attribute.String("grpc.service", "user.v1.UserService"),
		attribute.String("grpc.method", "CreateUser"),
		attribute.String("user.username", req.Username),
		attribute.String("user.email", req.Email),
	)

	logger.Info(ctx, "gRPC CreateUser called", logger.String("username", req.Username))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "invalid_argument"))
		return nil, status.Errorf(codes.InvalidArgument, "username, email, and password are required")
	}

	registerReq := &model.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := s.userService.Register(ctx, registerReq)
	if err != nil {
		logger.Error(ctx, "Failed to create user", logger.Error_(err))
		tracing.RecordError(ctx, err)
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "registration_failed"))
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	tracing.AddSpanAttributes(ctx,
		attribute.Int64("user.created_id", int64(user.ID)),
		attribute.Bool("operation.success", true),
	)

	return &pb.CreateUserResponse{
		User: convertUserToProto(user),
	}, nil
}

// UpdateUser updates an existing user
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	logger.Info(ctx, "gRPC UpdateUser called", logger.Int("user_id", int(req.Id)))

	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user ID must be positive")
	}

	// Get existing user
	user, err := s.userService.GetByID(ctx, uint(req.Id))
	if err != nil {
		logger.Error(ctx, "Failed to get user for update", logger.Error_(err))
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Update fields
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	err = s.userService.Update(ctx, user)
	if err != nil {
		logger.Error(ctx, "Failed to update user", logger.Error_(err))
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}

	return &pb.UpdateUserResponse{
		User: convertUserToProto(user),
	}, nil
}

// DeleteUser deletes a user
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	logger.Info(ctx, "gRPC DeleteUser called", logger.Int("user_id", int(req.Id)))

	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user ID must be positive")
	}

	err := s.userService.Delete(ctx, uint(req.Id))
	if err != nil {
		logger.Error(ctx, "Failed to delete user", logger.Error_(err))
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// ListUsers lists users with pagination (simplified implementation)
func (s *UserServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	logger.Info(ctx, "gRPC ListUsers called",
		logger.Int("page_size", int(req.PageSize)),
		logger.Int("page_number", int(req.PageNumber)))

	// Note: The current service interface doesn't have a ListUsers method
	// This is a simplified implementation that returns empty results
	// In a real implementation, you would add a ListUsers method to the service interface

	return &pb.ListUsersResponse{
		Users:      []*pb.User{},
		TotalCount: 0,
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Add tracing attributes
	tracing.AddSpanAttributes(ctx,
		attribute.String("grpc.service", "user.v1.UserService"),
		attribute.String("grpc.method", "Login"),
		attribute.String("user.username", req.Username),
	)

	logger.Info(ctx, "gRPC Login called", logger.String("username", req.Username))

	if req.Username == "" || req.Password == "" {
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "invalid_argument"))
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	loginReq := &model.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	user, err := s.userService.Login(ctx, loginReq)
	if err != nil {
		logger.Error(ctx, "Authentication failed", logger.Error_(err))
		tracing.RecordError(ctx, err)
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "authentication_failed"))
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(ctx, user.ID, user.Username)
	if err != nil {
		logger.Error(ctx, "Failed to generate access token", logger.Error_(err))
		tracing.RecordError(ctx, err)
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "token_generation_failed"))
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	// Generate refresh token (for simplicity, using the same method)
	refreshToken, err := s.jwtManager.GenerateToken(ctx, user.ID, user.Username)
	if err != nil {
		logger.Error(ctx, "Failed to generate refresh token", logger.Error_(err))
		tracing.RecordError(ctx, err)
		tracing.AddSpanAttributes(ctx, attribute.String("error.type", "refresh_token_generation_failed"))
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	tracing.AddSpanAttributes(ctx,
		attribute.Int64("user.id", int64(user.ID)),
		attribute.Bool("operation.success", true),
		attribute.Bool("tokens.generated", true),
	)

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         convertUserToProto(user),
	}, nil
}

// Check performs health check
func (s *UserServiceServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	logger.Debug(ctx, "gRPC Health check called", logger.String("service", req.Service))

	// Simple health check - you can add more sophisticated checks here
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}

// convertUserToProto converts internal user model to protobuf user
func convertUserToProto(user *model.User) *pb.User {
	return &pb.User{
		Id:        int64(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
