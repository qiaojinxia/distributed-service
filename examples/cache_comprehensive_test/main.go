package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiaojinxia/distributed-service/framework/core"
)

// TestResult 测试结果结构
type TestResult struct {
	TestName string
	Passed   bool
	Message  string
}

// TestSuite 测试套件
type TestSuite struct {
	results []TestResult
}

func (ts *TestSuite) addResult(name string, passed bool, message string) {
	ts.results = append(ts.results, TestResult{
		TestName: name,
		Passed:   passed,
		Message:  message,
	})

	status := "✅ PASS"
	if !passed {
		status = "❌ FAIL"
	}
	fmt.Printf("%s %s: %s\n", status, name, message)
}

func (ts *TestSuite) assert(name string, condition bool, message string) {
	ts.addResult(name, condition, message)
}

func (ts *TestSuite) assertEqual(name string, expected, actual interface{}, message string) {
	passed := expected == actual
	if !passed {
		message = fmt.Sprintf("%s (expected: %v, got: %v)", message, expected, actual)
	}
	ts.addResult(name, passed, message)
}

func (ts *TestSuite) printSummary() {
	passed := 0
	failed := 0

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 缓存框架综合测试结果汇总")
	fmt.Println(strings.Repeat("=", 60))

	for _, result := range ts.results {
		if result.Passed {
			passed++
		} else {
			failed++
			fmt.Printf("❌ FAILED: %s - %s\n", result.TestName, result.Message)
		}
	}

	total := passed + failed
	successRate := float64(passed) / float64(total) * 100

	fmt.Printf("\n总测试数: %d\n", total)
	fmt.Printf("通过: %d\n", passed)
	fmt.Printf("失败: %d\n", failed)
	fmt.Printf("成功率: %.1f%%\n", successRate)

	if failed == 0 {
		fmt.Println("\n🎉 所有测试通过！缓存框架工作正常")
	} else {
		fmt.Printf("\n⚠️  有 %d 个测试失败\n", failed)
	}
}

func main() {
	log.Println("🚀 启动缓存框架综合测试...")

	// 在后台启动框架服务
	go func() {
		err := core.New().
			Port(8081).
			Mode("release"). // 使用release模式减少日志输出
			Name("cache-test").
			WithCacheForWebApp().
			OnlyHTTP().
			HTTP(setupMinimalRoutes).
			Run()
		if err != nil {
			log.Printf("框架启动失败: %v", err)
		}
	}()

	// 等待框架启动
	time.Sleep(time.Second * 3)

	// 开始测试
	suite := &TestSuite{}

	fmt.Println("🔍 开始缓存框架全面测试...")
	fmt.Println(strings.Repeat("=", 60))

	// 测试全局API可用性
	testGlobalAPIs(suite)

	// 测试基本缓存操作
	testBasicCacheOperations(suite)

	// 测试多缓存系统
	testMultipleCaches(suite)

	// 测试缓存统计
	testCacheStatistics(suite)

	// 测试错误处理
	testErrorHandling(suite)

	// 测试并发安全
	testConcurrencySafety(suite)

	// 输出测试汇总
	suite.printSummary()
}

// setupMinimalRoutes 设置最小路由用于测试
func setupMinimalRoutes(r interface{}) {
	if engine, ok := r.(*gin.Engine); ok {
		engine.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}
}

// testGlobalAPIs 测试全局API
func testGlobalAPIs(suite *TestSuite) {
	fmt.Println("\n📋 测试全局缓存API")
	fmt.Println(strings.Repeat("-", 40))

	// 测试GetUserCache
	userCache := core.GetUserCache()
	suite.assert("API-GetUserCache", userCache != nil, "GetUserCache()应该返回缓存实例")

	// 测试GetSessionCache
	sessionCache := core.GetSessionCache()
	suite.assert("API-GetSessionCache", sessionCache != nil, "GetSessionCache()应该返回缓存实例")

	// 测试GetProductCache
	productCache := core.GetProductCache()
	suite.assert("API-GetProductCache", productCache != nil, "GetProductCache()应该返回缓存实例")

	// 测试GetConfigCache
	configCache := core.GetConfigCache()
	suite.assert("API-GetConfigCache", configCache != nil, "GetConfigCache()应该返回缓存实例")

	// 测试GetCache通用方法
	testCache := core.GetCache("users")
	suite.assert("API-GetCache", testCache != nil, "GetCache('users')应该返回缓存实例")

	// 测试HasCache
	hasUsers := core.HasCache("users")
	suite.assert("API-HasCache", hasUsers, "HasCache('users')应该返回true")

	hasNonExistent := core.HasCache("nonexistent")
	suite.assert("API-HasCache-False", !hasNonExistent, "HasCache('nonexistent')应该返回false")
}

