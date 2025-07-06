# 插件模块设计文档

## 📋 概述

插件模块是分布式服务框架的扩展核心组件，提供动态插件加载、生命周期管理和事件总线功能。支持热插拔、依赖管理、安全隔离和插件通信，实现框架的高度可扩展性。

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    应用程序层                            │
│                Application Layer                        │
│              Plugin Management API                      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   插件管理层                             │
│                Plugin Management Layer                  │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   插件注册表    │   生命周期管理   │   依赖解析器    │ │
│  │Plugin Registry │Lifecycle Manager │Dependency Resolver│ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   事件通信层                             │
│               Event Communication Layer                 │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │    事件总线     │   消息队列      │   调度器        │ │
│  │  Event Bus      │Message Queue    │  Scheduler      │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   插件实例层                             │
│                Plugin Instance Layer                    │
│  ┌─────────────────┬─────────────────┬─────────────────┐ │
│  │   业务插件      │   中间件插件    │   系统插件      │ │
│  │Business Plugin  │Middleware Plugin│ System Plugin   │ │
│  └─────────────────┴─────────────────┴─────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 🎯 核心特点

### 1. 动态加载
- **热插拔**: 运行时动态加载和卸载插件
- **版本管理**: 支持插件版本控制和升级
- **依赖解析**: 自动解析和管理插件依赖关系
- **隔离机制**: 插件之间的安全隔离和资源控制

### 2. 生命周期管理
- **初始化阶段**: 插件加载和配置验证
- **启动阶段**: 插件服务启动和资源分配
- **运行阶段**: 插件正常运行和事件处理
- **停止阶段**: 优雅关闭和资源清理

### 3. 事件驱动
- **事件总线**: 统一的事件发布和订阅机制
- **异步处理**: 支持异步事件处理和回调
- **事件路由**: 智能事件路由和负载均衡
- **错误处理**: 完整的事件处理错误机制

### 4. 安全控制
- **权限管理**: 基于角色的插件权限控制
- **资源限制**: CPU、内存等资源使用限制
- **沙箱机制**: 插件运行沙箱环境
- **审计日志**: 插件操作审计和监控

## 🚀 使用示例

### 创建基础插件

```go
package main

import (
    "context"
    "github.com/qiaojinxia/distributed-service/framework/plugin"
)

// 实现插件接口
type HelloPlugin struct {
    plugin.BasePlugin
    config *HelloConfig
}

type HelloConfig struct {
    Message string `yaml:"message"`
    Enabled bool   `yaml:"enabled"`
}

func (p *HelloPlugin) Name() string {
    return "hello-plugin"
}

func (p *HelloPlugin) Version() string {
    return "1.0.0"
}

func (p *HelloPlugin) Description() string {
    return "A simple hello world plugin"
}

func (p *HelloPlugin) Dependencies() []string {
    return []string{"logger-plugin"} // 依赖日志插件
}

func (p *HelloPlugin) Initialize(ctx context.Context, config interface{}) error {
    cfg, ok := config.(*HelloConfig)
    if !ok {
        return plugin.ErrInvalidConfig
    }
    
    p.config = cfg
    p.Logger().Info("Hello plugin initialized", 
        plugin.String("message", cfg.Message))
    return nil
}

func (p *HelloPlugin) Start(ctx context.Context) error {
    if !p.config.Enabled {
        return plugin.ErrPluginDisabled
    }
    
    // 订阅事件
    p.EventBus().Subscribe("user.login", p.handleUserLogin)
    p.EventBus().Subscribe("user.logout", p.handleUserLogout)
    
    p.Logger().Info("Hello plugin started")
    return nil
}

func (p *HelloPlugin) Stop(ctx context.Context) error {
    // 取消事件订阅
    p.EventBus().Unsubscribe("user.login", p.handleUserLogin)
    p.EventBus().Unsubscribe("user.logout", p.handleUserLogout)
    
    p.Logger().Info("Hello plugin stopped")
    return nil
}

func (p *HelloPlugin) handleUserLogin(event plugin.Event) {
    userID := event.Data["user_id"].(string)
    p.Logger().Info("User logged in", plugin.String("user_id", userID))
    
    // 发布欢迎事件
    welcomeEvent := plugin.Event{
        Type: "user.welcome",
        Data: map[string]interface{}{
            "user_id": userID,
            "message": p.config.Message,
        },
    }
    p.EventBus().Publish(welcomeEvent)
}

func (p *HelloPlugin) handleUserLogout(event plugin.Event) {
    userID := event.Data["user_id"].(string)
    p.Logger().Info("User logged out", plugin.String("user_id", userID))
}

// 插件工厂函数
func NewHelloPlugin() plugin.Plugin {
    return &HelloPlugin{}
}

// 注册插件
func init() {
    plugin.Register("hello-plugin", NewHelloPlugin)
}
```

