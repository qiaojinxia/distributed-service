package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"

	"distributed-service/framework/config"
	"distributed-service/framework/logger"
	"distributed-service/framework/protection"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SentinelProtectionMiddleware Sentinel保护中间件
type SentinelProtectionMiddleware struct {
	sentinelManager *protection.SentinelManager
	enabled         bool
}

// NewSentinelProtectionMiddleware 创建Sentinel保护中间件
func NewSentinelProtectionMiddleware(ctx context.Context, cfg *config.ProtectionConfig) (*SentinelProtectionMiddleware, error) {
	if cfg == nil || !cfg.Enabled {
		return &SentinelProtectionMiddleware{enabled: false}, nil
	}

	sentinelManager := protection.NewSentinelManager()

	// 初始化Sentinel
	if err := sentinelManager.Init(); err != nil {
		logger.Error(ctx, "Failed to initialize Sentinel manager", zap.Error(err))
		return &SentinelProtectionMiddleware{enabled: false}, nil
	}

	// 加载限流规则
	for _, rule := range cfg.RateLimitRules {
		if rule.Enabled {
			err := sentinelManager.ConfigureFlowRuleWithConfig(rule)
			if err != nil {
				logger.Error(ctx, "Failed to configure flow rule",
					zap.String("name", rule.Name),
					zap.String("resource", rule.Resource),
					zap.Float64("threshold", rule.Threshold),
					zap.Error(err))
			} else {
				logger.Info(ctx, "Flow rule configured",
					zap.String("name", rule.Name),
					zap.String("resource", rule.Resource),
					zap.Float64("threshold", rule.Threshold),
					zap.Uint32("interval_ms", rule.StatIntervalMs),
					zap.Bool("is_wildcard", strings.Contains(rule.Resource, "*") || strings.Contains(rule.Resource, ",")))
			}
		}
	}

	// 加载熔断器规则
	for _, cb := range cfg.CircuitBreakers {
		if cb.Enabled {
			err := sentinelManager.ConfigureCircuitBreakerWithConfig(cb)
			if err != nil {
				logger.Error(ctx, "Failed to configure circuit breaker rule",
					zap.String("name", cb.Name),
					zap.String("resource", cb.Resource),
					zap.String("strategy", cb.Strategy),
					zap.Error(err))
			} else {
				logger.Info(ctx, "Circuit breaker configured",
					zap.String("name", cb.Name),
					zap.String("resource", cb.Resource),
					zap.String("strategy", cb.Strategy),
					zap.Float64("threshold", cb.Threshold),
					zap.Bool("is_wildcard", strings.Contains(cb.Resource, "*") || strings.Contains(cb.Resource, ",")))
			}
		}
	}

	middleware := &SentinelProtectionMiddleware{
		sentinelManager: sentinelManager,
		enabled:         true,
	}

	logger.Info(ctx, "Sentinel protection middleware initialized",
		zap.Int("flow_rules", len(cfg.RateLimitRules)),
		zap.Int("circuit_breakers", len(cfg.CircuitBreakers)))

	return middleware, nil
}

// mapURLToResourceName 将URL路径映射为Sentinel资源名 - 统一的映射逻辑
func mapURLToResourceName(path string) string {
	// 如果是完整的URL路径（以/开头）
	if strings.HasPrefix(path, "/") {
		// 去掉开头的斜杠
		resource := strings.TrimPrefix(path, "/")

		// 如果是根路径，使用"root"
		if resource == "" {
			resource = "root"
		}

		// 支持通配符匹配的特殊处理
		if strings.HasPrefix(path, "/api/v1/users") && path != "/api/v1/users" {
			// 用户相关接口的子路径都映射到 api_v1_users
			resource = "api_v1_users"
		} else {
			// 将斜杠替换为下划线以符合Sentinel资源名规范
			resource = strings.ReplaceAll(resource, "/", "_")
		}

		return resource
	}

	// 其他格式（如grpc_user_service）保持原样，只替换特殊字符
	resource := strings.ReplaceAll(path, ":", "_")
	resource = strings.ReplaceAll(resource, "/", "_")
	return resource
}

