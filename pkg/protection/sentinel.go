package protection

import (
	"context"
	"distributed-service/pkg/logger"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/logging"
	"go.uber.org/zap"

	appconfig "distributed-service/pkg/config"
)

// ResourceMatcher 资源匹配器，支持通配符匹配
type ResourceMatcher struct {
	Pattern   string // 匹配模式
	Resource  string // 原始资源配置
	Priority  int    // 匹配优先级 (数字越小优先级越高)
	IsPattern bool   // 是否为通配符模式
}

// SentinelManager Sentinel管理器，支持通配符匹配的熔断和限流
type SentinelManager struct {
	initialized         bool
	flowRules           map[string]*flow.Rule                         // 存储所有限流规则 (key: 实际资源名)
	circuitBreakerRules map[string]*circuitbreaker.Rule               // 存储所有熔断规则 (key: 实际资源名)
	flowMatchers        []ResourceMatcher                             // 限流规则匹配器
	circuitMatchers     []ResourceMatcher                             // 熔断规则匹配器
	configFlowRules     map[string]appconfig.RateLimitRuleConfig      // 配置的限流规则
	configCircuitRules  map[string]appconfig.CircuitBreakerRuleConfig // 配置的熔断规则
}

// NewSentinelManager 创建Sentinel管理器
func NewSentinelManager() *SentinelManager {
	return &SentinelManager{
		flowRules:           make(map[string]*flow.Rule),
		circuitBreakerRules: make(map[string]*circuitbreaker.Rule),
		flowMatchers:        make([]ResourceMatcher, 0),
		circuitMatchers:     make([]ResourceMatcher, 0),
		configFlowRules:     make(map[string]appconfig.RateLimitRuleConfig),
		configCircuitRules:  make(map[string]appconfig.CircuitBreakerRuleConfig),
	}
}

// MatchResource 匹配资源名称，支持通配符和多模式匹配
func MatchResource(resource string, pattern string) bool {
	// 支持多模式匹配 (逗号分隔)
	patterns := strings.Split(pattern, ",")
	for _, p := range patterns {
		p = strings.TrimSpace(p)

		// 处理特殊的路径通配符（如 /api/* 应该匹配 /api/v1/users）
		if strings.HasSuffix(p, "/*") {
			prefix := strings.TrimSuffix(p, "/*")
			if strings.HasPrefix(resource, prefix+"/") {
				return true
			}
		}

		// 使用filepath.Match进行标准通配符匹配
		if matched, _ := filepath.Match(p, resource); matched {
			return true
		}
	}
	return false
}

// GetMatchingResource 获取匹配的资源配置，优先级：精确匹配 > 具体通配符 > 通用通配符
func (sm *SentinelManager) GetMatchingResource(resource string, matchers []ResourceMatcher) *ResourceMatcher {
	var matched []ResourceMatcher

	// 收集所有匹配的模式
	for _, matcher := range matchers {
		if matcher.Pattern == resource {
			// 精确匹配，直接返回
			return &matcher
		}
		if matcher.IsPattern && MatchResource(resource, matcher.Pattern) {
			matched = append(matched, matcher)
		}
	}

	if len(matched) == 0 {
		return nil
	}

	// 按优先级排序，优先级数字越小越优先
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Priority < matched[j].Priority
	})

	return &matched[0]
}

// CalculatePatternPriority 计算模式的优先级
func CalculatePatternPriority(pattern string) int {
	// 精确匹配（无通配符）
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, ",") {
		return 1
	}

	// 多模式匹配
	if strings.Contains(pattern, ",") {
		return 4
	}

	// 具体路径通配符（如 /api/v1/auth/*）
	parts := strings.Split(pattern, "/")
	if len(parts) >= 4 && strings.HasSuffix(pattern, "*") {
		return 2
	}

	// 通用通配符（如 /api/*）
	if strings.HasSuffix(pattern, "*") {
		return 3
	}

	// 其他模式
	return 5
}

// Init 初始化Sentinel
func (sm *SentinelManager) Init() error {
	if sm.initialized {
		return nil
	}

	// 配置Sentinel
	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()

	// 初始化Sentinel
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		return fmt.Errorf("failed to initialize sentinel: %w", err)
	}

	sm.initialized = true
	logger.Info(context.Background(), "Sentinel initialized successfully with wildcard matching support")
	return nil
}

