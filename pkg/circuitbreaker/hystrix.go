package circuitbreaker

import (
	"context"
	"distributed-service/pkg/logger"
	"distributed-service/pkg/metrics"
	"errors"
	"fmt"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CircuitBreaker 熔断器接口
type CircuitBreaker interface {
	// ExecuteCommand 执行被熔断器保护的命令
	ExecuteCommand(commandName string, run func() (interface{}, error), fallback func(error) (interface{}, error)) (interface{}, error)
	// ExecuteCommandAsync 异步执行被熔断器保护的命令
	ExecuteCommandAsync(commandName string, run func() (interface{}, error), fallback func(error) (interface{}, error)) <-chan interface{}
	// Middleware 熔断器中间件
	Middleware(commandName string, fallback gin.HandlerFunc) gin.HandlerFunc
	// GetCircuitState 获取熔断器状态
	GetCircuitState(commandName string) CircuitState
}

// CircuitState 熔断器状态
type CircuitState struct {
	Name                   string `json:"name"`
	IsOpen                 bool   `json:"is_open"`
	RequestCount           int64  `json:"request_count"`
	ErrorCount             int64  `json:"error_count"`
	ErrorPercentage        int    `json:"error_percentage"`
	RequestVolumeThreshold int    `json:"request_volume_threshold"`
	SleepWindow            int    `json:"sleep_window_ms"`
}

// circuitBreaker 熔断器实现
type circuitBreaker struct{}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker() CircuitBreaker {
	return &circuitBreaker{}
}

// ConfigureCommand 配置熔断器命令
func ConfigureCommand(commandName string, config Config) {
	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:                config.Timeout,
		MaxConcurrentRequests:  config.MaxConcurrentRequests,
		RequestVolumeThreshold: config.RequestVolumeThreshold,
		SleepWindow:            config.SleepWindow,
		ErrorPercentThreshold:  config.ErrorPercentThreshold,
	})

	logger.Info(context.Background(), "Circuit breaker configured",
		zap.String("command", commandName),
		zap.Int("timeout", config.Timeout),
		zap.Int("max_concurrent", config.MaxConcurrentRequests),
		zap.Int("volume_threshold", config.RequestVolumeThreshold),
		zap.Int("error_threshold", config.ErrorPercentThreshold))
}

// Config 熔断器配置
type Config struct {
	// Timeout 超时时间(毫秒)
	Timeout int
	// MaxConcurrentRequests 最大并发请求数
	MaxConcurrentRequests int
	// RequestVolumeThreshold 请求量阈值
	RequestVolumeThreshold int
	// SleepWindow 熔断器打开后的休眠窗口时间(毫秒)
	SleepWindow int
	// ErrorPercentThreshold 错误百分比阈值
	ErrorPercentThreshold int
}

// DefaultConfigs 默认配置
var DefaultConfigs = map[string]Config{
	"database": {
		Timeout:                5000,  // 5秒
		MaxConcurrentRequests:  100,   // 最大100个并发
		RequestVolumeThreshold: 20,    // 20个请求后开始统计
		SleepWindow:            10000, // 10秒休眠窗口
		ErrorPercentThreshold:  50,    // 50%错误率
	},
	"external_api": {
		Timeout:                3000, // 3秒
		MaxConcurrentRequests:  50,   // 最大50个并发
		RequestVolumeThreshold: 10,   // 10个请求后开始统计
		SleepWindow:            5000, // 5秒休眠窗口
		ErrorPercentThreshold:  30,   // 30%错误率
	},
	"cache": {
		Timeout:                1000, // 1秒
		MaxConcurrentRequests:  200,  // 最大200个并发
		RequestVolumeThreshold: 5,    // 5个请求后开始统计
		SleepWindow:            3000, // 3秒休眠窗口
		ErrorPercentThreshold:  60,   // 60%错误率
	},
}

// InitDefaultCircuitBreakers 初始化默认熔断器
func InitDefaultCircuitBreakers() {
	for name, config := range DefaultConfigs {
		ConfigureCommand(name, config)
	}

	logger.Info(context.Background(), "Default circuit breakers initialized",
		zap.Int("count", len(DefaultConfigs)))
}

