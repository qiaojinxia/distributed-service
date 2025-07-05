package main

import (
	"context"
	"github.com/qiaojinxia/distributed-service/framework/core"
	"log"
	"time"

	"github.com/qiaojinxia/distributed-service/framework/config"
	httpTransport "github.com/qiaojinxia/distributed-service/framework/transport/http"
	"github.com/qiaojinxia/distributed-service/pkg/etcd"
	"github.com/qiaojinxia/distributed-service/pkg/kafka"
	"github.com/qiaojinxia/distributed-service/pkg/redis_cluster"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 启动高级分布式服务示例...")
	log.Println("   包含：Redis Cluster + Kafka + Etcd")

	// 使用新增的三个高级服务启动框架
	err := core.New().
		Port(8080).
		Name("advanced-services-demo").
		Version("v3.0.0").
		Mode("debug").

		// 🔗 Redis集群配置
		WithRedisCluster(&config.RedisClusterConfig{
			Addrs:      []string{"localhost:7000", "localhost:7001", "localhost:7002"},
			Password:   "",
			MaxRetries: 3,

			// 连接池配置
			PoolSize:           20,
			MinIdleConns:       5,
			MaxConnAge:         60,  // 秒
			PoolTimeout:        10,  // 秒
			IdleTimeout:        300, // 秒
			IdleCheckFrequency: 60,  // 秒

			// 集群配置
			MaxRedirects:   8,
			ReadOnly:       false,
			RouteByLatency: true,
			RouteRandomly:  false,

			// 超时配置
			DialTimeout:  5, // 秒
			ReadTimeout:  3, // 秒
			WriteTimeout: 3, // 秒
		}).

		// 📊 Kafka配置
		WithKafka(&config.KafkaConfig{
			Brokers:       []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			ClientID:      "advanced-services-demo",
			Group:         "advanced-demo-group",
			Version:       "2.8.0",
			RetryBackoff:  100,
			RetryMax:      3,
			FlushMessages: 100,
			FlushBytes:    1024 * 1024,
			FlushTimeout:  100,
		}).

		// 🔑 Etcd配置
		WithEtcd(&config.EtcdConfig{
			Endpoints:   []string{"localhost:2379", "localhost:2380", "localhost:2381"},
			Username:    "",
			Password:    "",
			DialTimeout: 5,
		}).

		// 🌐 HTTP服务配置
		HTTP(setupAdvancedServicesRoutes).

		// 🔄 生命周期回调
		BeforeStart(func(ctx context.Context) error {
			log.Println("🔧 执行高级服务初始化...")
			log.Println("  - 连接Redis集群")
			log.Println("  - 初始化Kafka生产者和消费者")
			log.Println("  - 建立Etcd连接")

			// 初始化并测试服务
			go testRedisCluster(ctx)
			go testKafka(ctx)
			go testEtcd(ctx)

			return nil
		}).
		AfterStart(func(ctx context.Context) error {
			log.Println("✅ 高级分布式服务启动完成!")
			log.Println("📍 服务信息:")
			log.Println("  - HTTP API:        http://localhost:8080")
			log.Println("  - 健康检查:         http://localhost:8080/health")
			log.Println("  - Redis集群测试:    http://localhost:8080/api/redis-cluster/test")
			log.Println("  - Kafka测试:       http://localhost:8080/api/kafka/test")
			log.Println("  - Etcd测试:        http://localhost:8080/api/etcd/test")

			log.Println("🧩 已启用的高级服务:")
			log.Println("  - ✅ Redis Cluster    (分布式缓存集群)")
			log.Println("  - ✅ Apache Kafka     (分布式消息队列)")
			log.Println("  - ✅ Etcd             (分布式配置中心)")

			return nil
		}).
		BeforeStop(func(ctx context.Context) error {
			log.Println("🧹 执行高级服务清理...")
			log.Println("  - 关闭Redis集群连接")
			log.Println("  - 停止Kafka生产者和消费者")
			log.Println("  - 断开Etcd连接")
			return nil
		}).

		// 🚀 启动框架
		Run()

	if err != nil {
		log.Fatalf("高级分布式服务启动失败: %v", err)
	}
}

