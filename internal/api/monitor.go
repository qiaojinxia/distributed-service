package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"distributed-service/pkg/config"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/monitor"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonitorHandler handles monitoring-related requests
type MonitorHandler struct {
	systemMonitor  *monitor.SystemMonitor
	serviceMonitor *monitor.ServiceMonitor
}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler(cfg *config.Config) *MonitorHandler {
	ctx := context.Background()

	systemMonitor := monitor.NewSystemMonitor(ctx)

	// Construct connection strings from config
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		cfg.MySQL.Username, cfg.MySQL.Password,
		cfg.MySQL.Host, cfg.MySQL.Port,
		cfg.MySQL.Database, cfg.MySQL.Charset)

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.RabbitMQ.Username, cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host, cfg.RabbitMQ.Port,
		cfg.RabbitMQ.VHost)

	consulAddr := fmt.Sprintf("%s:%d", cfg.Consul.Host, cfg.Consul.Port)

	grpcAddr := fmt.Sprintf("localhost:%d", cfg.GRPC.Port)

	serviceConfig := monitor.ServiceConfig{
		MySQL: monitor.MySQLConfig{
			DSN:     mysqlDSN,
			Enabled: true,
		},
		Redis: monitor.RedisConfig{
			Addr:     redisAddr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			Enabled:  true,
		},
		RabbitMQ: monitor.RabbitMQConfig{
			URL:     rabbitmqURL,
			Enabled: true,
		},
		Consul: monitor.ConsulConfig{
			Address: consulAddr,
			Enabled: true,
		},
		GRPC: monitor.GRPCConfig{
			Address: grpcAddr,
			Enabled: true,
		},
	}

	serviceMonitor := monitor.NewServiceMonitor(ctx, serviceConfig)

	return &MonitorHandler{
		systemMonitor:  systemMonitor,
		serviceMonitor: serviceMonitor,
	}
}

// MonitorStats represents combined monitoring statistics
type MonitorStats struct {
	System   *monitor.SystemStats   `json:"system"`
	Services *monitor.ServicesStats `json:"services"`
	Process  *monitor.ProcessStats  `json:"process"`
}

// GetSystemStats returns system resource statistics
// @Summary Get system statistics
// @Description Get current system resource usage including CPU, memory, disk, and network
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} monitor.SystemStats
// @Failure 500 {object} gin.H
// @Router /api/v1/monitor/system [get]
func (h *MonitorHandler) GetSystemStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		logger.Error(ctx, "Failed to get system stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve system statistics",
		})
		return
	}

	logger.Info(ctx, "System stats retrieved successfully")
	c.JSON(http.StatusOK, stats)
}

// GetServicesStats returns service health statistics
// @Summary Get service health statistics
// @Description Get health status of all configured services (MySQL, Redis, RabbitMQ, Consul, gRPC)
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} monitor.ServicesStats
// @Failure 500 {object} gin.H
// @Router /api/v1/monitor/services [get]
func (h *MonitorHandler) GetServicesStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.serviceMonitor.GetServicesStats()
	if err != nil {
		logger.Error(ctx, "Failed to get services stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve services statistics",
		})
		return
	}

	logger.Info(ctx, "Services stats retrieved successfully",
		zap.Int("total_services", stats.Summary.Total),
		zap.Int("healthy", stats.Summary.Healthy),
		zap.Int("unhealthy", stats.Summary.Unhealthy))

	c.JSON(http.StatusOK, stats)
}

// GetProcessStats returns process-specific statistics
// @Summary Get process statistics
// @Description Get statistics for the current process including CPU usage, memory, and thread count
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} monitor.ProcessStats
// @Failure 500 {object} gin.H
// @Router /api/v1/monitor/process [get]
func (h *MonitorHandler) GetProcessStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.systemMonitor.GetProcessStats()
	if err != nil {
		logger.Error(ctx, "Failed to get process stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve process statistics",
		})
		return
	}

	logger.Info(ctx, "Process stats retrieved successfully")
	c.JSON(http.StatusOK, stats)
}

// GetOverallStats returns combined monitoring statistics
// @Summary Get overall monitoring statistics
// @Description Get combined system, services, and process statistics
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} MonitorStats
// @Failure 500 {object} gin.H
// @Router /api/v1/monitor/stats [get]
func (h *MonitorHandler) GetOverallStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get system stats
	systemStats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		logger.Error(ctx, "Failed to get system stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve system statistics",
		})
		return
	}

	// Get services stats
	servicesStats, err := h.serviceMonitor.GetServicesStats()
	if err != nil {
		logger.Error(ctx, "Failed to get services stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve services statistics",
		})
		return
	}

	// Get process stats
	processStats, err := h.systemMonitor.GetProcessStats()
	if err != nil {
		logger.Error(ctx, "Failed to get process stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve process statistics",
		})
		return
	}

	overallStats := MonitorStats{
		System:   systemStats,
		Services: servicesStats,
		Process:  processStats,
	}

	logger.Info(ctx, "Overall stats retrieved successfully")
	c.JSON(http.StatusOK, overallStats)
}