// HTTPMiddleware HTTP中间件
func (spm *SentinelProtectionMiddleware) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !spm.enabled {
			c.Next()
			return
		}

		// 使用URL路径作为资源名（支持通配符匹配）
		resource := c.Request.URL.Path

		logger.Debug(c.Request.Context(), "Processing request with Sentinel",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("resource", resource))

		// 使用Sentinel Entry进行保护
		entry, blockErr := spm.sentinelManager.Entry(c.Request.Context(), resource, base.Inbound)

		if blockErr != nil {
			// 根据错误类型区分限流和熔断
			spm.handleBlockError(c, resource, blockErr)
			return
		}

		defer entry.Exit()

		// 执行请求处理
		c.Next()

		// 检查响应状态，记录错误到Sentinel
		if c.Writer.Status() >= 400 {
			sentinel.TraceError(entry, fmt.Errorf("HTTP error: %d", c.Writer.Status()))
		}
	}
}

// handleBlockError 处理不同类型的阻塞错误
func (spm *SentinelProtectionMiddleware) handleBlockError(c *gin.Context, resource string, blockErr *base.BlockError) {
	switch blockErr.BlockType() {
	case base.BlockTypeFlow:
		// 限流错误 - 429 Too Many Requests
		logger.Warn(c.Request.Context(), "Request rate limited",
			zap.String("resource", resource),
			zap.String("path", c.Request.URL.Path),
			zap.String("rule_type", "rate_limit"))

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Rate Limit Exceeded",
			"code":        "RATE_LIMITED",
			"message":     fmt.Sprintf("Request rate limit exceeded for resource '%s'", resource),
			"path":        c.Request.URL.Path,
			"resource":    resource,
			"block_type":  "rate_limit",
			"retry_after": "Please retry after a moment",
		})

	case base.BlockTypeCircuitBreaking:
		// 熔断错误 - 503 Service Unavailable
		logger.Warn(c.Request.Context(), "Request circuit breaking",
			zap.String("resource", resource),
			zap.String("path", c.Request.URL.Path),
			zap.String("rule_type", "circuit_breaker"))

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":       "Service Circuit Breaker Activated",
			"code":        "CIRCUIT_BREAKER",
			"message":     fmt.Sprintf("Service temporarily unavailable due to circuit breaker for resource '%s'", resource),
			"path":        c.Request.URL.Path,
			"resource":    resource,
			"block_type":  "circuit_breaker",
			"retry_after": "Service will be available shortly",
		})

	case base.BlockTypeSystemFlow:
		// 系统规则错误 - 503 Service Unavailable
		logger.Warn(c.Request.Context(), "Request blocked by system rule",
			zap.String("resource", resource),
			zap.String("path", c.Request.URL.Path),
			zap.String("rule_type", "system_rule"))

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":       "System Protection Activated",
			"code":        "SYSTEM_PROTECTION",
			"message":     fmt.Sprintf("Request blocked by system protection for resource '%s'", resource),
			"path":        c.Request.URL.Path,
			"resource":    resource,
			"block_type":  "system_protection",
			"retry_after": "System is under high load, please retry later",
		})

	case base.BlockTypeHotSpotParamFlow:
		// 热点参数限流 - 429 Too Many Requests
		logger.Warn(c.Request.Context(), "Request blocked by hotspot param flow",
			zap.String("resource", resource),
			zap.String("path", c.Request.URL.Path),
			zap.String("rule_type", "hotspot_param"))

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Hotspot Parameter Rate Limit",
			"code":        "HOTSPOT_LIMITED",
			"message":     fmt.Sprintf("Hotspot parameter rate limit exceeded for resource '%s'", resource),
			"path":        c.Request.URL.Path,
			"resource":    resource,
			"block_type":  "hotspot_param",
			"retry_after": "Please retry with different parameters",
		})

	default:
		// 未知类型错误 - 503 Service Unavailable
		logger.Error(c.Request.Context(), "Request blocked by unknown rule",
			zap.String("resource", resource),
			zap.String("path", c.Request.URL.Path),
			zap.String("block_type", fmt.Sprintf("%v", blockErr.BlockType())),
			zap.String("error", blockErr.Error()))

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":       "Service Protection Activated",
			"code":        "UNKNOWN_PROTECTION",
			"message":     fmt.Sprintf("Request blocked by protection mechanism for resource '%s'", resource),
			"path":        c.Request.URL.Path,
			"resource":    resource,
			"block_type":  "unknown",
			"retry_after": "Please retry later",
		})
	}

	c.Abort()
}