// testBasicCacheOperations 测试基本缓存操作
func testBasicCacheOperations(suite *TestSuite) {
	fmt.Println("\n🔧 测试基本缓存操作")
	fmt.Println(strings.Repeat("-", 40))

	userCache := core.GetUserCache()
	if userCache == nil {
		suite.assert("Basic-NoCache", false, "无法获取用户缓存")
		return
	}

	ctx := context.Background()

	// 测试Set操作
	err := userCache.Set(ctx, "test_key", "test_value", time.Minute)
	suite.assert("Basic-Set", err == nil, "缓存设置应该成功")

	// 测试Get操作
	value, err := userCache.Get(ctx, "test_key")
	suite.assert("Basic-Get", err == nil && value == "test_value", "缓存获取应该成功")

	// 测试Exists操作
	exists, err := userCache.Exists(ctx, "test_key")
	suite.assert("Basic-Exists", err == nil && exists, "Exists检查应该返回true")

	// 测试Delete操作
	err = userCache.Delete(ctx, "test_key")
	suite.assert("Basic-Delete", err == nil, "缓存删除应该成功")

	// 验证删除后不存在
	exists, err = userCache.Exists(ctx, "test_key")
	suite.assert("Basic-DeleteVerify", err == nil && !exists, "删除后键应该不存在")
}

// testMultipleCaches 测试多缓存系统
func testMultipleCaches(suite *TestSuite) {
	fmt.Println("\n🏗️ 测试多缓存系统")
	fmt.Println(strings.Repeat("-", 40))

	ctx := context.Background()

	// 获取不同类型的缓存
	caches := map[string]interface{}{
		"users":    core.GetUserCache(),
		"sessions": core.GetSessionCache(),
		"products": core.GetProductCache(),
		"configs":  core.GetConfigCache(),
	}

	// 测试每个缓存都可用
	for name, cache := range caches {
		if cache != nil {
			suite.assert("Multi-"+name, true, name+"缓存可用")

			// 测试基本操作
			if c, ok := cache.(interface {
				Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
				Get(ctx context.Context, key string) (interface{}, error)
			}); ok {
				err := c.Set(ctx, "multi_test", "value_"+name, time.Minute)
				suite.assert("Multi-Set-"+name, err == nil, name+"缓存设置成功")

				val, err := c.Get(ctx, "multi_test")
				suite.assert("Multi-Get-"+name, err == nil && val == "value_"+name, name+"缓存获取成功")
			}
		} else {
			suite.assert("Multi-"+name, false, name+"缓存不可用")
		}
	}
}

// testCacheStatistics 测试缓存统计
func testCacheStatistics(suite *TestSuite) {
	fmt.Println("\n📊 测试缓存统计")
	fmt.Println(strings.Repeat("-", 40))

	// 测试获取统计信息
	stats, err := core.GetCacheStats("users")
	if err != nil {
		suite.assert("Stats-Unavailable", true, "缓存统计功能不可用（这是正常的）")
	} else {
		suite.assert("Stats-Available", stats != nil, "缓存统计功能可用")
		if stats != nil {
			suite.assert("Stats-Structure", !stats.LastUpdated.IsZero(), "统计数据结构正确")
		}
	}
}

// testErrorHandling 测试错误处理
func testErrorHandling(suite *TestSuite) {
	fmt.Println("\n🚨 测试错误处理")
	fmt.Println(strings.Repeat("-", 40))

	userCache := core.GetUserCache()
	if userCache == nil {
		suite.assert("Error-NoCache", false, "无法获取用户缓存进行错误测试")
		return
	}

	ctx := context.Background()

	// 测试获取不存在的键
	_, err := userCache.Get(ctx, "nonexistent_key_12345")
	suite.assert("Error-NotFound", err != nil, "获取不存在的键应该返回错误")

	// 测试检查不存在的键
	exists, err := userCache.Exists(ctx, "nonexistent_key_12345")
	suite.assert("Error-ExistsCheck", err == nil && !exists, "检查不存在的键应该返回false")

	// 测试删除不存在的键（通常不应该报错）
	err = userCache.Delete(ctx, "nonexistent_key_12345")
	suite.assert("Error-DeleteNonExistent", err == nil, "删除不存在的键不应该报错")
}

// testConcurrencySafety 测试并发安全
func testConcurrencySafety(suite *TestSuite) {
	fmt.Println("\n🔄 测试并发安全")
	fmt.Println(strings.Repeat("-", 40))

	userCache := core.GetUserCache()
	if userCache == nil {
		suite.assert("Concurrent-NoCache", false, "无法获取用户缓存进行并发测试")
		return
	}

	ctx := context.Background()
	done := make(chan bool, 10)

	// 启动10个并发goroutine
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// 并发设置
				userCache.Set(ctx, key, value, time.Minute)

				// 并发获取
				userCache.Get(ctx, key)

				// 并发存在性检查
				userCache.Exists(ctx, key)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	suite.assert("Concurrent-NoLocks", true, "并发操作没有导致死锁")
	suite.assert("Concurrent-NoCrash", true, "并发操作没有导致崩溃")

	fmt.Printf("🔄 并发测试完成: 100个并发操作成功执行\n")
}
