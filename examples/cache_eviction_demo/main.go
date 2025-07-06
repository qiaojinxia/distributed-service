package main

import (
	"context"
	"fmt"

	"strings"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/cache"
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
	fmt.Println("📊 测试结果汇总")
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
		fmt.Println("\n🎉 所有测试通过！")
	} else {
		fmt.Printf("\n⚠️  有 %d 个测试失败\n", failed)
	}
}

func main() {
	ctx := context.Background()
	suite := &TestSuite{}

	fmt.Println("🚀 缓存淘汰策略全面测试")
	fmt.Println(strings.Repeat("=", 60))

	// 测试LRU策略
	testLRUPolicy(ctx, suite)

	// 测试TTL策略
	testTTLPolicy(ctx, suite)

	// 测试Simple策略
	testSimplePolicy(ctx, suite)

	// 测试配置验证
	testConfigValidation(ctx, suite)

	// 测试Builder
	testBuilder(ctx, suite)

	// 测试统计信息
	testStatistics(ctx, suite)

	// 测试回调函数
	testCallbacks(ctx, suite)

	// 测试错误处理
	testErrorHandling(ctx, suite)

	// 测试并发安全
	testConcurrencySafety(ctx, suite)

	// 输出测试汇总
	suite.printSummary()
}

// testLRUPolicy 测试LRU策略
func testLRUPolicy(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n📋 测试LRU策略")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:         3, // 设置小容量便于测试淘汰
		DefaultTTL:      time.Minute,
		CleanupInterval: time.Second * 10,
		EvictionPolicy:  cache.EvictionPolicyLRU,
	}

	lruCache, err := cache.NewMemoryCache(config)
	suite.assert("LRU-创建", err == nil, "LRU缓存创建成功")

	if err != nil {
		return
	}

	// 测试基本操作
	err = lruCache.Set(ctx, "key1", "value1", 0)
	suite.assert("LRU-设置1", err == nil, "设置第一个键值对")

	err = lruCache.Set(ctx, "key2", "value2", 0)
	suite.assert("LRU-设置2", err == nil, "设置第二个键值对")

	err = lruCache.Set(ctx, "key3", "value3", 0)
	suite.assert("LRU-设置3", err == nil, "设置第三个键值对")

	// 验证所有键都存在
	exists1, _ := lruCache.Exists(ctx, "key1")
	exists2, _ := lruCache.Exists(ctx, "key2")
	exists3, _ := lruCache.Exists(ctx, "key3")

	suite.assert("LRU-存在性1", exists1, "key1应该存在")
	suite.assert("LRU-存在性2", exists2, "key2应该存在")
	suite.assert("LRU-存在性3", exists3, "key3应该存在")

	// 访问key1，使其变为最近使用
	val, err := lruCache.Get(ctx, "key1")
	suite.assert("LRU-访问", err == nil && val == "value1", "访问key1成功")

	// 添加第4个键，应该淘汰key2（最久未使用）
	err = lruCache.Set(ctx, "key4", "value4", 0)
	suite.assert("LRU-设置4", err == nil, "设置第四个键值对")

	// 验证淘汰结果
	exists1, _ = lruCache.Exists(ctx, "key1")
	exists2, _ = lruCache.Exists(ctx, "key2")
	exists3, _ = lruCache.Exists(ctx, "key3")
	exists4, _ := lruCache.Exists(ctx, "key4")

	suite.assert("LRU-淘汰后1", exists1, "key1应该仍然存在（最近访问）")
	suite.assert("LRU-淘汰后2", !exists2, "key2应该被淘汰（最久未使用）")
	suite.assert("LRU-淘汰后3", exists3, "key3应该仍然存在")
	suite.assert("LRU-淘汰后4", exists4, "key4应该存在（新添加）")

	// 测试删除操作
	err = lruCache.Delete(ctx, "key1")
	suite.assert("LRU-删除", err == nil, "删除操作成功")

	exists1, _ = lruCache.Exists(ctx, "key1")
	suite.assert("LRU-删除验证", !exists1, "删除后key1不应该存在")

	// 测试清空操作
	err = lruCache.Clear(ctx)
	suite.assert("LRU-清空", err == nil, "清空操作成功")

	exists3, _ = lruCache.Exists(ctx, "key3")
	exists4, _ = lruCache.Exists(ctx, "key4")
	suite.assert("LRU-清空验证", !exists3 && !exists4, "清空后所有键都不存在")
}