// ExecuteCommand 执行被熔断器保护的命令
func (cb *circuitBreaker) ExecuteCommand(commandName string, run func() (interface{}, error), fallback func(error) (interface{}, error)) (interface{}, error) {
	var result interface{}
	var cmdErr error

	err := hystrix.Do(commandName, func() error {
		var err error
		result, err = run()
		cmdErr = err
		return err
	}, func(err error) error {
		logger.Warn(context.Background(), "Circuit breaker fallback triggered",
			zap.String("command", commandName),
			zap.Error(err))

		if fallback != nil {
			result, cmdErr = fallback(err)
			return cmdErr
		}

		// 默认降级处理
		result = nil
		cmdErr = fmt.Errorf("service unavailable (circuit breaker open): %w", err)
		return cmdErr
	})

	// 记录指标
	if metrics.RequestCounter != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RequestCounter.WithLabelValues("circuit_breaker", commandName, status).Inc()
	}

	return result, err
}

// ExecuteCommandAsync 异步执行被熔断器保护的命令
func (cb *circuitBreaker) ExecuteCommandAsync(commandName string, run func() (interface{}, error), fallback func(error) (interface{}, error)) <-chan interface{} {
	resultChan := make(chan interface{}, 1)

	errorChan := hystrix.Go(commandName, func() error {
		result, err := run()
		if err == nil {
			resultChan <- result
		}
		return err
	}, func(err error) error {
		logger.Warn(context.Background(), "Async circuit breaker fallback triggered",
			zap.String("command", commandName),
			zap.Error(err))

		if fallback != nil {
			result, fallbackErr := fallback(err)
			if fallbackErr == nil {
				resultChan <- result
			}
			return fallbackErr
		}

		resultChan <- fmt.Errorf("service unavailable (circuit breaker open): %w", err)
		return err
	})

	// 监听错误
	go func() {
		select {
		case err := <-errorChan:
			if err != nil {
				logger.Error(context.Background(), "Async command failed",
					zap.String("command", commandName),
					zap.Error(err))
			}
		}
	}()

	return resultChan
}

// Middleware 熔断器中间件
func (cb *circuitBreaker) Middleware(commandName string, fallback gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := hystrix.Do(commandName, func() error {
			c.Next()
			// 检查响应状态码
			if c.Writer.Status() >= 500 {
				return errors.New("server error")
			}
			return nil
		}, func(err error) error {
			logger.Warn(c.Request.Context(), "HTTP circuit breaker fallback triggered",
				zap.String("command", commandName),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))

			// 清除已有的响应
			c.Header("Content-Type", "application/json")

			if fallback != nil {
				fallback(c)
			} else {
				// 默认降级响应
				c.JSON(503, gin.H{
					"error":   "Service Unavailable",
					"message": "The service is temporarily unavailable due to circuit breaker",
					"code":    "CIRCUIT_BREAKER_OPEN",
				})
			}
			return nil
		})

		if err != nil {
			logger.Error(c.Request.Context(), "Circuit breaker middleware error",
				zap.String("command", commandName),
				zap.Error(err))
		}
	}
}

// GetCircuitState 获取熔断器状态
func (cb *circuitBreaker) GetCircuitState(commandName string) CircuitState {
	circuit, _, _ := hystrix.GetCircuit(commandName)
	if circuit == nil {
		return CircuitState{
			Name:   commandName,
			IsOpen: false,
		}
	}

	// 获取基本状态信息
	return CircuitState{
		Name:            commandName,
		IsOpen:          circuit.IsOpen(),
		RequestCount:    0, // hystrix库限制，无法直接获取
		ErrorCount:      0,
		ErrorPercentage: 0,
	}
}

// MetricsStreamHandler 提供实时指标流
func MetricsStreamHandler() gin.HandlerFunc {
	streamHandler := hystrix.NewStreamHandler()
	return gin.WrapH(streamHandler)
}

// GetAllCircuitStates 获取所有熔断器状态
func GetAllCircuitStates() map[string]CircuitState {
	states := make(map[string]CircuitState)
	cb := &circuitBreaker{}

	for name := range DefaultConfigs {
		states[name] = cb.GetCircuitState(name)
	}

	return states
}

// HealthCheck 熔断器健康检查
func HealthCheck() map[string]interface{} {
	states := GetAllCircuitStates()
	openCircuits := make([]string, 0)

	for name, state := range states {
		if state.IsOpen {
			openCircuits = append(openCircuits, name)
		}
	}

	status := "healthy"
	if len(openCircuits) > 0 {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":         status,
		"open_circuits":  openCircuits,
		"total_circuits": len(states),
		"states":         states,
	}
}
