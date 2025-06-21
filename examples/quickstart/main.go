package main

import (
	"distributed-service/framework"
	"log"
)

func main() {
	log.Println("🚀 启动分布式服务框架示例...")

	// 最简单的启动方式 - 零配置
	if err := framework.Start(); err != nil {
		log.Fatalf("Framework start failed: %v", err)
	}
}