### 插件管理器使用

```go
package main

import (
    "context"
    "time"
    "github.com/qiaojinxia/distributed-service/framework/plugin"
)

func main() {
    // 创建插件管理器
    manager := plugin.NewManager(plugin.ManagerConfig{
        PluginDir:      "./plugins",
        ConfigDir:      "./config/plugins",
        MaxConcurrency: 10,
        Timeout:        30 * time.Second,
    })
    
    // 注册插件配置
    helloConfig := &HelloConfig{
        Message: "欢迎使用我们的服务！",
        Enabled: true,
    }
    
    // 加载插件
    ctx := context.Background()
    err := manager.LoadPlugin(ctx, "hello-plugin", helloConfig)
    if err != nil {
        panic(err)
    }
    
    // 启动所有插件
    err = manager.StartAll(ctx)
    if err != nil {
        panic(err)
    }
    
    // 模拟用户登录事件
    loginEvent := plugin.Event{
        Type: "user.login",
        Data: map[string]interface{}{
            "user_id": "12345",
            "ip":      "192.168.1.100",
        },
    }
    manager.EventBus().Publish(loginEvent)
    
    // 等待一段时间
    time.Sleep(5 * time.Second)
    
    // 优雅关闭
    manager.StopAll(ctx)
}
```

### 高级插件示例

```go
package main

import (
    "context"
    "time"
    "github.com/qiaojinxia/distributed-service/framework/plugin"
    "github.com/qiaojinxia/distributed-service/framework/database"
)

// 数据库插件
type DatabasePlugin struct {
    plugin.BasePlugin
    config *DatabaseConfig
    db     *database.MySQL
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

func (p *DatabasePlugin) Name() string {
    return "database-plugin"
}

func (p *DatabasePlugin) Version() string {
    return "2.1.0"
}

func (p *DatabasePlugin) Initialize(ctx context.Context, config interface{}) error {
    cfg := config.(*DatabaseConfig)
    p.config = cfg
    
    // 初始化数据库连接
    dbConfig := database.MySQLConfig{
        Host:     cfg.Host,
        Port:     cfg.Port,
        Database: cfg.Database,
        Username: cfg.Username,
        Password: cfg.Password,
    }
    
    db, err := database.NewMySQL(dbConfig)
    if err != nil {
        return err
    }
    
    p.db = db
    return nil
}

func (p *DatabasePlugin) Start(ctx context.Context) error {
    // 注册数据库服务
    p.Registry().RegisterService("database", p.db)
    
    // 订阅数据相关事件
    p.EventBus().Subscribe("data.save", p.handleSaveData)
    p.EventBus().Subscribe("data.query", p.handleQueryData)
    
    return nil
}

func (p *DatabasePlugin) handleSaveData(event plugin.Event) {
    table := event.Data["table"].(string)
    data := event.Data["data"].(map[string]interface{})
    
    // 保存数据到数据库
    result := p.db.Table(table).Create(data)
    if result.Error != nil {
        p.Logger().Error("Failed to save data", plugin.Err(result.Error))
        return
    }
    
    // 发布保存成功事件
    p.EventBus().Publish(plugin.Event{
        Type: "data.saved",
        Data: map[string]interface{}{
            "table": table,
            "id":    result.Statement.Dest,
        },
    })
}

func (p *DatabasePlugin) handleQueryData(event plugin.Event) {
    query := event.Data["query"].(string)
    params := event.Data["params"].([]interface{})
    
    var results []map[string]interface{}
    p.db.Raw(query, params...).Scan(&results)
    
    // 发布查询结果
    p.EventBus().Publish(plugin.Event{
        Type: "data.query_result",
        Data: map[string]interface{}{
            "query":   query,
            "results": results,
        },
    })
}
```

