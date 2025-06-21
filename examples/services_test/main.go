package main

import (
	"context"
	"distributed-service/framework"
	"distributed-service/framework/config"
	httpTransport "distributed-service/framework/transport/http"
	"distributed-service/pkg/etcd"
	"distributed-service/pkg/kafka"
	"distributed-service/pkg/redis_cluster"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 启动新增服务测试...")

	// 使用简化配置启动框架
	err := framework.New().
		Port(8080).
		Name("services-test").
		Version("v1.0.0").
		Mode("debug").

		// 配置三个新服务（使用默认值进行测试）
		WithRedisCluster(&config.RedisClusterConfig{
			Addrs:    []string{"localhost:7000", "localhost:7001", "localhost:7002"},
			PoolSize: 10,
		}).
		WithKafka(&config.KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
			Group:    "test-group",
			Version:  "2.8.0",
		}).
		WithEtcd(&config.EtcdConfig{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 5,
		}).

		// HTTP路由
		HTTP(setupTestRoutes).
		BeforeStart(func(ctx context.Context) error {
			log.Println("🔧 初始化测试服务...")
			return nil
		}).
		AfterStart(func(ctx context.Context) error {
			log.Println("✅ 服务测试启动完成!")
			log.Println("📍 测试端点:")
			log.Println("  - http://localhost:8080/test/redis-cluster")
			log.Println("  - http://localhost:8080/test/kafka")
			log.Println("  - http://localhost:8080/test/etcd")
			log.Println("  - http://localhost:8080/test/all")
			return nil
		}).
		Run()

	if err != nil {
		log.Fatalf("服务测试启动失败: %v", err)
	}
}

func setupTestRoutes(r interface{}) {
	if engine, ok := r.(*gin.Engine); ok {
		test := engine.Group("/test")
		{
			test.GET("/redis-cluster", testRedisCluster)
			test.GET("/kafka", testKafka)
			test.GET("/etcd", testEtcd)
			test.GET("/all", testAll)
			test.GET("/status", getStatus)
		}
	}
}

func testRedisCluster(c *gin.Context) {
	client := redis_cluster.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Redis Cluster client not initialized")
		return
	}

	ctx := c.Request.Context()
	testKey := "test:cluster:key"
	testValue := "Hello Redis Cluster!"

	// 测试设置
	if err := client.Set(ctx, testKey, testValue, time.Minute); err != nil {
		httpTransport.InternalError(c, "Failed to set Redis key: "+err.Error())
		return
	}

	// 测试获取
	value, err := client.Get(ctx, testKey)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get Redis key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"service":         "Redis Cluster",
		"status":          "success",
		"test_key":        testKey,
		"test_value":      testValue,
		"retrieved_value": value,
		"match":           value == testValue,
	})
}

func testKafka(c *gin.Context) {
	client := kafka.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Kafka client not initialized")
		return
	}

	ctx := c.Request.Context()
	testTopic := "test-topic"
	testMessage := "Hello Kafka!"

	// 测试发送消息
	if err := client.SendMessage(ctx, testTopic, []byte("test-key"), []byte(testMessage)); err != nil {
		httpTransport.InternalError(c, "Failed to send Kafka message: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"service":      "Apache Kafka",
		"status":       "success",
		"test_topic":   testTopic,
		"test_message": testMessage,
		"note":         "Message sent successfully",
	})
}

func testEtcd(c *gin.Context) {
	client := etcd.GetClient()
	if client == nil {
		httpTransport.InternalError(c, "Etcd client not initialized")
		return
	}

	ctx := c.Request.Context()
	testKey := "test/etcd/key"
	testValue := "Hello Etcd!"

	// 测试设置
	if err := client.Put(ctx, testKey, testValue); err != nil {
		httpTransport.InternalError(c, "Failed to put Etcd key: "+err.Error())
		return
	}

	// 测试获取
	value, err := client.Get(ctx, testKey)
	if err != nil {
		httpTransport.InternalError(c, "Failed to get Etcd key: "+err.Error())
		return
	}

	httpTransport.Success(c, gin.H{
		"service":         "Etcd",
		"status":          "success",
		"test_key":        testKey,
		"test_value":      testValue,
		"retrieved_value": value,
		"match":           value == testValue,
	})
}

func testAll(c *gin.Context) {
	results := make(map[string]interface{})

	// 测试Redis Cluster
	if client := redis_cluster.GetClient(); client != nil {
		ctx := c.Request.Context()
		testKey := "test:all:redis"
		err := client.Set(ctx, testKey, "test", time.Minute)
		results["redis_cluster"] = map[string]interface{}{
			"available":   true,
			"test_passed": err == nil,
			"error": func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}
	} else {
		results["redis_cluster"] = map[string]interface{}{
			"available":   false,
			"test_passed": false,
			"error":       "client not initialized",
		}
	}

	// 测试Kafka
	if client := kafka.GetClient(); client != nil {
		ctx := c.Request.Context()
		err := client.SendMessage(ctx, "test-all", []byte("key"), []byte("test"))
		results["kafka"] = map[string]interface{}{
			"available":   true,
			"test_passed": err == nil,
			"error": func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}
	} else {
		results["kafka"] = map[string]interface{}{
			"available":   false,
			"test_passed": false,
			"error":       "client not initialized",
		}
	}

	// 测试Etcd
	if client := etcd.GetClient(); client != nil {
		ctx := c.Request.Context()
		err := client.Put(ctx, "test/all/etcd", "test")
		results["etcd"] = map[string]interface{}{
			"available":   true,
			"test_passed": err == nil,
			"error": func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}
	} else {
		results["etcd"] = map[string]interface{}{
			"available":   false,
			"test_passed": false,
			"error":       "client not initialized",
		}
	}

	// 计算总体状态
	totalTests := len(results)
	passedTests := 0
	for _, result := range results {
		if r, ok := result.(map[string]interface{}); ok {
			if r["test_passed"].(bool) {
				passedTests++
			}
		}
	}

	httpTransport.Success(c, gin.H{
		"summary": gin.H{
			"total_tests":  totalTests,
			"passed_tests": passedTests,
			"success_rate": float64(passedTests) / float64(totalTests) * 100,
		},
		"results": results,
	})
}

func getStatus(c *gin.Context) {
	status := gin.H{
		"redis_cluster": redis_cluster.GetClient() != nil,
		"kafka":         kafka.GetClient() != nil,
		"etcd":          etcd.GetClient() != nil,
	}

	httpTransport.Success(c, gin.H{
		"timestamp": time.Now(),
		"services":  status,
	})
}
