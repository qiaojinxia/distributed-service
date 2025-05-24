#!/bin/bash

# 链路追踪功能验证脚本

set -e

echo "🔍 开始验证分布式链路追踪功能..."

# 服务地址
BASE_URL="http://localhost:8080"
JAEGER_UI="http://localhost:16686"

# 检查服务状态
echo "📊 检查服务状态..."
services=(
    "应用服务:$BASE_URL/health"
    "Jaeger UI:$JAEGER_UI"
)

all_services_up=true
for service in "${services[@]}"; do
    name=$(echo $service | cut -d: -f1)
    url=$(echo $service | cut -d: -f2-)
    
    if curl -f -s "$url" > /dev/null; then
        echo "✅ $name 运行正常"
    else
        echo "❌ $name 无法访问: $url"
        all_services_up=false
    fi
done

if [ "$all_services_up" = false ]; then
    echo "⚠️  部分服务未正常运行，请检查 docker-compose ps"
    exit 1
fi

echo ""
echo "🧪 执行追踪测试..."

# 生成唯一的测试数据
TIMESTAMP=$(date +%s)
TEST_USERNAME="tracetest_$TIMESTAMP"
TEST_EMAIL="trace_$TIMESTAMP@example.com"
TEST_PASSWORD="password123"

echo "📝 测试数据: $TEST_USERNAME / $TEST_EMAIL"

# 1. 健康检查（生成追踪数据）
echo "🏥 测试健康检查..."
HEALTH_RESPONSE=$(curl -s -X GET "$BASE_URL/health" \
  -H "X-Request-ID: verify-health-$TIMESTAMP")
echo "健康检查响应: $HEALTH_RESPONSE"

# 2. 用户注册（生成复杂追踪数据）
echo "👤 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: verify-register-$TIMESTAMP" \
  -d "{
    \"username\": \"$TEST_USERNAME\",
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q "error"; then
    echo "⚠️  注册可能失败: $REGISTER_RESPONSE"
else
    echo "✅ 注册成功"
fi

# 3. 用户登录（生成更多追踪数据）
echo "🔑 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: verify-login-$TIMESTAMP" \
  -d "{
    \"username\": \"$TEST_USERNAME\",
    \"password\": \"$TEST_PASSWORD\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo "✅ 登录成功"
else
    echo "⚠️  登录可能失败: $LOGIN_RESPONSE"
fi

echo ""
echo "✅ 追踪测试完成！"
echo ""
echo "🔍 验证步骤："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. 访问 Jaeger UI: $JAEGER_UI"
echo "2. 在 Service 下拉框选择 'distributed-service'"
echo "3. 设置时间范围为最近 5 分钟"
echo "4. 点击 'Find Traces' 按钮"
echo "5. 查找包含以下请求 ID 的追踪数据:"
echo "   - verify-health-$TIMESTAMP"
echo "   - verify-register-$TIMESTAMP"  
echo "   - verify-login-$TIMESTAMP"
echo ""
echo "🎯 应该看到的追踪层次结构:"
echo "   HTTP Request Span"
echo "   ├── userService.Register"
echo "   │   ├── userRepository.GetByUsername"
echo "   │   └── userRepository.Create"
echo "   └── userService.Login"
echo "       └── userRepository.GetByUsername"
echo ""
echo "📊 检查要点:"
echo "   ✓ 每个 span 都有正确的名称和属性"
echo "   ✓ HTTP span 包含方法、路径、状态码"
echo "   ✓ Service span 包含用户名、邮箱等业务属性"
echo "   ✓ Repository span 包含数据库操作信息"
echo "   ✓ 整个调用链路完整且时间合理"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📚 更多信息:"
echo "  - 详细文档: docs/TRACING.md"
echo "  - 完整测试: ./scripts/test-tracing.sh" 