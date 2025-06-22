package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/examples/http-grpc-test/client"
	user "github.com/qiaojinxia/distributed-service/examples/http-grpc-test/proto"
	"github.com/qiaojinxia/distributed-service/examples/http-grpc-test/service"
	"github.com/qiaojinxia/distributed-service/framework"
	"github.com/qiaojinxia/distributed-service/framework/logger"
	"github.com/qiaojinxia/distributed-service/framework/transport/grpc"
	"google.golang.org/grpc/status"
)

// 全局变量
var (
	grpcClient *client.GRPCClient
)

func main() {
	log := logger.GetLogger()

	log.Info(context.Background(), "🚀 启动HTTP + gRPC集成服务...")

	// 同时启动HTTP和gRPC服务
	err := framework.New().
		Port(8080).                // HTTP端口
		Name("http-grpc-service"). // 服务名称
		Version("v1.0.0").         // 版本
		Mode("debug").             // 运行模式
		EnableAll().               // 启用所有服务 (HTTP + gRPC + Metrics + Tracing)
		HTTP(setupHTTPRoutes).     // 注册HTTP路由
		GRPC(setupGRPCServices).   // 注册gRPC服务
		BeforeStart(func(ctx context.Context) error {
			log.Info(ctx, "🔧 初始化HTTP + gRPC服务...")
			return nil
		}).
		AfterStart(func(ctx context.Context) error {
			log.Info(ctx, "✅ HTTP + gRPC服务启动完成!")
			log.Info(ctx, "🌐 HTTP服务监听: http://localhost:8080")
			log.Info(ctx, "🔌 gRPC服务监听: localhost:9093")

			// 初始化gRPC客户端（用于HTTP到gRPC的调用）
			log.Info(ctx, "🔗 初始化gRPC客户端...")
			var err error
			grpcClient, err = client.NewGRPCClient("localhost:9093")
			if err != nil {
				log.Error(ctx, "Failed to create gRPC client", logger.Err(err))
				return err
			}
			log.Info(ctx, "✅ gRPC客户端初始化成功")

			log.Info(ctx, "📋 可用的HTTP端点:")
			log.Info(ctx, "  - GET  http://localhost:8080/health")
			log.Info(ctx, "  - GET  http://localhost:8080/ping")
			log.Info(ctx, "  - GET  http://localhost:8080/api/users")
			log.Info(ctx, "  - GET  http://localhost:8080/api/users/:id")
			log.Info(ctx, "  - POST http://localhost:8080/api/users")
			log.Info(ctx, "  - PUT  http://localhost:8080/api/users/:id")
			log.Info(ctx, "  - DELETE http://localhost:8080/api/users/:id")
			log.Info(ctx, "  - GET  http://localhost:8080/grpc/health")
			log.Info(ctx, "🔌 可用的gRPC服务:")
			log.Info(ctx, "  - UserService (user.UserService)")
			log.Info(ctx, "  - HealthCheck (grpc.health.v1.Health)")
			log.Info(ctx, "  - Server Reflection")
			return nil
		}).
		Run()

	if err != nil {
		log.Fatal(context.Background(), "服务启动失败", logger.Any("error", err))
	}

	// 清理资源
	if grpcClient != nil {
		err = grpcClient.Close()
	}
}

// setupHTTPRoutes 设置HTTP路由
func setupHTTPRoutes(r interface{}) {
	router := r.(*gin.Engine)
	log := logger.GetLogger()

	log.Info(context.Background(), "🌐 注册HTTP路由:")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "http-grpc-test",
			"version": "v1.0.0",
			"message": "HTTP service is working!",
		})
	})
	log.Info(context.Background(), "  ✅ GET /health")

	// Ping端点
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "pong",
			"timestamp": time.Now().Unix(),
		})
	})
	log.Info(context.Background(), "  ✅ GET /ping")

	// gRPC健康检查 (通过HTTP调用gRPC)
	router.GET("/grpc/health", grpcHealthHandler)
	log.Info(context.Background(), "  ✅ GET /grpc/health (calls gRPC)")

	// 用户API - 通过HTTP调用gRPC服务
	api := router.Group("/api")
	{
		// 列出用户
		api.GET("/users", listUsersHandler)
		log.Info(context.Background(), "  ✅ GET /api/users (calls gRPC)")

		// 获取用户详情
		api.GET("/users/:id", getUserHandler)
		log.Info(context.Background(), "  ✅ GET /api/users/:id (calls gRPC)")

		// 创建用户
		api.POST("/users", createUserHandler)
		log.Info(context.Background(), "  ✅ POST /api/users (calls gRPC)")

		// 更新用户
		api.PUT("/users/:id", updateUserHandler)
		log.Info(context.Background(), "  ✅ PUT /api/users/:id (calls gRPC)")

		// 删除用户
		api.DELETE("/users/:id", deleteUserHandler)
		log.Info(context.Background(), "  ✅ DELETE /api/users/:id (calls gRPC)")
	}
}

