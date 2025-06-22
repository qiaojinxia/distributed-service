package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/middleware"
)

// Server HTTP服务器
type Server struct {
	engine *gin.Engine
	server *http.Server
	config *Config
	logger logger.Logger
}

// Config HTTP服务器配置
type Config struct {
	Port         int           `yaml:"port" json:"port"`
	Mode         string        `yaml:"mode" json:"mode"` // debug, release, test
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	EnableTLS    bool          `yaml:"enable_tls" json:"enable_tls"`
	CertFile     string        `yaml:"cert_file" json:"cert_file"`
	KeyFile      string        `yaml:"key_file" json:"key_file"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:         8080,
		Mode:         "debug",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		EnableTLS:    false,
	}
}

// NewServer 创建HTTP服务器
func NewServer(cfg *Config) *Server {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 设置Gin模式
	gin.SetMode(cfg.Mode)

	engine := gin.New()

	// 基础中间件
	engine.Use(
		middleware.HTTPRecoveryMiddleware(),
		middleware.HTTPLoggingMiddleware(),
		middleware.HTTPCORSMiddleware(),
		middleware.HTTPMetricsMiddleware(),
	)

	return &Server{
		engine: engine,
		config: cfg,
		logger: logger.GetLogger(),
	}
}

// Engine 获取Gin引擎
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Use 添加中间件
func (s *Server) Use(middlewares ...gin.HandlerFunc) {
	s.engine.Use(middlewares...)
}

// AddRoutes 添加路由组
func (s *Server) AddRoutes(prefix string, routes func(*gin.RouterGroup)) {
	group := s.engine.Group(prefix)
	routes(group)
}

// GET 添加GET路由
func (s *Server) GET(path string, handlers ...gin.HandlerFunc) {
	s.engine.GET(path, handlers...)
}

// POST 添加POST路由
func (s *Server) POST(path string, handlers ...gin.HandlerFunc) {
	s.engine.POST(path, handlers...)
}

// PUT 添加PUT路由
func (s *Server) PUT(path string, handlers ...gin.HandlerFunc) {
	s.engine.PUT(path, handlers...)
}

// DELETE 添加DELETE路由
func (s *Server) DELETE(path string, handlers ...gin.HandlerFunc) {
	s.engine.DELETE(path, handlers...)
}

// PATCH 添加PATCH路由
func (s *Server) PATCH(path string, handlers ...gin.HandlerFunc) {
	s.engine.PATCH(path, handlers...)
}

// Start 启动HTTP服务器
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.logger.Infof(context.Background(), "🌐 HTTP Server starting on %s (mode: %s)", addr, s.config.Mode)

	// 启动服务器
	go func() {
		var err error
		if s.config.EnableTLS {
			err = s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorf(context.Background(), "HTTP Server start failed: %v", err)
		}
	}()

	return nil
}

// Stop 停止HTTP服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.logger.Info(context.Background(), "🛑 Stopping HTTP Server...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Errorf(context.Background(), "HTTP Server shutdown failed: %v", err)
		return err
	}

	s.logger.Info(context.Background(), "✅ HTTP Server stopped")
	return nil
}

// ConvertConfig 转换配置格式
func ConvertConfig(cfg *config.HTTPConfig) (*Config, error) {
	return &Config{
		Port:         cfg.Port,
		Mode:         cfg.Mode,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		EnableTLS:    cfg.EnableTLS,
		CertFile:     cfg.CertFile,
		KeyFile:      cfg.KeyFile,
	}, nil
}