### 插件间通信

```go
package main

import (
    "github.com/qiaojinxia/distributed-service/framework/plugin"
)

// 通知插件
type NotificationPlugin struct {
    plugin.BasePlugin
    emailService *EmailService
    smsService   *SMSService
}

func (p *NotificationPlugin) Start(ctx context.Context) error {
    // 监听欢迎事件
    p.EventBus().Subscribe("user.welcome", p.sendWelcomeNotification)
    
    // 监听订单事件
    p.EventBus().Subscribe("order.created", p.sendOrderConfirmation)
    
    return nil
}

func (p *NotificationPlugin) sendWelcomeNotification(event plugin.Event) {
    userID := event.Data["user_id"].(string)
    message := event.Data["message"].(string)
    
    // 获取用户信息（通过调用其他插件服务）
    userService := p.Registry().GetService("user-service")
    user := userService.GetUser(userID)
    
    // 发送欢迎邮件
    p.emailService.Send(EmailMessage{
        To:      user.Email,
        Subject: "欢迎注册",
        Body:    message,
    })
    
    p.Logger().Info("Welcome notification sent", 
        plugin.String("user_id", userID))
}

// 用户服务插件
type UserServicePlugin struct {
    plugin.BasePlugin
    userRepo *UserRepository
}

func (p *UserServicePlugin) Start(ctx context.Context) error {
    // 注册用户服务
    p.Registry().RegisterService("user-service", &UserService{
        repo: p.userRepo,
    })
    
    return nil
}

type UserService struct {
    repo *UserRepository
}

func (s *UserService) GetUser(userID string) *User {
    return s.repo.FindByID(userID)
}
```

## 🔧 配置选项

### 插件管理器配置

```go
type ManagerConfig struct {
    // 基础配置
    PluginDir      string        `yaml:"plugin_dir"`      // 插件目录
    ConfigDir      string        `yaml:"config_dir"`      // 配置目录
    DataDir        string        `yaml:"data_dir"`        // 数据目录
    
    // 并发配置
    MaxConcurrency int           `yaml:"max_concurrency"` // 最大并发数
    Timeout        time.Duration `yaml:"timeout"`         // 操作超时
    
    // 安全配置
    EnableSandbox  bool          `yaml:"enable_sandbox"`  // 启用沙箱
    MaxMemory      int64         `yaml:"max_memory"`      // 最大内存(MB)
    MaxCPU         float64       `yaml:"max_cpu"`         // 最大CPU使用率
    
    // 监控配置
    EnableMetrics  bool          `yaml:"enable_metrics"`  // 启用指标收集
    MetricsPort    int           `yaml:"metrics_port"`    // 指标端口
    
    // 热更新配置
    EnableHotReload bool         `yaml:"enable_hot_reload"` // 启用热重载
    WatchInterval   time.Duration `yaml:"watch_interval"`   // 监控间隔
}
```

### 插件配置

```go
type PluginConfig struct {
    // 基本信息
    Name        string            `yaml:"name"`
    Version     string            `yaml:"version"`
    Enabled     bool              `yaml:"enabled"`
    
    // 依赖配置
    Dependencies []string          `yaml:"dependencies"`
    Conflicts    []string          `yaml:"conflicts"`
    
    // 资源限制
    Resources   ResourceLimits    `yaml:"resources"`
    
    // 权限配置
    Permissions []Permission      `yaml:"permissions"`
    
    // 自定义配置
    Config      map[string]interface{} `yaml:"config"`
}

type ResourceLimits struct {
    Memory    int64   `yaml:"memory"`     // 内存限制(MB)
    CPU       float64 `yaml:"cpu"`        // CPU限制
    Goroutines int    `yaml:"goroutines"` // 协程数限制
    FileHandles int   `yaml:"file_handles"` // 文件句柄限制
}

type Permission struct {
    Resource string   `yaml:"resource"` // 资源类型
    Actions  []string `yaml:"actions"`  // 允许的操作
}
```

