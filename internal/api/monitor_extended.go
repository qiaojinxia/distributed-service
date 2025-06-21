package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSimpleDashboard serves the simple monitoring dashboard HTML page
// @Summary Get simple monitoring dashboard
// @Description Serve the simple monitoring dashboard web interface
// @Tags monitoring
// @Accept json
// @Produce text/html
// @Success 200 {string} string "HTML content"
// @Router /monitor/simple [get]
func (h *MonitorHandler) GetSimpleDashboard(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, getSimpleDashboardHTML())
}

// GetFullDashboard serves the full monitoring dashboard HTML page
// @Summary Get full monitoring dashboard
// @Description Serve the full monitoring dashboard web interface
// @Tags monitoring
// @Accept json
// @Produce text/html
// @Success 200 {string} string "HTML content"
// @Router /monitor/full [get]
func (h *MonitorHandler) GetFullDashboard(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, getDashboardHTML())
}

// GetDetailPage serves the detailed monitoring page for specific type
// @Summary Get detailed monitoring page
// @Description Serve detailed monitoring information for specific monitoring type
// @Tags monitoring
// @Accept json
// @Produce text/html
// @Param type path string true "Monitoring type (system, services, process)"
// @Success 200 {string} string "HTML content"
// @Router /monitor/details/{type} [get]
func (h *MonitorHandler) GetDetailPage(c *gin.Context) {
	monitorType := c.Param("type")

	c.Header("Content-Type", "text/html; charset=utf-8")

	switch monitorType {
	case "system":
		c.String(http.StatusOK, getSystemDetailHTML())
	case "services":
		c.String(http.StatusOK, getServicesDetailHTML())
	case "process":
		c.String(http.StatusOK, getProcessDetailHTML())
	default:
		c.String(http.StatusNotFound, getNotFoundHTML(monitorType))
	}
}

