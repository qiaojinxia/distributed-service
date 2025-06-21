package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Config holds gRPC server configuration
type Config struct {
	Port                  int           `mapstructure:"port"`
	MaxRecvMsgSize        int           `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize        int           `mapstructure:"max_send_msg_size"`
	ConnectionTimeout     time.Duration `mapstructure:"connection_timeout"`
	MaxConnectionIdle     time.Duration `mapstructure:"max_connection_idle"`
	MaxConnectionAge      time.Duration `mapstructure:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `mapstructure:"max_connection_age_grace"`
	Time                  time.Duration `mapstructure:"time"`
	Timeout               time.Duration `mapstructure:"timeout"`
	EnableReflection      bool          `mapstructure:"enable_reflection"`
	EnableHealthCheck     bool          `mapstructure:"enable_health_check"`
}

// Server wraps gRPC server with additional functionality
type Server struct {
	server    *grpc.Server
	listener  net.Listener
	config    *Config
	healthSrv *health.Server
}

// NewServer creates a new gRPC server with middleware and configuration
func NewServer(ctx context.Context, config *Config, interceptors ...grpc.ServerOption) (*Server, error) {
	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", config.Port, err)
	}

	// Server options
	opts := []grpc.ServerOption{
		// Message size limits
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),

		// Keep alive parameters
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     config.MaxConnectionIdle,
			MaxConnectionAge:      config.MaxConnectionAge,
			MaxConnectionAgeGrace: config.MaxConnectionAgeGrace,
			Time:                  config.Time,
			Timeout:               config.Timeout,
		}),

		// Enforcement policy
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),

		// Connection timeout
		grpc.ConnectionTimeout(config.ConnectionTimeout),
	}

	// 添加传入的拦截器选项
	opts = append(opts, interceptors...)

	// Create gRPC server
	server := grpc.NewServer(opts...)

	// Health check service
	var healthSrv *health.Server
	if config.EnableHealthCheck {
		healthSrv = health.NewServer()
		grpc_health_v1.RegisterHealthServer(server, healthSrv)
	}

	// Enable reflection for development
	if config.EnableReflection {
		reflection.Register(server)
	}

	logger.Info(ctx, "gRPC server created",
		logger.Int("port", config.Port),
		logger.Bool("health_check_enabled", config.EnableHealthCheck),
		logger.Bool("reflection_enabled", config.EnableReflection),
		logger.Int("interceptors_count", len(interceptors)))

	return &Server{
		server:    server,
		listener:  lis,
		config:    config,
		healthSrv: healthSrv,
	}, nil
}

// RegisterService registers a gRPC service
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
}

// SetHealthStatus sets the health status for a service
func (s *Server) SetHealthStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	if s.healthSrv != nil {
		s.healthSrv.SetServingStatus(service, status)
	}
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	logger.Info(ctx, "Starting gRPC server",
		logger.String("address", s.listener.Addr().String()),
		logger.Int("port", s.config.Port),
	)

	// Set overall health status
	if s.healthSrv != nil {
		s.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	// 在goroutine中启动服务器，使其非阻塞
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			logger.Error(ctx, "gRPC server stopped with error", logger.Any("error", err))
		}
	}()

	logger.Info(ctx, "gRPC server started successfully")
	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info(ctx, "Stopping gRPC server...")

	// Set health status to not serving
	if s.healthSrv != nil {
		s.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	// Create a channel to signal when graceful stop is complete
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or context timeout
	select {
	case <-stopped:
		logger.Info(ctx, "gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		logger.Warn(ctx, "gRPC server stop timeout, forcing shutdown")
		s.server.Stop()
		return ctx.Err()
	}
}

// GetServer returns the underlying gRPC server
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

// GetListener returns the server listener
func (s *Server) GetListener() net.Listener {
	return s.listener
}

// DefaultConfig returns default gRPC server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:                  9090,
		MaxRecvMsgSize:        4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:        4 * 1024 * 1024, // 4MB
		ConnectionTimeout:     5 * time.Second,
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
		EnableReflection:      true,
		EnableHealthCheck:     true,
	}
}

// NewServerWithInterceptors 创建带有预配置拦截器的gRPC服务器
func NewServerWithInterceptors(ctx context.Context, config *Config,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor) (*Server, error) {

	// 构建服务器选项
	var opts []grpc.ServerOption

	// 添加一元拦截器
	if len(unaryInterceptors) > 0 {
		if len(unaryInterceptors) == 1 {
			opts = append(opts, grpc.UnaryInterceptor(unaryInterceptors[0]))
		} else {
			opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
		}
	}

	// 添加流拦截器
	if len(streamInterceptors) > 0 {
		if len(streamInterceptors) == 1 {
			opts = append(opts, grpc.StreamInterceptor(streamInterceptors[0]))
		} else {
			opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
		}
	}

	return NewServer(ctx, config, opts...)
}