### 配置文件示例

```yaml
# config/plugin_manager.yaml
plugin_manager:
  plugin_dir: "./plugins"
  config_dir: "./config/plugins"
  data_dir: "./data/plugins"
  
  max_concurrency: 10
  timeout: "30s"
  
  enable_sandbox: true
  max_memory: 512      # 512MB
  max_cpu: 0.5         # 50% CPU
  
  enable_metrics: true
  metrics_port: 9090
  
  enable_hot_reload: true
  watch_interval: "5s"

# config/plugins/hello-plugin.yaml
hello-plugin:
  name: "hello-plugin"
  version: "1.0.0"
  enabled: true
  
  dependencies:
    - "logger-plugin"
    
  resources:
    memory: 64        # 64MB
    cpu: 0.1          # 10% CPU
    goroutines: 100
    file_handles: 50
    
  permissions:
    - resource: "event_bus"
      actions: ["publish", "subscribe"]
    - resource: "logger"
      actions: ["write"]
      
  config:
    message: "欢迎使用我们的服务！"
    max_notifications: 1000
    retry_count: 3
```

## 📊 监控与指标

### 插件运行指标

```go
type PluginMetrics struct {
    // 基础指标
    Name           string        `json:"name"`
    Status         string        `json:"status"`         // running, stopped, error
    Uptime         time.Duration `json:"uptime"`
    RestartCount   int           `json:"restart_count"`
    
    // 资源使用
    MemoryUsage    int64         `json:"memory_usage"`   // MB
    CPUUsage       float64       `json:"cpu_usage"`      // 百分比
    GoroutineCount int           `json:"goroutine_count"`
    
    // 事件处理
    EventsReceived int64         `json:"events_received"`
    EventsPublished int64        `json:"events_published"`
    EventErrors    int64         `json:"event_errors"`
    
    // 性能指标
    AvgProcessTime time.Duration `json:"avg_process_time"`
    MaxProcessTime time.Duration `json:"max_process_time"`
    ErrorRate      float64       `json:"error_rate"`
}
```

### 系统级指标

```go
type SystemMetrics struct {
    // 插件统计
    TotalPlugins    int `json:"total_plugins"`
    RunningPlugins  int `json:"running_plugins"`
    FailedPlugins   int `json:"failed_plugins"`
    
    // 事件总线
    EventQueueSize  int   `json:"event_queue_size"`
    EventThroughput int64 `json:"event_throughput"` // events/sec
    
    // 资源使用
    TotalMemoryUsage int64   `json:"total_memory_usage"`
    TotalCPUUsage    float64 `json:"total_cpu_usage"`
    
    // 依赖关系
    DependencyConflicts int `json:"dependency_conflicts"`
    CircularDependencies int `json:"circular_dependencies"`
}
```

## 🔍 最佳实践

### 1. 插件设计原则

```go
// ✅ 推荐：单一职责
type AuthPlugin struct {
    plugin.BasePlugin
    // 只负责认证相关功能
}

// ✅ 推荐：依赖注入
func (p *AuthPlugin) Initialize(ctx context.Context, config interface{}) error {
    // 通过依赖注入获取所需服务
    p.cache = p.Registry().GetService("cache")
    p.database = p.Registry().GetService("database")
    return nil
}

// ✅ 推荐：优雅错误处理
func (p *AuthPlugin) handleAuthEvent(event plugin.Event) {
    defer func() {
        if r := recover(); r != nil {
            p.Logger().Error("Auth event handler panic", 
                plugin.Any("panic", r))
        }
    }()
    
    // 处理逻辑...
}
```

### 2. 性能优化

