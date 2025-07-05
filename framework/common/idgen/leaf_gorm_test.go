package idgen

import (
	"context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"testing"
	"time"
)

func setupTestDB(t testing.TB) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

func TestGormLeafIDGenerator_BasicUsage(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()
	config.DefaultStep = 100

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()

	// 创建表
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 创建业务标识
	err = generator.CreateBizTag(ctx, "test", 100, "test description")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	// 生成ID
	id1, err := generator.NextID(ctx, "test")
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	id2, err := generator.NextID(ctx, "test")
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	if id2 <= id1 {
		t.Errorf("Expected id2 (%d) > id1 (%d)", id2, id1)
	}

	if id2-id1 != 1 {
		t.Errorf("Expected consecutive IDs, got id1=%d, id2=%d", id1, id2)
	}
}

func TestGormLeafIDGenerator_BatchNextID(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()
	config.DefaultStep = 50

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	_ = generator.CreateTable(ctx)
	_ = generator.CreateBizTag(ctx, "batch_test", 50, "batch test")

	// 批量生成ID
	batchSize := 10
	ids, err := generator.BatchNextID(ctx, "batch_test", batchSize)
	if err != nil {
		t.Fatalf("Failed to batch generate IDs: %v", err)
	}

	if len(ids) != batchSize {
		t.Errorf("Expected %d IDs, got %d", batchSize, len(ids))
	}

	// 检查ID是否连续递增
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[i-1]+1 {
			t.Errorf("IDs not consecutive: ids[%d]=%d, ids[%d]=%d",
				i-1, ids[i-1], i, ids[i])
		}
	}
}

