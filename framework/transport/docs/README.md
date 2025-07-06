# 传输模块设计文档

## 📋 概述

传输模块是分布式服务框架的通信核心组件，提供HTTP和gRPC的统一传输层支持。基于Gin和gRPC构建，支持负载均衡、服务发现、熔断降级和多协议互操作。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    客户端层                              │
│                  Client Layer                           │
│        HTTP Client | gRPC Client | WebSocket           │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   网关代理层                             │
│                Gateway Proxy Layer                      │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   负载均衡器    │   协议转换器    │   限流器        │ │
│  │Load Balancer    │Protocol Convert │Rate Limiter     │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   传输服务层                             │
│               Transport Service Layer                   │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   HTTP服务器    │   gRPC服务器    │   WebSocket     │ │
│  │  HTTP Server    │  gRPC Server    │  WebSocket      │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   业务处理层                             │
│               Business Logic Layer                      │
│              Handler | Service | Repository             │
└─────────────────────────────────────────────────────────┘
```

## 🎯 核心特点

### 1. 多协议支持
- **HTTP/1.1**: 基于Gin的高性能HTTP服务器
- **HTTP/2**: 原生HTTP/2支持和服务端推送
- **gRPC**: 高性能RPC通信协议
- **WebSocket**: 实时双向通信支持
- **协议转换**: HTTP到gRPC的自动转换

### 2. 服务发现
- **静态配置**: 基于配置文件的服务发现
- **动态发现**: 基于Consul/Etcd的动态服务发现
- **健康检查**: 自动健康检查和故障转移
- **服务注册**: 自动服务注册和注销

### 3. 负载均衡
- **轮询算法**: Round Robin负载均衡
- **加权轮询**: Weighted Round Robin
- **最少连接**: Least Connections
- **一致性哈希**: Consistent Hashing
- **自定义算法**: 支持自定义负载均衡策略

### 4. 可靠性保证
- **熔断器**: Circuit Breaker模式
- **重试机制**: 自动重试和退避策略
- **超时控制**: 请求超时和取消机制
- **优雅关闭**: 服务优雅启动和关闭

## 🚀 使用示例

### HTTP服务器

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/qiaojinxia/distributed-service/framework/transport/http"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
)

func main() {
    // 创建HTTP服务器配置
    config := http.ServerConfig{
        Port:            8080,
        Mode:            "release",
        MaxHeaderBytes:  1 << 20, // 1MB
        ReadTimeout:     30 * time.Second,
        WriteTimeout:    30 * time.Second,
        IdleTimeout:     60 * time.Second,
        ShutdownTimeout: 15 * time.Second,
    }
    
    // 创建HTTP服务器
    server := http.NewServer(config)
    
    // 添加全局中间件
    server.Use(middleware.CORS())
    server.Use(middleware.RequestID())
    server.Use(middleware.Logger())
    server.Use(middleware.Recovery())
    server.Use(middleware.Metrics())
    
    // 注册路由
    registerRoutes(server.Router())
    
    // 启动服务器
    if err := server.Start(); err != nil {
        panic(err)
    }
}

func registerRoutes(r *gin.Engine) {
    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "timestamp": time.Now(),
        })
    })
    
    // API路由组
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth("secret-key"))
    {
        // 用户相关
        users := api.Group("/users")
        {
            users.GET("", getUsersHandler)
            users.POST("", createUserHandler)
            users.GET("/:id", getUserHandler)
            users.PUT("/:id", updateUserHandler)
            users.DELETE("/:id", deleteUserHandler)
        }
        
        // 订单相关
        orders := api.Group("/orders")
        {
            orders.GET("", getOrdersHandler)
            orders.POST("", createOrderHandler)
            orders.GET("/:id", getOrderHandler)
        }
    }
    
    // 文件上传
    r.POST("/upload", uploadHandler)
    
    // WebSocket
    r.GET("/ws", websocketHandler)
}

// 处理器示例
func getUsersHandler(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    size := c.DefaultQuery("size", "10")
    
    users, total, err := userService.GetUsers(page, size)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "获取用户列表失败",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data":  users,
        "total": total,
        "page":  page,
        "size":  size,
    })
}
```

### gRPC服务器