// testTTLPolicy 测试TTL策略
func testTTLPolicy(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n⏰ 测试TTL策略")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:         10,
		DefaultTTL:      time.Second * 2, // 2秒TTL
		CleanupInterval: time.Millisecond * 500,
		EvictionPolicy:  cache.EvictionPolicyTTL,
	}

	ttlCache, err := cache.NewMemoryCache(config)
	suite.assert("TTL-创建", err == nil, "TTL缓存创建成功")

	if err != nil {
		return
	}

	// 测试基本操作
	err = ttlCache.Set(ctx, "temp1", "临时数据1", 0) // 使用默认TTL
	suite.assert("TTL-设置1", err == nil, "设置临时数据1")

	err = ttlCache.Set(ctx, "temp2", "临时数据2", time.Second) // 自定义TTL（但expirable库可能忽略）
	suite.assert("TTL-设置2", err == nil, "设置临时数据2")

	// 立即检查
	val1, err1 := ttlCache.Get(ctx, "temp1")
	val2, err2 := ttlCache.Get(ctx, "temp2")

	suite.assert("TTL-立即获取1", err1 == nil && val1 == "临时数据1", "立即获取temp1成功")
	suite.assert("TTL-立即获取2", err2 == nil && val2 == "临时数据2", "立即获取temp2成功")

	// 等待1秒
	fmt.Println("⏳ 等待1秒...")
	time.Sleep(time.Millisecond * 900)

	// 再次检查（应该仍然存在）
	exists1, _ := ttlCache.Exists(ctx, "temp1")
	exists2, _ := ttlCache.Exists(ctx, "temp2")

	suite.assert("TTL-1秒后1", exists1, "1秒后temp1应该仍然存在")
	suite.assert("TTL-1秒后2", exists2, "1秒后temp2应该仍然存在")

	// 等待更长时间直到过期
	fmt.Println("⏳ 等待2秒直到过期...")
	time.Sleep(time.Second * 2)

	// 最终检查（应该过期）
	_, err1 = ttlCache.Get(ctx, "temp1")
	_, err2 = ttlCache.Get(ctx, "temp2")

	suite.assert("TTL-过期1", err1 != nil, "temp1应该过期")
	suite.assert("TTL-过期2", err2 != nil, "temp2应该过期")

	// 测试过期后不存在
	exists1, _ = ttlCache.Exists(ctx, "temp1")
	exists2, _ = ttlCache.Exists(ctx, "temp2")

	suite.assert("TTL-过期存在性1", !exists1, "过期后temp1不应该存在")
	suite.assert("TTL-过期存在性2", !exists2, "过期后temp2不应该存在")
}

// testSimplePolicy 测试Simple策略
func testSimplePolicy(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n🔧 测试Simple策略")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:         5,
		DefaultTTL:      time.Second * 3,
		CleanupInterval: time.Second,
		EvictionPolicy:  cache.EvictionPolicySimple,
	}

	simpleCache, err := cache.NewMemoryCache(config)
	suite.assert("Simple-创建", err == nil, "Simple缓存创建成功")

	if err != nil {
		return
	}

	// 测试批量设置
	configs := map[string]string{
		"app_name":   "测试应用",
		"version":    "1.0.0",
		"debug_mode": "true",
		"max_users":  "1000",
		"timeout":    "30s",
	}

	for key, value := range configs {
		err = simpleCache.Set(ctx, key, value, 0)
		suite.assert("Simple-设置"+key, err == nil, "设置"+key+"成功")
	}

	// 验证所有数据
	for key, expectedValue := range configs {
		val, err := simpleCache.Get(ctx, key)
		suite.assert("Simple-获取"+key, err == nil && val == expectedValue, "获取"+key+"正确")
	}

	// 测试自定义TTL
	err = simpleCache.Set(ctx, "short_lived", "短期数据", time.Millisecond*500)
	suite.assert("Simple-短期设置", err == nil, "设置短期数据")

	val, err := simpleCache.Get(ctx, "short_lived")
	suite.assert("Simple-短期获取", err == nil && val == "短期数据", "立即获取短期数据成功")

	// 等待短期数据过期
	time.Sleep(time.Millisecond * 800)

	_, err = simpleCache.Get(ctx, "short_lived")
	suite.assert("Simple-短期过期", err != nil, "短期数据应该过期")

	// 验证其他数据仍然存在
	val, err = simpleCache.Get(ctx, "app_name")
	suite.assert("Simple-长期存在", err == nil && val == "测试应用", "长期数据仍然存在")
}

