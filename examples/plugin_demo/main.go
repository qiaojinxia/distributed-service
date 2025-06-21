package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"distributed-service/framework/plugin"
	httpTransport "distributed-service/framework/transport/http"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🧩 启动插件化架构演示...")
	log.Println("   展示：插件生命周期管理 + 事件驱动架构 + 定时任务调度")

	// 创建插件管理器
	manager := plugin.NewDefaultManager(&plugin.ManagerConfig{
		EnableDependencyCheck: true,
		HealthCheckInterval:   30 * time.Second,
		MaxStartupTime:        60 * time.Second,
		EnableMetrics:         true,
	})

	// 设置日志
	logger := plugin.NewSimplePluginLogger("plugin-demo")
	manager.SetLogger(logger)

	// 设置插件工厂
	factory := plugin.NewDefaultPluginFactory()
	manager.SetFactory(factory)

	// 🔧 使用插件工厂创建插件
	log.Println("🔧 使用插件工厂创建插件...")

	// 创建配置
	redisConfig := plugin.NewConfigBuilder().
		Set("addrs", []string{"localhost:7000", "localhost:7001", "localhost:7002"}).
		SetInt("pool_size", 10).
		SetInt("max_retries", 3).
		Build()

	kafkaConfig := plugin.NewConfigBuilder().
		Set("brokers", []string{"localhost:9092"}).
		SetString("client_id", "plugin-demo-client").
		SetString("group", "plugin-demo-group").
		SetString("version", "2.8.0").
		Build()

	etcdConfig := plugin.NewConfigBuilder().
		Set("endpoints", []string{"localhost:2379"}).
		SetInt("dial_timeout", 5).
		Build()

	schedulerConfig := plugin.NewConfigBuilder().
		SetBool("auto_start", true).
		SetInt("max_concurrent_tasks", 10).
		Build()

	// 🚀 设置事件监听
	log.Println("📡 设置插件事件监听器...")
	setupEventListeners(manager)

	// 📊 显示插件工厂能力
	displayFactoryCapabilities(factory)

	// ⏰ 初始化定时任务插件
	log.Println("⏰ 初始化定时任务调度器...")
	schedulerPlugin := initializeScheduler(manager, schedulerConfig)

	// 🌐 启动HTTP服务器展示插件管理功能
	log.Println("🌐 启动HTTP服务器...")

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	// 设置插件管理API路由
	setupPluginManagementRoutes(engine, manager)

	// 设置定时任务管理路由
	setupSchedulerRoutes(engine, schedulerPlugin)

	// 设置演示路由
	setupDemoRoutes(engine, redisConfig, kafkaConfig, etcdConfig, schedulerConfig)

	// 启动服务器
	log.Println("✅ 插件化架构演示启动完成!")
	log.Println("📍 服务信息:")
	log.Println("  - HTTP API:          http://localhost:8080")
	log.Println("  - 插件管理:           http://localhost:8080/plugins")
	log.Println("  - 定时任务:           http://localhost:8080/scheduler")
	log.Println("  - 演示页面:           http://localhost:8080/demo")
	log.Println("  - 配置演示:           http://localhost:8080/demo/config")
	log.Println("  - 工厂演示:           http://localhost:8080/demo/factory")
	log.Println("  - 任务演示:           http://localhost:8080/demo/scheduler")

	// 优雅关闭
	defer func() {
		log.Println("🛑 关闭插件系统...")
		if err := manager.StopAll(); err != nil {
			log.Printf("Failed to stop plugins: %v", err)
		}
	}()

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// initializeScheduler 初始化定时任务调度器
func initializeScheduler(manager *plugin.DefaultManager, config plugin.Config) *plugin.SchedulerPlugin {
	// 创建调度器插件
	schedulerPlugin := plugin.NewSchedulerPlugin()

	// 手动注册插件 - 创建注册表实例
	registry := plugin.NewDefaultRegistry()
	manager.SetRegistry(registry)

	if err := registry.Register(schedulerPlugin); err != nil {
		log.Fatalf("Failed to register scheduler plugin: %v", err)
	}

	// 初始化插件
	if err := manager.InitializePlugin("scheduler", config); err != nil {
		log.Fatalf("Failed to initialize scheduler plugin: %v", err)
	}

	// 启动插件
	if err := manager.StartPlugin("scheduler"); err != nil {
		log.Fatalf("Failed to start scheduler plugin: %v", err)
	}

	// 创建一些示例任务
	createExampleTasks(schedulerPlugin)

	return schedulerPlugin
}

