package main

import (
	"context"
	"github.com/qiaojinxia/distributed-service/framework/core"
	"log"
)

func main() {
	log.Println("🌐 启动Web应用示例...")

	// 链式配置启动Web服务
	err := core.New().
		Port(8080).                                  // 设置端口
		Mode("debug").                               // 开发模式
		Name("web-demo").                            // 应用名称
		Version("v1.0.0").                           // 应用版本
		OnlyHTTP().                                  // 只启用HTTP服务
		HTTP(setupRoutes).                           // 设置路由
		AfterStart(func(ctx context.Context) error { // 启动完成回调
			log.Println("✅ Web服务启动完成!")
			log.Println("📍 访问: http://localhost:8080")
			log.Println("🔍 健康检查: http://localhost:8080/health")
			log.Println("📊 信息接口: http://localhost:8080/info")
			return nil
		}).
		Run() // 启动服务

	if err != nil {
		log.Fatalf("Web服务启动失败: %v", err)
	}
}

// setupRoutes 设置HTTP路由
func setupRoutes(r interface{}) {
	// 这里会在实现传输层后进行具体的路由设置
	// 暂时作为示例展示API设计
	log.Println("📡 设置HTTP路由...")
}