// setupAdvancedServicesRoutes 设置高级服务HTTP路由
func setupAdvancedServicesRoutes(r interface{}) {
	if engine, ok := r.(*gin.Engine); ok {
		log.Println("📡 设置高级服务HTTP路由...")

		// API路由组
		api := engine.Group("/api")
		{
			// Redis集群API
			redisCluster := api.Group("/redis-cluster")
			{
				redisCluster.GET("/test", testRedisClusterHandler)
				redisCluster.POST("/set", setRedisClusterHandler)
				redisCluster.GET("/get/:key", getRedisClusterHandler)
				redisCluster.DELETE("/del/:key", deleteRedisClusterHandler)
				redisCluster.GET("/info", redisClusterInfoHandler)
			}

			// Kafka API
			kafkaGroup := api.Group("/kafka")
			{
				kafkaGroup.GET("/test", testKafkaHandler)
				kafkaGroup.POST("/send", sendKafkaMessageHandler)
				kafkaGroup.GET("/topics", getKafkaTopicsHandler)
				kafkaGroup.GET("/consume/:topic", consumeKafkaHandler)
			}

			// Etcd API
			etcdGroup := api.Group("/etcd")
			{
				etcdGroup.GET("/test", testEtcdHandler)
				etcdGroup.POST("/put", putEtcdHandler)
				etcdGroup.GET("/get/:key", getEtcdHandler)
				etcdGroup.DELETE("/del/:key", deleteEtcdHandler)
				etcdGroup.GET("/watch/:key", watchEtcdHandler)
				etcdGroup.POST("/lock", lockEtcdHandler)
			}

			// 综合测试API
			api.GET("/test/all", testAllServicesHandler)
			api.GET("/status", getServicesStatusHandler)
		}

		log.Println("✅ 高级服务HTTP路由配置完成")
		log.Println("  - Redis Cluster API: /api/redis-cluster/*")
		log.Println("  - Kafka API:         /api/kafka/*")
		log.Println("  - Etcd API:          /api/etcd/*")
		log.Println("  - 综合测试:           /api/test/all")
	}
}

// ================================
// 🗃️ Redis Cluster 处理器
// ================================

func testRedisClusterHandler(c *gin.Context) {
	client := redis_cluster.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Redis Cluster client not initialized")
		return
	}

	ctx := c.Request.Context()

	// 测试设置和获取
	testKey := "test:redis-cluster"
	testValue := "Hello Redis Cluster!"

	if err := client.Set(ctx, testKey, testValue, time.Minute); err != nil {
		httpTransport.InternalError(c, "Failed to set Redis Cluster key: "+err.Error())
		return
	}

	value, err := client.Get(ctx, testKey)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get Redis Cluster key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message":         "Redis Cluster test successful",
		"test_key":        testKey,
		"test_value":      testValue,
		"retrieved_value": value,
		"cluster_info":    "Connected to Redis Cluster",
	})
}

func setRedisClusterHandler(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		TTL   int    `json:"ttl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpTransport.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	client := redis_cluster.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Redis Cluster client not initialized")
		return
	}

	ctx := c.Request.Context()
	ttl := time.Duration(req.TTL) * time.Second
	if req.TTL == 0 {
		ttl = 0 // 永不过期
	}

	if err := client.Set(ctx, req.Key, req.Value, ttl); err != nil {
		httpTransport.InternalError(c, "Failed to set key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message": "Key set successfully",
		"key":     req.Key,
		"ttl":     req.TTL,
	})
}

func getRedisClusterHandler(c *gin.Context) {
	key := c.Param("key")

	client := redis_cluster.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Redis Cluster client not initialized")
		return
	}

	ctx := c.Request.Context()
	value, err := client.Get(ctx, key)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get key: "+err.Error())
		return
	}

	if value == "" {
		httpTransport.NotFound(c, "Key not found")
		return
	}

	httpTransport.Success(c, gin.H{
		"key":   key,
		"value": value,
	})
}

func deleteRedisClusterHandler(c *gin.Context) {
	key := c.Param("key")

	client := redis_cluster.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Redis Cluster client not initialized")
		return
	}

	ctx := c.Request.Context()
	if err := client.Del(ctx, key); err != nil {
		httpTransport.InternalError(c, "Failed to delete key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message": "Key deleted successfully",
		"key":     key,
	})
}

func redisClusterInfoHandler(c *gin.Context) {
	client := redis_cluster.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Redis Cluster client not initialized")
		return
	}

	ctx := c.Request.Context()
	info, err := client.ClusterInfo(ctx)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get cluster info: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"cluster_info": info,
	})
}

// ================================
// 📊 Kafka 处理器
// ================================

func testKafkaHandler(c *gin.Context) {
	client := kafka.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Kafka client not initialized")
		return
	}

	ctx := c.Request.Context()

	// 发送测试消息
	testTopic := "test-topic"
	testMessage := "Hello Kafka!"

	if err := client.SendMessage(ctx, testTopic, []byte("test-key"), []byte(testMessage)); err != nil {
		httpTransport.InternalError(c, "Failed to send Kafka message: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message":      "Kafka test successful",
		"test_topic":   testTopic,
		"test_message": testMessage,
	})
}

func sendKafkaMessageHandler(c *gin.Context) {
	var req struct {
		Topic   string            `json:"topic" binding:"required"`
		Key     string            `json:"key"`
		Message string            `json:"message" binding:"required"`
		Headers map[string]string `json:"headers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpTransport.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	client := kafka.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Kafka client not initialized")
		return
	}

	ctx := c.Request.Context()

	// 转换headers
	headers := make(map[string][]byte)
	for k, v := range req.Headers {
		headers[k] = []byte(v)
	}

	if len(headers) > 0 {
		err := client.SendMessageWithHeaders(ctx, req.Topic, []byte(req.Key), []byte(req.Message), headers)
		if err != nil {
			httpTransport.InternalError(c, "Failed to send message: "+err.Error())
			return
		}
	} else {
		err := client.SendMessage(ctx, req.Topic, []byte(req.Key), []byte(req.Message))
		if err != nil {
			httpTransport.InternalError(c, "Failed to send message: "+err.Error())
			return
		}
	}

	httpTransport.Success(c, gin.H{
		"message": "Message sent successfully",
		"topic":   req.Topic,
	})
}

