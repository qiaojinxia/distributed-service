package core

import (
	"github.com/qiaojinxia/distributed-service/framework/app"
	"github.com/qiaojinxia/distributed-service/framework/cache"
	idgen2 "github.com/qiaojinxia/distributed-service/framework/common/idgen"
	"github.com/qiaojinxia/distributed-service/framework/component"
)

// 包初始化时设置缓存注册回调
func init() {
	component.SetCacheRegistryCallback(initGlobalCacheSystem)
}

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

// ================================
// 🆔 分布式ID服务
// ================================

// NewIDGenerator 创建分布式ID生成器
//
// 使用示例：
//
//	idGen, err := framework.NewIDGenerator(idgen.Config{
//	  Type: "leaf",
//	  TableName: "leaf_alloc",
//	  Database: &idgen.DatabaseConfig{
//	    Driver: "mysql",
//	    Host: "localhost",
//	    Port: 3306,
//	    Database: "test",
//	    Username: "root",
//	    Password: "password",
//	    Charset: "utf8mb4",
//	  },
//	})
func NewIDGenerator(config idgen2.Config) (idgen2.IDGenerator, error) {
	return idgen2.NewIDGeneratorFromConfig(config)
}

// ================================
// 💾 统一缓存系统
// ================================

// frameworkCacheService 框架缓存服务实例
var frameworkCacheService *cache.FrameworkCacheService

// initGlobalCacheSystem 初始化全局缓存系统（内部使用）
func initGlobalCacheSystem(fcs *cache.FrameworkCacheService) error {
	frameworkCacheService = fcs
	return nil
}

// NewCacheManager 创建缓存管理器（已废弃，使用 GetCache 系列方法）
//
// 推荐使用：
//   - GetCache(name) - 获取命名缓存
//   - GetUserCache() - 获取用户缓存
//   - GetSessionCache() - 获取会话缓存
//
// Deprecated: 使用新的统一缓存API
func NewCacheManager() *cache.Manager {
	manager := cache.NewManager()

	// 注册默认的缓存构建器
	manager.RegisterBuilder(cache.TypeMemory, &cache.MemoryBuilder{})
	manager.RegisterBuilder(cache.TypeRedis, &cache.RedisBuilder{})
	manager.RegisterBuilder(cache.TypeHybrid, &cache.HybridBuilder{})

	return manager
}

// ================================
// 🎯 简化缓存访问API
// ================================

// GetCache 获取指定名称的缓存（推荐使用）
//
// 使用示例：
//
//	userCache := framework.GetCache("users")
//	if userCache != nil {
//	  userCache.Set(ctx, "key", "value", time.Hour)
//	}
func GetCache(name string) cache.Cache {
	if frameworkCacheService == nil {
		return nil
	}
	c, _ := frameworkCacheService.GetNamedCache(name)
	return c
}

// GetUserCache 获取用户缓存
func GetUserCache() cache.Cache {
	return GetCache("users")
}

// GetSessionCache 获取会话缓存
func GetSessionCache() cache.Cache {
	return GetCache("sessions")
}

// GetProductCache 获取产品缓存
func GetProductCache() cache.Cache {
	return GetCache("products")
}

// GetConfigCache 获取配置缓存
func GetConfigCache() cache.Cache {
	return GetCache("configs")
}

// HasCache 检查指定缓存是否存在
func HasCache(name string) bool {
	return GetCache(name) != nil
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats(name string) (*cache.Stats, error) {
	c := GetCache(name)
	if c == nil {
		return nil, cache.ErrCacheNotFound
	}
	if statsCache, ok := c.(cache.StatsCache); ok {
		stats := statsCache.GetStats()
		return &stats, nil
	}
	return nil, cache.ErrStatsNotSupported
}