```go
package main

import (
    "context"
    "net"
    "google.golang.org/grpc"
    "github.com/qiaojinxia/distributed-service/framework/transport/grpc"
    "github.com/qiaojinxia/distributed-service/framework/middleware"
    pb "your-project/proto"
)

func main() {
    // 创建gRPC服务器配置
    config := grpc.ServerConfig{
        Port:                 50051,
        MaxConcurrentStreams: 1000,
        MaxReceiveMessageSize: 4 * 1024 * 1024, // 4MB
        MaxSendMessageSize:    4 * 1024 * 1024, // 4MB
        ConnectionTimeout:     60 * time.Second,
        KeepaliveTime:        30 * time.Second,
        KeepaliveTimeout:     5 * time.Second,
    }
    
    // 创建gRPC服务器
    server := grpc.NewServer(config)
    
    // 添加拦截器
    server.AddUnaryInterceptor(middleware.UnaryRequestID())
    server.AddUnaryInterceptor(middleware.UnaryLogger())
    server.AddUnaryInterceptor(middleware.UnaryAuth("secret-key"))
    server.AddUnaryInterceptor(middleware.UnaryMetrics())
    server.AddUnaryInterceptor(middleware.UnaryRecovery())
    
    server.AddStreamInterceptor(middleware.StreamLogger())
    server.AddStreamInterceptor(middleware.StreamAuth("secret-key"))
    server.AddStreamInterceptor(middleware.StreamMetrics())
    
    // 注册服务
    pb.RegisterUserServiceServer(server.Server(), &userServiceImpl{})
    pb.RegisterOrderServiceServer(server.Server(), &orderServiceImpl{})
    
    // 启动服务器
    if err := server.Start(); err != nil {
        panic(err)
    }
}

// 用户服务实现
type userServiceImpl struct {
    pb.UnimplementedUserServiceServer
    userRepo *UserRepository
}

func (s *userServiceImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := s.userRepo.GetByID(req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "用户不存在: %v", err)
    }
    
    return &pb.GetUserResponse{
        User: &pb.User{
            Id:    user.ID,
            Name:  user.Name,
            Email: user.Email,
            CreatedAt: timestamppb.New(user.CreatedAt),
        },
    }, nil
}

func (s *userServiceImpl) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, status.Errorf(codes.Internal, "创建用户失败: %v", err)
    }
    
    return &pb.CreateUserResponse{
        User: &pb.User{
            Id:    user.ID,
            Name:  user.Name,
            Email: user.Email,
            CreatedAt: timestamppb.New(user.CreatedAt),
        },
    }, nil
}

// 流式服务示例
func (s *userServiceImpl) StreamUsers(req *pb.StreamUsersRequest, stream pb.UserService_StreamUsersServer) error {
    users, err := s.userRepo.GetAll()
    if err != nil {
        return status.Errorf(codes.Internal, "获取用户列表失败: %v", err)
    }
    
    for _, user := range users {
        if err := stream.Send(&pb.User{
            Id:    user.ID,
            Name:  user.Name,
            Email: user.Email,
            CreatedAt: timestamppb.New(user.CreatedAt),
        }); err != nil {
            return err
        }
        
        // 模拟实时数据流
        time.Sleep(100 * time.Millisecond)
    }
    
    return nil
}
```

### 客户端使用

```go
package main

import (
    "context"
    "time"
    "github.com/qiaojinxia/distributed-service/framework/transport/client"
)

func main() {
    // HTTP客户端
    httpClient := client.NewHTTPClient(client.HTTPConfig{
        BaseURL:    "http://api.example.com",
        Timeout:    30 * time.Second,
        MaxRetries: 3,
        RetryDelay: time.Second,
    })
    
    // 发送HTTP请求
    resp, err := httpClient.Get("/api/v1/users", client.WithHeaders(map[string]string{
        "Authorization": "Bearer " + token,
    }))
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    // gRPC客户端
    grpcClient, err := client.NewGRPCClient(client.GRPCConfig{
        Address:           "localhost:50051",
        DialTimeout:       5 * time.Second,
        MaxRetries:        3,
        EnableLoadBalance: true,
        LoadBalancePolicy: "round_robin",
    })
    if err != nil {
        panic(err)
    }
    defer grpcClient.Close()
    
    // 调用gRPC服务
    userClient := pb.NewUserServiceClient(grpcClient.Conn())
    user, err := userClient.GetUser(context.Background(), &pb.GetUserRequest{
        UserId: "12345",
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("用户信息: %+v\n", user)
}
```

### 网关代理

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/transport/gateway"
)

