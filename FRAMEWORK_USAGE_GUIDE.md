# 🚀 分布式服务框架使用指南

## 📦 引用框架

### 1. 添加依赖

在您的项目中添加框架依赖：

```bash
go mod init your-project-name
go get github.com/qiaojinxia/distributed_service
```

### 2. 基本导入

```go
import (
    "github.com/qiaojinxia/distributed_service/framework"
    "github.com/qiaojinxia/distributed_service/framework/config"
)
```

## 🎯 快速开始

### 零配置启动

```go
package main

import "github.com/qiaojinxia/distributed_service/framework"

func main() {
    // 一行代码启动完整服务
    framework.Start()
}
```

### HTTP服务

```go
package main

import (
    "github.com/qiaojinxia/distributed_service/framework"
    "github.com/gin-gonic/gin"
)

func main() {
    framework.Web(8080, func(r *gin.Engine) {
        r.GET("/hello", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "Hello World!"})
        })
    })
}
```

### gRPC服务

```go
package main

import (
    "github.com/qiaojinxia/distributed_service/framework"
    "google.golang.org/grpc"
)

func main() {
    framework.Micro(9000, func(s interface{}) {
        grpcServer := s.(*grpc.Server)
        // 注册您的gRPC服务
        // pb.RegisterYourServiceServer(grpcServer, &YourService{})
    })
}
```

## 🔧 配置示例

### 数据库 + Redis + JWT

```go
package main

import (
    "github.com/qiaojinxia/distributed_service/framework"
    "github.com/qiaojinxia/distributed_service/framework/config"
)

func main() {
    framework.New().
        WithDatabase(&config.MySQLConfig{
            Host:     "localhost",
            Port:     3306,
            Username: "root",
            Password: "password",
            Database: "myapp",
        }).
        WithRedis(&config.RedisConfig{
            Host: "localhost",
            Port: 6379,
        }).
        WithAuth(&config.JWTConfig{
            SecretKey: "your-secret-key",
            Issuer:    "your-app",
        }).
        HTTP(setupRoutes).
        Run()
}

func setupRoutes(r interface{}) {
    // 设置路由
}
```

### 微服务架构

```go
framework.New().
    Port(8080).
    Name("user-service").
    Version("v1.0.0").
    WithRegistry(&config.ConsulConfig{
        Host: "localhost",
        Port: 8500,
    }).
    WithMetrics(&config.MetricsConfig{
        Enabled: true,
        PrometheusPort: 9090,
    }).
    WithTracing(&config.TracingConfig{
        Enabled: true,
        ServiceName: "user-service",
    }).
    EnableAll().
    HTTP(setupHTTPRoutes).
    GRPC(setupGRPCServices).
    Run()
```

## 📊 高级服务

### Redis Cluster + Kafka + Etcd

```go
framework.New().
    WithRedisCluster(&config.RedisClusterConfig{
        Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
        PoolSize: 20,
    }).
    WithKafka(&config.KafkaConfig{
        Brokers: []string{"localhost:9092"},
        ClientID: "my-app",
        Group: "my-group",
    }).
    WithEtcd(&config.EtcdConfig{
        Endpoints: []string{"localhost:2379"},
    }).
    HTTP(setupRoutes).
    Run()
```

### MongoDB + Elasticsearch

```go
framework.New().
    WithMongoDB(&config.MongoDBConfig{
        URI: "mongodb://localhost:27017",
        Database: "myapp",
    }).
    WithElasticsearch(&config.ElasticsearchConfig{
        Addresses: []string{"http://localhost:9200"},
    }).
    HTTP(setupRoutes).
    Run()
```

## 🛡️ 保护与监控

### Sentinel 保护

```go
framework.New().
    WithProtection(&config.ProtectionConfig{
        Enabled: true,
        RateLimitRules: []config.RateLimitRuleConfig{
            {
                Name:           "api-limit",
                Resource:       "/api/*",
                Threshold:      100,
                StatIntervalMs: 1000,
                Enabled:        true,
            },
        },
    }).
    HTTP(setupRoutes).
    Run()
```

### 完整监控

```go
framework.New().
    WithMetrics(&config.MetricsConfig{
        Enabled: true,
        PrometheusPort: 9090,
    }).
    WithTracing(&config.TracingConfig{
        Enabled: true,
        ServiceName: "my-service",
        ExporterType: "otlp",
        Endpoint: "http://localhost:4318",
    }).
    WithLogger(&config.LoggerConfig{
        Level: "info",
        Encoding: "json",
    }).
    HTTP(setupRoutes).
    Run()
```

## 🔄 生命周期管理

```go
framework.New().
    BeforeStart(func(ctx context.Context) error {
        log.Println("服务启动前...")
        return nil
    }).
    AfterStart(func(ctx context.Context) error {
        log.Println("服务启动完成!")
        return nil
    }).
    BeforeStop(func(ctx context.Context) error {
        log.Println("服务停止前...")
        return nil
    }).
    HTTP(setupRoutes).
    Run()
```

## 🎨 便捷方法

### 开发模式

```go
// 开发模式 - 自动启用所有功能
framework.Dev()

// 或者
framework.New().
    Dev().
    HTTP(setupRoutes).
    Run()
```

### 生产模式

```go
// 生产模式
framework.Prod()

// 或者
framework.New().
    Prod().
    HTTP(setupRoutes).
    Run()
```

### 智能检测

```go
// 自动检测环境配置
framework.New().
    AutoDetect().
    WithEnv().
    HTTP(setupRoutes).
    Run()
```

## 📝 配置文件

支持从配置文件加载：

```yaml
# config.yaml
server:
  port: 8080
  name: "my-service"
  
mysql:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "myapp"

redis:
  host: "localhost"
  port: 6379
```

```go
framework.New().
    Config("config.yaml").
    HTTP(setupRoutes).
    Run()
```

## 🌍 环境变量

支持环境变量覆盖：

```bash
export PORT=9000
export GIN_MODE=release
export CONFIG_PATH=config/prod.yaml
```

```go
framework.New().
    WithEnv().  // 从环境变量读取配置
    HTTP(setupRoutes).
    Run()
```

## 🧩 组件管理

### 获取组件实例

```go
builder := framework.New().
    WithDatabase(&config.MySQLConfig{...}).
    WithRedis(&config.RedisConfig{...})

// 获取组件管理器
manager := builder.GetComponentManager()

// 获取具体组件
config := manager.GetConfig()
auth := manager.GetAuth()
registry := manager.GetRegistry()
```

### 禁用组件

```go
framework.New().
    EnableAll().
    DisableComponents("metrics", "tracing").
    HTTP(setupRoutes).
    Run()
```

## 📚 更多示例

查看 `examples/` 目录下的完整示例：

- `examples/quickstart/` - 快速开始
- `examples/web/` - Web应用
- `examples/microservice/` - 微服务
- `examples/advanced_services/` - 高级服务
- `examples/components/` - 组件配置

## 🚀 部署

### Docker

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-service
  template:
    metadata:
      labels:
        app: my-service
    spec:
      containers:
      - name: my-service
        image: my-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: GIN_MODE
          value: "release"
```

## 🔗 相关链接

- **项目主页**: https://github.com/qiaojinxia/distributed_service
- **问题反馈**: https://github.com/qiaojinxia/distributed_service/issues
- **文档**: 查看项目 README.md

## 📄 许可证

本项目基于 MIT 许可证开源。 