func getKafkaTopicsHandler(c *gin.Context) {
	client := kafka.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Kafka client not initialized")
		return
	}

	ctx := c.Request.Context()
	topics, err := client.GetTopics(ctx)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get topics: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"topics": topics,
		"count":  len(topics),
	})
}

func consumeKafkaHandler(c *gin.Context) {
	topic := c.Param("topic")

	httpTransport.Success(c, gin.H{
		"message": "Kafka consumer endpoint",
		"topic":   topic,
		"note":    "Consumer running in background",
	})
}

// ================================
// 🔑 Etcd 处理器
// ================================

func testEtcdHandler(c *gin.Context) {
	client := etcd.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Etcd client not initialized")
		return
	}

	ctx := c.Request.Context()

	// 测试设置和获取
	testKey := "test/etcd/key"
	testValue := "Hello Etcd!"

	if err := client.Put(ctx, testKey, testValue); err != nil {
		httpTransport.InternalError(c, "Failed to put Etcd key: "+err.Error())
		return
	}

	value, err := client.Get(ctx, testKey)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get Etcd key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message":         "Etcd test successful",
		"test_key":        testKey,
		"test_value":      testValue,
		"retrieved_value": value,
	})
}

func putEtcdHandler(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		TTL   int    `json:"ttl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpTransport.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	client := etcd.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Etcd client not initialized")
		return
	}

	ctx := c.Request.Context()

	if req.TTL > 0 {
		ttl := time.Duration(req.TTL) * time.Second
		err := client.PutWithTTL(ctx, req.Key, req.Value, ttl)
		if err != nil {
			httpTransport.InternalError(c, "Failed to put key with TTL: "+err.Error())
			return
		}
	} else {
		err := client.Put(ctx, req.Key, req.Value)
		if err != nil {
			httpTransport.InternalError(c, "Failed to put key: "+err.Error())
			return
		}
	}

	httpTransport.Success(c, gin.H{
		"message": "Key set successfully",
		"key":     req.Key,
		"ttl":     req.TTL,
	})
}

func getEtcdHandler(c *gin.Context) {
	key := c.Param("key")

	client := etcd.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Etcd client not initialized")
		return
	}

	ctx := c.Request.Context()
	value, err := client.Get(ctx, key)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get key: "+err.Error())
		return
	}

	if value == "" {
		httpTransport.NotFound(c, "Key not found")
		return
	}

	httpTransport.Success(c, gin.H{
		"key":   key,
		"value": value,
	})
}

func deleteEtcdHandler(c *gin.Context) {
	key := c.Param("key")

	client := etcd.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Etcd client not initialized")
		return
	}

	ctx := c.Request.Context()
	if err := client.Delete(ctx, key); err != nil {
		httpTransport.InternalError(c, "Failed to delete key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message": "Key deleted successfully",
		"key":     key,
	})
}

func watchEtcdHandler(c *gin.Context) {
	key := c.Param("key")

	httpTransport.Success(c, gin.H{
		"message": "Etcd watch endpoint",
		"key":     key,
		"note":    "Watch running in background",
	})
}

func lockEtcdHandler(c *gin.Context) {
	var req struct {
		Key string `json:"key" binding:"required"`
		TTL int    `json:"ttl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httpTransport.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	client := etcd.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Etcd client not initialized")
		return
	}

	ctx := c.Request.Context()
	ttl := time.Duration(req.TTL) * time.Second
	if req.TTL == 0 {
		ttl = 10 * time.Second // 默认10秒
	}

	resp, err := client.Lock(ctx, req.Key, ttl)
	if err != nil {
		httpTransport.InternalError(c, "Failed to acquire lock: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"message": "Lock acquired",
		"key":     req.Key,
		"ttl":     req.TTL,
		"success": resp.Succeeded,
	})
}