// FlowRuleConfig 限流规则配置
type FlowRuleConfig struct {
	Resource               string  `json:"resource"`                 // 资源名
	Threshold              float64 `json:"threshold"`                // 统计窗口内允许的最大请求数量
	StatIntervalInMs       uint32  `json:"stat_interval_in_ms"`      // 统计时间窗口(毫秒)，实际QPS = (Threshold × 1000) / StatIntervalInMs
	TokenCalculateStrategy int     `json:"token_calculate_strategy"` // 令牌计算策略
	ControlBehavior        int     `json:"control_behavior"`         // 控制行为
}

// ConfigureFlowRule 配置限流规则 - 简化版本，直接指定每秒阈值
func (sm *SentinelManager) ConfigureFlowRule(resource string, thresholdPerSecond float64) error {
	// 创建配置对象
	cfg := appconfig.RateLimitRuleConfig{
		Name:           fmt.Sprintf("auto_%s", resource),
		Resource:       resource,
		Threshold:      thresholdPerSecond,
		StatIntervalMs: 1000, // 默认1秒统计窗口
		Enabled:        true,
		Description:    fmt.Sprintf("Auto-generated rule for %s", resource),
	}
	return sm.ConfigureFlowRuleWithConfig(cfg)
}

// ConfigureFlowRuleWithConfig 根据配置文件配置限流规则，支持通配符
func (sm *SentinelManager) ConfigureFlowRuleWithConfig(cfg appconfig.RateLimitRuleConfig) error {
	if !sm.initialized {
		if err := sm.Init(); err != nil {
			return err
		}
	}

	// 保存配置规则
	sm.configFlowRules[cfg.Name] = cfg

	// 创建资源匹配器
	matcher := ResourceMatcher{
		Pattern:   cfg.Resource,
		Resource:  cfg.Resource,
		Priority:  CalculatePatternPriority(cfg.Resource),
		IsPattern: strings.Contains(cfg.Resource, "*") || strings.Contains(cfg.Resource, ","),
	}

	// 添加或更新匹配器
	found := false
	for i, m := range sm.flowMatchers {
		if m.Pattern == cfg.Resource {
			sm.flowMatchers[i] = matcher
			found = true
			break
		}
	}
	if !found {
		sm.flowMatchers = append(sm.flowMatchers, matcher)
	}

	// 对于非通配符规则，直接创建Sentinel规则
	if !matcher.IsPattern {
		err := sm.createFlowRule(cfg.Resource, cfg)
		if err != nil {
			return err
		}
	}

	logger.Info(context.Background(), "Flow rule configured",
		zap.String("name", cfg.Name),
		zap.String("pattern", cfg.Resource),
		zap.Int("priority", matcher.Priority),
		zap.Bool("is_pattern", matcher.IsPattern))
	return nil
}

// createFlowRule 创建具体的Sentinel限流规则
func (sm *SentinelManager) createFlowRule(resource string, cfg appconfig.RateLimitRuleConfig) error {
	rule := &flow.Rule{
		Resource:               resource,
		Threshold:              cfg.Threshold,
		StatIntervalInMs:       cfg.StatIntervalMs,
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        flow.Reject,
	}

	// 存储规则
	sm.flowRules[resource] = rule

	// 重新加载所有规则
	err := sm.reloadFlowRules()
	if err != nil {
		delete(sm.flowRules, resource)
		return fmt.Errorf("failed to load flow rules: %w", err)
	}

	return nil
}

// reloadFlowRules 重新加载所有限流规则
func (sm *SentinelManager) reloadFlowRules() error {
	rules := make([]*flow.Rule, 0, len(sm.flowRules))
	for _, rule := range sm.flowRules {
		rules = append(rules, rule)
	}

	_, err := flow.LoadRules(rules)
	return err
}

