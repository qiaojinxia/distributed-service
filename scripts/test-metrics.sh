#!/bin/bash

echo "🧪 测试数据库指标记录功能"
echo "========================================"

# 定义服务地址
APP_URL="http://localhost:8080"
METRICS_URL="http://localhost:9090"

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 5

# 清理之前的数据
echo "🧹 清理之前的测试数据..."
curl -s --max-time 5 --connect-timeout 3 -X DELETE "$APP_URL/api/v1/users/999" > /dev/null 2>&1

echo "📊 开始测试数据库操作..."

echo ""
echo "1️⃣ 测试用户注册 (CREATE 操作)"
REGISTER_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 -X POST "$APP_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: metrics-test-register" \
  -d '{
    "username": "metricsuser",
    "email": "metrics@test.com",
    "password": "test123"
  }')

if echo "$REGISTER_RESPONSE" | grep -q "token"; then
    echo "✅ 用户注册成功"
    # 提取用户ID和token
    USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)
    TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "   用户ID: $USER_ID"
else
    echo "❌ 用户注册失败: $REGISTER_RESPONSE"
fi

echo ""
echo "2️⃣ 测试用户登录 (SELECT 操作)"
LOGIN_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 -X POST "$APP_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: metrics-test-login" \
  -d '{
    "username": "metricsuser",
    "password": "test123"
  }')

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo "✅ 用户登录成功"
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
else
    echo "❌ 用户登录失败: $LOGIN_RESPONSE"
fi

echo ""
echo "3️⃣ 测试获取用户信息 (SELECT 操作)"
if [ -n "$USER_ID" ]; then
    USER_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 -X GET "$APP_URL/api/v1/users/$USER_ID" \
      -H "X-Request-ID: metrics-test-getuser")
    
    if echo "$USER_RESPONSE" | grep -q "metricsuser"; then
        echo "✅ 获取用户信息成功"
    else
        echo "❌ 获取用户信息失败: $USER_RESPONSE"
    fi
fi

echo ""
echo "4️⃣ 测试获取当前用户 (SELECT 操作)"
if [ -n "$TOKEN" ]; then
    ME_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 -X GET "$APP_URL/api/v1/users/me" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-Request-ID: metrics-test-getme")
    
    if echo "$ME_RESPONSE" | grep -q "metricsuser"; then
        echo "✅ 获取当前用户信息成功"
    else
        echo "❌ 获取当前用户信息失败: $ME_RESPONSE"
    fi
fi

echo ""
echo "5️⃣ 测试修改密码 (UPDATE 操作)"
if [ -n "$TOKEN" ]; then
    CHANGE_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 -X POST "$APP_URL/api/v1/auth/change-password" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-Request-ID: metrics-test-changepass" \
      -d '{
        "old_password": "test123",
        "new_password": "newtest123"
      }')
    
    if echo "$CHANGE_RESPONSE" | grep -q "success" || [ -z "$CHANGE_RESPONSE" ]; then
        echo "✅ 修改密码成功"
    else
        echo "❌ 修改密码失败: $CHANGE_RESPONSE"
    fi
fi

echo ""
echo "6️⃣ 测试删除用户 (DELETE 操作)"
if [ -n "$USER_ID" ] && [ -n "$TOKEN" ]; then
    DELETE_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 -X DELETE "$APP_URL/api/v1/users/$USER_ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-Request-ID: metrics-test-delete")
    
    if [ -z "$DELETE_RESPONSE" ] || echo "$DELETE_RESPONSE" | grep -q "204"; then
        echo "✅ 删除用户成功"
    else
        echo "⚠️  删除用户响应: $DELETE_RESPONSE"
    fi
fi

echo ""
echo "📈 检查 Prometheus 指标..."
sleep 2

# 检查指标是否存在
METRICS_RESPONSE=$(curl -s --max-time 10 --connect-timeout 5 "$METRICS_URL/metrics" | grep "database_query_duration_seconds")

if [ -n "$METRICS_RESPONSE" ]; then
    echo "✅ 数据库查询指标已记录"
    echo ""
    echo "📊 数据库指标详情:"
    echo "$METRICS_RESPONSE" | grep "database_query_duration_seconds" | head -10
    
    echo ""
    echo "📈 指标统计:"
    echo "CREATE 操作: $(echo "$METRICS_RESPONSE" | grep -c 'operation="CREATE"')"
    echo "SELECT 操作: $(echo "$METRICS_RESPONSE" | grep -c 'operation="SELECT"')" 
    echo "UPDATE 操作: $(echo "$METRICS_RESPONSE" | grep -c 'operation="UPDATE"')"
    echo "DELETE 操作: $(echo "$METRICS_RESPONSE" | grep -c 'operation="DELETE"')"
else
    echo "❌ 未找到数据库查询指标"
fi

echo ""
echo "🎯 测试完成！"
echo "👉 访问 Prometheus: http://localhost:9091"
echo "👉 查询指标: database_query_duration_seconds"
echo "👉 访问 Grafana: http://localhost:3000 (admin/admin123)" 