// GRPCUnaryInterceptor gRPC一元拦截器
func (spm *SentinelProtectionMiddleware) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !spm.enabled {
			return handler(ctx, req)
		}

		// 构建gRPC资源名
		resource := mapGRPCMethodToResourceName(info.FullMethod)

		var resp interface{}
		var handlerErr error

		// 执行保护逻辑
		err := spm.sentinelManager.Execute(ctx, resource, func() error {
			resp, handlerErr = handler(ctx, req)
			return handlerErr
		})

		if err != nil && strings.Contains(err.Error(), "blocked by sentinel") {
			logger.Info(ctx, "gRPC request blocked by Sentinel",
				zap.String("resource", resource),
				zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.ResourceExhausted, "Rate limited: %v", err)
		}

		return resp, handlerErr
	}
}

// GRPCStreamInterceptor gRPC流拦截器
func (spm *SentinelProtectionMiddleware) GRPCStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !spm.enabled {
			return handler(srv, ss)
		}

		// 构建gRPC资源名
		resource := mapGRPCMethodToResourceName(info.FullMethod)

		// 执行保护逻辑
		err := spm.sentinelManager.Execute(ss.Context(), resource, func() error {
			return handler(srv, ss)
		})

		if err != nil && strings.Contains(err.Error(), "blocked by sentinel") {
			return status.Errorf(codes.ResourceExhausted, "Rate limited: %v", err)
		}

		return err
	}
}

// mapGRPCMethodToResourceName 构建gRPC资源名
func mapGRPCMethodToResourceName(fullMethod string) string {
	// 解析gRPC方法格式: /package.Service/Method
	parts := strings.Split(strings.TrimPrefix(fullMethod, "/"), "/")
	if len(parts) != 2 {
		// 如果格式不对，使用原来的逻辑
		resource := strings.TrimPrefix(fullMethod, "/")
		resource = strings.ReplaceAll(resource, "/", "_")
		resource = strings.ReplaceAll(resource, ".", "_")
		return resource
	}

	serviceWithPackage := parts[0] // 例如: user.UserService
	methodName := parts[1]         // 例如: GetUser

	// 提取服务名（去掉包名）
	serviceParts := strings.Split(serviceWithPackage, ".")
	serviceName := serviceParts[len(serviceParts)-1] // UserService

	// 转换为小写+下划线格式
	serviceName = strings.ToLower(strings.ReplaceAll(serviceName, "Service", "_service"))
	methodName = camelToSnake(methodName)

	// 构建资源名: /grpc/service_name/method_name
	return fmt.Sprintf("/grpc/%s/%s", serviceName, methodName)
}

// camelToSnake 将驼峰命名转换为下划线命名
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, rune(strings.ToLower(string(r))[0]))
	}
	return string(result)
}

// IsEnabled 检查是否启用
func (spm *SentinelProtectionMiddleware) IsEnabled() bool {
	return spm.enabled
}

// GetStats 获取统计信息
func (spm *SentinelProtectionMiddleware) GetStats(_ context.Context) map[string]interface{} {
	if !spm.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	return map[string]interface{}{
		"enabled": true,
		"type":    "sentinel",
	}
}

// Close 关闭中间件
func (spm *SentinelProtectionMiddleware) Close() error {
	if spm.sentinelManager != nil {
		_ = spm.sentinelManager.Close()
	}
	return nil
}
