package main

import (
	"context"
	"fmt"
	"github.com/qiaojinxia/distributed-service/framework/core"
	"log"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/common/idgen"
)

func main() {
	fmt.Println("🚀 基于GORM的美团Leaf分布式ID生成器演示")
	fmt.Println("============================================")

	// 演示基本功能
	demoBasicUsage()

	// 演示配置构建器
	demoConfigBuilder()

	// 演示性能测试
	demoPerformanceTest()

	// 演示监控和指标
	demoMetricsAndMonitoring()
}

func demoBasicUsage() {
	fmt.Println("\n📖 基本用法演示")
	fmt.Println("=================")

	// 使用SQLite作为演示数据库（不需要外部依赖）
	config := idgen.SQLiteConfig("./demo.db")

	// 创建ID生成器
	idGen, err := core.NewIDGenerator(config)
	if err != nil {
		log.Fatalf("创建ID生成器失败: %v", err)
	}

	// 确保实现了正确的接口
	gormGen, ok := idGen.(*idgen.GormLeafIDGenerator)
	if !ok {
		log.Fatalf("类型断言失败")
	}
	defer func(gormGen *idgen.GormLeafIDGenerator) {
		err = gormGen.Close()
	}(gormGen)

	ctx := context.Background()

	// 创建表
	err = gormGen.CreateTable(ctx)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("✅ 数据库表创建成功")

	// 创建业务标识
	err = gormGen.CreateBizTag(ctx, "user", 1000, "用户ID生成器")
	if err != nil {
		fmt.Printf("⚠️  创建用户业务标识失败（可能已存在）: %v\n", err)
	} else {
		fmt.Println("✅ 用户业务标识创建成功")
	}

	err = gormGen.CreateBizTag(ctx, "order", 2000, "订单ID生成器")
	if err != nil {
		fmt.Printf("⚠️  创建订单业务标识失败（可能已存在）: %v\n", err)
	} else {
		fmt.Println("✅ 订单业务标识创建成功")
	}

	// 生成ID
	fmt.Println("\n🔢 生成ID演示:")
	for i := 0; i < 5; i++ {
		userID, err := gormGen.NextID(ctx, "user")
		if err != nil {
			log.Printf("生成用户ID失败: %v", err)
			continue
		}

		orderID, err := gormGen.NextID(ctx, "order")
		if err != nil {
			log.Printf("生成订单ID失败: %v", err)
			continue
		}

		fmt.Printf("   用户ID: %d, 订单ID: %d\n", userID, orderID)
	}

	// 批量生成ID
	fmt.Println("\n📦 批量生成ID演示:")
	userIDs, err := gormGen.BatchNextID(ctx, "user", 10)
	if err != nil {
		log.Printf("批量生成用户ID失败: %v", err)
	} else {
		fmt.Printf("   批量用户ID: %v\n", userIDs)
	}
}

func demoConfigBuilder() {
	fmt.Println("\n🔧 配置构建器演示")
	fmt.Println("==================")

	// 使用配置构建器创建自定义配置
	config := idgen.NewConfigBuilder().
		WithSQLite("./custom.db").
		WithLeafConfig(&idgen.LeafConfig{
			DefaultStep:      500,
			PreloadThreshold: 0.8,
			CleanupInterval:  time.Minute * 30,
			MaxStepSize:      50000,
			MinStepSize:      50,
			StepAdjustRatio:  1.5,
		}).
		WithLogLevel("info").
		Build()

	fmt.Printf("🛠️  自定义配置:\n")
	fmt.Printf("   数据库: %s\n", config.Database.Database)
	fmt.Printf("   默认步长: %d\n", config.Leaf.DefaultStep)
	fmt.Printf("   预加载阈值: %.1f%%\n", config.Leaf.PreloadThreshold*100)
	fmt.Printf("   清理间隔: %v\n", config.Leaf.CleanupInterval)

	// MySQL配置示例（注释掉，因为需要MySQL服务）
	/*
		mysqlConfig := idgen.MySQLConfig("localhost", 3306, "test_db", "root", "password")
		fmt.Printf("🐬 MySQL配置示例:\n")
		fmt.Printf("   DSN: %s\n", mysqlConfig.Database.BuildDSN())
	*/

	// PostgreSQL配置示例（注释掉，因为需要PostgreSQL服务）
	/*
		pgConfig := idgen.PostgreSQLConfig("localhost", 5432, "test_db", "postgres", "password")
		fmt.Printf("🐘 PostgreSQL配置示例:\n")
		fmt.Printf("   DSN: %s\n", pgConfig.Database.BuildDSN())
	*/
}

