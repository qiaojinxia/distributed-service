package framework

import (
	"github.com/qiaojinxia/distributed-service/framework/app"
)

// 🚀 分布式服务框架 - 主API

// New 创建新的应用构建器
//
// 使用示例：
//
//	framework.New().
//	  Port(8080).
//	  HTTP(routes).
//	  Run()
func New() *app.Builder {
	return app.New()
}

// ================================
// 🎯 零配置快速启动
// ================================

// Start 零配置启动 - 自动检测环境
//
// 使用示例：
//
//	framework.Start() // 一行启动完整服务
func Start() error {
	return New().AutoDetect().Run()
}

// Quick 快速启动 - 使用默认配置
//
// 使用示例：
//
//	framework.Quick() // 8080端口，生产模式
func Quick() error {
	return New().Port(8080).Mode("release").Run()
}

// ================================
// 🎨 便捷启动方法
// ================================

// Web 快速Web服务器
//
// 使用示例：
//
//	framework.Web(8080, func(r *gin.Engine) {
//	  r.GET("/", handler)
//	})
func Web(port int, routes ...app.HTTPHandler) error {
	builder := New().Port(port).OnlyHTTP()
	for _, route := range routes {
		builder.HTTP(route)
	}
	return builder.Run()
}

// Micro 微服务模式启动 - 只启用gRPC
//
// 使用示例：
//
//	framework.Micro(9000, grpcService1, grpcService2)
func Micro(port int, services ...app.GRPCHandler) error {
	builder := New().Port(port).OnlyGRPC()
	for _, service := range services {
		builder.GRPC(service)
	}
	return builder.Run()
}

// ================================
// 🛠️ 开发便捷方法
// ================================

// Dev 开发模式启动
func Dev() error {
	return New().
		Port(8080).
		Mode("debug").
		EnableAll().
		HTTP(defaultRoutes).
		Run()
}

// Prod 生产模式启动
func Prod() error {
	return New().
		Port(80).
		Mode("release").
		EnableAll().
		Run()
}

// defaultRoutes 默认路由 - 提供基础的健康检查和信息接口
func defaultRoutes(r interface{}) {
	// 这里会在 transport/http 中实现
}