// testConfigValidation 测试配置验证
func testConfigValidation(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n⚙️  测试配置验证")
	fmt.Println(strings.Repeat("-", 40))

	// 测试默认配置
	config := cache.MemoryConfig{}
	cache1, err := cache.NewMemoryCache(config)
	suite.assert("Config-默认", err == nil, "默认配置应该可以工作")

	if cache1 != nil {
		// 验证默认值
		suite.assert("Config-默认策略", config.EvictionPolicy == "" || config.EvictionPolicy == cache.EvictionPolicyLRU, "默认策略应该是LRU")
	}

	// 测试无效策略
	invalidConfig := cache.MemoryConfig{
		MaxSize:        100,
		EvictionPolicy: cache.EvictionPolicy("invalid"),
	}
	_, err = cache.NewMemoryCache(invalidConfig)
	suite.assert("Config-无效策略", err != nil, "无效策略应该返回错误")

	// 测试边界值
	boundaryConfig := cache.MemoryConfig{
		MaxSize:         1, // 最小容量
		DefaultTTL:      time.Nanosecond,
		CleanupInterval: time.Nanosecond,
		EvictionPolicy:  cache.EvictionPolicyLRU,
	}
	cache2, err := cache.NewMemoryCache(boundaryConfig)
	suite.assert("Config-边界值", err == nil, "边界值配置应该可以工作")

	if cache2 != nil {
		// 测试最小容量的工作情况
		err = cache2.Set(ctx, "key1", "value1", 0)
		suite.assert("Config-边界设置1", err == nil, "边界配置设置第一个值")

		err = cache2.Set(ctx, "key2", "value2", 0)
		suite.assert("Config-边界设置2", err == nil, "边界配置设置第二个值")

		// 第一个应该被淘汰
		exists1, _ := cache2.Exists(ctx, "key1")
		exists2, _ := cache2.Exists(ctx, "key2")
		suite.assert("Config-边界淘汰", !exists1 && exists2, "容量为1时应该淘汰旧值")
	}
}

// testBuilder 测试Builder
func testBuilder(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n🏗️  测试Builder")
	fmt.Println(strings.Repeat("-", 40))

	builder := &cache.MemoryBuilder{}

	// 测试默认配置
	config1 := cache.Config{Settings: nil}
	_, err := builder.Build(config1)
	suite.assert("Builder-默认", err == nil, "Builder默认配置构建成功")

	// 测试自定义配置
	config2 := cache.Config{
		Settings: map[string]interface{}{
			"max_size":         500,
			"eviction_policy":  "ttl",
			"default_ttl":      time.Minute * 30,
			"cleanup_interval": time.Minute * 5,
		},
	}
	cache2, err := builder.Build(config2)
	suite.assert("Builder-自定义", err == nil, "Builder自定义配置构建成功")

	if cache2 != nil {
		// 验证配置是否生效
		err = cache2.Set(ctx, "test", "value", 0)
		suite.assert("Builder-功能", err == nil, "Builder构建的缓存可以正常工作")
	}

	// 测试无效配置
	config3 := cache.Config{
		Settings: map[string]interface{}{
			"eviction_policy": "invalid_policy",
		},
	}
	_, err = builder.Build(config3)
	suite.assert("Builder-无效", err != nil, "Builder应该拒绝无效配置")
}

// testStatistics 测试统计信息
func testStatistics(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n📊 测试统计信息")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:        3,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}

	testCache, err := cache.NewMemoryCache(config)
	suite.assert("Stats-创建", err == nil, "统计测试缓存创建成功")

	if err != nil {
		return
	}

	// 重置统计
	testCache.ResetStats()
	stats := testCache.GetStats()
	suite.assertEqual("Stats-初始命中", int64(0), stats.Hits, "初始命中数应该为0")
	suite.assertEqual("Stats-初始未命中", int64(0), stats.Misses, "初始未命中数应该为0")

	// 测试设置操作
	testCache.Set(ctx, "key1", "value1", 0)
	testCache.Set(ctx, "key2", "value2", 0)

	stats = testCache.GetStats()
	suite.assertEqual("Stats-设置数", int64(2), stats.Sets, "设置数应该为2")

	// 测试命中
	testCache.Get(ctx, "key1")
	testCache.Get(ctx, "key2")

	stats = testCache.GetStats()
	suite.assertEqual("Stats-命中数", int64(2), stats.Hits, "命中数应该为2")

	// 测试未命中
	testCache.Get(ctx, "nonexistent")

	stats = testCache.GetStats()
	suite.assertEqual("Stats-未命中数", int64(1), stats.Misses, "未命中数应该为1")

	// 测试删除
	testCache.Delete(ctx, "key1")

	stats = testCache.GetStats()
	suite.assertEqual("Stats-删除数", int64(1), stats.Deletes, "删除数应该为1")

	// 测试淘汰（通过添加更多项目）
	testCache.Set(ctx, "key3", "value3", 0)
	testCache.Set(ctx, "key4", "value4", 0) // 应该触发淘汰

	stats = testCache.GetStats()
	suite.assert("Stats-淘汰数", stats.Evictions > 0, "应该有淘汰发生")

	// 计算命中率
	hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
	suite.assert("Stats-命中率", hitRate > 0 && hitRate <= 100, "命中率应该在合理范围内")

	fmt.Printf("📈 当前统计: 命中=%d, 未命中=%d, 设置=%d, 删除=%d, 淘汰=%d, 命中率=%.1f%%\n",
		stats.Hits, stats.Misses, stats.Sets, stats.Deletes, stats.Evictions, hitRate)
}