// ConfigureCircuitBreakerRule 配置熔断器规则，支持通配符
func (sm *SentinelManager) ConfigureCircuitBreakerRule(resource string, cfg CircuitBreakerConfig) error {
	if !sm.initialized {
		if err := sm.Init(); err != nil {
			return err
		}
	}

	// 解析策略
	var strategy circuitbreaker.Strategy
	switch cfg.Strategy {
	case "ErrorRatio":
		strategy = circuitbreaker.ErrorRatio
	case "ErrorCount":
		strategy = circuitbreaker.ErrorCount
	case "SlowRequestRatio":
		strategy = circuitbreaker.SlowRequestRatio
	default:
		strategy = circuitbreaker.ErrorRatio // 默认错误比例策略
	}

	// 创建新规则
	rule := &circuitbreaker.Rule{
		Resource:                     resource,
		Strategy:                     strategy,
		RetryTimeoutMs:               cfg.RetryTimeoutMs,
		MinRequestAmount:             cfg.MinRequestAmount,
		StatIntervalMs:               cfg.StatIntervalMs,
		StatSlidingWindowBucketCount: cfg.StatSlidingWindowBucketCount,
		Threshold:                    cfg.Threshold,
		ProbeNum:                     cfg.ProbeNum,
	}

	// 只有慢调用策略才设置最大允许响应时间
	if strategy == circuitbreaker.SlowRequestRatio {
		rule.MaxAllowedRtMs = cfg.MaxAllowedRtMs
	}

	// 存储规则（如果资源已存在，会覆盖旧规则）
	sm.circuitBreakerRules[resource] = rule

	// 重新加载所有规则
	err := sm.reloadCircuitBreakerRules()
	if err != nil {
		// 如果加载失败，从map中移除这个规则
		delete(sm.circuitBreakerRules, resource)
		return fmt.Errorf("failed to load circuit breaker rules: %w", err)
	}

	logger.Info(context.Background(), "Circuit breaker rule configured",
		zap.String("resource", resource),
		zap.String("strategy", cfg.Strategy),
		zap.Float64("threshold", cfg.Threshold))
	return nil
}

// ConfigureCircuitBreakerWithConfig 根据配置文件配置熔断器规则，支持通配符
func (sm *SentinelManager) ConfigureCircuitBreakerWithConfig(cfg appconfig.CircuitBreakerRuleConfig) error {
	if !sm.initialized {
		if err := sm.Init(); err != nil {
			return err
		}
	}

	// 保存配置规则
	sm.configCircuitRules[cfg.Name] = cfg

	// 创建资源匹配器
	matcher := ResourceMatcher{
		Pattern:   cfg.Resource,
		Resource:  cfg.Resource,
		Priority:  CalculatePatternPriority(cfg.Resource),
		IsPattern: strings.Contains(cfg.Resource, "*") || strings.Contains(cfg.Resource, ","),
	}

	// 添加或更新匹配器
	found := false
	for i, m := range sm.circuitMatchers {
		if m.Pattern == cfg.Resource {
			sm.circuitMatchers[i] = matcher
			found = true
			break
		}
	}
	if !found {
		sm.circuitMatchers = append(sm.circuitMatchers, matcher)
	}

	// 对于非通配符规则，直接创建Sentinel规则
	if !matcher.IsPattern {
		circuitCfg := CircuitBreakerConfig{
			Strategy:                     cfg.Strategy,
			RetryTimeoutMs:               cfg.RetryTimeoutMs,
			MinRequestAmount:             cfg.MinRequestAmount,
			StatIntervalMs:               cfg.StatIntervalMs,
			StatSlidingWindowBucketCount: cfg.StatSlidingWindowBucketCount,
			MaxAllowedRtMs:               cfg.MaxAllowedRtMs,
			Threshold:                    cfg.Threshold,
			ProbeNum:                     cfg.ProbeNum,
		}
		err := sm.ConfigureCircuitBreakerRule(cfg.Resource, circuitCfg)
		if err != nil {
			return err
		}
	}

	logger.Info(context.Background(), "Circuit breaker rule configured",
		zap.String("name", cfg.Name),
		zap.String("pattern", cfg.Resource),
		zap.Int("priority", matcher.Priority),
		zap.Bool("is_pattern", matcher.IsPattern))
	return nil
}

// reloadCircuitBreakerRules 重新加载所有熔断规则
func (sm *SentinelManager) reloadCircuitBreakerRules() error {
	rules := make([]*circuitbreaker.Rule, 0, len(sm.circuitBreakerRules))
	for _, rule := range sm.circuitBreakerRules {
		rules = append(rules, rule)
	}

	_, err := circuitbreaker.LoadRules(rules)
	return err
}

