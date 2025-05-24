#!/bin/bash

# 分布式链路追踪测试脚本

set -e

echo "🚀 开始测试分布式链路追踪功能..."

# 服务地址
BASE_URL="http://localhost:8080"
JAEGER_UI="http://localhost:16686"

# 测试数据
USERNAME="testuser_$(date +%s)"
EMAIL="test_$(date +%s)@example.com"
PASSWORD="password123"

echo "📝 测试数据:"
echo "  用户名: $USERNAME"
echo "  邮箱: $EMAIL"

# 1. 测试用户注册
echo "🔐 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-register-$(date +%s)" \
  -d "{
    \"username\": \"$USERNAME\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "注册响应: $REGISTER_RESPONSE"

# 2. 测试用户登录
echo "🔑 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-login-$(date +%s)" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\"
  }")

echo "登录响应: $LOGIN_RESPONSE"

# 提取 JWT token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ 无法获取 JWT token"
  exit 1
fi

echo "✅ 获取到 JWT token: ${TOKEN:0:20}..."

# 3. 测试获取用户信息
echo "👤 测试获取用户信息..."
USER_INFO_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: test-userinfo-$(date +%s)")

echo "用户信息响应: $USER_INFO_RESPONSE"

# 4. 测试修改密码
echo "🔒 测试修改密码..."
NEW_PASSWORD="newpassword123"
CHANGE_PASSWORD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/change-password" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: test-changepass-$(date +%s)" \
  -d "{
    \"old_password\": \"$PASSWORD\",
    \"new_password\": \"$NEW_PASSWORD\"
  }")

echo "修改密码响应: $CHANGE_PASSWORD_RESPONSE"

# 5. 测试健康检查
echo "🏥 测试健康检查..."
HEALTH_RESPONSE=$(curl -s -X GET "$BASE_URL/health" \
  -H "X-Request-ID: test-health-$(date +%s)")

echo "健康检查响应: $HEALTH_RESPONSE"

# 6. 测试指标端点
echo "📊 测试指标端点..."
METRICS_RESPONSE=$(curl -s -X GET "http://localhost:9090/metrics" | head -10)
echo "指标响应 (前10行):"
echo "$METRICS_RESPONSE"

echo ""
echo "✅ 所有测试完成！"
echo ""
echo "🔍 查看追踪信息:"
echo "  Jaeger UI: $JAEGER_UI"
echo "  在 Jaeger UI 中搜索服务: distributed-service"
echo ""
echo "📊 查看监控信息:"
echo "  Prometheus: http://localhost:9091"
echo "  Grafana: http://localhost:3000 (admin/admin123)"
echo ""
echo "🎯 追踪功能验证要点:"
echo "  1. 检查 Jaeger UI 中是否有追踪数据"
echo "  2. 验证请求链路是否完整 (HTTP -> Service -> Repository)"
echo "  3. 检查 span 属性是否正确设置"
echo "  4. 验证错误追踪是否正常工作" 