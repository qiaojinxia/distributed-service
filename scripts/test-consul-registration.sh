#!/bin/bash

echo "🔍 测试 Consul 服务注册和健康检查..."

# 等待服务启动
echo "⏳ 等待服务完全启动..."
sleep 5

# 检查 Consul 是否运行
echo "1️⃣  检查 Consul 服务状态..."
if curl -f -s http://localhost:8500/v1/status/leader > /dev/null; then
    echo "✅ Consul 服务运行正常"
else
    echo "❌ Consul 服务未运行"
    exit 1
fi

# 检查应用健康状态
echo "2️⃣  检查应用健康状态..."
if curl -f -s http://localhost:8080/health > /dev/null; then
    echo "✅ 应用健康检查通过"
else
    echo "❌ 应用健康检查失败"
fi

# 查看 Consul 中注册的服务
echo "3️⃣  查看 Consul 中注册的服务..."
echo "📋 所有注册的服务："
curl -s http://localhost:8500/v1/catalog/services | jq '.'

echo ""
echo "📋 distributed-service 服务详情："
curl -s http://localhost:8500/v1/health/service/distributed-service | jq '.'

# 检查服务注册详情
echo ""
echo "4️⃣  检查服务注册详情..."
SERVICE_INFO=$(curl -s http://localhost:8500/v1/health/service/distributed-service)

if echo "$SERVICE_INFO" | jq -e '.[0]' > /dev/null 2>&1; then
    SERVICE_ADDRESS=$(echo "$SERVICE_INFO" | jq -r '.[0].Service.Address')
    SERVICE_PORT=$(echo "$SERVICE_INFO" | jq -r '.[0].Service.Port')
    HEALTH_STATUS=$(echo "$SERVICE_INFO" | jq -r '.[0].Checks[1].Status')
    HEALTH_URL=$(echo "$SERVICE_INFO" | jq -r '.[0].Checks[1].HTTP')
    
    echo "✅ 服务已成功注册到 Consul"
    echo "   - 服务地址: $SERVICE_ADDRESS"
    echo "   - 服务端口: $SERVICE_PORT"
    echo "   - 健康检查URL: $HEALTH_URL"
    echo "   - 健康状态: $HEALTH_STATUS"
    
    # 验证地址是否正确
    if [[ "$SERVICE_ADDRESS" == "app" ]]; then
        echo "✅ 服务地址配置正确 (使用容器名称 'app')"
    elif [[ "$SERVICE_ADDRESS" == "localhost" ]]; then
        echo "⚠️  服务地址使用 localhost (可能是开发环境)"
    else
        echo "❓ 服务地址: $SERVICE_ADDRESS"
    fi
    
    # 验证健康检查 URL
    if [[ "$HEALTH_URL" == *"app:8080"* ]]; then
        echo "✅ 健康检查URL配置正确 (使用容器网络)"
    elif [[ "$HEALTH_URL" == *"localhost:8080"* ]]; then
        echo "⚠️  健康检查URL使用 localhost (可能是开发环境)"
    else
        echo "❓ 健康检查URL: $HEALTH_URL"
    fi
    
    # 检查健康状态
    if [[ "$HEALTH_STATUS" == "passing" ]]; then
        echo "✅ 服务健康检查通过"
    else
        echo "❌ 服务健康检查失败: $HEALTH_STATUS"
    fi
else
    echo "❌ 服务未在 Consul 中找到"
    exit 1
fi

echo ""
echo "5️⃣  测试通过 Consul 发现服务..."
SERVICE_INSTANCES=$(curl -s "http://localhost:8500/v1/health/service/distributed-service?passing")
INSTANCE_COUNT=$(echo "$SERVICE_INSTANCES" | jq '. | length')

echo "发现 $INSTANCE_COUNT 个健康的服务实例"

if [[ "$INSTANCE_COUNT" -gt 0 ]]; then
    echo "✅ 服务发现功能正常"
else
    echo "❌ 没有发现健康的服务实例"
fi

echo ""
echo "🎉 Consul 测试完成！" 