// EnsureResourceRules 确保资源有对应的规则（动态创建通配符匹配的规则）
func (sm *SentinelManager) EnsureResourceRules(resource string) {
	// 检查限流规则
	if _, exists := sm.flowRules[resource]; !exists {
		if matcher := sm.GetMatchingResource(resource, sm.flowMatchers); matcher != nil {
			// 使用匹配器中的原始资源名查找配置
			for _, cfg := range sm.configFlowRules {
				if cfg.Resource == matcher.Resource {
					err := sm.createFlowRule(resource, cfg)
					if err != nil {
						logger.Error(context.Background(), "Failed to create dynamic flow rule",
							zap.String("resource", resource),
							zap.String("pattern", matcher.Pattern),
							zap.Error(err))
					} else {
						logger.Info(context.Background(), "Created dynamic flow rule",
							zap.String("resource", resource),
							zap.String("pattern", matcher.Pattern))
					}
					break
				}
			}
		}
	}

	// 检查熔断规则
	if _, exists := sm.circuitBreakerRules[resource]; !exists {
		if matcher := sm.GetMatchingResource(resource, sm.circuitMatchers); matcher != nil {
			// 使用匹配器中的原始资源名查找配置
			for _, cfg := range sm.configCircuitRules {
				if cfg.Resource == matcher.Resource {
					circuitCfg := CircuitBreakerConfig{
						Strategy:                     cfg.Strategy,
						RetryTimeoutMs:               cfg.RetryTimeoutMs,
						MinRequestAmount:             cfg.MinRequestAmount,
						StatIntervalMs:               cfg.StatIntervalMs,
						StatSlidingWindowBucketCount: cfg.StatSlidingWindowBucketCount,
						MaxAllowedRtMs:               cfg.MaxAllowedRtMs,
						Threshold:                    cfg.Threshold,
						ProbeNum:                     cfg.ProbeNum,
					}
					err := sm.ConfigureCircuitBreakerRule(resource, circuitCfg)
					if err != nil {
						logger.Error(context.Background(), "Failed to configure circuit breaker rule",
							zap.String("resource", resource),
							zap.String("pattern", matcher.Pattern),
							zap.Error(err))
					} else {
						logger.Info(context.Background(), "Created dynamic circuit breaker rule",
							zap.String("resource", resource),
							zap.String("pattern", matcher.Pattern))
					}
					break
				}
			}
		}
	}
}

// Entry 执行带保护的操作，支持通配符匹配
func (sm *SentinelManager) Entry(_ context.Context, resource string, entryType base.TrafficType) (*base.SentinelEntry, *base.BlockError) {
	// 确保资源有对应的规则
	sm.EnsureResourceRules(resource)

	return sentinel.Entry(resource, sentinel.WithTrafficType(entryType))
}

// Execute 执行带保护的函数，支持通配符匹配
func (sm *SentinelManager) Execute(_ context.Context, resource string, fn func() error) error {
	if !sm.initialized {
		if err := sm.Init(); err != nil {
			return err
		}
	}

	// 确保资源有对应的规则
	sm.EnsureResourceRules(resource)

	entry, err := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound))
	if err != nil {
		// 被限流或熔断
		return fmt.Errorf("blocked by sentinel: %w", err)
	}
	defer entry.Exit()

	// 执行业务逻辑
	bizErr := fn()
	if bizErr != nil {
		// 记录异常
		sentinel.TraceError(entry, bizErr)
		return bizErr
	}

	return nil
}

// RemoveFlowRule 移除指定资源的限流规则
func (sm *SentinelManager) RemoveFlowRule(resource string) error {
	if !sm.initialized {
		return fmt.Errorf("sentinel not initialized")
	}

	delete(sm.flowRules, resource)
	return sm.reloadFlowRules()
}

// RemoveCircuitBreakerRule 移除指定资源的熔断规则
func (sm *SentinelManager) RemoveCircuitBreakerRule(resource string) error {
	if !sm.initialized {
		return fmt.Errorf("sentinel not initialized")
	}

	delete(sm.circuitBreakerRules, resource)
	return sm.reloadCircuitBreakerRules()
}

// GetAllRules 获取所有配置的规则
func (sm *SentinelManager) GetAllRules() map[string]interface{} {
	flowRuleNames := make([]string, 0, len(sm.flowRules))
	for resource := range sm.flowRules {
		flowRuleNames = append(flowRuleNames, resource)
	}

	circuitBreakerNames := make([]string, 0, len(sm.circuitBreakerRules))
	for resource := range sm.circuitBreakerRules {
		circuitBreakerNames = append(circuitBreakerNames, resource)
	}

	flowMatchers := make([]string, 0, len(sm.flowMatchers))
	for _, matcher := range sm.flowMatchers {
		flowMatchers = append(flowMatchers, fmt.Sprintf("%s (priority: %d)", matcher.Pattern, matcher.Priority))
	}

	circuitMatchers := make([]string, 0, len(sm.circuitMatchers))
	for _, matcher := range sm.circuitMatchers {
		circuitMatchers = append(circuitMatchers, fmt.Sprintf("%s (priority: %d)", matcher.Pattern, matcher.Priority))
	}

	return map[string]interface{}{
		"flow_rules":            flowRuleNames,
		"circuit_breakers":      circuitBreakerNames,
		"flow_rule_count":       len(sm.flowRules),
		"circuit_breaker_count": len(sm.circuitBreakerRules),
		"flow_matchers":         flowMatchers,
		"circuit_matchers":      circuitMatchers,
		"wildcard_support":      true,
	}
}