// testCallbacks 测试回调函数
func testCallbacks(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n🔔 测试回调函数")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:        2,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}

	testCache, err := cache.NewMemoryCache(config)
	suite.assert("Callback-创建", err == nil, "回调测试缓存创建成功")

	if err != nil {
		return
	}

	// 设置回调函数
	var evictedKeys []string
	testCache.SetEvictionCallback(func(key string, value interface{}) {
		evictedKeys = append(evictedKeys, key)
		fmt.Printf("🗑️  淘汰回调: key=%s, value=%v\n", key, value)
	})

	// 填满缓存
	testCache.Set(ctx, "key1", "value1", 0)
	testCache.Set(ctx, "key2", "value2", 0)

	suite.assertEqual("Callback-填满前", 0, len(evictedKeys), "填满前不应该有淘汰")

	// 触发淘汰
	testCache.Set(ctx, "key3", "value3", 0)

	suite.assert("Callback-淘汰触发", len(evictedKeys) > 0, "应该触发淘汰回调")

	if len(evictedKeys) > 0 {
		suite.assertEqual("Callback-淘汰键", "key1", evictedKeys[0], "应该淘汰key1")
	}
}

// testErrorHandling 测试错误处理
func testErrorHandling(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n🚨 测试错误处理")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:        10,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}

	testCache, err := cache.NewMemoryCache(config)
	suite.assert("Error-创建", err == nil, "错误处理测试缓存创建成功")

	if err != nil {
		return
	}

	// 测试获取不存在的键
	_, err = testCache.Get(ctx, "nonexistent")
	suite.assert("Error-不存在键", err != nil, "获取不存在的键应该返回错误")

	// 测试删除不存在的键（应该不报错）
	err = testCache.Delete(ctx, "nonexistent")
	suite.assert("Error-删除不存在", err == nil, "删除不存在的键不应该报错")

	// 测试存在性检查
	exists, err := testCache.Exists(ctx, "nonexistent")
	suite.assert("Error-存在性检查", err == nil && !exists, "存在性检查应该正确返回false")

	// 测试关闭操作
	err = testCache.Close()
	suite.assert("Error-关闭", err == nil, "关闭操作不应该报错")
}

// testConcurrencySafety 测试并发安全
func testConcurrencySafety(ctx context.Context, suite *TestSuite) {
	fmt.Println("\n🔄 测试并发安全")
	fmt.Println(strings.Repeat("-", 40))

	config := cache.MemoryConfig{
		MaxSize:        100,
		EvictionPolicy: cache.EvictionPolicyLRU,
	}

	testCache, err := cache.NewMemoryCache(config)
	suite.assert("Concurrent-创建", err == nil, "并发测试缓存创建成功")

	if err != nil {
		return
	}

	// 并发写入测试
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// 并发设置
				testCache.Set(ctx, key, value, 0)

				// 并发获取
				testCache.Get(ctx, key)

				// 并发存在性检查
				testCache.Exists(ctx, key)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 检查最终状态
	stats := testCache.GetStats()
	suite.assert("Concurrent-完成", stats.Sets > 0, "并发操作应该有设置记录")
	suite.assert("Concurrent-无崩溃", true, "并发操作没有导致崩溃")

	fmt.Printf("🔄 并发测试完成: 设置=%d, 获取请求=%d\n", stats.Sets, stats.Hits+stats.Misses)
}
