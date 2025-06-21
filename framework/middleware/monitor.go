package middleware

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled    bool   `json:"enabled"`
	Path       string `json:"path"`
	Dashboard  bool   `json:"dashboard"`
	DetailView bool   `json:"detail_view"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       CPUStats  `json:"cpu"`
	Memory    MemStats  `json:"memory"`
	Runtime   GoStats   `json:"runtime"`
}

// CPUStats CPU统计
type CPUStats struct {
	Count int     `json:"count"`
	Usage float64 `json:"usage_percent"`
}

// MemStats 内存统计
type MemStats struct {
	Alloc      uint64 `json:"alloc_bytes"`
	TotalAlloc uint64 `json:"total_alloc_bytes"`
	Sys        uint64 `json:"sys_bytes"`
	NumGC      uint32 `json:"num_gc"`
}

// GoStats Go运行时统计
type GoStats struct {
	Version    string `json:"version"`
	Goroutines int    `json:"goroutines"`
	GOOS       string `json:"goos"`
	GOARCH     string `json:"goarch"`
}

// DefaultMonitorConfig 默认监控配置
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		Enabled:    true,
		Path:       "/monitor",
		Dashboard:  true,
		DetailView: true,
	}
}

// MonitoringRoutes 添加监控路由
func MonitoringRoutes(r *gin.Engine, config ...MonitorConfig) {
	cfg := DefaultMonitorConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	if !cfg.Enabled {
		return
	}

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"framework": "distributed-service",
			"version":   "v2.1.0",
		})
	})

	// 系统统计API
	r.GET("/api/monitor/stats", func(c *gin.Context) {
		stats := getSystemStats()
		c.JSON(http.StatusOK, stats)
	})

	// 监控仪表盘
	if cfg.Dashboard {
		r.GET(cfg.Path, func(c *gin.Context) {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, getMonitorDashboard())
		})
	}

	// 详细视图
	if cfg.DetailView {
		r.GET(cfg.Path+"/details", func(c *gin.Context) {
			stats := getSystemStats()
			c.JSON(http.StatusOK, gin.H{
				"system": stats,
				"components": gin.H{
					"http":    true,
					"grpc":    true,
					"metrics": true,
					"tracing": true,
					"lock":    true,
				},
			})
		})
	}
}

// getSystemStats 获取系统统计信息
func getSystemStats() SystemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemStats{
		Timestamp: time.Now(),
		CPU: CPUStats{
			Count: runtime.NumCPU(),
			Usage: 0.0, // 简化版本，可以后续扩展
		},
		Memory: MemStats{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		Runtime: GoStats{
			Version:    runtime.Version(),
			Goroutines: runtime.NumGoroutine(),
			GOOS:       runtime.GOOS,
			GOARCH:     runtime.GOARCH,
		},
	}
}

// getMonitorDashboard 获取监控仪表盘HTML
func getMonitorDashboard() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🚀 分布式服务框架 - 监控面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            color: white;
            margin-bottom: 30px;
        }
        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
        }
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
        }
        .card {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 25px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        .card h3 {
            color: #333;
            margin-bottom: 20px;
            font-size: 1.3rem;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .stat-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 12px;
            padding: 8px 12px;
            background: rgba(102, 126, 234, 0.1);
            border-radius: 6px;
            font-size: 0.9rem;
        }
        .stat-label {
            font-weight: 600;
            color: #555;
        }
        .stat-value {
            font-weight: bold;
            color: #333;
        }
        .refresh-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.9rem;
            margin-top: 15px;
        }
        .refresh-btn:hover {
            background: #5a67d8;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 分布式服务框架</h1>
            <p>系统监控面板 v2.1.0</p>
        </div>
        
        <div class="dashboard-grid">
            <div class="card">
                <h3>📊 系统信息</h3>
                <div id="system-stats">
                    <div class="stat-item">
                        <span class="stat-label">状态</span>
                        <span class="stat-value">正在加载...</span>
                    </div>
                </div>
                <button class="refresh-btn" onclick="refreshStats()">刷新数据</button>
            </div>
            
            <div class="card">
                <h3>🔧 框架组件</h3>
                <div id="components">
                    <div class="stat-item">
                        <span class="stat-label">组件状态</span>
                        <span class="stat-value">正在加载...</span>
                    </div>
                </div>
            </div>
            
            <div class="card">
                <h3>🔗 快速链接</h3>
                <div class="stat-item">
                    <span class="stat-label">健康检查</span>
                    <span class="stat-value"><a href="/health" target="_blank">查看</a></span>
                </div>
                <div class="stat-item">
                    <span class="stat-label">API文档</span>
                    <span class="stat-value"><a href="/swagger/" target="_blank">查看</a></span>
                </div>
                <div class="stat-item">
                    <span class="stat-label">详细监控</span>
                    <span class="stat-value"><a href="/monitor/details" target="_blank">查看</a></span>
                </div>
            </div>
        </div>
    </div>

    <script>
        function formatBytes(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function refreshStats() {
            fetch('/api/monitor/stats')
                .then(response => response.json())
                .then(data => {
                    const systemStats = document.getElementById('system-stats');
                    systemStats.innerHTML = ` + "`" + `
                        <div class="stat-item">
                            <span class="stat-label">Go版本</span>
                            <span class="stat-value">${data.runtime.version}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">CPU核心</span>
                            <span class="stat-value">${data.cpu.count}核</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">内存使用</span>
                            <span class="stat-value">${formatBytes(data.memory.alloc)}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Goroutines</span>
                            <span class="stat-value">${data.runtime.goroutines}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">GC次数</span>
                            <span class="stat-value">${data.memory.num_gc}</span>
                        </div>
                    ` + "`" + `;
                })
                .catch(error => {
                    console.error('获取统计信息失败:', error);
                });

            fetch('/monitor/details')
                .then(response => response.json())
                .then(data => {
                    const components = document.getElementById('components');
                    const comp = data.components;
                    components.innerHTML = ` + "`" + `
                        <div class="stat-item">
                            <span class="stat-label">HTTP服务</span>
                            <span class="stat-value">${comp.http ? '✅ 启用' : '❌ 禁用'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">gRPC服务</span>
                            <span class="stat-value">${comp.grpc ? '✅ 启用' : '❌ 禁用'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">监控指标</span>
                            <span class="stat-value">${comp.metrics ? '✅ 启用' : '❌ 禁用'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">链路追踪</span>
                            <span class="stat-value">${comp.tracing ? '✅ 启用' : '❌ 禁用'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">分布式锁</span>
                            <span class="stat-value">${comp.lock ? '✅ 启用' : '❌ 禁用'}</span>
                        </div>
                    ` + "`" + `;
                })
                .catch(error => {
                    console.error('获取组件信息失败:', error);
                });
        }

        // 页面加载时刷新数据
        document.addEventListener('DOMContentLoaded', refreshStats);
        
        // 每30秒自动刷新
        setInterval(refreshStats, 30000);
    </script>
</body>
</html>`
}
