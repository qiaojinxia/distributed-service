package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/qiaojinxia/distributed-service/framework/config"
	"github.com/qiaojinxia/distributed-service/framework/protection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFixedGRPCMapping 测试修复后的gRPC映射逻辑
func TestFixedGRPCMapping(t *testing.T) {
	// 创建Sentinel管理器
	manager := protection.NewSentinelManager()
	require.NoError(t, manager.Init())

	t.Run("ConfigureRulesFromConfig", func(t *testing.T) {
		// 使用config.yaml中的实际配置
		rules := []config.RateLimitRuleConfig{
			{
				Name:           "grpc_user_service_limiter",
				Resource:       "/grpc/user_service/*",
				Threshold:      50,
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "gRPC用户服务限流",
			},
			{
				Name:           "grpc_read_operations_limiter",
				Resource:       "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*",
				Threshold:      80,
				StatIntervalMs: 1000,
				Enabled:        true,
				Description:    "gRPC读操作限流",
			},
			{
				Name:           "grpc_write_operations_limiter",
				Resource:       "/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*",
				Threshold:      10,
				StatIntervalMs: 60000,
				Enabled:        true,
				Description:    "gRPC写操作限流",
			},
		}

		for _, rule := range rules {
			require.NoError(t, manager.ConfigureFlowRuleWithConfig(rule))
		}

		t.Log("✅ 规则配置完成")
	})

	t.Run("TestGRPCMappingAndMatching", func(t *testing.T) {
		testCases := []struct {
			grpcMethod             string
			expectedResource       string
			shouldMatchRead        bool
			shouldMatchWrite       bool
			shouldMatchUserService bool
			description            string
		}{
			{
				grpcMethod:             "/user.UserService/GetUser",
				expectedResource:       "/grpc/user_service/get_user",
				shouldMatchRead:        true,
				shouldMatchWrite:       false,
				shouldMatchUserService: true,
				description:            "获取用户方法",
			},
			{
				grpcMethod:             "/user.UserService/ListUsers",
				expectedResource:       "/grpc/user_service/list_users",
				shouldMatchRead:        true,
				shouldMatchWrite:       false,
				shouldMatchUserService: true,
				description:            "列出用户方法",
			},
			{
				grpcMethod:             "/user.UserService/FindUserByEmail",
				expectedResource:       "/grpc/user_service/find_user_by_email",
				shouldMatchRead:        true,
				shouldMatchWrite:       false,
				shouldMatchUserService: true,
				description:            "根据邮箱查找用户方法",
			},
			{
				grpcMethod:             "/user.UserService/CreateUser",
				expectedResource:       "/grpc/user_service/create_user",
				shouldMatchRead:        false,
				shouldMatchWrite:       true,
				shouldMatchUserService: true,
				description:            "创建用户方法",
			},
			{
				grpcMethod:             "/user.UserService/UpdateUser",
				expectedResource:       "/grpc/user_service/update_user",
				shouldMatchRead:        false,
				shouldMatchWrite:       true,
				shouldMatchUserService: true,
				description:            "更新用户方法",
			},
			{
				grpcMethod:             "/user.UserService/DeleteUser",
				expectedResource:       "/grpc/user_service/delete_user",
				shouldMatchRead:        false,
				shouldMatchWrite:       true,
				shouldMatchUserService: true,
				description:            "删除用户方法",
			},
			{
				grpcMethod:             "/order.OrderService/GetOrder",
				expectedResource:       "/grpc/order_service/get_order",
				shouldMatchRead:        true,
				shouldMatchWrite:       false,
				shouldMatchUserService: false,
				description:            "获取订单方法",
			},
			{
				grpcMethod:             "/order.OrderService/CreateOrder",
				expectedResource:       "/grpc/order_service/create_order",
				shouldMatchRead:        false,
				shouldMatchWrite:       true,
				shouldMatchUserService: false,
				description:            "创建订单方法",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// 使用修复后的映射逻辑
				resource := mapGRPCMethodToResourceNameFixed(tc.grpcMethod)

				t.Logf("📝 gRPC方法: %s", tc.grpcMethod)
				t.Logf("📝 映射资源: %s", resource)
				t.Logf("📝 期望资源: %s", tc.expectedResource)

				assert.Equal(t, tc.expectedResource, resource, "资源映射应该正确")

				// 测试通配符匹配
				readPattern := "/grpc/*/get*,/grpc/*/list*,/grpc/*/find*"
				writePattern := "/grpc/*/create*,/grpc/*/update*,/grpc/*/delete*"
				userServicePattern := "/grpc/user_service/*"

				readMatch := protection.MatchResource(resource, readPattern)
				writeMatch := protection.MatchResource(resource, writePattern)
				userServiceMatch := protection.MatchResource(resource, userServicePattern)

				assert.Equal(t, tc.shouldMatchRead, readMatch, "读操作匹配")
				assert.Equal(t, tc.shouldMatchWrite, writeMatch, "写操作匹配")
				assert.Equal(t, tc.shouldMatchUserService, userServiceMatch, "用户服务匹配")

				t.Logf("✅ 匹配结果 - 读操作:%v, 写操作:%v, 用户服务:%v",
					readMatch, writeMatch, userServiceMatch)

				// 测试动态规则创建
				manager.EnsureResourceRules(resource)
				stats := manager.GetResourceStats(resource)

				hasRule := false
				ruleInfo := ""

				if flowRule, exists := stats["flow_rule"]; exists {
					hasRule = true
					ruleMap := flowRule.(map[string]interface{})
					threshold := ruleMap["threshold"].(float64)
					ruleInfo = fmt.Sprintf("具体规则(阈值:%.0f)", threshold)
				} else if matcher, exists := stats["flow_rule_matcher"]; exists {
					hasRule = true
					matcherMap := matcher.(map[string]interface{})
					pattern := matcherMap["pattern"].(string)
					priority := matcherMap["priority"].(int)
					ruleInfo = fmt.Sprintf("匹配器(模式:%s, 优先级:%d)", pattern, priority)
				}

				assert.True(t, hasRule, "应该有匹配的规则")
				t.Logf("📊 规则信息: %s", ruleInfo)
			})
		}
	})
}

// mapGRPCMethodToResourceNameFixed 修复后的gRPC方法名映射逻辑
func mapGRPCMethodToResourceNameFixed(fullMethod string) string {
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
	methodName = camelToSnakeCase(methodName)

	// 构建资源名: /grpc/service_name/method_name
	return fmt.Sprintf("/grpc/%s/%s", serviceName, methodName)
}

// camelToSnakeCase 将驼峰命名转换为下划线命名
func camelToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, rune(strings.ToLower(string(r))[0]))
	}
	return string(result)
}
