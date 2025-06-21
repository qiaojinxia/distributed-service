package test

import (
	"testing"

	"distributed-service/framework/config"
	"distributed-service/framework/protection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWildcardMatchingExample 通配符匹配示例测试
func TestWildcardMatchingExample(t *testing.T) {
	// 创建Sentinel管理器
	manager := protection.NewSentinelManager()
	require.NoError(t, manager.Init())

	t.Run("ConfigureWildcardRules", func(t *testing.T) {
		// 配置通配符限流规则
		authRule := config.RateLimitRuleConfig{
			Name:           "auth_api_limiter",
			Resource:       "/api/v1/auth/*",
			Threshold:      10,
			StatIntervalMs: 60000,
			Enabled:        true,
			Description:    "认证接口通配符限流",
		}

		usersRule := config.RateLimitRuleConfig{
			Name:           "users_api_limiter",
			Resource:       "/api/v1/users/*",
			Threshold:      30,
			StatIntervalMs: 60000,
			Enabled:        true,
			Description:    "用户接口通配符限流",
		}

		generalRule := config.RateLimitRuleConfig{
			Name:           "api_general_limiter",
			Resource:       "/api/*",
			Threshold:      100,
			StatIntervalMs: 60000,
			Enabled:        true,
			Description:    "API通用兜底限流",
		}

		grpcReadRule := config.RateLimitRuleConfig{
			Name:           "grpc_read_operations",
			Resource:       "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*",
			Threshold:      80,
			StatIntervalMs: 1000,
			Enabled:        true,
			Description:    "gRPC读操作多模式匹配",
		}

		// 配置规则
		require.NoError(t, manager.ConfigureFlowRuleWithConfig(authRule))
		require.NoError(t, manager.ConfigureFlowRuleWithConfig(usersRule))
		require.NoError(t, manager.ConfigureFlowRuleWithConfig(generalRule))
		require.NoError(t, manager.ConfigureFlowRuleWithConfig(grpcReadRule))

		t.Log("✅ 通配符规则配置完成")
	})

	t.Run("TestResourceMatching", func(t *testing.T) {
		// 测试资源匹配
		testCases := []struct {
			resource        string
			expectedPattern string
			description     string
		}{
			{
				resource:        "/api/v1/auth/login",
				expectedPattern: "/api/v1/auth/*",
				description:     "登录接口应该匹配认证通配符",
			},
			{
				resource:        "/api/v1/auth/register",
				expectedPattern: "/api/v1/auth/*",
				description:     "注册接口应该匹配认证通配符",
			},
			{
				resource:        "/api/v1/users/123",
				expectedPattern: "/api/v1/users/*",
				description:     "用户详情接口应该匹配用户通配符",
			},
			{
				resource:        "/api/v1/users/profile",
				expectedPattern: "/api/v1/users/*",
				description:     "用户资料接口应该匹配用户通配符",
			},
			{
				resource:        "/api/v1/orders/123",
				expectedPattern: "/api/*",
				description:     "订单接口应该匹配通用API兜底",
			},
			{
				resource:        "/grpc/user_service/get_user",
				expectedPattern: "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*",
				description:     "gRPC获取用户应该匹配读操作多模式",
			},
			{
				resource:        "/grpc/order_service/list_orders",
				expectedPattern: "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*",
				description:     "gRPC列出订单应该匹配读操作多模式",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// 获取资源统计信息来验证匹配
				stats := manager.GetResourceStats(tc.resource)

				t.Logf("📊 资源 %s 的统计信息: %+v", tc.resource, stats)

				// 检查是否有匹配的规则
				hasRule := false
				if _, exists := stats["flow_rule"]; exists {
					hasRule = true
					t.Logf("✅ 资源 %s 已有具体规则", tc.resource)
				} else if matcher, exists := stats["flow_rule_matcher"]; exists {
					hasRule = true
					matcherMap := matcher.(map[string]interface{})
					pattern := matcherMap["pattern"].(string)
					priority := matcherMap["priority"].(int)

					t.Logf("✅ 资源 %s 匹配到模式: %s (优先级: %d)", tc.resource, pattern, priority)
					assert.Equal(t, tc.expectedPattern, pattern, "匹配的模式应该正确")
				}

				assert.True(t, hasRule, "资源应该有匹配的规则")
			})
		}
	})

	t.Run("TestMatchingPriority", func(t *testing.T) {
		// 测试匹配优先级
		// 精确匹配 > 具体通配符 > 通用通配符

		// 添加一个精确匹配规则
		exactRule := config.RateLimitRuleConfig{
			Name:           "exact_login_limiter",
			Resource:       "/api/v1/auth/login",
			Threshold:      5,
			StatIntervalMs: 60000,
			Enabled:        true,
			Description:    "登录接口精确限流",
		}
		require.NoError(t, manager.ConfigureFlowRuleWithConfig(exactRule))

		// 测试精确匹配优先于通配符匹配
		stats := manager.GetResourceStats("/api/v1/auth/login")

		// 应该有具体的规则，而不是匹配器
		if flowRule, exists := stats["flow_rule"]; exists {
			ruleMap := flowRule.(map[string]interface{})
			threshold := ruleMap["threshold"].(float64)
			assert.Equal(t, float64(5), threshold, "应该使用精确匹配的阈值")
			t.Log("✅ 精确匹配优先级测试通过")
		} else {
			t.Error("❌ 精确匹配规则应该存在")
		}
	})

	t.Run("TestDynamicRuleCreation", func(t *testing.T) {
		// 测试动态规则创建
		newResource := "/api/v1/auth/change-password"

		// 第一次访问该资源，应该动态创建规则
		manager.EnsureResourceRules(newResource)

		// 检查是否创建了规则
		stats := manager.GetResourceStats(newResource)
		hasRule := false

		if _, exists := stats["flow_rule"]; exists {
			hasRule = true
			t.Logf("✅ 为资源 %s 动态创建了规则", newResource)
		}

		assert.True(t, hasRule, "应该为新资源动态创建规则")
	})

	t.Run("TestAllRulesStatus", func(t *testing.T) {
		// 获取所有规则状态
		allRules := manager.GetAllRules()

		t.Logf("📊 所有规则状态:")
		t.Logf("   • 限流规则数量: %d", allRules["flow_rule_count"])
		t.Logf("   • 熔断器数量: %d", allRules["circuit_breaker_count"])
		t.Logf("   • 通配符支持: %v", allRules["wildcard_support"])

		if flowMatchers, exists := allRules["flow_matchers"]; exists {
			matchers := flowMatchers.([]string)
			t.Logf("   • 限流匹配器: %v", matchers)
		}

		assert.True(t, allRules["wildcard_support"].(bool), "应该支持通配符")
		assert.Greater(t, allRules["flow_rule_count"].(int), 0, "应该有限流规则")
	})
}
