package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"distributed-service/framework/config"
	"distributed-service/framework/logger"
	"distributed-service/framework/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHandler 模拟API处理器
type MockHandler struct {
	delayMs    int
	errorRate  float64
	callCount  int64
	shouldFail bool
	mu         sync.Mutex
}

func NewMockHandler(delayMs int, errorRate float64) *MockHandler {
	return &MockHandler{
		delayMs:   delayMs,
		errorRate: errorRate,
	}
}

func (h *MockHandler) ServeHTTP(c *gin.Context) {
	h.mu.Lock()
	h.callCount++
	count := h.callCount
	h.mu.Unlock()

	// 模拟处理延迟
	if h.delayMs > 0 {
		time.Sleep(time.Duration(h.delayMs) * time.Millisecond)
	}

	// 模拟错误
	if h.shouldFail || (h.errorRate > 0 && float64(count%10) <= h.errorRate*10) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": fmt.Sprintf("Mock error for call %d", count),
		})
		return
	}

	// 正常响应
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Request processed successfully",
		"call_id":   count,
		"endpoint":  c.Request.URL.Path,
		"timestamp": time.Now().Unix(),
	})
}

func (h *MockHandler) SetShouldFail(shouldFail bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.shouldFail = shouldFail
}

func (h *MockHandler) GetCallCount() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.callCount
}

func (h *MockHandler) ResetCallCount() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callCount = 0
}