func main() {
    // 创建网关配置
    config := gateway.Config{
        Port: 8080,
        
        // 上游服务配置
        Upstreams: map[string]gateway.Upstream{
            "user-service": {
                Protocol: "grpc",
                Endpoints: []string{
                    "user-service-1:50051",
                    "user-service-2:50051",
                    "user-service-3:50051",
                },
                LoadBalance: "round_robin",
                HealthCheck: gateway.HealthCheck{
                    Enabled:  true,
                    Interval: 30 * time.Second,
                    Timeout:  5 * time.Second,
                    Path:     "/health",
                },
            },
            "order-service": {
                Protocol: "http",
                Endpoints: []string{
                    "http://order-service-1:8080",
                    "http://order-service-2:8080",
                },
                LoadBalance: "least_conn",
            },
        },
        
        // 路由配置
        Routes: []gateway.Route{
            {
                Path:     "/api/v1/users/*",
                Methods:  []string{"GET", "POST", "PUT", "DELETE"},
                Upstream: "user-service",
                Timeout:  30 * time.Second,
                Retry:    3,
            },
            {
                Path:     "/api/v1/orders/*",
                Methods:  []string{"GET", "POST"},
                Upstream: "order-service",
                Timeout:  15 * time.Second,
                Retry:    2,
            },
        },
        
        // 中间件配置
        Middleware: gateway.MiddlewareConfig{
            RateLimit: gateway.RateLimitConfig{
                Enabled: true,
                Rate:    1000, // 1000 req/min
                Burst:   100,
            },
            Auth: gateway.AuthConfig{
                Enabled: true,
                Type:    "jwt",
                Secret:  "your-secret-key",
            },
            CORS: gateway.CORSConfig{
                Enabled: true,
                Origins: []string{"*"},
                Methods: []string{"GET", "POST", "PUT", "DELETE"},
                Headers: []string{"Content-Type", "Authorization"},
            },
        },
    }
    
    // 创建并启动网关
    gw := gateway.New(config)
    if err := gw.Start(); err != nil {
        panic(err)
    }
}
```

## 🔧 配置选项

### HTTP服务器配置

```go
type ServerConfig struct {
    // 基础配置
    Host            string        `yaml:"host"`
    Port            int           `yaml:"port"`
    Mode            string        `yaml:"mode"`        // debug, release, test
    
    // 性能配置
    MaxHeaderBytes  int           `yaml:"max_header_bytes"`
    ReadTimeout     time.Duration `yaml:"read_timeout"`
    WriteTimeout    time.Duration `yaml:"write_timeout"`
    IdleTimeout     time.Duration `yaml:"idle_timeout"`
    
    // 关闭配置
    ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
    
    // TLS配置
    TLS             TLSConfig     `yaml:"tls"`
    
    // 静态文件配置
    Static          StaticConfig  `yaml:"static"`
}

type TLSConfig struct {
    Enabled  bool   `yaml:"enabled"`
    CertFile string `yaml:"cert_file"`
    KeyFile  string `yaml:"key_file"`
}