// GetMonitoringDashboard serves the monitoring dashboard HTML page
// @Summary Get monitoring dashboard
// @Description Serve the monitoring dashboard web interface
// @Tags monitoring
// @Accept json
// @Produce text/html
// @Success 200 {string} string "HTML content"
// @Router /monitor [get]
func (h *MonitorHandler) GetMonitoringDashboard(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, getDashboardHTML())
}

// HealthCheck provides a detailed health check endpoint
// @Summary Detailed health check
// @Description Enhanced health check with detailed service status
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 503 {object} gin.H
// @Router /api/v1/monitor/health [get]
func (h *MonitorHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	servicesStats, err := h.serviceMonitor.GetServicesStats()
	if err != nil {
		logger.Error(ctx, "Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Failed to check services health",
		})
		return
	}

	status := "healthy"
	if servicesStats.Summary.Unhealthy > 0 {
		status = "degraded"
	}
	if servicesStats.Summary.Healthy == 0 {
		status = "unhealthy"
	}

	response := gin.H{
		"status":    status,
		"timestamp": time.Now(),
		"services":  servicesStats.Services,
		"summary":   servicesStats.Summary,
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	logger.Info(ctx, "Health check completed",
		zap.String("status", status),
		zap.Int("healthy_services", servicesStats.Summary.Healthy),
		zap.Int("unhealthy_services", servicesStats.Summary.Unhealthy))

	c.JSON(statusCode, response)
}

// GetMetricsHistory returns historical metrics data
// @Summary Get metrics history
// @Description Get historical metrics data for charts (placeholder for future implementation)
// @Tags monitoring
// @Accept json
// @Produce json
// @Param duration query string false "Duration (e.g., 1h, 24h, 7d)" default(1h)
// @Param interval query string false "Data interval (e.g., 1m, 5m, 1h)" default(1m)
// @Success 200 {object} gin.H
// @Router /api/v1/monitor/metrics/history [get]
func (h *MonitorHandler) GetMetricsHistory(c *gin.Context) {
	duration := c.DefaultQuery("duration", "1h")
	interval := c.DefaultQuery("interval", "1m")

	// This is a placeholder implementation
	// In a real application, you would query a time-series database
	response := gin.H{
		"duration": duration,
		"interval": interval,
		"data": gin.H{
			"cpu_usage":    generateMockTimeSeriesData(),
			"memory_usage": generateMockTimeSeriesData(),
			"disk_usage":   generateMockTimeSeriesData(),
			"network_io":   generateMockTimeSeriesData(),
		},
		"message": "This is a placeholder implementation. Integrate with a time-series database for real data.",
	}

	c.JSON(http.StatusOK, response)
}

// generateMockTimeSeriesData generates mock time series data for demonstration
func generateMockTimeSeriesData() []gin.H {
	data := make([]gin.H, 60) // 60 data points
	baseTime := time.Now().Add(-time.Hour)

	for i := 0; i < 60; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		value := 50.0 + float64(i%20) // Mock varying values

		data[i] = gin.H{
			"timestamp": timestamp.Unix(),
			"value":     value,
		}
	}

	return data
}