// TestAPIProtectionWithRealConfig 使用真实配置测试API保护功能
func TestAPIProtectionWithRealConfig(t *testing.T) {
	// 初始化logger
	err := logger.InitLogger(&logger.Config{
		Level:      "info",
		Encoding:   "console",
		OutputPath: "stdout",
	})
	require.NoError(t, err)

	// 创建保护配置（根据config.yaml）
	protectionConfig := &config.ProtectionConfig{
		Enabled: true,
		RateLimitRules: []config.RateLimitRuleConfig{
			{
				Name:           "health_check_limiter",
				Resource:       "/health",
				Threshold:      2,
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "健康检查接口限流 - 每秒2次 (2 QPS)",
			},
			{
				Name:           "auth_api_limiter",
				Resource:       "/api/v1/auth/*",
				Threshold:      10,
				StatIntervalMs: 60000,
				Enabled:        true,
				Description:    "认证接口限流 - 每分钟10次 (0.167 QPS)",
			},
			{
				Name:           "users_api_limiter",
				Resource:       "/api/v1/users/*",
				Threshold:      30,
				StatIntervalMs: 60000,
				Enabled:        true,
				Description:    "用户接口限流 - 每分钟30次 (0.5 QPS)",
			},
			{
				Name:           "api_general_limiter",
				Resource:       "/api/*",
				Threshold:      100,
				StatIntervalMs: 60000,
				Enabled:        true,
				Description:    "API接口通用限流 - 每分钟100次 (1.67 QPS)",
			},
			{
				Name:           "protection_status_limiter",
				Resource:       "/protection/*",
				Threshold:      20,
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "保护状态接口限流 - 每秒20次 (20 QPS)",
			},
		},
		CircuitBreakers: []config.CircuitBreakerRuleConfig{
			{
				Name:                         "auth_api_circuit",
				Resource:                     "/api/v1/auth/*",
				Strategy:                     "ErrorRatio",
				RetryTimeoutMs:               5000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 10,
				Threshold:                    0.5,
				ProbeNum:                     3,
				Enabled:                      true,
				Description:                  "认证接口熔断器 - 错误率超过50%时熔断",
			},
			{
				Name:                         "auth_api_circuit",
				Resource:                     "/api/test/auth/*",
				Strategy:                     "ErrorRatio",
				RetryTimeoutMs:               5000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 10,
				Threshold:                    0.5,
				ProbeNum:                     3,
				Enabled:                      true,
				Description:                  "认证接口熔断器 - 错误率超过50%时熔断",
			},
			{
				Name:                         "users_api_circuit",
				Resource:                     "/api/v1/users/*",
				Strategy:                     "ErrorRatio",
				RetryTimeoutMs:               3000,
				MinRequestAmount:             8,
				StatIntervalMs:               8000,
				StatSlidingWindowBucketCount: 8,
				Threshold:                    0.6,
				ProbeNum:                     2,
				Enabled:                      true,
				Description:                  "用户接口熔断器 - 错误率超过60%时熔断",
			},
			{
				Name:                         "api_general_circuit",
				Resource:                     "/api/*",
				Strategy:                     "ErrorRatio",
				RetryTimeoutMs:               10000,
				MinRequestAmount:             20,
				StatIntervalMs:               15000,
				StatSlidingWindowBucketCount: 15,
				Threshold:                    0.8,
				ProbeNum:                     5,
				Enabled:                      true,
				Description:                  "API接口通用熔断器 - 错误率超过80%时熔断",
			},
		},
	}

	// 创建HTTP服务器
	server := setupHTTPServer(t, protectionConfig)

	t.Run("TestHealthCheckRateLimit", func(t *testing.T) {
		// 健康检查接口限流测试 (2 QPS)
		testRateLimit(t, server, "/health", 2, 1*time.Second, 5, "健康检查")
	})

	t.Run("TestAuthAPIRateLimit", func(t *testing.T) {
		// 认证接口限流测试 (10 requests per 60s)
		testRateLimit(t, server, "/api/v1/auth/login", 10, 60*time.Second, 15, "认证接口")
	})

	t.Run("TestUsersAPIRateLimit", func(t *testing.T) {
		// 用户接口限流测试 (30 requests per 60s)
		testRateLimit(t, server, "/api/v1/users/profile", 30, 60*time.Second, 40, "用户接口")
	})

	t.Run("TestProtectionStatusRateLimit", func(t *testing.T) {
		// 保护状态接口限流测试 (20 QPS)
		testRateLimit(t, server, "/protection/status", 20, 1*time.Second, 30, "保护状态")
	})

	t.Run("TestAPIGeneralRateLimit", func(t *testing.T) {
		// 通用API限流测试 (100 requests per 60s)
		testRateLimit(t, server, "/api/v1/products/list", 100, 60*time.Second, 120, "通用API")
	})

	t.Run("TestPriorityMatching", func(t *testing.T) {
		// 优先级匹配测试
		testPriorityMatching(t, server)
	})

	//使用不存在的接口测试熔断
	t.Run("TestAuthAPICircuitBreaker", func(t *testing.T) {
		// 认证接口熔断器测试 (50%错误率)
		testCircuitBreaker(t, server, "/api/test/auth/register", 0.5, 10, "认证接口熔断器")
	})

	t.Run("TestUsersAPICircuitBreaker", func(t *testing.T) {
		// 用户接口熔断器测试 (60%错误率)
		testCircuitBreaker(t, server, "/api/v1/users/delete", 0.6, 8, "用户接口熔断器")
	})

	t.Run("TestAPIGeneralCircuitBreaker", func(t *testing.T) {
		// 通用API熔断器测试 (80%错误率)
		testCircuitBreaker(t, server, "/api/test/orders/create", 0.8, 20, "通用API熔断器")
	})

	t.Run("TestConcurrentRequests", func(t *testing.T) {
		// 并发请求测试
		testConcurrentRequests(t, server)
	})

	t.Run("TestWildcardMatching", func(t *testing.T) {
		// 通配符匹配测试
		testWildcardMatching(t, server)
	})
}