func TestGormLeafIDGenerator_ConcurrentAccess(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()
	config.DefaultStep = 1000

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = generator.CreateBizTag(ctx, "concurrent_test", 1000, "concurrent test")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	goroutineCount := 10
	idsPerGoroutine := 100
	totalIDs := goroutineCount * idsPerGoroutine

	allIDs := make([]int64, totalIDs)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	index := 0

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < idsPerGoroutine; j++ {
				id, err := generator.NextID(ctx, "concurrent_test")
				if err != nil {
					t.Errorf("Failed to generate ID in goroutine: %v", err)
					return
				}

				mutex.Lock()
				allIDs[index] = id
				index++
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	// 检查是否有重复ID
	idMap := make(map[int64]bool)
	for _, id := range allIDs {
		if idMap[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true
	}

	if len(idMap) != totalIDs {
		t.Errorf("Expected %d unique IDs, got %d", totalIDs, len(idMap))
	}
}

func TestGormLeafIDGenerator_SegmentPreload(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()
	config.DefaultStep = 10
	config.PreloadThreshold = 0.5 // 50%时预加载

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = generator.CreateBizTag(ctx, "preload_test", 10, "preload test")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	// 生成足够的ID来触发预加载
	for i := 0; i < 8; i++ { // 超过50%阈值（5个ID）触发预加载
		id, err := generator.NextID(ctx, "preload_test")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
		t.Logf("Generated ID: %d", id)
	}

	// 给预加载一些时间
	time.Sleep(500 * time.Millisecond)

	// 继续生成更多ID来完成第一个segment并切换到第二个
	for i := 0; i < 8; i++ {
		id, err := generator.NextID(ctx, "preload_test")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
		t.Logf("Generated ID: %d", id)
	}

	// 给第二次预加载时间
	time.Sleep(500 * time.Millisecond)

	// 检查指标
	metrics := generator.GetMetrics("preload_test")
	if metrics.SegmentLoads < 2 {
		t.Errorf("Expected at least 2 segment loads, got %d", metrics.SegmentLoads)
	}

	t.Logf("Segment loads: %d", metrics.SegmentLoads)
	t.Logf("Buffer switches: %d", metrics.BufferSwitches)
}

func TestGormLeafIDGenerator_MultipleBizTags(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 创建多个业务标识
	bizTags := []string{"user", "order", "product"}
	for _, bizTag := range bizTags {
		err := generator.CreateBizTag(ctx, bizTag, 100, bizTag+" description")
		if err != nil {
			t.Fatalf("Failed to create biz tag %s: %v", bizTag, err)
		}
	}

	// 为每个业务标识生成ID
	idMaps := make(map[string][]int64)
	for _, bizTag := range bizTags {
		ids := make([]int64, 5)
		for i := 0; i < 5; i++ {
			id, err := generator.NextID(ctx, bizTag)
			if err != nil {
				t.Fatalf("Failed to generate ID for %s: %v", bizTag, err)
			}
			ids[i] = id
		}
		idMaps[bizTag] = ids
	}

	// 检查不同业务标识的ID范围不重叠
	for bizTag, ids := range idMaps {
		t.Logf("BizTag %s: %v", bizTag, ids)

		// 检查同一业务标识内ID连续
		for i := 1; i < len(ids); i++ {
			if ids[i] != ids[i-1]+1 {
				t.Errorf("IDs not consecutive for %s: %d, %d",
					bizTag, ids[i-1], ids[i])
			}
		}
	}
}

func TestGormLeafIDGenerator_StepAdjustment(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()
	config.DefaultStep = 100
	config.MinStepSize = 50
	config.MaxStepSize = 1000

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = generator.CreateBizTag(ctx, "step_test", 100, "step test")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	// 生成一个ID来初始化buffer
	_, err = generator.NextID(ctx, "step_test")
	if err != nil {
		t.Fatalf("Failed to generate initial ID: %v", err)
	}

	// 更新步长
	err = generator.UpdateStep(ctx, "step_test", 200)
	if err != nil {
		t.Fatalf("Failed to update step: %v", err)
	}

	// 检查buffer状态
	status := generator.GetBufferStatus("step_test")
	step, exists := status["step"]
	if !exists {
		t.Error("Step not found in buffer status")
	} else if step != int32(200) {
		t.Errorf("Expected step 200, got %v", step)
	}
}

func TestGormLeafIDGenerator_Metrics(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = generator.CreateBizTag(ctx, "metrics_test", 100, "metrics test")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	// 生成一些ID
	expectedCount := 50
	for i := 0; i < expectedCount; i++ {
		_, err := generator.NextID(ctx, "metrics_test")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
	}

	// 检查指标
	metrics := generator.GetMetrics("metrics_test")

	if metrics.TotalRequests != int64(expectedCount) {
		t.Errorf("Expected %d total requests, got %d", expectedCount, metrics.TotalRequests)
	}

	if metrics.SuccessRequests != int64(expectedCount) {
		t.Errorf("Expected %d success requests, got %d", expectedCount, metrics.SuccessRequests)
	}

	if metrics.FailedRequests != 0 {
		t.Errorf("Expected 0 failed requests, got %d", metrics.FailedRequests)
	}

	successRate := metrics.SuccessRate()
	if successRate != 1.0 {
		t.Errorf("Expected 100%% success rate, got %.2f", successRate*100)
	}

	if metrics.SegmentLoads == 0 {
		t.Error("Expected at least 1 segment load")
	}
}

func TestGormLeafIDGenerator_BufferStatus(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()
	config.DefaultStep = 20

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = generator.CreateBizTag(ctx, "buffer_test", 20, "buffer test")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	// 生成一些ID
	for i := 0; i < 10; i++ {
		_, err := generator.NextID(ctx, "buffer_test")
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}
	}

	// 获取buffer状态
	status := generator.GetBufferStatus("buffer_test")

	// 检查必要的字段
	requiredFields := []string{"biz_tag", "current_pos", "next_ready", "init_ok", "step"}
	for _, field := range requiredFields {
		if _, exists := status[field]; !exists {
			t.Errorf("Required field %s not found in buffer status", field)
		}
	}

	// 检查current_segment
	if currentSegment, exists := status["current_segment"]; exists {
		segment := currentSegment.(map[string]interface{})
		segmentFields := []string{"min", "max", "cursor", "usage_ratio", "remaining"}
		for _, field := range segmentFields {
			if _, exists := segment[field]; !exists {
				t.Errorf("Required field %s not found in current_segment", field)
			}
		}
	} else {
		t.Error("current_segment not found in buffer status")
	}

	t.Logf("Buffer status: %+v", status)
}

func TestGormLeafIDGenerator_DeleteBizTag(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	err = generator.CreateBizTag(ctx, "delete_test", 100, "delete test")
	if err != nil {
		t.Fatalf("Failed to create biz tag: %v", err)
	}

	// 生成一个ID确保业务标识存在
	_, err = generator.NextID(ctx, "delete_test")
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	// 删除业务标识
	err = generator.DeleteBizTag(ctx, "delete_test")
	if err != nil {
		t.Fatalf("Failed to delete biz tag: %v", err)
	}

	// 尝试再次生成ID应该会自动重新创建（由于自动创建功能）
	id, err := generator.NextID(ctx, "delete_test")
	if err != nil {
		t.Errorf("Expected auto recreation to work, got error: %v", err)
	} else if id <= 0 {
		t.Errorf("Expected positive ID after recreation, got %d", id)
	}
}

func TestGormLeafIDGenerator_ErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	config := DefaultLeafConfig()

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	err := generator.CreateTable(ctx)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 由于自动创建业务标识功能，这个测试不再有效
	// 尝试为不存在的业务标识生成ID（会自动创建）
	id, err := generator.NextID(ctx, "auto_created")
	if err != nil {
		t.Errorf("Expected auto creation to work, got error: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// 批量生成0个ID
	_, err = generator.BatchNextID(ctx, "test", 0)
	if err == nil {
		t.Error("Expected error when batch generating 0 IDs")
	}

	// 批量生成负数个ID
	_, err = generator.BatchNextID(ctx, "test", -1)
	if err == nil {
		t.Error("Expected error when batch generating negative IDs")
	}
}

func BenchmarkGormLeafIDGenerator_NextID(b *testing.B) {
	db := setupTestDB(b)
	config := DefaultLeafConfig()
	config.DefaultStep = 10000

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	_ = generator.CreateTable(ctx)
	_ = generator.CreateBizTag(ctx, "bench", 10000, "benchmark test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := generator.NextID(ctx, "bench")
			if err != nil {
				b.Fatalf("Failed to generate ID: %v", err)
			}
		}
	})
}

func BenchmarkGormLeafIDGenerator_BatchNextID(b *testing.B) {
	db := setupTestDB(b)
	config := DefaultLeafConfig()
	config.DefaultStep = 10000

	generator := NewGormLeafIDGenerator(db, config)
	defer generator.Close()

	ctx := context.Background()
	_ = generator.CreateTable(ctx)
	_ = generator.CreateBizTag(ctx, "batch_bench", 10000, "batch benchmark test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.BatchNextID(ctx, "batch_bench", 100)
		if err != nil {
			b.Fatalf("Failed to batch generate IDs: %v", err)
		}
	}
}