// createExampleTasks 创建示例任务
func createExampleTasks(schedulerPlugin *plugin.SchedulerPlugin) {
	log.Println("📝 创建示例定时任务...")

	// 1. 每分钟执行的日志任务
	logTask := plugin.NewTaskBuilder("log-task", "系统日志任务").
		Description("每分钟输出系统状态日志").
		Cron("@every 1m").
		Handler(func(ctx context.Context, task *plugin.Task) error {
			log.Printf("📊 [定时任务] %s 执行中 - 执行次数: %d", task.Name, task.RunCount)
			return nil
		}).
		Metadata("category", "logging").
		Build()

	if err := schedulerPlugin.ScheduleTask(logTask); err != nil {
		log.Printf("Failed to schedule log task: %v", err)
	}

	// 2. 30秒后执行一次的清理任务
	cleanupTask := plugin.NewTaskBuilder("cleanup-task", "系统清理任务").
		Description("30秒后执行一次性系统清理").
		Once(30*time.Second).
		Handler(func(ctx context.Context, task *plugin.Task) error {
			log.Printf("🧹 [一次性任务] %s 执行清理操作", task.Name)
			return nil
		}).
		Metadata("category", "maintenance").
		Build()

	if err := schedulerPlugin.ScheduleTask(cleanupTask); err != nil {
		log.Printf("Failed to schedule cleanup task: %v", err)
	}

	// 3. 每5秒执行的监控任务（最多执行10次）
	monitorTask := plugin.NewTaskBuilder("monitor-task", "系统监控任务").
		Description("每5秒检查系统状态，最多执行10次").
		Interval(5*time.Second).
		MaxRuns(10).
		Handler(func(ctx context.Context, task *plugin.Task) error {
			log.Printf("🔍 [监控任务] %s 检查系统状态 - 剩余执行次数: %d", task.Name, task.Schedule.MaxRuns-task.RunCount)
			return nil
		}).
		Metadata("category", "monitoring").
		Build()

	if err := schedulerPlugin.ScheduleTask(monitorTask); err != nil {
		log.Printf("Failed to schedule monitor task: %v", err)
	}

	log.Println("✅ 示例任务创建完成")
}