// getDashboardHTML returns the HTML content for the monitoring dashboard
func getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Enhanced System Monitoring Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            font-size: 14px;
        }
        
        .container {
            max-width: 1600px;
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
        
        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }
        
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .wide-grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 20px;
            margin-bottom: 30px;
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
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        
        .expand-btn {
            background: none;
            border: none;
            font-size: 1.2rem;
            cursor: pointer;
            color: #667eea;
            transition: transform 0.2s;
        }
        
        .expand-btn:hover {
            transform: scale(1.1);
        }
        
        .collapsible {
            max-height: 500px;
            overflow: hidden;
            transition: max-height 0.3s ease;
        }
        
        .collapsed {
            max-height: 80px;
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
        
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-left: 10px;
        }

        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e0e0e0;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 5px;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #4CAF50, #8BC34A);
            transition: width 0.3s ease;
        }
        
        .progress-fill.warning {
            background: linear-gradient(90deg, #ff9800, #ffc107);
        }
        
        .progress-fill.danger {
            background: linear-gradient(90deg, #f44336, #ff5722);
        }
        
        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 15px;
        }
        
        .service-card {
            background: rgba(255, 255, 255, 0.9);
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
            border-left: 4px solid transparent;
        }
        
        .service-card.healthy {
            border-left-color: #4CAF50;
        }
        
        .service-card.unhealthy {
            border-left-color: #f44336;
        }
        
        .service-card.degraded {
            border-left-color: #ff9800;
        }
        
        .service-name {
            font-weight: bold;
            margin-bottom: 15px;
            color: #333;
            font-size: 1.1rem;
        }
        
        .service-status {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 10px;
        }
        
        .service-latency {
            font-size: 0.9rem;
            color: #666;
            background: rgba(102, 126, 234, 0.1);
            padding: 4px 8px;
            border-radius: 4px;
        }
        
        .service-details {
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid #eee;
        }
        
        .detail-group {
            margin-bottom: 12px;
        }
        
        .detail-group-title {
            font-weight: 600;
            color: #555;
            margin-bottom: 6px;
            font-size: 0.9rem;
        }
        
        .detail-item {
            display: flex;
            justify-content: space-between;
            padding: 4px 8px;
            background: rgba(0, 0, 0, 0.03);
            border-radius: 4px;
            font-size: 0.8rem;
            margin-bottom: 3px;
        }
        
        .cpu-cores {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 10px;
            margin-top: 10px;
        }
        
        .cpu-core {
            text-align: center;
            padding: 8px;
            background: rgba(102, 126, 234, 0.1);
            border-radius: 6px;
            font-size: 0.8rem;
        }
        
        .network-interfaces {
            display: grid;
            gap: 10px;
            margin-top: 10px;
        }
        
        .network-interface {
            background: rgba(102, 126, 234, 0.1);
            padding: 10px;
            border-radius: 6px;
        }
        
        .refresh-btn {
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            cursor: pointer;
            font-size: 1rem;
            font-weight: 600;
            transition: transform 0.2s ease;
            margin: 0 10px 10px 0;
        }
        
        .refresh-btn:hover {
            transform: translateY(-2px);
        }
        
        .auto-refresh {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            margin-top: 20px;
            color: white;
        }
        
        .auto-refresh input[type="checkbox"] {
            transform: scale(1.2);
        }
        
        .loading {
            opacity: 0.7;
            pointer-events: none;
        }
        
        .loading-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.3);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 9999;
        }
        
        .loading-spinner {
            background: white;
            padding: 30px;
            border-radius: 15px;
            text-align: center;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
        }
        
        .spinner {
            width: 40px;
            height: 40px;
            border: 4px solid #f3f3f3;
            border-top: 4px solid #667eea;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 0 auto 15px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .metric-badge {
            background: #667eea;
            color: white;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.7rem;
            font-weight: bold;
        }
        
        .error-message {
            color: #f44336;
            font-size: 0.8rem;
            margin-top: 5px;
            font-style: italic;
        }
        
        @media (max-width: 768px) {
            .dashboard-grid {
                grid-template-columns: 1fr;
            }
            
            .services-grid {
                grid-template-columns: 1fr;
            }
            
            .header h1 {
                font-size: 2rem;
            }
            
            .card {
                padding: 20px;
            }
        }
        
        @media (max-width: 480px) {
            .cpu-cores {
                grid-template-columns: repeat(2, 1fr);
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖥️ Enhanced System Monitoring Dashboard</h1>
            <p>Real-time detailed system and service monitoring with connection pool insights</p>
        </div>
        
        <div class="dashboard-grid">
            <div class="card">
                <h3>💻 System Resources <button class="expand-btn" onclick="toggleSection('system-details')">📊</button></h3>
                <div id="system-summary">
                    <div class="stat-item">
                        <span class="stat-label">Loading...</span>
                        <span class="stat-value">Please wait</span>
                    </div>
                </div>
                <div id="system-details" class="collapsible">
                    <div id="cpu-cores" class="cpu-cores"></div>
                    <div id="network-interfaces" class="network-interfaces"></div>
                </div>
            </div>
            
            <div class="card">
                <h3>🔧 Services Health <button class="expand-btn" onclick="toggleSection('services-summary-details')">📈</button></h3>
                <div id="services-summary">
                    <div class="stat-item">
                        <span class="stat-label">Loading...</span>
                        <span class="stat-value">Please wait</span>
                    </div>
                </div>
                <div id="services-summary-details" class="collapsible collapsed">
                    <div id="service-performance-metrics"></div>
                </div>
            </div>
            
            <div class="card">
                <h3>⚙️ Process Runtime <button class="expand-btn" onclick="toggleSection('process-details')">🔍</button></h3>
                <div id="process-summary">
                    <div class="stat-item">
                        <span class="stat-label">Loading...</span>
                        <span class="stat-value">Please wait</span>
                    </div>
                </div>
                <div id="process-details" class="collapsible collapsed">
                    <div id="go-runtime-stats"></div>
                </div>
            </div>
        </div>
        
        <div class="wide-grid">
            <div class="card">
                <h3>🚦 Detailed Service Monitoring <button class="expand-btn" onclick="toggleSection('all-services-details')">🔽</button></h3>
                <div id="services-detail" class="services-grid">
                    <div class="service-card">
                        <div class="service-name">Loading...</div>
                        <div class="service-status">Please wait</div>
                    </div>
                </div>
                <div id="all-services-details" class="collapsible collapsed">
                    <div id="detailed-service-info"></div>
                </div>
            </div>
        </div>
        
        <div class="auto-refresh">
            <button class="refresh-btn" onclick="refreshData()">🔄 Refresh Now</button>
            <button class="refresh-btn" onclick="toggleAllSections()">📋 Toggle All Details</button>
            <label>
                <input type="checkbox" id="auto-refresh" checked> Auto-refresh (30s)
            </label>
        </div>
    </div>

    <script>
        let autoRefreshInterval;
        let allExpanded = false;
        
        function formatBytes(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }
        
        function formatUptime(duration) {
            const seconds = Math.floor(duration / 1000000000);
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            return days + 'd ' + hours + 'h ' + minutes + 'm';
        }
        
        function formatNumber(num) {
            return new Intl.NumberFormat().format(num);
        }
        
        function getProgressClass(percentage) {
            if (percentage > 80) return 'danger';
            if (percentage > 60) return 'warning';
            return '';
        }
        
        function toggleSection(sectionId) {
            const section = document.getElementById(sectionId);
            section.classList.toggle('collapsed');
        }
        
        function toggleAllSections() {
            const sections = document.querySelectorAll('.collapsible');
            allExpanded = !allExpanded;
            
            sections.forEach(section => {
                if (allExpanded) {
                    section.classList.remove('collapsed');
                } else {
                    section.classList.add('collapsed');
                }
            });
            
            document.querySelector('button[onclick="toggleAllSections()"]').textContent = 
                allExpanded ? '📋 Collapse All' : '📋 Expand All';
        }
        
        function updateSystemStats(data) {
            // Check if data exists and has required properties
            if (!data || !data.cpu || !data.memory || !data.disk) {
                console.error('Invalid system stats data:', data);
                return;
            }
            
            // Update summary
            const summaryContainer = document.getElementById('system-summary');
            const cpuUsage = data.cpu.usage || 0;
            const memUsedPercent = data.memory.used_percent || 0;
            const diskUsedPercent = data.disk.used_percent || 0;
            
            const cpuClass = getProgressClass(cpuUsage);
            const memClass = getProgressClass(memUsedPercent);
            const diskClass = getProgressClass(diskUsedPercent);
            
            summaryContainer.innerHTML = 
                '<div class="stat-item">' +
                '<span class="stat-label">CPU Usage</span>' +
                '<span class="stat-value">' + cpuUsage.toFixed(1) + '% <span class="metric-badge">' + (data.cpu.logical_cores || 0) + ' cores</span></span>' +
                '</div>' +
                '<div class="progress-bar">' +
                '<div class="progress-fill ' + cpuClass + '" style="width: ' + cpuUsage + '%"></div>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Memory Usage</span>' +
                '<span class="stat-value">' + formatBytes(data.memory.used || 0) + ' / ' + formatBytes(data.memory.total || 0) + '</span>' +
                '</div>' +
                '<div class="progress-bar">' +
                '<div class="progress-fill ' + memClass + '" style="width: ' + memUsedPercent + '%"></div>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Disk Usage</span>' +
                '<span class="stat-value">' + formatBytes(data.disk.used || 0) + ' / ' + formatBytes(data.disk.total || 0) + '</span>' +
                '</div>' +
                '<div class="progress-bar">' +
                '<div class="progress-fill ' + diskClass + '" style="width: ' + diskUsedPercent + '%"></div>' +
                '</div>';
            
            // Update CPU cores
            const cpuContainer = document.getElementById('cpu-cores');
            let cpuHtml = '<div class="detail-group-title">CPU Cores Usage</div>';
            if (data.cpu && data.cpu.usage_per_cpu && Array.isArray(data.cpu.usage_per_cpu)) {
                data.cpu.usage_per_cpu.forEach((core, index) => {
                    let coreUsage = 0;
                    
                    // Safely extract core usage value
                    if (core && typeof core === 'object') {
                        const values = Object.values(core);
                        if (values.length > 0 && typeof values[0] === 'number') {
                            coreUsage = values[0];
                        } else if (core.usage !== undefined && typeof core.usage === 'number') {
                            coreUsage = core.usage;
                        } else if (core.percent !== undefined && typeof core.percent === 'number') {
                            coreUsage = core.percent;
                        }
                    } else if (typeof core === 'number') {
                        // If core is directly a number value
                        coreUsage = core;
                    }
                    
                    // Ensure coreUsage is a valid number
                    if (isNaN(coreUsage) || coreUsage < 0) {
                        coreUsage = 0;
                    }
                    if (coreUsage > 100) {
                        coreUsage = 100;
                    }
                    
                    cpuHtml += '<div class="cpu-core">' +
                        '<div>Core ' + index + '</div>' +
                        '<div style="font-weight: bold; color: ' + (coreUsage > 80 ? '#f44336' : coreUsage > 60 ? '#ff9800' : '#4CAF50') + '">' +
                        coreUsage.toFixed(1) + '%</div>' +
                        '</div>';
                });
            } else {
                cpuHtml += '<div class="cpu-core">No CPU core data available</div>';
            }
            cpuContainer.innerHTML = cpuHtml;
            
            // Update network interfaces
            const networkContainer = document.getElementById('network-interfaces');
            let networkHtml = '<div class="detail-group-title">Network Interfaces</div>';
            if (data.network && data.network.interfaces && typeof data.network.interfaces === 'object') {
                Object.keys(data.network.interfaces).forEach(ifaceName => {
                    const iface = data.network.interfaces[ifaceName];
                    networkHtml += '<div class="network-interface">' +
                        '<div style="font-weight: bold; margin-bottom: 8px;">' + ifaceName + '</div>' +
                        '<div class="detail-item">' +
                        '<span>Sent:</span><span>' + formatBytes(iface.bytes_sent) + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Received:</span><span>' + formatBytes(iface.bytes_recv) + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Packets Sent:</span><span>' + formatNumber(iface.packets_sent) + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Packets Received:</span><span>' + formatNumber(iface.packets_recv) + '</span>' +
                        '</div>' +
                        '</div>';
                });
            } else {
                networkHtml += '<div class="network-interface">No network interface data available</div>';
            }
            networkContainer.innerHTML = networkHtml;
        }
        
        function updateServicesStats(data) {
            // Check if data exists and has required properties
            if (!data || !data.summary) {
                console.error('Invalid services stats data:', data);
                return;
            }
            
            const container = document.getElementById('services-summary');
            const total = data.summary.total || 0;
            const healthy = data.summary.healthy || 0;
            const unhealthy = data.summary.unhealthy || 0;
            const unknown = data.summary.unknown || 0;
            const healthyPercentage = total > 0 ? (healthy / total * 100) : 0;
            
            container.innerHTML = 
                '<div class="stat-item">' +
                '<span class="stat-label">Total Services</span>' +
                '<span class="stat-value">' + total + ' <span class="metric-badge">active</span></span>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Healthy</span>' +
                '<span class="stat-value">' + healthy + ' <span class="status-indicator healthy"></span></span>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Issues</span>' +
                '<span class="stat-value">' + (unhealthy + unknown) + ' <span class="status-indicator ' + (unhealthy > 0 ? 'unhealthy' : 'unknown') + '"></span></span>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Health Score</span>' +
                '<span class="stat-value">' + healthyPercentage.toFixed(1) + '%</span>' +
                '</div>';
            
            // Update performance metrics
            const perfContainer = document.getElementById('service-performance-metrics');
            let perfHtml = '<div class="detail-group-title">Response Time Performance</div>';
            let totalLatency = 0;
            let healthyCount = 0;
            
            if (data.services && Array.isArray(data.services)) {
                data.services.forEach(service => {
                    if (service.status === 'healthy') {
                        totalLatency += service.latency;
                        healthyCount++;
                    }
                    
                    let perfClass = 'healthy';
                    if (service.latency > 50) perfClass = 'unhealthy';
                    else if (service.latency > 20) perfClass = 'degraded';
                    
                    perfHtml += '<div class="detail-item">' +
                        '<span>' + service.name + '</span>' +
                        '<span><span class="status-indicator ' + perfClass + '"></span> ' + service.latency + 'ms</span>' +
                        '</div>';
                });
            } else {
                perfHtml += '<div class="detail-item">No service data available</div>';
            }
            
            const avgLatency = healthyCount > 0 ? (totalLatency / healthyCount).toFixed(1) : 0;
            perfHtml = '<div class="stat-item">' +
                '<span class="stat-label">Average Response Time</span>' +
                '<span class="stat-value">' + avgLatency + 'ms</span>' +
                '</div>' + perfHtml;
            
            perfContainer.innerHTML = perfHtml;
        }
        
        function updateProcessStats(data) {
            // Check if data exists and has required properties
            if (!data) {
                console.error('Invalid process stats data:', data);
                return;
            }
            
            const summaryContainer = document.getElementById('process-summary');
            summaryContainer.innerHTML = 
                '<div class="stat-item">' +
                '<span class="stat-label">Process ID</span>' +
                '<span class="stat-value">' + (data.pid || 'N/A') + '</span>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">CPU Usage</span>' +
                '<span class="stat-value">' + (data.cpu_percent ? data.cpu_percent.toFixed(1) : '0.0') + '%</span>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Memory (RSS)</span>' +
                '<span class="stat-value">' + formatBytes(data.memory_rss || 0) + '</span>' +
                '</div>' +
                '<div class="stat-item">' +
                '<span class="stat-label">Uptime</span>' +
                '<span class="stat-value">' + (data.uptime ? formatUptime(data.uptime) : 'N/A') + '</span>' +
                '</div>';
            
            // Update Go runtime details
            const runtimeContainer = document.getElementById('go-runtime-stats');
            if (data.runtime) {
                runtimeContainer.innerHTML = 
                    '<div class="detail-group-title">Go Runtime Statistics</div>' +
                    '<div class="detail-item">' +
                    '<span>Goroutines:</span><span>' + formatNumber(data.runtime.num_goroutines || 0) + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Heap Allocated:</span><span>' + formatBytes(data.runtime.heap_alloc || 0) + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Heap System:</span><span>' + formatBytes(data.runtime.heap_sys || 0) + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>GC Runs:</span><span>' + formatNumber(data.runtime.num_gc || 0) + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Memory VMS:</span><span>' + formatBytes(data.memory_vms || 0) + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Threads:</span><span>' + (data.num_threads || 0) + '</span>' +
                    '</div>';
            } else {
                runtimeContainer.innerHTML = '<div class="detail-group-title">Go Runtime Statistics</div><div class="detail-item">No runtime data available</div>';
            }
        }
        
        function updateServicesDetail(data) {
            const container = document.getElementById('services-detail');
            let html = '';
            
            if (data.services && Array.isArray(data.services)) {
                data.services.forEach(service => {
                    html += '<div class="service-card ' + service.status + '">' +
                        '<div class="service-name">' + service.name + '</div>' +
                        '<div class="service-status">' +
                        '<div>' +
                        '<span class="status-indicator ' + service.status + '"></span>' +
                        '<span>' + service.status.charAt(0).toUpperCase() + service.status.slice(1) + '</span>' +
                        '</div>' +
                        '<div class="service-latency">' + service.latency + 'ms</div>' +
                        '</div>' +
                        '<div style="font-size: 0.8rem; color: #666; margin-bottom: 10px;">' +
                        service.message +
                        '</div>';
                    
                    // Add detailed service information
                    if (service.details && Object.keys(service.details).length > 0) {
                        html += '<div class="service-details">' +
                            generateServiceDetails(service) +
                            '</div>';
                    }
                    
                    html += '</div>';
                });
            } else {
                html = '<div class="service-card"><div class="service-name">No service data available</div></div>';
            }
            
            container.innerHTML = html;
        }
        
        function generateServiceDetails(service) {
            let detailsHtml = '';
            const details = service.details;
            
            if (service.name === 'MySQL' && details.connection_pool) {
                detailsHtml += '<div class="detail-group">' +
                    '<div class="detail-group-title">🔗 Connection Pool</div>' +
                    '<div class="detail-item">' +
                    '<span>Open Connections:</span><span>' + details.connection_pool.open_connections + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>In Use:</span><span>' + details.connection_pool.in_use + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Idle:</span><span>' + details.connection_pool.idle + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Wait Count:</span><span>' + formatNumber(details.connection_pool.wait_count) + '</span>' +
                    '</div>' +
                    '</div>';
                
                if (details.mysql_version) {
                    detailsHtml += '<div class="detail-group">' +
                        '<div class="detail-group-title">📊 Database Info</div>' +
                        '<div class="detail-item">' +
                        '<span>Version:</span><span>' + details.mysql_version + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Query Test:</span><span>' + (details.query_test === 'passed' ? '✅ Passed' : '❌ Failed') + '</span>' +
                        '</div>' +
                        '</div>';
                }
            }
            
            if (service.name === 'Redis' && details.connection_pool) {
                detailsHtml += '<div class="detail-group">' +
                    '<div class="detail-group-title">🔗 Connection Pool</div>' +
                    '<div class="detail-item">' +
                    '<span>Total Connections:</span><span>' + details.connection_pool.total_conns + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Idle Connections:</span><span>' + details.connection_pool.idle_conns + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Hits:</span><span>' + formatNumber(details.connection_pool.hits) + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Misses:</span><span>' + formatNumber(details.connection_pool.misses) + '</span>' +
                    '</div>' +
                    '</div>';
                
                if (details.redis_info) {
                    detailsHtml += '<div class="detail-group">' +
                        '<div class="detail-group-title">📊 Redis Info</div>' +
                        '<div class="detail-item">' +
                        '<span>Version:</span><span>' + (details.redis_info.redis_version || 'N/A') + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Connected Clients:</span><span>' + (details.redis_info.connected_clients || 'N/A') + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Used Memory:</span><span>' + (details.redis_info.used_memory ? formatBytes(details.redis_info.used_memory) : 'N/A') + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Keyspace Hits:</span><span>' + formatNumber(details.redis_info.keyspace_hits || 0) + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Keyspace Misses:</span><span>' + formatNumber(details.redis_info.keyspace_misses || 0) + '</span>' +
                        '</div>' +
                        '</div>';
                }
                
                if (details.write_test && details.read_test) {
                    detailsHtml += '<div class="detail-group">' +
                        '<div class="detail-group-title">🧪 Operation Tests</div>' +
                        '<div class="detail-item">' +
                        '<span>Write Test:</span><span>' + (details.write_test === 'passed' ? '✅ Passed' : '❌ Failed') + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Read Test:</span><span>' + (details.read_test === 'passed' ? '✅ Passed' : '❌ Failed') + '</span>' +
                        '</div>' +
                        '</div>';
                }
            }
            
            if (service.name === 'RabbitMQ' && details.connection_info) {
                detailsHtml += '<div class="detail-group">' +
                    '<div class="detail-group-title">🔗 Connection Info</div>' +
                    '<div class="detail-item">' +
                    '<span>Local Address:</span><span>' + details.connection_info.local_addr + '</span>' +
                    '</div>' +
                    '<div class="detail-item">' +
                    '<span>Remote Address:</span><span>' + details.connection_info.remote_addr + '</span>' +
                    '</div>' +
                    '</div>';
                
                if (details.queue_test) {
                    detailsHtml += '<div class="detail-group">' +
                        '<div class="detail-group-title">🧪 Queue Operations</div>' +
                        '<div class="detail-item">' +
                        '<span>Queue Test:</span><span>' + (details.queue_test === 'passed' ? '✅ Passed' : '❌ Failed') + '</span>' +
                        '</div>' +
                        '<div class="detail-item">' +
                        '<span>Publish Test:</span><span>' + (details.publish_test === 'passed' ? '✅ Passed' : '❌ Failed') + '</span>' +
                        '</div>' +
                        '</div>';
                }
            }
            
            if (service.name === 'gRPC' && details.connection_state) {
                detailsHtml += '<div class="detail-group">' +
                    '<div class="detail-group-title">🔗 Connection State</div>' +
                    '<div class="detail-item">' +
                    '<span>State:</span><span>' + details.connection_state + '</span>' +
                    '</div>';
                
                if (details.health_check_available !== undefined) {
                    detailsHtml += '<div class="detail-item">' +
                        '<span>Health Check:</span><span>' + (details.health_check_available ? '✅ Available' : '❌ Not Available') + '</span>' +
                        '</div>';
                }
                
                if (details.connection_info && details.connection_info.address) {
                    detailsHtml += '<div class="detail-item">' +
                        '<span>Address:</span><span>' + details.connection_info.address + '</span>' +
                        '</div>';
                }
                
                detailsHtml += '</div>';
            }
            
            if (service.name === 'Consul' && details.connection_string) {
                detailsHtml += '<div class="detail-group">' +
                    '<div class="detail-group-title">🔗 Connection Info</div>' +
                    '<div class="detail-item">' +
                    '<span>Address:</span><span>' + details.connection_string + '</span>' +
                    '</div>' +
                    '</div>';
            }
            
            return detailsHtml;
        }
        
        async function refreshData(retryCount = 0) {
            const refreshBtn = document.querySelector('.refresh-btn');
            const originalText = refreshBtn.textContent;
            refreshBtn.textContent = '🔄 Refreshing...';
            refreshBtn.classList.add('loading');
            
            // Only show overlay for manual refresh or first load
            if (retryCount === 0) {
                showLoadingOverlay();
            }
            
            try {
                // Create AbortController for timeout control
                const controller = new AbortController();
                const timeoutId = setTimeout(() => controller.abort(), 15000); // Increased to 15 seconds
                
                const response = await fetch('/api/v1/monitor/stats', {
                    signal: controller.signal,
                    headers: {
                        'Accept': 'application/json',
                        'Cache-Control': 'no-cache'
                    }
                });
                
                clearTimeout(timeoutId);
                
                if (!response.ok) {
                    throw new Error('HTTP error! status: ' + response.status);
                }
                
                const data = await response.json();
                
                // Safely update each section with null/undefined checks
                if (data.system) {
                    updateSystemStats(data.system);
                } else {
                    console.warn('System data missing from API response');
                }
                
                if (data.services) {
                    updateServicesStats(data.services);
                    updateServicesDetail(data.services);
                } else {
                    console.warn('Services data missing from API response');
                }
                
                if (data.process) {
                    updateProcessStats(data.process);
                } else {
                    console.warn('Process data missing from API response');
                }
                
                console.log('Enhanced dashboard updated successfully');
                
                // Clear any previous error messages
                const errorElements = document.querySelectorAll('.error-message, .global-error-message');
                errorElements.forEach(el => el.remove());
                
                // Reset retry count on success
                window.refreshRetryCount = 0;
                
            } catch (error) {
                console.error('Failed to refresh data (attempt ' + (retryCount + 1) + '):', error);
                
                // Only show errors for manual refresh or after multiple auto-refresh failures
                if (retryCount === 0 || retryCount >= 2) {
                    if (error.name === 'AbortError') {
                        showErrorMessage('⏱️ 请求超时。服务器可能繁忙，请稍后再试。', 'warning');
                    } else if (error.message.includes('TypeError') || error.message.includes('NetworkError')) {
                        showErrorMessage('🌐 网络连接问题。请检查网络连接后重试。', 'error');
                    } else {
                        showErrorMessage('❌ 监控数据刷新失败。请检查服务状态。', 'error');
                    }
                }
                
                // Auto-retry for network errors (up to 3 times)
                if (retryCount < 2 && !error.name === 'AbortError') {
                    console.log('Retrying refresh in 3 seconds...');
                    setTimeout(() => {
                        refreshData(retryCount + 1);
                    }, 3000);
                    return;
                }
                
            } finally {
                // Hide loading overlay
                if (retryCount === 0) {
                    hideLoadingOverlay();
                }
                refreshBtn.textContent = originalText;
                refreshBtn.classList.remove('loading');
            }
        }
        
        function showLoadingOverlay() {
            // Remove existing overlay if present
            const existingOverlay = document.querySelector('.loading-overlay');
            if (existingOverlay) {
                existingOverlay.remove();
            }
            
            // Create loading overlay
            const overlay = document.createElement('div');
            overlay.className = 'loading-overlay';
            overlay.innerHTML = '' +
                '<div class="loading-spinner">' +
                    '<div class="spinner"></div>' +
                    '<div style="font-weight: 600; color: #333; margin-bottom: 10px;">正在刷新监控数据...</div>' +
                    '<div style="font-size: 0.9rem; color: #666;">系统正在收集实时数据，请稍候（约2-3秒）</div>' +
                '</div>';
            
            document.body.appendChild(overlay);
        }
        
        function hideLoadingOverlay() {
            const overlay = document.querySelector('.loading-overlay');
            if (overlay) {
                overlay.remove();
            }
        }
        
        function showErrorMessage(message, type = 'error') {
            // Remove any existing error messages
            const existingErrors = document.querySelectorAll('.global-error-message');
            existingErrors.forEach(el => el.remove());
            
            // Don't show too many error messages
            if (window.lastErrorTime && Date.now() - window.lastErrorTime < 10000) {
                return;
            }
            window.lastErrorTime = Date.now();
            
            // Create new error message
            const errorDiv = document.createElement('div');
            errorDiv.className = 'global-error-message';
            
            let backgroundColor = 'rgba(244, 67, 54, 0.9)'; // error
            if (type === 'warning') {
                backgroundColor = 'rgba(255, 152, 0, 0.9)';
            } else if (type === 'info') {
                backgroundColor = 'rgba(33, 150, 243, 0.9)';
            }
            
            errorDiv.style.cssText = '' +
                'position: fixed;' +
                'top: 20px;' +
                'right: 20px;' +
                'background: ' + backgroundColor + ';' +
                'color: white;' +
                'padding: 15px 20px;' +
                'border-radius: 8px;' +
                'box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);' +
                'z-index: 1000;' +
                'max-width: 400px;' +
                'animation: slideIn 0.3s ease;' +
                'cursor: pointer;';
            errorDiv.textContent = message;
            
            // Add animation keyframes if not already present
            if (!document.querySelector('#error-animations')) {
                const style = document.createElement('style');
                style.id = 'error-animations';
                style.textContent = '' +
                    '@keyframes slideIn {' +
                        'from { transform: translateX(100%); opacity: 0; }' +
                        'to { transform: translateX(0); opacity: 1; }' +
                    '}';
                document.head.appendChild(style);
            }
            
            document.body.appendChild(errorDiv);
            
            // Auto-remove after 8 seconds for errors, 5 seconds for warnings
            const removeDelay = type === 'error' ? 8000 : 5000;
            setTimeout(() => {
                if (errorDiv.parentNode) {
                    errorDiv.remove();
                }
            }, removeDelay);
            
            // Click to dismiss
            errorDiv.addEventListener('click', () => errorDiv.remove());
        }
        
        function toggleAutoRefresh() {
            const checkbox = document.getElementById('auto-refresh');
            
            if (checkbox.checked) {
                // Clear any existing interval
                if (autoRefreshInterval) {
                    clearInterval(autoRefreshInterval);
                }
                autoRefreshInterval = setInterval(() => {
                    // Only auto-refresh if not already refreshing
                    const refreshBtn = document.querySelector('.refresh-btn');
                    if (!refreshBtn.classList.contains('loading')) {
                        refreshData(0); // Start with retry count 0 for auto-refresh
                    }
                }, 30000);
                console.log('Auto-refresh enabled (30s interval)');
            } else {
                if (autoRefreshInterval) {
                    clearInterval(autoRefreshInterval);
                    autoRefreshInterval = null;
                }
                console.log('Auto-refresh disabled');
            }
        }
        
        // Initialize enhanced dashboard
        document.addEventListener('DOMContentLoaded', function() {
            console.log('Initializing Enhanced Monitoring Dashboard...');
            refreshData();
            toggleAutoRefresh();
            
            document.getElementById('auto-refresh').addEventListener('change', toggleAutoRefresh);
            
            // Add keyboard shortcuts
            document.addEventListener('keydown', function(e) {
                if (e.key === 'r' && (e.ctrlKey || e.metaKey)) {
                    e.preventDefault();
                    refreshData();
                }
                if (e.key === 'e' && (e.ctrlKey || e.metaKey)) {
                    e.preventDefault();
                    toggleAllSections();
                }
            });
            
            console.log('Enhanced dashboard initialized. Use Ctrl+R to refresh, Ctrl+E to expand/collapse all.');
        });
    </script>
</body>
</html>`
}