// setupGRPCServices 设置gRPC服务
func setupGRPCServices(s interface{}) {
	// gRPC服务器实例
	server := s.(*grpc.Server)

	log := logger.GetLogger()
	log.Info(context.Background(), "🔌 注册gRPC服务:")

	// 注册用户服务
	log.Info(context.Background(), "  📝 Creating UserService instance...")
	userService := service.NewUserService()

	log.Info(context.Background(), "  📝 Registering UserService with gRPC server...")
	user.RegisterUserServiceServer(server.GetServer(), userService)
	log.Info(context.Background(), "  ✅ UserService (user.UserService) registered successfully")
	log.Info(context.Background(), "    - GetUser")
	log.Info(context.Background(), "    - ListUsers")
	log.Info(context.Background(), "    - CreateUser")
	log.Info(context.Background(), "    - UpdateUser")
	log.Info(context.Background(), "    - DeleteUser")
	log.Info(context.Background(), "    - HealthCheck")

	log.Info(context.Background(), "  ✅ HealthCheck服务 (自动注册)")
	log.Info(context.Background(), "  ✅ Server Reflection (自动注册)")
	log.Info(context.Background(), "🎉 All gRPC services registered successfully!")
}

// ================================
// 🌐 HTTP处理器 (调用gRPC服务)
// ================================

// grpcHealthHandler gRPC健康检查
func grpcHealthHandler(c *gin.Context) {
	if grpcClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC client not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.HealthCheck(ctx)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "gRPC health check failed",
			"message": st.Message(),
			"code":    st.Code().String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    resp.Status,
		"message":   resp.Message,
		"timestamp": resp.Timestamp,
		"source":    "gRPC UserService",
	})
}

// listUsersHandler 列出用户
func listUsersHandler(c *gin.Context) {
	if grpcClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC client not available",
		})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.ListUsers(ctx, int32(page), int32(pageSize), search)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list users",
			"message": st.Message(),
			"code":    st.Code().String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":   resp.Users,
		"total":   resp.Total,
		"message": resp.Message,
		"source":  "gRPC UserService",
	})
}

// getUserHandler 获取用户详情
func getUserHandler(c *gin.Context) {
	if grpcClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC client not available",
		})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.GetUser(ctx, userID)
	if err != nil {
		st := status.Convert(err)
		statusCode := http.StatusInternalServerError
		if st.Code().String() == "NotFound" {
			statusCode = http.StatusNotFound
		} else if st.Code().String() == "InvalidArgument" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get user",
			"message": st.Message(),
			"code":    st.Code().String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    resp.User,
		"message": resp.Message,
		"source":  "gRPC UserService",
	})
}

// createUserHandler 创建用户
func createUserHandler(c *gin.Context) {
	if grpcClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC client not available",
		})
		return
	}

	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.CreateUser(ctx, req.Name, req.Email, req.Phone)
	if err != nil {
		st := status.Convert(err)
		statusCode := http.StatusInternalServerError
		if st.Code().String() == "AlreadyExists" {
			statusCode = http.StatusConflict
		} else if st.Code().String() == "InvalidArgument" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to create user",
			"message": st.Message(),
			"code":    st.Code().String(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    resp.User,
		"message": resp.Message,
		"source":  "gRPC UserService",
	})
}

// updateUserHandler 更新用户
func updateUserHandler(c *gin.Context) {
	if grpcClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC client not available",
		})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.UpdateUser(ctx, userID, req.Name, req.Email, req.Phone)
	if err != nil {
		st := status.Convert(err)
		statusCode := http.StatusInternalServerError
		if st.Code().String() == "NotFound" {
			statusCode = http.StatusNotFound
		} else if st.Code().String() == "AlreadyExists" {
			statusCode = http.StatusConflict
		} else if st.Code().String() == "InvalidArgument" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to update user",
			"message": st.Message(),
			"code":    st.Code().String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    resp.User,
		"message": resp.Message,
		"source":  "gRPC UserService",
	})
}

// deleteUserHandler 删除用户
func deleteUserHandler(c *gin.Context) {
	if grpcClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC client not available",
		})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.DeleteUser(ctx, userID)
	if err != nil {
		st := status.Convert(err)
		statusCode := http.StatusInternalServerError
		if st.Code().String() == "NotFound" {
			statusCode = http.StatusNotFound
		} else if st.Code().String() == "InvalidArgument" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to delete user",
			"message": st.Message(),
			"code":    st.Code().String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
		"source":  "gRPC UserService",
	})
}
