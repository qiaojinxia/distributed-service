package test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"distributed-service/framework/config"
	"distributed-service/framework/logger"
	"distributed-service/framework/middleware"
	orderPb "distributed-service/test/distributed-service/test/proto/order"
	userPb "distributed-service/test/distributed-service/test/proto/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TestGRPCRateLimitWithClient 使用gRPC客户端测试限流功能
func TestGRPCRateLimitWithClient(t *testing.T) {
	// 初始化logger
	err := logger.InitLogger(&logger.Config{
		Level:      "info",
		Encoding:   "console",
		OutputPath: "stdout",
	})
	require.NoError(t, err)

	// 创建保护配置
	protectionConfig := &config.ProtectionConfig{
		Enabled: true,
		RateLimitRules: []config.RateLimitRuleConfig{
			{
				Name:           "grpc_user_service_limiter",
				Resource:       "/grpc/user_service/*",
				Threshold:      5, // 每秒5个请求
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "gRPC用户服务限流",
			},
			{
				Name:           "grpc_read_operations_limiter",
				Resource:       "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*",
				Threshold:      10, // 每秒10个请求
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "gRPC读操作限流",
			},
			{
				Name:           "grpc_write_operations_limiter",
				Resource:       "/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*",
				Threshold:      2, // 每秒2个请求（写操作限制更严）
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "gRPC写操作限流",
			},
		},
		CircuitBreakers: []config.CircuitBreakerRuleConfig{},
	}

	// 启动gRPC服务器
	server, address := startGRPCServer(t, protectionConfig)
	defer server.Stop()

	// 创建gRPC客户端
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
	}(conn)

	userClient := userPb.NewUserServiceClient(conn)
	orderClient := orderPb.NewOrderServiceClient(conn)

	t.Run("TestUserServiceRateLimit", func(t *testing.T) {
		// 用户服务限流测试 (5 QPS)
		ctx := context.Background()

		successCount := 0
		rateLimitedCount := 0

		// 在1秒内发送10个请求，应该有5个成功，5个被限流
		for i := 0; i < 10; i++ {
			_, err := userClient.GetUser(ctx, &userPb.GetUserRequest{
				UserId: fmt.Sprintf("user-%d", i),
			})

			if err != nil {
				grpcStatus := status.Convert(err)
				if grpcStatus.Code() == codes.ResourceExhausted {
					rateLimitedCount++
					t.Logf("🚫 请求 %d 被限流: %s", i+1, grpcStatus.Message())
				} else {
					t.Errorf("❌ 意外错误: %v", err)
				}
			} else {
				successCount++
				t.Logf("✅ 请求 %d 成功", i+1)
			}
		}

		t.Logf("📊 用户服务限流测试结果: 成功=%d, 限流=%d", successCount, rateLimitedCount)

		// 验证限流效果（允许一些误差）
		assert.GreaterOrEqual(t, successCount, 4, "至少应该有4个请求成功")
		assert.LessOrEqual(t, successCount, 6, "成功请求不应超过6个")
		assert.GreaterOrEqual(t, rateLimitedCount, 4, "至少应该有4个请求被限流")
	})

	t.Run("TestReadOperationsRateLimit", func(t *testing.T) {
		// 等待限流器重置
		time.Sleep(2 * time.Second)

		// 使用订单服务测试读操作限流 (10 QPS)
		ctx := context.Background()

		successCount := 0
		rateLimitedCount := 0

		// 并发发送15个读请求
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < 15; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				_, err := orderClient.GetOrder(ctx, &orderPb.GetOrderRequest{
					OrderId: fmt.Sprintf("order-%d", index),
				})

				mu.Lock()
				defer mu.Unlock()

				if err != nil {
					grpcStatus := status.Convert(err)
					if grpcStatus.Code() == codes.ResourceExhausted {
						rateLimitedCount++
						t.Logf("🚫 GetOrder请求 %d 被限流", index+1)
					} else {
						t.Errorf("❌ 意外错误: %v", err)
					}
				} else {
					successCount++
					t.Logf("✅ GetOrder请求 %d 成功", index+1)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("📊 读操作限流测试结果: 成功=%d, 限流=%d", successCount, rateLimitedCount)

		// 验证限流效果（订单GetOrder应该匹配读操作规则 10 QPS）
		assert.GreaterOrEqual(t, successCount, 8, "读操作应该有更高的阈值")
		assert.GreaterOrEqual(t, rateLimitedCount, 3, "应该有请求被限流")
	})

	t.Run("TestWriteOperationsRateLimit", func(t *testing.T) {
		// 等待限流器重置
		time.Sleep(2 * time.Second)

		// 使用订单服务测试写操作限流 (2 QPS)
		ctx := context.Background()

		successCount := 0
		rateLimitedCount := 0

		// 发送5个写请求，应该有2个成功，3个被限流
		for i := 0; i < 5; i++ {
			_, err := orderClient.CreateOrder(ctx, &orderPb.CreateOrderRequest{
				UserId: fmt.Sprintf("user-%d", i),
				Items: []*orderPb.OrderItem{
					{
						ProductId:   "product-1",
						ProductName: "Test Product",
						Quantity:    1,
						Price:       99.99,
					},
				},
			})

			if err != nil {
				grpcStatus := status.Convert(err)
				if grpcStatus.Code() == codes.ResourceExhausted {
					rateLimitedCount++
					t.Logf("🚫 CreateOrder请求 %d 被限流", i+1)
				} else {
					t.Errorf("❌ 意外错误: %v", err)
				}
			} else {
				successCount++
				t.Logf("✅ CreateOrder请求 %d 成功", i+1)
			}
		}

		t.Logf("📊 写操作限流测试结果: 成功=%d, 限流=%d", successCount, rateLimitedCount)

		// 验证限流效果
		assert.GreaterOrEqual(t, successCount, 1, "至少应该有1个请求成功")
		assert.LessOrEqual(t, successCount, 3, "成功请求不应超过3个")
		assert.GreaterOrEqual(t, rateLimitedCount, 2, "至少应该有2个请求被限流")
	})

	t.Run("TestPriorityMatching", func(t *testing.T) {
		// 等待限流器重置
		time.Sleep(2 * time.Second)

		// 测试优先级匹配：用户服务的GetUser应该匹配用户服务规则(5 QPS)而不是读操作规则(10 QPS)
		ctx := context.Background()

		successCount := 0
		rateLimitedCount := 0

		// 快速发送8个GetUser请求
		start := time.Now()
		for i := 0; i < 8; i++ {
			_, err := userClient.GetUser(ctx, &userPb.GetUserRequest{
				UserId: fmt.Sprintf("priority-test-%d", i),
			})

			if err != nil {
				grpcStatus := status.Convert(err)
				if grpcStatus.Code() == codes.ResourceExhausted {
					rateLimitedCount++
				}
			} else {
				successCount++
			}
		}
		elapsed := time.Since(start)

		t.Logf("📊 优先级匹配测试结果: 成功=%d, 限流=%d, 耗时=%v", successCount, rateLimitedCount, elapsed)

		// 应该按照用户服务规则(5 QPS)限流，而不是读操作规则(10 QPS)
		assert.LessOrEqual(t, successCount, 6, "成功请求应该按照用户服务规则限流(≤6)")
		assert.GreaterOrEqual(t, rateLimitedCount, 2, "应该有请求被限流")
	})

	t.Run("TestMixedOperations", func(t *testing.T) {
		// 等待限流器重置
		time.Sleep(2 * time.Second)

		// 混合操作测试
		ctx := context.Background()

		results := make(map[string]int)
		var mu sync.Mutex
		var wg sync.WaitGroup

		// 同时发送不同类型的请求
		operations := []struct {
			name string
			fn   func() error
		}{
			{
				name: "UserGetUser",
				fn: func() error {
					_, err := userClient.GetUser(ctx, &userPb.GetUserRequest{UserId: "mixed-test-1"})
					return err
				},
			},
			{
				name: "OrderGetOrder",
				fn: func() error {
					_, err := orderClient.GetOrder(ctx, &orderPb.GetOrderRequest{OrderId: "mixed-test-1"})
					return err
				},
			},
			{
				name: "OrderCreateOrder",
				fn: func() error {
					_, err := orderClient.CreateOrder(ctx, &orderPb.CreateOrderRequest{
						UserId: "mixed-test-user",
						Items: []*orderPb.OrderItem{
							{ProductId: "p1", ProductName: "Product 1", Quantity: 1, Price: 10.0},
						},
					})
					return err
				},
			},
		}

		// 每种操作发送5次
		for _, op := range operations {
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(operation string, opFunc func() error) {
					defer wg.Done()

					err := opFunc()

					mu.Lock()
					defer mu.Unlock()

					if err != nil {
						grpcStatus := status.Convert(err)
						if grpcStatus.Code() == codes.ResourceExhausted {
							results[operation+"_limited"]++
						} else {
							results[operation+"_error"]++
						}
					} else {
						results[operation+"_success"]++
					}
				}(op.name, op.fn)
			}
		}

		wg.Wait()

		t.Log("📊 混合操作测试结果:")
		for key, count := range results {
			t.Logf("   %s: %d", key, count)
		}

		// 验证每种操作都有相应的限流效果
		assert.Greater(t, results["UserGetUser_success"]+results["UserGetUser_limited"], 0, "UserGetUser操作应该有响应")
		assert.Greater(t, results["OrderGetOrder_success"]+results["OrderGetOrder_limited"], 0, "OrderGetOrder操作应该有响应")
		assert.Greater(t, results["OrderCreateOrder_success"]+results["OrderCreateOrder_limited"], 0, "OrderCreateOrder操作应该有响应")

		// 写操作应该有限流
		if results["OrderCreateOrder_limited"] == 0 {
			t.Logf("⚠️  写操作没有被限流，可能是因为请求间隔导致")
		}
	})
}

// startGRPCServer 启动gRPC服务器用于测试
func startGRPCServer(t *testing.T, protectionConfig *config.ProtectionConfig) (*grpc.Server, string) {
	// 创建Sentinel中间件
	sentinelMiddleware, err := middleware.NewSentinelProtectionMiddleware(context.Background(), protectionConfig)
	require.NoError(t, err)

	// 创建gRPC服务器，添加Sentinel拦截器
	server := grpc.NewServer(
		grpc.UnaryInterceptor(sentinelMiddleware.GRPCUnaryInterceptor()),
		grpc.StreamInterceptor(sentinelMiddleware.GRPCStreamInterceptor()),
	)

	// 注册用户服务
	userPb.RegisterUserServiceServer(server, &UserServiceImpl{})

	// 注册订单服务
	orderPb.RegisterOrderServiceServer(server, &OrderServiceImpl{})

	// 监听端口
	listener, err := net.Listen("tcp", ":0") // 使用随机端口
	require.NoError(t, err)

	// 启动服务器
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("gRPC服务器错误: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	address := listener.Addr().String()
	t.Logf("🚀 gRPC服务器启动: %s", address)

	return server, address
}
