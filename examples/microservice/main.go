package main

import (
	"context"
	"github.com/qiaojinxia/distributed-service/framework/core"
	"log"
)

func main() {
	log.Println("🔌 启动微服务示例...")

	// 微服务模式 - 启用gRPC、监控、链路追踪
	err := core.New().
		Port(9000).                                   // gRPC端口
		Name("user-service").                         // 服务名称
		Version("v1.2.0").                            // 服务版本
		Mode("release").                              // 生产模式
		Enable("grpc", "metrics", "tracing").         // 启用指定组件
		Disable("http").                              // 禁用HTTP
		GRPC(setupGRPCServices).                      // 设置gRPC服务
		BeforeStart(func(ctx context.Context) error { // 启动前初始化
			log.Println("🔧 初始化数据库连接...")
			log.Println("🔧 初始化缓存...")
			return nil
		}).
		AfterStart(func(ctx context.Context) error { // 启动完成回调
			log.Println("✅ 微服务启动完成!")
			log.Println("🔌 gRPC服务: localhost:9000")
			log.Println("📊 监控指标: 已启用")
			log.Println("🔍 链路追踪: 已启用")
			return nil
		}).
		BeforeStop(func(ctx context.Context) error { // 停止时清理
			log.Println("🧹 清理资源...")
			return nil
		}).
		Run() // 启动服务

	if err != nil {
		log.Fatalf("微服务启动失败: %v", err)
	}
}

// setupGRPCServices 设置gRPC服务
func setupGRPCServices(s interface{}) {
	// 这里会在实现传输层后进行具体的gRPC服务注册
	// 暂时作为示例展示API设计
	log.Println("🔌 注册gRPC服务...")
	log.Println("  - UserService")
	log.Println("  - OrderService")
}