type StaticConfig struct {
    Enabled bool   `yaml:"enabled"`
    Root    string `yaml:"root"`
    Index   string `yaml:"index"`
}
```

### gRPC服务器配置

```go
type ServerConfig struct {
    // 基础配置
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
    
    // 连接配置
    MaxConcurrentStreams  uint32        `yaml:"max_concurrent_streams"`
    MaxReceiveMessageSize int           `yaml:"max_receive_message_size"`
    MaxSendMessageSize    int           `yaml:"max_send_message_size"`
    ConnectionTimeout     time.Duration `yaml:"connection_timeout"`
    
    // Keepalive配置
    KeepaliveTime    time.Duration `yaml:"keepalive_time"`
    KeepaliveTimeout time.Duration `yaml:"keepalive_timeout"`
    
    // TLS配置
    TLS TLSConfig `yaml:"tls"`
    
    // 反射配置
    EnableReflection bool `yaml:"enable_reflection"`
}
```

### 客户端配置

```go
type HTTPConfig struct {
    BaseURL    string            `yaml:"base_url"`
    Timeout    time.Duration     `yaml:"timeout"`
    MaxRetries int               `yaml:"max_retries"`
    RetryDelay time.Duration     `yaml:"retry_delay"`
    Headers    map[string]string `yaml:"headers"`
    
    // 连接池配置
    MaxIdleConns        int           `yaml:"max_idle_conns"`
    MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
    IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`
}

type GRPCConfig struct {
    Address           string        `yaml:"address"`
    DialTimeout       time.Duration `yaml:"dial_timeout"`
    MaxRetries        int           `yaml:"max_retries"`
    EnableLoadBalance bool          `yaml:"enable_load_balance"`
    LoadBalancePolicy string        `yaml:"load_balance_policy"`
    
    // 连接池配置
    MaxConnections int `yaml:"max_connections"`
    
    // TLS配置
    TLS TLSConfig `yaml:"tls"`
}
```

### 配置文件示例

```yaml
# config/transport.yaml
transport:
  http:
    host: "0.0.0.0"
    port: 8080
    mode: "release"
    
    max_header_bytes: 1048576  # 1MB
    read_timeout: "30s"
    write_timeout: "30s"
    idle_timeout: "60s"
    shutdown_timeout: "15s"
    
    tls:
      enabled: false
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"
      
    static:
      enabled: true
      root: "./static"
      index: "index.html"
      
  grpc:
    host: "0.0.0.0"
    port: 50051
    
    max_concurrent_streams: 1000
    max_receive_message_size: 4194304  # 4MB
    max_send_message_size: 4194304     # 4MB
    connection_timeout: "60s"
    
    keepalive_time: "30s"
    keepalive_timeout: "5s"
    
    enable_reflection: true
    
  gateway:
    port: 8080
    
    upstreams:
      user-service:
        protocol: "grpc"
        endpoints:
          - "user-service:50051"
        load_balance: "round_robin"
        health_check:
          enabled: true
          interval: "30s"
          timeout: "5s"
          
      order-service:
        protocol: "http"
        endpoints:
          - "http://order-service:8080"
        load_balance: "least_conn"
        
    routes:
      - path: "/api/v1/users/*"
        methods: ["GET", "POST", "PUT", "DELETE"]
        upstream: "user-service"
        timeout: "30s"
        retry: 3
        
    middleware:
      rate_limit:
        enabled: true
        rate: 1000
        burst: 100
      auth:
        enabled: true
        type: "jwt"
        secret: "your-secret-key"
      cors:
        enabled: true
        origins: ["*"]
        methods: ["GET", "POST", "PUT", "DELETE"]
```

## 📊 监控与指标

### HTTP指标

```go
type HTTPMetrics struct {
    // 请求统计
    RequestsTotal    int64         `json:"requests_total"`
    RequestsActive   int64         `json:"requests_active"`
    ResponseTime     time.Duration `json:"avg_response_time"`
    
    // 状态码分布
    Status2xx        int64         `json:"status_2xx"`
    Status3xx        int64         `json:"status_3xx"`
    Status4xx        int64         `json:"status_4xx"`
    Status5xx        int64         `json:"status_5xx"`
    
    // 吞吐量
    RequestsPerSecond float64      `json:"requests_per_second"`
    BytesPerSecond    float64      `json:"bytes_per_second"`
    
    // 连接统计
    ActiveConnections int          `json:"active_connections"`
    TotalConnections  int64        `json:"total_connections"`
}
```

### gRPC指标

```go
type GRPCMetrics struct {
    // RPC统计
    RPCsTotal        int64         `json:"rpcs_total"`
    RPCsActive       int64         `json:"rpcs_active"`
    RPCDuration      time.Duration `json:"avg_rpc_duration"`
    
    // 状态分布
    RPCsSuccessful   int64         `json:"rpcs_successful"`
    RPCsFailed       int64         `json:"rpcs_failed"`
    
    // 流统计
    StreamsTotal     int64         `json:"streams_total"`
    StreamsActive    int64         `json:"streams_active"`
    
    // 消息统计
    MessagesSent     int64         `json:"messages_sent"`
    MessagesReceived int64         `json:"messages_received"`
}
```

## 🔍 最佳实践

### 1. 错误处理

```go
// ✅ 推荐：统一错误响应
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func handleError(c *gin.Context, err error) {
    var resp ErrorResponse
    
    switch e := err.(type) {
    case *ValidationError:
        resp = ErrorResponse{
            Code:    "VALIDATION_ERROR",
            Message: "输入参数验证失败",
            Details: e.Fields,
        }
        c.JSON(http.StatusBadRequest, resp)
    case *NotFoundError:
        resp = ErrorResponse{
            Code:    "NOT_FOUND",
            Message: "资源不存在",
        }
        c.JSON(http.StatusNotFound, resp)
    default:
        resp = ErrorResponse{
            Code:    "INTERNAL_ERROR",
            Message: "内部服务器错误",
        }
        c.JSON(http.StatusInternalServerError, resp)
    }
}

// ✅ 推荐：gRPC错误处理
func handleGRPCError(err error) error {
    switch e := err.(type) {
    case *ValidationError:
        return status.Errorf(codes.InvalidArgument, "参数验证失败: %v", e.Message)
    case *NotFoundError:
        return status.Errorf(codes.NotFound, "资源不存在: %v", e.Message)
    case *PermissionError:
        return status.Errorf(codes.PermissionDenied, "权限不足: %v", e.Message)
    default:
        return status.Errorf(codes.Internal, "内部错误: %v", err)
    }
}
```

### 2. 性能优化

```go
// ✅ 推荐：连接池复用
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

// ✅ 推荐：gRPC连接复用
var grpcConn *grpc.ClientConn

func init() {
    var err error
    grpcConn, err = grpc.Dial("localhost:50051",
        grpc.WithInsecure(),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:    30 * time.Second,
            Timeout: 5 * time.Second,
        }))
    if err != nil {
        panic(err)
    }
}

// ✅ 推荐：批量处理
func batchProcessUsers(users []User) error {
    const batchSize = 100
    
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        
        batch := users[i:end]
        if err := processBatch(batch); err != nil {
            return err
        }
    }
    return nil
}
```

### 3. 安全最佳实践

```go
// ✅ 推荐：输入验证
func validateUserInput(user *User) error {
    if len(user.Name) < 2 || len(user.Name) > 50 {
        return errors.New("用户名长度必须在2-50字符之间")
    }
    
    if !isValidEmail(user.Email) {
        return errors.New("邮箱格式不正确")
    }
    
    if len(user.Password) < 8 {
        return errors.New("密码长度不能少于8位")
    }
    
    return nil
}

// ✅ 推荐：请求限制
func requestSizeLimit(maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.ContentLength > maxSize {
            c.AbortWithStatusJSON(413, gin.H{
                "error": "请求体过大",
            })
            return
        }
        
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
        c.Next()
    }
}

// ✅ 推荐：HTTPS重定向
func httpsRedirect() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.GetHeader("X-Forwarded-Proto") != "https" {
            httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
            c.Redirect(http.StatusMovedPermanently, httpsURL)
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 🚨 故障排查

### 常见问题

**Q1: HTTP连接超时**
```go
// 检查超时配置
func checkHTTPTimeouts() {
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            DialTimeout:         5 * time.Second,
            TLSHandshakeTimeout: 5 * time.Second,
            ResponseHeaderTimeout: 10 * time.Second,
        },
    }
}