// GetResourceStats 获取资源统计信息
func (sm *SentinelManager) GetResourceStats(resource string) map[string]interface{} {
	stats := map[string]interface{}{
		"resource":    resource,
		"initialized": sm.initialized,
	}

	if sm.initialized {
		// 检查是否有限流规则
		if flowRule, exists := sm.flowRules[resource]; exists {
			stats["flow_rule"] = map[string]interface{}{
				"threshold":          flowRule.Threshold,
				"stat_interval_ms":   flowRule.StatIntervalInMs,
				"control_behavior":   flowRule.ControlBehavior.String(),
				"calculate_strategy": flowRule.TokenCalculateStrategy.String(),
			}
		} else {
			// 检查是否有匹配的通配符规则
			if matcher := sm.GetMatchingResource(resource, sm.flowMatchers); matcher != nil {
				stats["flow_rule_matcher"] = map[string]interface{}{
					"pattern":  matcher.Pattern,
					"priority": matcher.Priority,
				}
			}
		}

		// 检查是否有熔断规则
		if cbRule, exists := sm.circuitBreakerRules[resource]; exists {
			stats["circuit_breaker_rule"] = map[string]interface{}{
				"strategy":                         cbRule.Strategy.String(),
				"retry_timeout_ms":                 cbRule.RetryTimeoutMs,
				"min_request_amount":               cbRule.MinRequestAmount,
				"stat_interval_ms":                 cbRule.StatIntervalMs,
				"stat_sliding_window_bucket_count": cbRule.StatSlidingWindowBucketCount,
				"threshold":                        cbRule.Threshold,
			}
		} else {
			// 检查是否有匹配的通配符规则
			if matcher := sm.GetMatchingResource(resource, sm.circuitMatchers); matcher != nil {
				stats["circuit_breaker_matcher"] = map[string]interface{}{
					"pattern":  matcher.Pattern,
					"priority": matcher.Priority,
				}
			}
		}
	}

	return stats
}

// Close 关闭Sentinel
func (sm *SentinelManager) Close() error {
	// Sentinel没有显式的关闭方法
	sm.initialized = false
	return nil
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Strategy                     string  `json:"strategy"`                         // 熔断策略
	RetryTimeoutMs               uint32  `json:"retry_timeout_ms"`                 // 熔断后重试超时时间(毫秒)
	MinRequestAmount             uint64  `json:"min_request_amount"`               // 触发熔断的最小请求数
	StatIntervalMs               uint32  `json:"stat_interval_ms"`                 // 统计时间窗口(毫秒)
	StatSlidingWindowBucketCount uint32  `json:"stat_sliding_window_bucket_count"` // 滑动窗口桶数
	MaxAllowedRtMs               uint64  `json:"max_allowed_rt_ms"`                // 最大允许响应时间(毫秒)
	Threshold                    float64 `json:"threshold"`                        // 熔断阈值
	ProbeNum                     uint64  `json:"probe_num"`                        // 半开状态探测请求数量
}

// FlowConfig 限流配置
type FlowConfig struct {
	Resource  string  `json:"resource"`  // 资源名
	Threshold float64 `json:"threshold"` // 限流阈值
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Strategy:                     "ErrorRatio", // 默认错误比例策略
		RetryTimeoutMs:               5000,         // 5秒重试
		MinRequestAmount:             10,           // 最少10个请求
		StatIntervalMs:               10000,        // 10秒统计窗口
		StatSlidingWindowBucketCount: 10,           // 10个桶
		Threshold:                    0.5,          // 50%错误率
		ProbeNum:                     3,            // 3个探测请求
	}
}

// ParseDuration 解析持续时间配置
func ParseDuration(duration string) (time.Duration, error) {
	return time.ParseDuration(duration)
}