```go
// ✅ 推荐：事件过滤
func (p *MyPlugin) Start(ctx context.Context) error {
    // 只订阅需要的事件
    p.EventBus().SubscribeWithFilter("user.*", p.handleUserEvents, func(event plugin.Event) bool {
        return event.Data["priority"] == "high"
    })
    
    return nil
}

// ✅ 推荐：批量处理
func (p *MyPlugin) handleBatchEvents(events []plugin.Event) {
    // 批量处理提高效率
    for _, event := range events {
        p.processEvent(event)
    }
}

// ✅ 推荐：资源池化
type ConnectionPool struct {
    connections chan *Connection
    maxSize     int
}

func (p *DatabasePlugin) Initialize(ctx context.Context, config interface{}) error {
    p.connectionPool = NewConnectionPool(10)
    return nil
}
```

### 3. 安全考虑

```go
// ✅ 推荐：权限检查
func (p *FilePlugin) writeFile(path string, data []byte) error {
    if !p.HasPermission("file.write") {
        return plugin.ErrPermissionDenied
    }
    
    // 检查路径安全性
    if !isSecurePath(path) {
        return plugin.ErrInvalidPath
    }
    
    return os.WriteFile(path, data, 0644)
}

// ✅ 推荐：输入验证
func (p *APIPlugin) handleAPICall(event plugin.Event) {
    // 验证输入
    if err := p.validateInput(event.Data); err != nil {
        p.Logger().Warn("Invalid input", plugin.Err(err))
        return
    }
    
    // 处理请求...
}
```

## 🚨 故障排查

### 常见问题

**Q1: 插件加载失败**
```go
// 检查插件依赖
func (m *Manager) validateDependencies(plugin Plugin) error {
    for _, dep := range plugin.Dependencies() {
        if !m.IsLoaded(dep) {
            return fmt.Errorf("dependency %s not loaded", dep)
        }
    }
    return nil
}

// 检查版本兼容性
func (m *Manager) checkCompatibility(plugin Plugin) error {
    version := plugin.Version()
    if !m.isVersionCompatible(version) {
        return fmt.Errorf("incompatible version: %s", version)
    }
    return nil
}
```

**Q2: 内存泄漏问题**
```go
// 监控内存使用
func (m *Manager) monitorMemory() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        for _, plugin := range m.plugins {
            usage := plugin.GetMemoryUsage()
            if usage > plugin.GetMemoryLimit() {
                m.Logger().Warn("Plugin memory usage high",
                    plugin.String("plugin", plugin.Name()),
                    plugin.Int64("usage", usage))
            }
        }
    }
}

// 强制垃圾回收
func (p *BasePlugin) forceGC() {
    runtime.GC()
    runtime.GC() // 强制两次GC
}
```

**Q3: 事件处理死锁**
```go
// 带超时的事件处理
func (p *MyPlugin) handleEventWithTimeout(event plugin.Event) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    done := make(chan struct{})
    go func() {
        defer close(done)
        p.processEvent(event)
    }()
    
    select {
    case <-done:
        // 处理完成
    case <-ctx.Done():
        p.Logger().Error("Event processing timeout")
    }
}
```

## 🔮 高级功能

### 插件热更新

```go
func (m *Manager) HotReload(pluginName string) error {
    // 停止旧版本
    if err := m.StopPlugin(pluginName); err != nil {
        return err
    }
    
    // 重新加载
    if err := m.ReloadPlugin(pluginName); err != nil {
        return err
    }
    
    // 启动新版本
    return m.StartPlugin(pluginName)
}
```

### 插件集群

```go
type ClusterManager struct {
    localManager  *Manager
    remoteManagers map[string]*RemoteManager
    coordinator   *Coordinator
}

func (cm *ClusterManager) DistributePlugin(plugin Plugin, nodes []string) error {
    for _, node := range nodes {
        if err := cm.deployToNode(plugin, node); err != nil {
            return err
        }
    }
    return nil
}
```

---

> 插件模块为框架提供了强大的扩展能力，支持动态加载、热更新和分布式部署，实现了真正的微内核架构。