// setupSchedulerRoutes 设置定时任务管理路由
func setupSchedulerRoutes(engine *gin.Engine, schedulerPlugin *plugin.SchedulerPlugin) {
	schedulerAPI := engine.Group("/scheduler")
	{
		// 获取所有任务
		schedulerAPI.GET("/tasks", func(c *gin.Context) {
			if schedulerPlugin.GetScheduler() == nil {
				httpTransport.InternalError(c, "Scheduler not available")
				return
			}

			tasks := schedulerPlugin.GetScheduler().GetAllTasks()
			httpTransport.Success(c, gin.H{
				"tasks": tasks,
				"total": len(tasks),
			})
		})

		// 获取指定任务
		schedulerAPI.GET("/tasks/:id", func(c *gin.Context) {
			taskID := c.Param("id")
			if schedulerPlugin.GetScheduler() == nil {
				httpTransport.InternalError(c, "Scheduler not available")
				return
			}

			task := schedulerPlugin.GetScheduler().GetTask(taskID)
			if task == nil {
				httpTransport.NotFound(c, "Task not found")
				return
			}

			httpTransport.Success(c, gin.H{
				"task": task,
			})
		})

		// 创建新任务
		schedulerAPI.POST("/tasks", func(c *gin.Context) {
			var req struct {
				ID          string `json:"id" binding:"required"`
				Name        string `json:"name" binding:"required"`
				Description string `json:"description"`
				Schedule    struct {
					Type     string `json:"type" binding:"required"` // "interval", "once", "cron"
					Interval string `json:"interval,omitempty"`      // e.g., "30s"
					Delay    string `json:"delay,omitempty"`         // e.g., "10s"
					Cron     string `json:"cron,omitempty"`          // e.g., "@every 1m"
					MaxRuns  int64  `json:"max_runs,omitempty"`
				} `json:"schedule" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				httpTransport.BadRequest(c, "Invalid request: "+err.Error())
				return
			}

			// 创建任务
			taskBuilder := plugin.NewTaskBuilder(req.ID, req.Name).
				Description(req.Description).
				Handler(func(ctx context.Context, task *plugin.Task) error {
					log.Printf("🎯 [API任务] %s 执行中", task.Name)
					return nil
				})

			// 设置调度
			switch req.Schedule.Type {
			case "interval":
				if req.Schedule.Interval == "" {
					httpTransport.BadRequest(c, "Interval is required for interval type")
					return
				}
				interval, err := time.ParseDuration(req.Schedule.Interval)
				if err != nil {
					httpTransport.BadRequest(c, "Invalid interval format: "+err.Error())
					return
				}
				taskBuilder.Interval(interval)

			case "once":
				delay := time.Duration(0)
				if req.Schedule.Delay != "" {
					var err error
					delay, err = time.ParseDuration(req.Schedule.Delay)
					if err != nil {
						httpTransport.BadRequest(c, "Invalid delay format: "+err.Error())
						return
					}
				}
				taskBuilder.Once(delay)

			case "cron":
				if req.Schedule.Cron == "" {
					httpTransport.BadRequest(c, "Cron expression is required for cron type")
					return
				}
				taskBuilder.Cron(req.Schedule.Cron)

			default:
				httpTransport.BadRequest(c, "Unsupported schedule type: "+req.Schedule.Type)
				return
			}

			if req.Schedule.MaxRuns > 0 {
				taskBuilder.MaxRuns(req.Schedule.MaxRuns)
			}

			task := taskBuilder.Build()

			// 调度任务
			if err := schedulerPlugin.ScheduleTask(task); err != nil {
				httpTransport.InternalError(c, "Failed to schedule task: "+err.Error())
				return
			}

			httpTransport.Success(c, gin.H{
				"message": "Task scheduled successfully",
				"task":    task,
			})
		})

		// 取消任务
		schedulerAPI.POST("/tasks/:id/cancel", func(c *gin.Context) {
			taskID := c.Param("id")
			if err := schedulerPlugin.GetScheduler().CancelTask(taskID); err != nil {
				httpTransport.InternalError(c, "Failed to cancel task: "+err.Error())
				return
			}

			httpTransport.Success(c, gin.H{
				"message": "Task canceled successfully",
				"task_id": taskID,
			})
		})

		// 暂停任务
		schedulerAPI.POST("/tasks/:id/pause", func(c *gin.Context) {
			taskID := c.Param("id")
			if err := schedulerPlugin.GetScheduler().PauseTask(taskID); err != nil {
				httpTransport.InternalError(c, "Failed to pause task: "+err.Error())
				return
			}

			httpTransport.Success(c, gin.H{
				"message": "Task paused successfully",
				"task_id": taskID,
			})
		})

		// 恢复任务
		schedulerAPI.POST("/tasks/:id/resume", func(c *gin.Context) {
			taskID := c.Param("id")
			if err := schedulerPlugin.GetScheduler().ResumeTask(taskID); err != nil {
				httpTransport.InternalError(c, "Failed to resume task: "+err.Error())
				return
			}

			httpTransport.Success(c, gin.H{
				"message": "Task resumed successfully",
				"task_id": taskID,
			})
		})

		// 根据状态获取任务
		schedulerAPI.GET("/tasks/status/:status", func(c *gin.Context) {
			statusStr := c.Param("status")
			statusInt, err := strconv.Atoi(statusStr)
			if err != nil {
				httpTransport.BadRequest(c, "Invalid status format")
				return
			}

			tasks := schedulerPlugin.GetScheduler().GetTasksByStatus(plugin.TaskStatus(statusInt))
			httpTransport.Success(c, gin.H{
				"tasks":  tasks,
				"status": statusStr,
				"total":  len(tasks),
			})
		})

		// 获取调度器状态
		schedulerAPI.GET("/status", func(c *gin.Context) {
			isRunning := schedulerPlugin.GetScheduler().IsRunning()
			allTasks := schedulerPlugin.GetScheduler().GetAllTasks()

			statusCounts := make(map[string]int)
			for _, task := range allTasks {
				status := task.Status.String()
				statusCounts[status]++
			}

			httpTransport.Success(c, gin.H{
				"running":       isRunning,
				"total_tasks":   len(allTasks),
				"status_counts": statusCounts,
			})
		})
	}
}

// setupEventListeners 设置事件监听器
func setupEventListeners(manager *plugin.DefaultManager) {
	// 监听插件生命周期事件
	manager.SubscribeEvent(plugin.EventPluginStarted, func(event *plugin.Event) error {
		log.Printf("🟢 插件启动: %s", event.Source)
		return nil
	})

	manager.SubscribeEvent(plugin.EventPluginStopped, func(event *plugin.Event) error {
		log.Printf("🔴 插件停止: %s", event.Source)
		return nil
	})

	manager.SubscribeEvent(plugin.EventPluginFailed, func(event *plugin.Event) error {
		log.Printf("❌ 插件失败: %s - %v", event.Source, event.Data)
		return nil
	})

	// 监听定时任务事件
	manager.SubscribeEvent("scheduler.task.scheduled", func(event *plugin.Event) error {
		if taskEvent, ok := event.Data.(*plugin.TaskEvent); ok {
			log.Printf("📅 任务已调度: %s", taskEvent.TaskName)
		}
		return nil
	})

	manager.SubscribeEvent("scheduler.task.started", func(event *plugin.Event) error {
		if taskEvent, ok := event.Data.(*plugin.TaskEvent); ok {
			log.Printf("▶️  任务开始: %s", taskEvent.TaskName)
		}
		return nil
	})

	manager.SubscribeEvent("scheduler.task.completed", func(event *plugin.Event) error {
		if taskEvent, ok := event.Data.(*plugin.TaskEvent); ok {
			log.Printf("✅ 任务完成: %s", taskEvent.TaskName)
		}
		return nil
	})

	manager.SubscribeEvent("scheduler.task.failed", func(event *plugin.Event) error {
		if taskEvent, ok := event.Data.(*plugin.TaskEvent); ok {
			log.Printf("❌ 任务失败: %s - %v", taskEvent.TaskName, taskEvent.Error)
		}
		return nil
	})

	// 发布系统启动事件
	manager.PublishEvent(plugin.NewSystemEvent(plugin.EventSystemStarted, map[string]interface{}{
		"timestamp": time.Now(),
		"message":   "Plugin system with scheduler started",
	}))
}

// displayFactoryCapabilities 显示工厂能力
func displayFactoryCapabilities(factory plugin.PluginFactory) {
	log.Println("🏭 插件工厂支持的类型:")

	types := factory.GetSupportedTypes()
	for _, pluginType := range types {
		log.Printf("  - %s", pluginType)
	}
}

// setupPluginManagementRoutes 设置插件管理路由
func setupPluginManagementRoutes(engine *gin.Engine, manager *plugin.DefaultManager) {
	pluginAPI := engine.Group("/plugins")
	{
		// 获取插件管理器状态
		pluginAPI.GET("/status", func(c *gin.Context) {
			statuses := manager.GetPluginsStatus()

			httpTransport.Success(c, gin.H{
				"plugins": statuses,
				"total":   len(statuses),
			})
		})

		// 发布测试事件
		pluginAPI.POST("/events/test", func(c *gin.Context) {
			event := plugin.NewEventBuilder().
				Type("test.demo").
				Source("demo-api").
				Data(gin.H{
					"message": "Test event from API",
					"time":    time.Now(),
				}).
				Build()

			if err := manager.PublishEvent(event); err != nil {
				httpTransport.InternalError(c, "Failed to publish event: "+err.Error())
				return
			}

			httpTransport.Success(c, gin.H{
				"message": "Test event published successfully",
				"event":   event,
			})
		})
	}
}

// setupDemoRoutes 设置演示路由
func setupDemoRoutes(engine *gin.Engine, redisConfig, kafkaConfig, etcdConfig, schedulerConfig plugin.Config) {
	demo := engine.Group("/demo")
	{
		// 演示主页
		demo.GET("/", func(c *gin.Context) {
			httpTransport.Success(c, gin.H{
				"title": "Plugin Architecture Demo with Scheduler",
				"features": []string{
					"Plugin lifecycle management",
					"Event-driven communication",
					"Configuration management",
					"Factory pattern",
					"Health monitoring",
					"Task scheduling",
					"Cron support",
					"Task state management",
				},
			})
		})

		// 配置演示
		demo.GET("/config", func(c *gin.Context) {
			httpTransport.Success(c, gin.H{
				"redis_config":     redisConfig.All(),
				"kafka_config":     kafkaConfig.All(),
				"etcd_config":      etcdConfig.All(),
				"scheduler_config": schedulerConfig.All(),
			})
		})

		// 工厂演示
		demo.GET("/factory", func(c *gin.Context) {
			factory := plugin.NewDefaultPluginFactory()

			httpTransport.Success(c, gin.H{
				"supported_types": factory.GetSupportedTypes(),
				"description":     "Plugin factory can create instances of these plugin types",
			})
		})

		// 定时任务演示
		demo.GET("/scheduler", func(c *gin.Context) {
			httpTransport.Success(c, gin.H{
				"title": "Task Scheduler Demo",
				"features": []string{
					"Cron expression scheduling",
					"Interval-based scheduling",
					"One-time delayed tasks",
					"Task lifecycle management",
					"Event-driven notifications",
				},
				"schedule_types": map[string]string{
					"cron":     "Use cron expressions like '@every 1m', '@daily'",
					"interval": "Use time.Duration like '30s', '5m', '1h'",
					"once":     "Execute once after specified delay",
				},
				"endpoints": []string{
					"GET /scheduler/tasks - Get all tasks",
					"POST /scheduler/tasks - Create new task",
					"GET /scheduler/tasks/:id - Get specific task",
					"POST /scheduler/tasks/:id/cancel - Cancel task",
					"POST /scheduler/tasks/:id/pause - Pause task",
					"POST /scheduler/tasks/:id/resume - Resume task",
				},
			})
		})

		// 事件系统演示
		demo.GET("/events", func(c *gin.Context) {
			httpTransport.Success(c, gin.H{
				"predefined_events": gin.H{
					"plugin_lifecycle": []string{
						plugin.EventPluginLoaded,
						plugin.EventPluginStarted,
						plugin.EventPluginStopped,
						plugin.EventPluginFailed,
					},
					"system_events": []string{
						plugin.EventSystemStarted,
						plugin.EventSystemStopped,
					},
					"scheduler_events": []string{
						"scheduler.task.scheduled",
						"scheduler.task.started",
						"scheduler.task.completed",
						"scheduler.task.failed",
						"scheduler.task.canceled",
						"scheduler.task.paused",
						"scheduler.task.resumed",
					},
				},
				"custom_events": "You can create custom events using EventBuilder",
			})
		})

		// 插件基础类演示
		demo.GET("/base-plugin", func(c *gin.Context) {
			// 创建一个演示插件
			demoPlugin := plugin.NewPluginBuilder("demo-plugin", "v1.0.0", "Demonstration plugin").
				Dependencies([]string{"logger"}).
				OnInitialize(func(ctx context.Context, config plugin.Config) error {
					log.Println("Demo plugin initialized")
					return nil
				}).
				OnStart(func(ctx context.Context) error {
					log.Println("Demo plugin started")
					return nil
				}).
				Build()

			httpTransport.Success(c, gin.H{
				"plugin_info": gin.H{
					"name":         demoPlugin.Name(),
					"version":      demoPlugin.Version(),
					"description":  demoPlugin.Description(),
					"dependencies": demoPlugin.Dependencies(),
					"status":       demoPlugin.Status().String(),
					"health":       demoPlugin.Health(),
				},
			})
		})
	}
}