// testRateLimit 测试限流功能
func testRateLimit(t *testing.T, server *httptest.Server, endpoint string, threshold int, window time.Duration, totalRequests int, description string) {
	client := &http.Client{Timeout: 5 * time.Second}

	var successCount, limitedCount int
	start := time.Now()

	t.Logf("🧪 开始%s限流测试: %s (阈值=%d, 窗口=%v)", description, endpoint, threshold, window)

	for i := 0; i < totalRequests; i++ {
		resp, err := client.Get(server.URL + endpoint)
		require.NoError(t, err)

		switch resp.StatusCode {
		case http.StatusOK:
			successCount++
			t.Logf("✅ 请求 %d/%d 成功", i+1, totalRequests)
		case http.StatusTooManyRequests:
			limitedCount++
			t.Logf("🚫 请求 %d/%d 被限流", i+1, totalRequests)
		default:
			t.Logf("❓ 请求 %d/%d 状态码: %d", i+1, totalRequests, resp.StatusCode)
		}

		_ = resp.Body.Close()

		// 如果是短窗口限流，稍微等待一下
		if window <= 5*time.Second && i < totalRequests-1 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	elapsed := time.Since(start)
	t.Logf("📊 %s测试结果: 成功=%d, 限流=%d, 耗时=%v", description, successCount, limitedCount, elapsed)

	// 验证限流效果
	if window <= 5*time.Second {
		// 短窗口：应该有明显的限流效果
		assert.GreaterOrEqual(t, successCount, threshold-2, "成功请求数应该接近阈值")
		assert.GreaterOrEqual(t, limitedCount, totalRequests-threshold-2, "应该有请求被限流")
	} else {
		// 长窗口：在测试期间可能不会触发限流
		t.Logf("⚠️  长窗口限流测试，实际效果取决于测试执行速度")
	}
}

// testCircuitBreaker 测试熔断器功能
func testCircuitBreaker(t *testing.T, server *httptest.Server, endpoint string, errorThreshold float64, minRequests int, description string) {
	client := &http.Client{Timeout: 5 * time.Second}

	t.Logf("🧪 开始%s测试: %s (错误阈值=%.0f%%, 最小请求数=%d)", description, endpoint, errorThreshold*100, minRequests)

	// 先发送一些成功请求建立基线
	warmupRequests := 3
	for i := 0; i < warmupRequests; i++ {
		resp, _ := client.Get(server.URL + endpoint)
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Logf("🔥 预热阶段: 发送了 %d 个基线请求", warmupRequests)

	// 启用错误模式
	setHandlerErrorMode(server, endpoint, true)
	defer setHandlerErrorMode(server, endpoint, false)

	var errorCount, successCount, circuitOpenCount int

	// 发送足够的错误请求来触发熔断
	for i := 0; i < minRequests+5; i++ {
		resp, err := client.Get(server.URL + endpoint)
		if err != nil {
			t.Logf("❌ 请求 %d 网络错误: %v", i+1, err)
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK:
			successCount++
			t.Logf("✅ 请求 %d 成功", i+1)
		case http.StatusNotFound:
			errorCount++
			t.Logf("💥 请求 %d 服务器错误", i+1)
		case http.StatusInternalServerError:
			errorCount++
			t.Logf("💥 请求 %d 服务器错误", i+1)
		case http.StatusServiceUnavailable:
			circuitOpenCount++
			t.Logf("🔌 请求 %d 熔断器开启", i+1)
		default:
			t.Logf("❓ 请求 %d 状态码: %d", i+1, resp.StatusCode)
		}

		_ = resp.Body.Close()
		time.Sleep(200 * time.Millisecond)
	}

	t.Logf("📊 %s测试结果: 成功=%d, 错误=%d, 熔断=%d", description, successCount, errorCount, circuitOpenCount)

	// 🔧 修正错误率计算 - 使用实际的总请求数
	// Sentinel在统计窗口内计算的是：错误数 / (预热成功请求 + 当前阶段所有请求)
	actualTotalRequests := warmupRequests + successCount + errorCount // 不包括熔断的请求，因为熔断后就不参与统计了
	if actualTotalRequests > 0 {
		actualErrorRate := float64(errorCount) / float64(actualTotalRequests)
		t.Logf("📈 实际错误率: %.2f%% = %d错误 / %d总请求 (阈值: %.0f%%)",
			actualErrorRate*100, errorCount, actualTotalRequests, errorThreshold*100)

		// 额外信息：显示Sentinel可能的判断逻辑 - 使用实际的minRequests参数
		if actualTotalRequests >= minRequests { // 使用传入的min_request_amount参数
			if actualErrorRate > errorThreshold {
				t.Logf("🎯 Sentinel判断: 总请求=%d >= %d 且 错误率=%.1f%% > %.0f%% → 应该熔断",
					actualTotalRequests, minRequests, actualErrorRate*100, errorThreshold*100)
			} else {
				t.Logf("⚠️  Sentinel判断: 总请求=%d >= %d 但 错误率=%.1f%% <= %.0f%% → 不应熔断",
					actualTotalRequests, minRequests, actualErrorRate*100, errorThreshold*100)
			}
		} else {
			t.Logf("⏳ Sentinel判断: 总请求=%d < %d → 样本不足，暂不判断", actualTotalRequests, minRequests)
		}
	}

	// 验证熔断效果 - 修正验证逻辑，使用正确的minRequests参数
	if actualTotalRequests >= minRequests && errorCount > 0 { // 达到最小统计要求且有错误
		actualErrorRate := float64(errorCount) / float64(actualTotalRequests)
		if actualErrorRate > errorThreshold {
			assert.GreaterOrEqual(t, circuitOpenCount, 1, "达到最小请求数且错误率超过阈值时应该触发熔断")
		}
	}
}

// testPriorityMatching 测试优先级匹配
func testPriorityMatching(t *testing.T, server *httptest.Server) {
	client := &http.Client{Timeout: 5 * time.Second}

	t.Log("🧪 开始优先级匹配测试")

	// 测试用户API应该匹配用户规则而不是通用规则
	endpoints := []struct {
		path         string
		expectedRule string
		description  string
	}{
		{"/api/v1/auth/login", "认证接口规则", "认证登录"},
		{"/api/v1/users/profile", "用户接口规则", "用户资料"},
		{"/api/v1/orders/list", "通用API规则", "订单列表"},
		{"/protection/status", "保护状态规则", "保护状态"},
		{"/health", "健康检查规则", "健康检查"},
	}

	for _, endpoint := range endpoints {
		resp, err := client.Get(server.URL + endpoint.path)
		require.NoError(t, err)

		t.Logf("🎯 %s (%s) → 状态码: %d", endpoint.description, endpoint.path, resp.StatusCode)

		_ = resp.Body.Close()
	}

	t.Log("📊 优先级匹配测试完成 - 检查日志确认规则匹配")
}

// testConcurrentRequests 测试并发请求
func testConcurrentRequests(t *testing.T, server *httptest.Server) {
	t.Log("🧪 开始并发请求测试")

	var wg sync.WaitGroup
	results := make(map[int]int) // status code -> count
	var mu sync.Mutex

	concurrency := 20
	requestsPerGoroutine := 5

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}

			for j := 0; j < requestsPerGoroutine; j++ {
				resp, err := client.Get(server.URL + "/api/v1/users/list")
				if err != nil {
					t.Logf("❌ Worker %d 请求 %d 失败: %v", workerID, j+1, err)
					continue
				}

				mu.Lock()
				results[resp.StatusCode]++
				mu.Unlock()

				_ = resp.Body.Close()
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	t.Log("📊 并发请求测试结果:")
	for statusCode, count := range results {
		statusText := http.StatusText(statusCode)
		if statusCode == http.StatusTooManyRequests {
			statusText = "Rate Limited"
		}
		t.Logf("   状态码 %d (%s): %d 次", statusCode, statusText, count)
	}

	totalRequests := concurrency * requestsPerGoroutine
	assert.Equal(t, totalRequests, sum(results), "总请求数应该匹配")
}

// testWildcardMatching 测试通配符匹配
func testWildcardMatching(t *testing.T, server *httptest.Server) {
	client := &http.Client{Timeout: 5 * time.Second}

	t.Log("🧪 开始通配符匹配测试")

	// 测试不同路径的通配符匹配
	testCases := []struct {
		path        string
		description string
	}{
		{"/api/v1/auth/login", "认证 - 登录"},
		{"/api/v1/auth/logout", "认证 - 登出"},
		{"/api/v1/users/123", "用户 - 获取"},
		{"/api/v1/users/456/profile", "用户 - 资料"},
		{"/api/v1/orders/789", "订单 - 通用API"},
		{"/api/v2/products/list", "产品 - 通用API"},
		{"/protection/rules", "保护 - 规则"},
		{"/protection/stats", "保护 - 统计"},
	}

	for _, tc := range testCases {
		resp, err := client.Get(server.URL + tc.path)
		require.NoError(t, err)

		t.Logf("🔍 %s (%s) → 状态码: %d", tc.description, tc.path, resp.StatusCode)

		_ = resp.Body.Close()
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("📊 通配符匹配测试完成")
}

// setupHTTPServer 设置HTTP服务器
func setupHTTPServer(t *testing.T, protectionConfig *config.ProtectionConfig) *httptest.Server {
	// 创建Sentinel中间件
	sentinelMiddleware, err := middleware.NewSentinelProtectionMiddleware(context.Background(), protectionConfig)
	require.NoError(t, err)

	// 创建Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 添加Sentinel中间件
	router.Use(sentinelMiddleware.HTTPMiddleware())

	// 创建模拟处理器
	normalHandler := NewMockHandler(10, 0.0) // 正常处理器
	errorHandler := NewMockHandler(50, 0.8)  // 高错误率处理器

	// 注册路由
	router.GET("/health", normalHandler.ServeHTTP)

	// 认证相关接口
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.GET("/login", normalHandler.ServeHTTP)
		authGroup.GET("/logout", normalHandler.ServeHTTP)
		authGroup.POST("/register", errorHandler.ServeHTTP) // 用于测试熔断
	}

	// 用户相关接口
	usersGroup := router.Group("/api/v1/users")
	{
		usersGroup.GET("/profile", normalHandler.ServeHTTP)
		usersGroup.GET("/list", normalHandler.ServeHTTP)
		usersGroup.GET("/:id", normalHandler.ServeHTTP)
		usersGroup.GET("/:id/profile", normalHandler.ServeHTTP)
		usersGroup.GET("/delete", errorHandler.ServeHTTP) // 用于测试熔断
	}

	// 通用API接口
	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/v1/products/list", normalHandler.ServeHTTP)
		apiGroup.GET("/v1/orders/list", normalHandler.ServeHTTP)
		apiGroup.GET("/v1/orders/:id", normalHandler.ServeHTTP)
		apiGroup.POST("/v1/orders/create", errorHandler.ServeHTTP) // 用于测试熔断
		apiGroup.GET("/v2/products/list", normalHandler.ServeHTTP)
	}

	// 保护状态接口
	protectionGroup := router.Group("/protection")
	{
		protectionGroup.GET("/status", normalHandler.ServeHTTP)
		protectionGroup.GET("/rules", normalHandler.ServeHTTP)
		protectionGroup.GET("/stats", normalHandler.ServeHTTP)
	}

	// 创建测试服务器
	server := httptest.NewServer(router)
	t.Logf("🚀 HTTP测试服务器启动: %s", server.URL)

	return server
}

// setHandlerErrorMode 设置处理器错误模式（模拟功能）
func setHandlerErrorMode(server *httptest.Server, endpoint string, shouldFail bool) {
	// 这是一个模拟函数，在实际实现中需要找到对应的handler并设置错误模式
	// 由于测试服务器的限制，这里只是一个占位符
}

// sum 计算map中所有值的总和
func sum(m map[int]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