func demoPerformanceTest() {
	fmt.Println("\n⚡ 性能测试演示")
	fmt.Println("================")

	config := idgen.SQLiteConfig("./performance.db")
	idGen, err := core.NewIDGenerator(config)
	if err != nil {
		log.Printf("创建性能测试ID生成器失败: %v", err)
		return
	}

	gormGen := idGen.(*idgen.GormLeafIDGenerator)
	defer func(gormGen *idgen.GormLeafIDGenerator) {
		err = gormGen.Close()
	}(gormGen)

	ctx := context.Background()

	// 创建表
	_ = gormGen.CreateTable(ctx)
	_ = gormGen.CreateBizTag(ctx, "perf_test", 5000, "性能测试")

	// 预热
	for i := 0; i < 100; i++ {
		_, _ = gormGen.NextID(ctx, "perf_test")
	}

	// 性能测试
	count := 10000
	start := time.Now()

	for i := 0; i < count; i++ {
		_, err := gormGen.NextID(ctx, "perf_test")
		if err != nil {
			log.Printf("性能测试生成ID失败: %v", err)
			break
		}
	}

	duration := time.Since(start)
	qps := float64(count) / duration.Seconds()

	fmt.Printf("📊 性能测试结果:\n")
	fmt.Printf("   生成数量: %d\n", count)
	fmt.Printf("   耗时: %v\n", duration)
	fmt.Printf("   QPS: %.0f\n", qps)
	fmt.Printf("   平均延迟: %v\n", duration/time.Duration(count))

	// 获取buffer状态
	status := gormGen.GetBufferStatus("perf_test")
	fmt.Printf("📈 Buffer状态:\n")
	for key, value := range status {
		fmt.Printf("   %s: %v\n", key, value)
	}
}