// 监控连接状态
func monitorConnections() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        stats := getConnectionStats()
        if stats.ActiveConnections > 1000 {
            logger.Warn("High connection count", 
                logger.Int("active", stats.ActiveConnections))
        }
    }
}
```

**Q2: gRPC连接失败**
```go
// 连接健康检查
func healthCheck(conn *grpc.ClientConn) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    client := grpc_health_v1.NewHealthClient(conn)
    resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
    if err != nil {
        return err
    }
    
    if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
        return errors.New("service not healthy")
    }
    
    return nil
}

// 重连机制
func reconnectGRPC(address string) (*grpc.ClientConn, error) {
    backoff := []time.Duration{
        1 * time.Second,
        2 * time.Second,
        4 * time.Second,
        8 * time.Second,
    }
    
    for i, delay := range backoff {
        conn, err := grpc.Dial(address, grpc.WithInsecure())
        if err == nil {
            return conn, nil
        }
        
        if i < len(backoff)-1 {
            time.Sleep(delay)
        }
    }
    
    return nil, errors.New("failed to connect after retries")
}
```

## 🔮 高级功能

### 协议转换

```go
// HTTP到gRPC转换
func httpToGRPCProxy(grpcConn *grpc.ClientConn) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 解析HTTP请求
        method := c.Request.Method
        path := c.Request.URL.Path
        
        // 转换为gRPC调用
        grpcMethod := convertHTTPToGRPCMethod(method, path)
        
        // 调用gRPC服务
        resp, err := invokeGRPCMethod(grpcConn, grpcMethod, c.Request.Body)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        // 返回响应
        c.JSON(200, resp)
    }
}
```

### 服务网格集成

```go
// Istio集成
func istioHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 传递Istio相关头信息
        headers := []string{
            "x-request-id",
            "x-b3-traceid",
            "x-b3-spanid",
            "x-b3-parentspanid",
            "x-b3-sampled",
            "x-b3-flags",
        }
        
        for _, header := range headers {
            if value := c.GetHeader(header); value != "" {
                c.Header(header, value)
            }
        }
        
        c.Next()
    }
}
```

---

> 传输模块为框架提供了完整的网络通信能力，支持多协议、高性能、高可靠的分布式服务通信需求。