// ================================
// 🧪 综合测试处理器
// ================================

func testAllServicesHandler(c *gin.Context) {
	results := make(map[string]interface{})

	// 测试Redis Cluster
	if client := redis_cluster.GetClient(); client != nil {
		ctx := c.Request.Context()
		testKey := "test:all:redis"
		err := client.Set(ctx, testKey, "test", time.Minute)
		results["redis_cluster"] = map[string]interface{}{
			"available": true,
			"test":      err == nil,
			"error":     getErrorString(err),
		}
	} else {
		results["redis_cluster"] = map[string]interface{}{
			"available": false,
			"test":      false,
			"error":     "client not initialized",
		}
	}

	// 测试Kafka
	if client := kafka.GetClient(); client != nil {
		ctx := c.Request.Context()
		err := client.SendMessage(ctx, "test-all", []byte("key"), []byte("test"))
		results["kafka"] = map[string]interface{}{
			"available": true,
			"test":      err == nil,
			"error":     getErrorString(err),
		}
	} else {
		results["kafka"] = map[string]interface{}{
			"available": false,
			"test":      false,
			"error":     "client not initialized",
		}
	}

	// 测试Etcd
	if client := etcd.GetClient(); client != nil {
		ctx := c.Request.Context()
		err := client.Put(ctx, "test/all/etcd", "test")
		results["etcd"] = map[string]interface{}{
			"available": true,
			"test":      err == nil,
			"error":     getErrorString(err),
		}
	} else {
		results["etcd"] = map[string]interface{}{
			"available": false,
			"test":      false,
			"error":     "client not initialized",
		}
	}

	httpTransport.Success(c, gin.H{
		"message": "All services test completed",
		"results": results,
	})
}

func getServicesStatusHandler(c *gin.Context) {
	status := gin.H{
		"redis_cluster": redis_cluster.GetClient() != nil,
		"kafka":         kafka.GetClient() != nil,
		"etcd":          etcd.GetClient() != nil,
	}

	httpTransport.Success(c, gin.H{
		"status": status,
	})
}

func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// ================================
// 🧪 后台测试服务
// ================================

func testRedisCluster(ctx context.Context) {
	client := redis_cluster.GetClient()
	if client == nil {
		log.Println("⚠️  Redis Cluster client not available for testing")
		return
	}

	log.Println("🧪 Testing Redis Cluster...")
	testKey := "framework:test:cluster"

	if err := client.Set(ctx, testKey, "Hello Redis Cluster!", time.Minute); err != nil {
		log.Printf("❌ Redis Cluster test failed: %v", err)
		return
	}

	value, err := client.Get(ctx, testKey)
	if err != nil {
		log.Printf("❌ Redis Cluster get failed: %v", err)
		return
	}

	log.Printf("✅ Redis Cluster test successful: %s", value)
}

func testKafka(ctx context.Context) {
	client := kafka.GetClient()
	if client == nil {
		log.Println("⚠️  Kafka client not available for testing")
		return
	}

	log.Println("🧪 Testing Kafka...")
	testTopic := "framework-test"

	if err := client.SendMessage(ctx, testTopic, []byte("test-key"), []byte("Hello Kafka!")); err != nil {
		log.Printf("❌ Kafka test failed: %v", err)
		return
	}

	log.Println("✅ Kafka test successful")
}

func testEtcd(ctx context.Context) {
	client := etcd.GetClient()
	if client == nil {
		log.Println("⚠️  Etcd client not available for testing")
		return
	}

	log.Println("🧪 Testing Etcd...")
	testKey := "framework/test/etcd"

	if err := client.Put(ctx, testKey, "Hello Etcd!"); err != nil {
		log.Printf("❌ Etcd test failed: %v", err)
		return
	}

	value, err := client.Get(ctx, testKey)
	if err != nil {
		log.Printf("❌ Etcd get failed: %v", err)
		return
	}

	log.Printf("✅ Etcd test successful: %s", value)
}