// getSimpleDashboardHTML returns the HTML content for the simple monitoring dashboard
func getSimpleDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>系统监控概览</title>
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
        
        .nav-links {
            display: flex;
            justify-content: center;
            gap: 15px;
            margin-bottom: 30px;
        }
        
        .nav-link {
            background: rgba(255, 255, 255, 0.2);
            color: white;
            text-decoration: none;
            padding: 10px 20px;
            border-radius: 25px;
            transition: all 0.3s ease;
            backdrop-filter: blur(10px);
        }
        
        .nav-link:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
        
        .overview-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .overview-card {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: 15px;
            padding: 20px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
            border: 1px solid rgba(255, 255, 255, 0.2);
            cursor: pointer;
            transition: transform 0.3s ease;
        }
        
        .overview-card:hover {
            transform: translateY(-5px);
        }
        
        .card-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 15px;
        }
        
        .card-title {
            font-size: 1.3rem;
            color: #333;
            font-weight: 600;
        }
        
        .card-icon {
            font-size: 2rem;
        }
        
        .metric-value {
            font-size: 2rem;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 5px;
        }
        
        .metric-label {
            color: #666;
            font-size: 0.9rem;
        }
        
        .services-summary {
            display: flex;
            gap: 20px;
            align-items: center;
        }
        
        .service-stat {
            text-align: center;
        }
        
        .service-count {
            font-size: 1.5rem;
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .refresh-controls {
            text-align: center;
            margin-top: 30px;
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
            margin: 0 10px;
        }
        
        .refresh-btn:hover {
            transform: translateY(-2px);
        }
        
        .auto-refresh {
            color: white;
            margin-top: 15px;
        }
        
        .loading {
            opacity: 0.7;
            pointer-events: none;
        }
        
        @media (max-width: 768px) {
            .overview-grid {
                grid-template-columns: 1fr;
            }
            
            .nav-links {
                flex-direction: column;
                align-items: center;
            }
            
            .services-summary {
                justify-content: space-around;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖥️ 系统监控概览</h1>
            <p>简洁的系统和服务状态监控</p>
        </div>
        
        <div class="nav-links">
            <a href="/monitor/full" class="nav-link">📊 完整监控</a>
            <a href="/monitor/details/system" class="nav-link">💻 系统详情</a>
            <a href="/monitor/details/services" class="nav-link">🔧 服务详情</a>
            <a href="/monitor/details/process" class="nav-link">⚙️ 进程详情</a>
        </div>
        
        <div class="overview-grid">
            <div class="overview-card" onclick="window.location.href='/monitor/details/system'">
                <div class="card-header">
                    <span class="card-title">系统资源</span>
                    <span class="card-icon">💻</span>
                </div>
                <div id="system-summary">
                    <div class="metric-value">-</div>
                    <div class="metric-label">CPU 使用率</div>
                </div>
            </div>
            
            <div class="overview-card" onclick="window.location.href='/monitor/details/services'">
                <div class="card-header">
                    <span class="card-title">服务状态</span>
                    <span class="card-icon">🔧</span>
                </div>
                <div id="services-summary" class="services-summary">
                    <div class="service-stat">
                        <div class="service-count">-</div>
                        <div>健康</div>
                    </div>
                    <div class="service-stat">
                        <div class="service-count">-</div>
                        <div>异常</div>
                    </div>
                </div>
            </div>
            
            <div class="overview-card" onclick="window.location.href='/monitor/details/process'">
                <div class="card-header">
                    <span class="card-title">进程状态</span>
                    <span class="card-icon">⚙️</span>
                </div>
                <div id="process-summary">
                    <div class="metric-value">-</div>
                    <div class="metric-label">内存使用</div>
                </div>
            </div>
        </div>
        
        <div class="refresh-controls">
            <button class="refresh-btn" onclick="refreshData()">🔄 刷新数据</button>
            <div class="auto-refresh">
                <label>
                    <input type="checkbox" id="auto-refresh" checked> 自动刷新 (30秒)
                </label>
            </div>
        </div>
    </div>

    <script>
        let autoRefreshInterval;
        
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }
        
        function updateSystemSummary(data) {
            if (!data || !data.cpu) return;
            
            const container = document.getElementById('system-summary');
            const cpuUsage = (data.cpu.usage || 0).toFixed(1);
            
            container.innerHTML = 
                '<div class="metric-value">' + cpuUsage + '%</div>' +
                '<div class="metric-label">CPU 使用率</div>';
        }
        
        function updateServicesSummary(data) {
            if (!data || !data.summary) return;
            
            const container = document.getElementById('services-summary');
            const healthy = data.summary.healthy || 0;
            const unhealthy = data.summary.unhealthy || 0;
            
            container.innerHTML = 
                '<div class="service-stat">' +
                    '<div class="service-count" style="color: #4CAF50;">' + healthy + '</div>' +
                    '<div>健康</div>' +
                '</div>' +
                '<div class="service-stat">' +
                    '<div class="service-count" style="color: #f44336;">' + unhealthy + '</div>' +
                    '<div>异常</div>' +
                '</div>';
        }
        
        function updateProcessSummary(data) {
            if (!data) return;
            
            const container = document.getElementById('process-summary');
            const memoryMB = data.memory_rss ? (data.memory_rss / 1024 / 1024).toFixed(0) : 0;
            
            container.innerHTML = 
                '<div class="metric-value">' + memoryMB + ' MB</div>' +
                '<div class="metric-label">内存使用</div>';
        }
        
        async function refreshData() {
            const refreshBtn = document.querySelector('.refresh-btn');
            const originalText = refreshBtn.textContent;
            refreshBtn.textContent = '🔄 刷新中...';
            refreshBtn.classList.add('loading');
            
            try {
                const response = await fetch('/api/v1/monitor/stats', {
                    headers: { 'Accept': 'application/json' }
                });
                
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                
                const data = await response.json();
                
                if (data.system) updateSystemSummary(data.system);
                if (data.services) updateServicesSummary(data.services);
                if (data.process) updateProcessSummary(data.process);
                
            } catch (error) {
                console.error('Failed to refresh data:', error);
            } finally {
                refreshBtn.textContent = originalText;
                refreshBtn.classList.remove('loading');
            }
        }
        
        function toggleAutoRefresh() {
            const checkbox = document.getElementById('auto-refresh');
            
            if (checkbox.checked) {
                if (autoRefreshInterval) clearInterval(autoRefreshInterval);
                autoRefreshInterval = setInterval(refreshData, 30000);
            } else {
                if (autoRefreshInterval) {
                    clearInterval(autoRefreshInterval);
                    autoRefreshInterval = null;
                }
            }
        }
        
        // Initialize
        document.addEventListener('DOMContentLoaded', function() {
            refreshData();
            toggleAutoRefresh();
            document.getElementById('auto-refresh').addEventListener('change', toggleAutoRefresh);
        });
    </script>
</body>
</html>`
}

// getSystemDetailHTML returns the HTML content for system detail page
func getSystemDetailHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>系统资源详情</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            margin: 0;
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
        
        .back-link {
            color: white;
            text-decoration: none;
            margin-bottom: 20px;
            display: inline-block;
            padding: 10px 20px;
            background: rgba(255, 255, 255, 0.2);
            border-radius: 25px;
            transition: all 0.3s ease;
        }
        
        .back-link:hover {
            background: rgba(255, 255, 255, 0.3);
        }
        
        .detail-card {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 20px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }
        
        .metric-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        
        .metric-item {
            padding: 15px;
            background: rgba(102, 126, 234, 0.1);
            border-radius: 8px;
            text-align: center;
        }
        
        .metric-value {
            font-size: 1.5rem;
            font-weight: bold;
            color: #667eea;
        }
        
        .metric-label {
            color: #666;
            margin-top: 5px;
        }
        
        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e0e0e0;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 8px;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #4CAF50, #8BC34A);
            transition: width 0.3s ease;
        }
        
        .progress-fill.warning { background: linear-gradient(90deg, #ff9800, #ffc107); }
        .progress-fill.danger { background: linear-gradient(90deg, #f44336, #ff5722); }
        
        .refresh-btn {
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            cursor: pointer;
            font-size: 1rem;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="/monitor" class="back-link">← 返回概览</a>
        
        <div class="header">
            <h1>💻 系统资源详情</h1>
        </div>
        
        <div class="detail-card">
            <h3>CPU 信息</h3>
            <div id="cpu-details" class="metric-grid">
                <div class="metric-item">
                    <div class="metric-value">-</div>
                    <div class="metric-label">总体使用率</div>
                    <div class="progress-bar"><div class="progress-fill" style="width: 0%"></div></div>
                </div>
            </div>
        </div>
        
        <div class="detail-card">
            <h3>内存信息</h3>
            <div id="memory-details" class="metric-grid">
                <div class="metric-item">
                    <div class="metric-value">-</div>
                    <div class="metric-label">已用内存</div>
                    <div class="progress-bar"><div class="progress-fill" style="width: 0%"></div></div>
                </div>
            </div>
        </div>
        
        <div class="detail-card">
            <h3>磁盘信息</h3>
            <div id="disk-details" class="metric-grid">
                <div class="metric-item">
                    <div class="metric-value">-</div>
                    <div class="metric-label">磁盘使用</div>
                    <div class="progress-bar"><div class="progress-fill" style="width: 0%"></div></div>
                </div>
            </div>
        </div>
        
        <button class="refresh-btn" onclick="refreshData()">🔄 刷新数据</button>
    </div>

    <script>
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }
        
        function getProgressClass(percentage) {
            if (percentage > 80) return 'danger';
            if (percentage > 60) return 'warning';
            return '';
        }
        
        async function refreshData() {
            try {
                const response = await fetch('/api/v1/monitor/system');
                const data = await response.json();
                
                // Update CPU
                if (data.cpu) {
                    const cpuContainer = document.getElementById('cpu-details');
                    const cpuUsage = data.cpu.usage || 0;
                    const cpuClass = getProgressClass(cpuUsage);
                    
                    cpuContainer.innerHTML = 
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + cpuUsage.toFixed(1) + '%</div>' +
                            '<div class="metric-label">总体使用率</div>' +
                            '<div class="progress-bar"><div class="progress-fill ' + cpuClass + '" style="width: ' + cpuUsage + '%"></div></div>' +
                        '</div>' +
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + (data.cpu.cores || 0) + '</div>' +
                            '<div class="metric-label">CPU 核心数</div>' +
                        '</div>';
                }
                
                // Update Memory
                if (data.memory) {
                    const memContainer = document.getElementById('memory-details');
                    const memUsedPercent = data.memory.used_percent || 0;
                    const memClass = getProgressClass(memUsedPercent);
                    
                    memContainer.innerHTML = 
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + formatBytes(data.memory.used || 0) + '</div>' +
                            '<div class="metric-label">已用内存</div>' +
                            '<div class="progress-bar"><div class="progress-fill ' + memClass + '" style="width: ' + memUsedPercent + '%"></div></div>' +
                        '</div>' +
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + formatBytes(data.memory.total || 0) + '</div>' +
                            '<div class="metric-label">总内存</div>' +
                        '</div>' +
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + formatBytes(data.memory.available || 0) + '</div>' +
                            '<div class="metric-label">可用内存</div>' +
                        '</div>';
                }
                
                // Update Disk
                if (data.disk) {
                    const diskContainer = document.getElementById('disk-details');
                    const diskUsedPercent = data.disk.used_percent || 0;
                    const diskClass = getProgressClass(diskUsedPercent);
                    
                    diskContainer.innerHTML = 
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + formatBytes(data.disk.used || 0) + '</div>' +
                            '<div class="metric-label">已用空间</div>' +
                            '<div class="progress-bar"><div class="progress-fill ' + diskClass + '" style="width: ' + diskUsedPercent + '%"></div></div>' +
                        '</div>' +
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + formatBytes(data.disk.total || 0) + '</div>' +
                            '<div class="metric-label">总空间</div>' +
                        '</div>' +
                        '<div class="metric-item">' +
                            '<div class="metric-value">' + formatBytes(data.disk.free || 0) + '</div>' +
                            '<div class="metric-label">空闲空间</div>' +
                        '</div>';
                }
                
            } catch (error) {
                console.error('Failed to refresh data:', error);
            }
        }
        
        // Initialize
        document.addEventListener('DOMContentLoaded', refreshData);
        setInterval(refreshData, 30000); // Auto refresh every 30 seconds
    </script>
</body>
</html>`
}

// getServicesDetailHTML returns the HTML content for services detail page
func getServicesDetailHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>服务状态详情</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', 'SF Pro Display', -apple-system, BlinkMacSystemFont, Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            margin: 0;
            line-height: 1.5;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        
        .header {
            text-align: center;
            color: white;
            margin-bottom: 30px;
        }
        
        .header h1 {
            font-size: 2.2rem;
            font-weight: 600;
            margin-bottom: 8px;
        }
        
        .header p {
            font-size: 1rem;
            opacity: 0.9;
        }
        
        .back-link {
            color: white;
            text-decoration: none;
            margin-bottom: 20px;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 10px 20px;
            background: rgba(255, 255, 255, 0.15);
            border-radius: 25px;
            font-size: 1rem;
            transition: all 0.3s ease;
        }
        
        .back-link:hover {
            background: rgba(255, 255, 255, 0.25);
        }
        
        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .service-card {
            background: white;
            border-radius: 12px;
            padding: 20px;
            border-left: 4px solid transparent;
            transition: all 0.3s ease;
        }
        
        .service-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }
        
        .service-card.healthy { 
            border-left-color: #10b981; 
        }
        .service-card.unhealthy { 
            border-left-color: #ef4444; 
        }
        .service-card.degraded { 
            border-left-color: #f59e0b; 
        }
        
        .service-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 15px;
            padding-bottom: 12px;
            border-bottom: 1px solid #e5e7eb;
        }
        
        .service-name {
            font-weight: 600;
            font-size: 1.5rem;
            color: #1f2937;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .service-icon {
            font-size: 1.5rem;
        }
        
        .service-status-badge {
            display: flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            border-radius: 15px;
            font-weight: 500;
            font-size: 0.9rem;
        }
        
        .service-status-badge.healthy {
            background: #f8f9fa;
            color: #6c757d;
            border: 1px solid #dee2e6;
        }
        
        .service-status-badge.unhealthy {
            background: #fef2f2;
            color: #dc2626;
            border: 1px solid #fecaca;
        }
        
        .service-status-badge.degraded {
            background: #fffbeb;
            color: #d97706;
            border: 1px solid #fed7aa;
        }
        
        .status-indicator {
            width: 8px;
            height: 8px;
            border-radius: 50%;
        }
        

        
        .service-latency {
            background: #f9fafb;
            color: #6b7280;
            padding: 5px 10px;
            border-radius: 6px;
            font-size: 0.9rem;
            margin-top: 8px;
            display: inline-flex;
            align-items: center;
            gap: 4px;
        }
        
        .service-message {
            color: #6b7280;
            font-size: 0.9rem;
            margin: 10px 0;
            padding: 8px 12px;
            background: #f9fafb;
            border-radius: 6px;
            border-left: 3px solid #d1d5db;
        }
        
        .detail-sections {
            margin-top: 15px;
        }
        
        .detail-section {
            margin-bottom: 12px;
            background: #fafafa;
            border-radius: 6px;
            overflow: hidden;
        }
        
        .section-header {
            background: #f3f4f6;
            color: #374151;
            padding: 8px 12px;
            font-weight: 500;
            font-size: 0.9rem;
            display: flex;
            align-items: center;
            gap: 6px;
        }
        
        .section-content {
            padding: 12px;
            background: white;
        }
        
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 8px;
        }
        
        .info-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 12px;
            background: #f9fafb;
            border-radius: 4px;
        }
        
        .info-label {
            font-weight: 500;
            color: #6b7280;
            font-size: 0.9rem;
            min-width: 90px;
        }
        
        .info-value {
            font-weight: 500;
            color: #374151;
            font-size: 0.9rem;
            text-align: right;
            flex: 1;
        }
        
        .test-result {
            display: inline-flex;
            align-items: center;
            gap: 3px;
            padding: 3px 6px;
            border-radius: 10px;
            font-size: 0.8rem;
            font-weight: 500;
        }
        
        .test-passed {
            background: #f0fdf4;
            color: #166534;
        }
        
        .test-failed {
            background: #fef2f2;
            color: #dc2626;
        }
        
        .connection-string {
            font-family: 'Monaco', 'Menlo', 'Cascadia Code', 'Courier New', monospace;
            background: #f3f4f6;
            color: #374151;
            padding: 8px 10px;
            border-radius: 4px;
            font-size: 0.85rem;
            word-break: break-all;
            line-height: 1.3;
        }
        
        .refresh-btn {
            background: linear-gradient(135deg, #3b82f6, #1d4ed8);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            cursor: pointer;
            font-size: 1rem;
            font-weight: 600;
            margin: 30px 0;
            transition: all 0.3s ease;
        }
        
        .refresh-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
        }
        
        .loading {
            opacity: 0.6;
            pointer-events: none;
        }
        
        @media (max-width: 1024px) {
            .services-grid {
                grid-template-columns: 1fr;
            }
            
            .service-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 10px;
            }
            
            .info-grid {
                grid-template-columns: 1fr;
            }
            
            .info-item {
                flex-direction: column;
                align-items: flex-start;
                gap: 4px;
            }
            
            .info-value {
                text-align: left;
            }
        }
        
        @media (max-width: 768px) {
            body {
                padding: 15px;
            }
            
            .header h1 {
                font-size: 1.8rem;
            }
            
            .service-name {
                font-size: 1.3rem;
            }
            
            .service-card {
                padding: 16px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="/monitor" class="back-link">← 返回概览</a>
        
        <div class="header">
            <h1>🔧 服务状态详情</h1>
            <p>详细的服务健康状态和连接信息</p>
        </div>
        
        <div id="services-container" class="services-grid">
            <div class="service-card">
                <div class="service-header">
                    <div class="service-name">🔄 加载中...</div>
                </div>
                <div class="service-message">正在获取服务状态和详细信息...</div>
            </div>
        </div>
        
        <div style="text-align: center;">
            <button class="refresh-btn" onclick="refreshData()">🔄 刷新服务数据</button>
        </div>
    </div>

    <script>
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }
        
        function formatNumber(num) {
            return new Intl.NumberFormat().format(num);
        }
        
        function getServiceIcon(serviceName) {
            const icons = {
                'MySQL': '🗄️',
                'Redis': '🔴',
                'RabbitMQ': '🐰',
                'Consul': '🏛️',
                'gRPC': '⚡'
            };
            return icons[serviceName] || '🔧';
        }
        
        function generateServiceDetails(service) {
            let detailsHtml = '';
            const details = service.details || {};
            
            if (service.name === 'MySQL' && details.connection_pool) {
                detailsHtml += 
                    '<div class="detail-section">' +
                        '<div class="section-header">🔗 数据库连接池</div>' +
                        '<div class="section-content">' +
                            '<div class="info-grid">' +
                                '<div class="info-item">' +
                                    '<span class="info-label">开放连接</span>' +
                                    '<span class="info-value">' + (details.connection_pool.open_connections || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">使用中</span>' +
                                    '<span class="info-value">' + (details.connection_pool.in_use || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">空闲连接</span>' +
                                    '<span class="info-value">' + (details.connection_pool.idle || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">等待次数</span>' +
                                    '<span class="info-value">' + formatNumber(details.connection_pool.wait_count || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">等待时长</span>' +
                                    '<span class="info-value">' + ((details.connection_pool.wait_duration_ns || 0) / 1000000).toFixed(2) + ' ms</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">最大连接</span>' +
                                    '<span class="info-value">' + (details.connection_pool.max_open_connections || 0) + '</span>' +
                                '</div>' +
                            '</div>' +
                        '</div>' +
                    '</div>';
                
                if (details.mysql_version || details.query_test) {
                    detailsHtml += 
                        '<div class="detail-section">' +
                            '<div class="section-header">📊 数据库信息</div>' +
                            '<div class="section-content">' +
                            '<div class="info-grid">';
                    
                    if (details.mysql_version) {
                        detailsHtml += 
                            '<div class="info-item">' +
                                '<span class="info-label">MySQL 版本</span>' +
                                '<span class="info-value">' + details.mysql_version + '</span>' +
                            '</div>';
                    }
                    
                    if (details.query_test) {
                        const testClass = details.query_test === 'passed' ? 'test-passed' : 'test-failed';
                        const testIcon = details.query_test === 'passed' ? '✅' : '❌';
                        detailsHtml += 
                            '<div class="info-item">' +
                                '<span class="info-label">查询测试</span>' +
                                '<span class="test-result ' + testClass + '">' + testIcon + ' ' + 
                                (details.query_test === 'passed' ? '通过' : '失败') + '</span>' +
                            '</div>';
                    }
                    
                    detailsHtml += '</div></div></div>';
                }
            }
            
            if (service.name === 'Redis' && details.connection_pool) {
                detailsHtml += 
                    '<div class="detail-section">' +
                        '<div class="section-header">🔗 Redis 连接池</div>' +
                        '<div class="section-content">' +
                            '<div class="info-grid">' +
                                '<div class="info-item">' +
                                    '<span class="info-label">总连接数</span>' +
                                    '<span class="info-value">' + (details.connection_pool.total_conns || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">空闲连接</span>' +
                                    '<span class="info-value">' + (details.connection_pool.idle_conns || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">过期连接</span>' +
                                    '<span class="info-value">' + (details.connection_pool.stale_conns || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">命中次数</span>' +
                                    '<span class="info-value">' + formatNumber(details.connection_pool.hits || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">未命中</span>' +
                                    '<span class="info-value">' + formatNumber(details.connection_pool.misses || 0) + '</span>' +
                                '</div>' +
                                '<div class="info-item">' +
                                    '<span class="info-label">超时次数</span>' +
                                    '<span class="info-value">' + formatNumber(details.connection_pool.timeouts || 0) + '</span>' +
                                '</div>' +
                            '</div>' +
                        '</div>' +
                    '</div>';
                
                if (details.redis_info) {
                    const redisInfo = details.redis_info;
                    detailsHtml += 
                        '<div class="detail-section">' +
                            '<div class="section-header">📊 Redis 服务器信息</div>' +
                            '<div class="section-content">' +
                                '<div class="info-grid">' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">Redis 版本</span>' +
                                        '<span class="info-value">' + (redisInfo.redis_version || 'N/A') + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">运行模式</span>' +
                                        '<span class="info-value">' + (redisInfo.redis_mode || 'N/A') + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">连接客户端</span>' +
                                        '<span class="info-value">' + (redisInfo.connected_clients || 0) + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">已用内存</span>' +
                                        '<span class="info-value">' + (redisInfo.used_memory ? formatBytes(redisInfo.used_memory) : 'N/A') + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">峰值内存</span>' +
                                        '<span class="info-value">' + (redisInfo.used_memory_peak ? formatBytes(redisInfo.used_memory_peak) : 'N/A') + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">总连接数</span>' +
                                        '<span class="info-value">' + formatNumber(redisInfo.total_connections_received || 0) + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">键空间命中</span>' +
                                        '<span class="info-value">' + formatNumber(redisInfo.keyspace_hits || 0) + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">键空间未命中</span>' +
                                        '<span class="info-value">' + formatNumber(redisInfo.keyspace_misses || 0) + '</span>' +
                                    '</div>' +
                                '</div>' +
                            '</div>' +
                        '</div>';
                }
                
                if (details.write_test || details.read_test) {
                    detailsHtml += 
                        '<div class="detail-section">' +
                            '<div class="section-header">🧪 功能测试</div>' +
                            '<div class="section-content">' +
                                '<div class="info-grid">';
                    
                    if (details.write_test) {
                        const writeClass = details.write_test === 'passed' ? 'test-passed' : 'test-failed';
                        const writeIcon = details.write_test === 'passed' ? '✅' : '❌';
                        detailsHtml += 
                            '<div class="info-item">' +
                                '<span class="info-label">写入测试</span>' +
                                '<span class="test-result ' + writeClass + '">' + writeIcon + ' ' + 
                                (details.write_test === 'passed' ? '通过' : '失败') + '</span>' +
                            '</div>';
                    }
                    
                    if (details.read_test) {
                        const readClass = details.read_test === 'passed' ? 'test-passed' : 'test-failed';
                        const readIcon = details.read_test === 'passed' ? '✅' : '❌';
                        detailsHtml += 
                            '<div class="info-item">' +
                                '<span class="info-label">读取测试</span>' +
                                '<span class="test-result ' + readClass + '">' + readIcon + ' ' + 
                                (details.read_test === 'passed' ? '通过' : '失败') + '</span>' +
                            '</div>';
                    }
                    
                    detailsHtml += '</div></div></div>';
                }
            }
            
            if (service.name === 'RabbitMQ') {
                if (details.connection_info) {
                    detailsHtml += 
                        '<div class="detail-section">' +
                            '<div class="section-header">🔗 连接信息</div>' +
                            '<div class="section-content">' +
                                '<div class="info-grid">' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">本地地址</span>' +
                                        '<span class="info-value connection-string">' + (details.connection_info.local_addr || 'N/A') + '</span>' +
                                    '</div>' +
                                    '<div class="info-item">' +
                                        '<span class="info-label">远程地址</span>' +
                                        '<span class="info-value connection-string">' + (details.connection_info.remote_addr || 'N/A') + '</span>' +
                                    '</div>';
                    
                    if (details.connection_info.url_masked) {
                        detailsHtml += 
                            '<div class="info-item" style="grid-column: 1 / -1;">' +
                                '<span class="info-label">连接URL</span>' +
                                '<div class="connection-string">' + details.connection_info.url_masked + '</div>' +
                            '</div>';
                    }
                    
                    detailsHtml += '</div></div></div>';
                }
                
                if (details.queue_test || details.publish_test) {
                    detailsHtml += 
                        '<div class="detail-section">' +
                            '<div class="section-header">🧪 队列操作测试</div>' +
                            '<div class="section-content">' +
                                '<div class="info-grid">';
                    
                    if (details.queue_test) {
                        const queueClass = details.queue_test === 'passed' ? 'test-passed' : 'test-failed';
                        const queueIcon = details.queue_test === 'passed' ? '✅' : '❌';
                        detailsHtml += 
                            '<div class="info-item">' +
                                '<span class="info-label">队列声明</span>' +
                                '<span class="test-result ' + queueClass + '">' + queueIcon + ' ' + 
                                (details.queue_test === 'passed' ? '通过' : '失败') + '</span>' +
                            '</div>';
                    }
                    
                    if (details.publish_test) {
                        const publishClass = details.publish_test === 'passed' ? 'test-passed' : 'test-failed';
                        const publishIcon = details.publish_test === 'passed' ? '✅' : '❌';
                        detailsHtml += 
                            '<div class="info-item">' +
                                '<span class="info-label">消息发布</span>' +
                                '<span class="test-result ' + publishClass + '">' + publishIcon + ' ' + 
                                (details.publish_test === 'passed' ? '通过' : '失败') + '</span>' +
                            '</div>';
                    }
                    
                    detailsHtml += '</div></div></div>';
                }
            }
            
            if (service.name === 'gRPC') {
                detailsHtml += 
                    '<div class="detail-section">' +
                        '<div class="section-header">🔗 gRPC 连接状态</div>' +
                        '<div class="section-content">' +
                            '<div class="info-grid">' +
                                '<div class="info-item">' +
                                    '<span class="info-label">连接状态</span>' +
                                    '<span class="info-value">' + (details.connection_state || 'Unknown') + '</span>' +
                                '</div>';
                
                if (details.connection_info && details.connection_info.address) {
                    detailsHtml += 
                        '<div class="info-item">' +
                            '<span class="info-label">服务地址</span>' +
                            '<span class="info-value connection-string">' + details.connection_info.address + '</span>' +
                        '</div>';
                }
                
                if (details.health_check_available !== undefined) {
                    const healthClass = details.health_check_available ? 'test-passed' : 'test-failed';
                    const healthIcon = details.health_check_available ? '✅' : '❌';
                    detailsHtml += 
                        '<div class="info-item">' +
                            '<span class="info-label">健康检查</span>' +
                            '<span class="test-result ' + healthClass + '">' + healthIcon + ' ' + 
                            (details.health_check_available ? '可用' : '不可用') + '</span>' +
                        '</div>';
                }
                
                detailsHtml += '</div></div></div>';
                
                if (details.health_check_error && details.health_check_error !== '') {
                    detailsHtml += 
                        '<div class="detail-section">' +
                            '<div class="section-header">⚠️ 健康检查错误</div>' +
                            '<div class="section-content">' +
                                '<div class="connection-string" style="color: #ef4444; background: #fef2f2;">' + details.health_check_error + '</div>' +
                            '</div>' +
                        '</div>';
                }
            }
            
            if (service.name === 'Consul' && details.connection_string) {
                detailsHtml += 
                    '<div class="detail-section">' +
                        '<div class="section-header">🔗 Consul 连接</div>' +
                        '<div class="section-content">' +
                            '<div class="info-item">' +
                                '<span class="info-label">服务地址</span>' +
                                '<span class="info-value connection-string">' + details.connection_string + '</span>' +
                            '</div>' +
                        '</div>' +
                    '</div>';
                }
                
                return detailsHtml;
            }
            
            async function refreshData() {
                const refreshBtn = document.querySelector('.refresh-btn');
                const originalText = refreshBtn.textContent;
                refreshBtn.textContent = '🔄 刷新中...';
                refreshBtn.classList.add('loading');
                
                try {
                    const response = await fetch('/api/v1/monitor/services');
                    const data = await response.json();
                    
                    const container = document.getElementById('services-container');
                    let html = '';
                    
                    if (data.services && Array.isArray(data.services)) {
                        data.services.forEach(service => {
                            const serviceIcon = getServiceIcon(service.name);
                            const statusText = service.status.charAt(0).toUpperCase() + service.status.slice(1);
                            const latencyColor = service.latency > 100 ? '#ef4444' : service.latency > 50 ? '#f59e0b' : '#10b981';
                            
                            html += 
                                '<div class="service-card ' + service.status + '">' +
                                    '<div class="service-header">' +
                                        '<div class="service-name">' +
                                            '<span class="service-icon">' + serviceIcon + '</span>' +
                                            service.name +
                                        '</div>' +
                                        '<div class="service-status-badge ' + service.status + '">' +
                                            '<span class="status-indicator ' + service.status + '"></span>' +
                                            statusText +
                                        '</div>' +
                                    '</div>' +
                                    '<div class="service-message">' + service.message + '</div>' +
                                    '<div class="service-latency" style="border-color: ' + latencyColor + ';">⚡ 响应时间: ' + service.latency + ' ms</div>';
                            
                            // Add detailed service information
                            const detailsHtml = generateServiceDetails(service);
                            if (detailsHtml) {
                                html += '<div class="detail-sections">' + detailsHtml + '</div>';
                            }
                            
                            html += '</div>';
                        });
                    } else {
                        html = 
                            '<div class="service-card">' +
                                '<div class="service-header">' +
                                    '<div class="service-name">❌ 无服务数据</div>' +
                                '</div>' +
                                '<div class="service-message">无法获取服务状态信息</div>' +
                            '</div>';
                    }
                    
                    container.innerHTML = html;
                    
                } catch (error) {
                    console.error('Failed to refresh data:', error);
                    document.getElementById('services-container').innerHTML = 
                        '<div class="service-card unhealthy">' +
                            '<div class="service-header">' +
                                '<div class="service-name">❌ 获取数据失败</div>' +
                            '</div>' +
                            '<div class="service-message">网络错误或服务不可用</div>' +
                        '</div>';
                } finally {
                    refreshBtn.textContent = originalText;
                    refreshBtn.classList.remove('loading');
                }
            }
            
            // Initialize
            document.addEventListener('DOMContentLoaded', refreshData);
            setInterval(refreshData, 30000); // Auto refresh every 30 seconds
        </script>
    </body>
</html>`
}

// getProcessDetailHTML returns the HTML content for process detail page
func getProcessDetailHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>进程状态详情</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            margin: 0;
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
        
        .back-link {
            color: white;
            text-decoration: none;
            margin-bottom: 20px;
            display: inline-block;
            padding: 10px 20px;
            background: rgba(255, 255, 255, 0.2);
            border-radius: 25px;
            transition: all 0.3s ease;
        }
        
        .back-link:hover {
            background: rgba(255, 255, 255, 0.3);
        }
        
        .detail-card {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 20px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }
        
        .metric-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        
        .metric-item {
            padding: 15px;
            background: rgba(102, 126, 234, 0.1);
            border-radius: 8px;
            text-align: center;
        }
        
        .metric-value {
            font-size: 1.5rem;
            font-weight: bold;
            color: #667eea;
        }
        
        .metric-label {
            color: #666;
            margin-top: 5px;
        }
        
        .refresh-btn {
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            cursor: pointer;
            font-size: 1rem;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="/monitor" class="back-link">← 返回概览</a>
        
        <div class="header">
            <h1>⚙️ 进程状态详情</h1>
        </div>
        
        <div class="detail-card">
            <h3>进程基本信息</h3>
            <div id="process-basic" class="metric-grid">
                <div class="metric-item">
                    <div class="metric-value">-</div>
                    <div class="metric-label">进程 ID</div>
                </div>
            </div>
        </div>
        
        <div class="detail-card">
            <h3>资源使用情况</h3>
            <div id="process-resources" class="metric-grid">
                <div class="metric-item">
                    <div class="metric-value">-</div>
                    <div class="metric-label">CPU 使用率</div>
                </div>
            </div>
        </div>
        
        <div class="detail-card">
            <h3>Go 运行时信息</h3>
            <div id="go-runtime" class="metric-grid">
                <div class="metric-item">
                    <div class="metric-value">-</div>
                    <div class="metric-label">Goroutines</div>
                </div>
            </div>
        </div>
        
        <button class="refresh-btn" onclick="refreshData()">🔄 刷新数据</button>
    </div>

    <script>
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }
        
        function formatUptime(nanoseconds) {
            const seconds = Math.floor(nanoseconds / 1000000000);
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            return days + '天 ' + hours + '小时 ' + minutes + '分钟';
        }
        
        async function refreshData() {
            try {
                const response = await fetch('/api/v1/monitor/process');
                const data = await response.json();
                
                // Update basic info
                const basicContainer = document.getElementById('process-basic');
                basicContainer.innerHTML = 
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + (data.pid || 'N/A') + '</div>' +
                        '<div class="metric-label">进程 ID</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + (data.num_threads || 0) + '</div>' +
                        '<div class="metric-label">线程数</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + (data.uptime ? formatUptime(data.uptime) : 'N/A') + '</div>' +
                        '<div class="metric-label">运行时间</div>' +
                    '</div>';
                
                // Update resources
                const resourcesContainer = document.getElementById('process-resources');
                resourcesContainer.innerHTML = 
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + (data.cpu_percent ? data.cpu_percent.toFixed(1) : '0.0') + '%</div>' +
                        '<div class="metric-label">CPU 使用率</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + formatBytes(data.memory_rss || 0) + '</div>' +
                        '<div class="metric-label">物理内存</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + formatBytes(data.memory_vms || 0) + '</div>' +
                        '<div class="metric-label">虚拟内存</div>' +
                    '</div>';

                
                // Update Go runtime - fetch from separate endpoint
                try {
                    const runtimeResponse = await fetch('/api/v1/monitor/stats');
                    const runtimeData = await runtimeResponse.json();
                    
                    if (runtimeData.system && runtimeData.system.runtime) {
                        const runtime = runtimeData.system.runtime;
                        const runtimeContainer = document.getElementById('go-runtime');
                        runtimeContainer.innerHTML = 
                            '<div class="metric-item">' +
                                '<div class="metric-value">' + (runtime.goroutines || 0) + '</div>' +
                                '<div class="metric-label">Goroutines</div>' +
                            '</div>' +
                            '<div class="metric-item">' +
                                '<div class="metric-value">' + formatBytes(runtime.heap_alloc || 0) + '</div>' +
                                '<div class="metric-label">堆内存分配</div>' +
                            '</div>' +
                            '<div class="metric-item">' +
                                '<div class="metric-value">' + formatBytes(runtime.heap_sys || 0) + '</div>' +
                                '<div class="metric-label">堆系统内存</div>' +
                            '</div>' +
                            '<div class="metric-item">' +
                                '<div class="metric-value">' + (runtime.num_gc || 0) + '</div>' +
                                '<div class="metric-label">GC 次数</div>' +
                            '</div>';
                    } else {
                        const runtimeContainer = document.getElementById('go-runtime');
                        runtimeContainer.innerHTML = 
                            '<div class="metric-item">' +
                                '<div class="metric-value">N/A</div>' +
                                '<div class="metric-label">运行时数据不可用</div>' +
                            '</div>';
                    }
                } catch (runtimeError) {
                    console.error('Failed to fetch runtime data:', runtimeError);
                    const runtimeContainer = document.getElementById('go-runtime');
                    runtimeContainer.innerHTML = 
                        '<div class="metric-item">' +
                            '<div class="metric-value">-</div>' +
                            '<div class="metric-label">获取运行时数据失败</div>' +
                        '</div>';
                }
                
            } catch (error) {
                console.error('Failed to refresh data:', error);
            }
        }
        
        // Initialize
        document.addEventListener('DOMContentLoaded', refreshData);
        setInterval(refreshData, 30000); // Auto refresh every 30 seconds
    </script>
</body>
</html>`
}

// getNotFoundHTML returns the HTML content for not found page
func getNotFoundHTML(monitorType string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>页面未找到</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            margin: 0;
        }
        
        .error-container {
            text-align: center;
            color: white;
            padding: 40px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            backdrop-filter: blur(10px);
        }
        
        .error-code {
            font-size: 4rem;
            font-weight: bold;
            margin-bottom: 20px;
        }
        
        .error-message {
            font-size: 1.2rem;
            margin-bottom: 30px;
        }
        
        .back-link {
            color: white;
            text-decoration: none;
            padding: 12px 25px;
            background: rgba(255, 255, 255, 0.2);
            border-radius: 25px;
            transition: all 0.3s ease;
        }
        
        .back-link:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-code">404</div>
        <div class="error-message">
            监控类型 "` + monitorType + `" 不存在<br>
            支持的类型: system, services, process
        </div>
        <a href="/monitor" class="back-link">返回监控概览</a>
    </div>
</body>
</html>`
}
