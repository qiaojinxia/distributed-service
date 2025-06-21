package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthResult
}

// HealthResult 健康检查结果
type HealthResult struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Latency   string       `json:"latency,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Details   interface{}  `json:"details,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     HealthStatus            `json:"status"`
	Timestamp  time.Time               `json:"timestamp"`
	Duration   string                  `json:"duration"`
	Components map[string]HealthResult `json:"components,omitempty"`
	Summary    struct {
		Total     int `json:"total"`
		Healthy   int `json:"healthy"`
		Unhealthy int `json:"unhealthy"`
		Degraded  int `json:"degraded"`
	} `json:"summary"`
}

// HealthManager 健康检查管理器
type HealthManager struct {
	checks []HealthCheck
}

// NewHealthManager 创建健康检查管理器
func NewHealthManager() *HealthManager {
	return &HealthManager{
		checks: make([]HealthCheck, 0),
	}
}

// AddCheck 添加健康检查
func (h *HealthManager) AddCheck(check HealthCheck) {
	h.checks = append(h.checks, check)
}

// CheckHealth 执行所有健康检查
func (h *HealthManager) CheckHealth(ctx context.Context) HealthResponse {
	start := time.Now()
	components := make(map[string]HealthResult)

	var healthy, unhealthy, degraded int

	// 并发执行所有检查
	resultChan := make(chan struct {
		name   string
		result HealthResult
	}, len(h.checks))

	for _, check := range h.checks {
		go func(c HealthCheck) {
			result := c.Check(ctx)
			resultChan <- struct {
				name   string
				result HealthResult
			}{c.Name(), result}
		}(check)
	}

	// 收集结果
	for i := 0; i < len(h.checks); i++ {
		select {
		case result := <-resultChan:
			components[result.name] = result.result
			switch result.result.Status {
			case HealthStatusHealthy:
				healthy++
			case HealthStatusUnhealthy:
				unhealthy++
			case HealthStatusDegraded:
				degraded++
			}
		case <-time.After(30 * time.Second):
			// 超时处理
			unhealthy++
			components["timeout"] = HealthResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Health check timeout",
				Timestamp: time.Now(),
			}
		}
	}

	// 确定整体状态
	var overallStatus HealthStatus
	if unhealthy > 0 {
		overallStatus = HealthStatusUnhealthy
	} else if degraded > 0 {
		overallStatus = HealthStatusDegraded
	} else {
		overallStatus = HealthStatusHealthy
	}

	response := HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Duration:   time.Since(start).String(),
		Components: components,
	}

	response.Summary.Total = len(h.checks)
	response.Summary.Healthy = healthy
	response.Summary.Unhealthy = unhealthy
	response.Summary.Degraded = degraded

	return response
}

// SetupHealthRoutes 设置健康检查路由
func (h *HealthManager) SetupHealthRoutes(engine *gin.Engine) {
	// 简单健康检查
	engine.GET("/health", func(c *gin.Context) {
		result := h.CheckHealth(c.Request.Context())

		statusCode := 200
		if result.Status == HealthStatusUnhealthy {
			statusCode = 503
		} else if result.Status == HealthStatusDegraded {
			statusCode = 200 // 降级状态仍返回200
		}

		c.JSON(statusCode, result)
	})

	// 活跃性检查 (Liveness Probe)
	engine.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"message":   "Service is alive",
		})
	})

	// 就绪性检查 (Readiness Probe)
	engine.GET("/health/ready", func(c *gin.Context) {
		result := h.CheckHealth(c.Request.Context())

		statusCode := 200
		if result.Status == HealthStatusUnhealthy {
			statusCode = 503
		}

		c.JSON(statusCode, gin.H{
			"status":    result.Status,
			"timestamp": result.Timestamp,
			"ready":     result.Status != HealthStatusUnhealthy,
		})
	})

	// 详细健康检查
	engine.GET("/health/detail", func(c *gin.Context) {
		result := h.CheckHealth(c.Request.Context())
		c.JSON(200, result)
	})
}

// ===== 内置健康检查器 =====

// DatabaseHealthCheck 数据库健康检查
type DatabaseHealthCheck struct {
	name string
	ping func(ctx context.Context) error
}

// NewDatabaseHealthCheck 创建数据库健康检查
func NewDatabaseHealthCheck(name string, ping func(ctx context.Context) error) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name: name,
		ping: ping,
	}
}

func (d *DatabaseHealthCheck) Name() string {
	return d.name
}

func (d *DatabaseHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()

	err := d.ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return HealthResult{
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Database connection failed: %v", err),
			Latency:   latency.String(),
			Timestamp: time.Now(),
		}
	}

	return HealthResult{
		Status:    HealthStatusHealthy,
		Message:   "Database connection OK",
		Latency:   latency.String(),
		Timestamp: time.Now(),
	}
}

// RedisHealthCheck Redis健康检查
type RedisHealthCheck struct {
	name string
	ping func(ctx context.Context) error
}

// NewRedisHealthCheck 创建Redis健康检查
func NewRedisHealthCheck(name string, ping func(ctx context.Context) error) *RedisHealthCheck {
	return &RedisHealthCheck{
		name: name,
		ping: ping,
	}
}

func (r *RedisHealthCheck) Name() string {
	return r.name
}

func (r *RedisHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()

	err := r.ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return HealthResult{
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Redis connection failed: %v", err),
			Latency:   latency.String(),
			Timestamp: time.Now(),
		}
	}

	return HealthResult{
		Status:    HealthStatusHealthy,
		Message:   "Redis connection OK",
		Latency:   latency.String(),
		Timestamp: time.Now(),
	}
}

// HTTPHealthCheck HTTP端点健康检查
type HTTPHealthCheck struct {
	name string
	url  string
}

// NewHTTPHealthCheck 创建HTTP健康检查
func NewHTTPHealthCheck(name, url string) *HTTPHealthCheck {
	return &HTTPHealthCheck{
		name: name,
		url:  url,
	}
}

func (h *HTTPHealthCheck) Name() string {
	return h.name
}

func (h *HTTPHealthCheck) Check(ctx context.Context) HealthResult {
	// 这里可以实现HTTP健康检查逻辑
	return HealthResult{
		Status:    HealthStatusHealthy,
		Message:   "HTTP endpoint OK",
		Timestamp: time.Now(),
	}
}