func demoMetricsAndMonitoring() {
	fmt.Println("\n📊 监控和指标演示")
	fmt.Println("==================")

	config := idgen.SQLiteConfig("./metrics.db")
	idGen, err := core.NewIDGenerator(config)
	if err != nil {
		log.Printf("创建监控测试ID生成器失败: %v", err)
		return
	}

	gormGen := idGen.(*idgen.GormLeafIDGenerator)
	defer func(gormGen *idgen.GormLeafIDGenerator) {
		err = gormGen.Close()
	}(gormGen)

	ctx := context.Background()

	// 创建表和业务标识
	err = gormGen.CreateTable(ctx)
	if err != nil {
		return
	}
	err = gormGen.CreateBizTag(ctx, "metrics_test", 1000, "监控测试")
	if err != nil {
		return
	}

	// 生成一些ID来产生指标数据
	fmt.Println("🔄 生成测试数据...")
	for i := 0; i < 1000; i++ {
		_, _ = gormGen.NextID(ctx, "metrics_test")
		if i%100 == 0 {
			fmt.Printf("   已生成 %d 个ID\n", i+1)
		}
	}

	// 获取指标
	metrics := gormGen.GetMetrics("metrics_test")
	fmt.Printf("\n📈 业务指标 (metrics_test):\n")
	fmt.Printf("   总请求数: %d\n", metrics.TotalRequests)
	fmt.Printf("   成功请求数: %d\n", metrics.SuccessRequests)
	fmt.Printf("   失败请求数: %d\n", metrics.FailedRequests)
	fmt.Printf("   成功率: %.2f%%\n", metrics.SuccessRate()*100)
	fmt.Printf("   号段加载次数: %d\n", metrics.SegmentLoads)
	fmt.Printf("   缓冲区切换次数: %d\n", metrics.BufferSwitches)
	fmt.Printf("   平均QPS: %.2f\n", metrics.AverageQPS)
	fmt.Printf("   最后更新: %v\n", metrics.LastUpdateTime.Format("2006-01-02 15:04:05"))

	// 获取所有指标
	allMetrics := gormGen.GetAllMetrics()
	fmt.Printf("\n📊 全局指标汇总:\n")
	totalRequests := int64(0)
	totalSuccess := int64(0)

	for bizTag, metric := range allMetrics {
		fmt.Printf("   %s: 请求=%d, 成功率=%.1f%%, QPS=%.1f\n",
			bizTag, metric.TotalRequests, metric.SuccessRate()*100, metric.AverageQPS)
		totalRequests += metric.TotalRequests
		totalSuccess += metric.SuccessRequests
	}

	if totalRequests > 0 {
		overallSuccessRate := float64(totalSuccess) / float64(totalRequests) * 100
		fmt.Printf("   整体成功率: %.2f%%\n", overallSuccessRate)
	}

	// 更新步长演示
	fmt.Println("\n⚙️  动态步长调整演示:")
	oldStep := gormGen.GetBufferStatus("metrics_test")["step"]
	fmt.Printf("   当前步长: %v\n", oldStep)

	err = gormGen.UpdateStep(ctx, "metrics_test", 3000)
	if err != nil {
		fmt.Printf("   更新步长失败: %v\n", err)
	} else {
		newStep := gormGen.GetBufferStatus("metrics_test")["step"]
		fmt.Printf("   新步长: %v\n", newStep)
		fmt.Println("   ✅ 步长更新成功")
	}

	fmt.Println("\n🎯 关键特性总结:")
	fmt.Println("   ✅ 基于GORM，支持多种数据库")
	fmt.Println("   ✅ 双缓冲区机制，高性能ID生成")
	fmt.Println("   ✅ 自动预加载，避免服务阻塞")
	fmt.Println("   ✅ 动态步长调整，适应不同负载")
	fmt.Println("   ✅ 完整的监控指标和状态查看")
	fmt.Println("   ✅ 支持多业务标识，业务隔离")
	fmt.Println("   ✅ 优雅关闭和资源清理")
}

// 并发测试示例
func demoConcurrencyTest() {
	fmt.Println("\n🔀 并发测试演示")
	fmt.Println("================")

	config := idgen.SQLiteConfig("./concurrency.db")
	idGen, err := core.NewIDGenerator(config)
	if err != nil {
		log.Printf("创建并发测试ID生成器失败: %v", err)
		return
	}

	gormGen := idGen.(*idgen.GormLeafIDGenerator)
	defer func(gormGen *idgen.GormLeafIDGenerator) {
		err = gormGen.Close()
	}(gormGen)

	ctx := context.Background()
	_ = gormGen.CreateTable(ctx)
	_ = gormGen.CreateBizTag(ctx, "concurrent_test", 1000, "并发测试")

	// 并发生成ID
	goroutineCount := 10
	idsPerGoroutine := 1000

	start := time.Now()
	done := make(chan bool, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(goroutineID int) {
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gormGen.NextID(ctx, "concurrent_test")
				if err != nil {
					log.Printf("协程%d生成ID失败: %v", goroutineID, err)
					break
				}
				_ = id // 使用ID避免编译器警告
			}
			done <- true
		}(i)
	}

	// 等待所有协程完成
	for i := 0; i < goroutineCount; i++ {
		<-done
	}

	duration := time.Since(start)
	totalIDs := goroutineCount * idsPerGoroutine
	qps := float64(totalIDs) / duration.Seconds()

	fmt.Printf("🔀 并发测试结果:\n")
	fmt.Printf("   协程数: %d\n", goroutineCount)
	fmt.Printf("   每协程ID数: %d\n", idsPerGoroutine)
	fmt.Printf("   总ID数: %d\n", totalIDs)
	fmt.Printf("   耗时: %v\n", duration)
	fmt.Printf("   并发QPS: %.0f\n", qps)

	// 检查指标
	metrics := gormGen.GetMetrics("concurrent_test")
	fmt.Printf("   成功率: %.2f%%\n", metrics.SuccessRate()*100)
}
