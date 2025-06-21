#!/bin/bash

echo "🚀 Starting HTTP + gRPC integrated service test..."

# 启动服务
./http-grpc-test &
SERVICE_PID=$!

echo "📍 Service PID: $SERVICE_PID"

# 等待5秒让服务启动
echo "⏳ Waiting 5 seconds for services to start..."
sleep 5

# 检查HTTP端口
echo "🔍 Checking HTTP port 8080..."
HTTP_PORT_CHECK=$(lsof -i :8080 | grep LISTEN)
if [ -n "$HTTP_PORT_CHECK" ]; then
    echo "✅ HTTP service is listening on port 8080"
    echo "$HTTP_PORT_CHECK"
else
    echo "❌ HTTP service is not listening on port 8080"
fi

# 检查gRPC端口
echo "🔍 Checking gRPC port 9093..."
GRPC_PORT_CHECK=$(lsof -i :9093 | grep LISTEN)
if [ -n "$GRPC_PORT_CHECK" ]; then
    echo "✅ gRPC service is listening on port 9093"
    echo "$GRPC_PORT_CHECK"
else
    echo "❌ gRPC service is not listening on port 9093"
fi

# 测试HTTP端点
echo "🧪 Testing HTTP health endpoint..."
HTTP_RESPONSE=$(curl -s -w "%{http_code}" http://localhost:8080/health)
if [[ "$HTTP_RESPONSE" =~ 200$ ]]; then
    echo "✅ HTTP health check passed"
    echo "${HTTP_RESPONSE%200}"
else
    echo "❌ HTTP health check failed"
    echo "$HTTP_RESPONSE"
fi

echo "🧪 Testing HTTP ping endpoint..."
PING_RESPONSE=$(curl -s -w "%{http_code}" http://localhost:8080/ping)
if [[ "$PING_RESPONSE" =~ 200$ ]]; then
    echo "✅ HTTP ping endpoint working"
    echo "${PING_RESPONSE%200}"
else
    echo "❌ HTTP ping endpoint failed"
    echo "$PING_RESPONSE"
fi

# 测试gRPC健康检查 (如果grpcurl可用)
if command -v grpcurl &> /dev/null; then
    echo "🧪 Testing gRPC health check..."
    GRPC_HEALTH=$(grpcurl -plaintext localhost:9093 grpc.health.v1.Health/Check 2>&1)
    if [[ "$GRPC_HEALTH" =~ "SERVING" ]]; then
        echo "✅ gRPC health check passed"
        echo "$GRPC_HEALTH"
    else
        echo "⚠️  gRPC health check result: $GRPC_HEALTH"
    fi
    
    echo "🧪 Testing gRPC service reflection..."
    GRPC_SERVICES=$(grpcurl -plaintext localhost:9093 list 2>&1)
    if [[ "$GRPC_SERVICES" =~ "grpc.health.v1.Health" ]]; then
        echo "✅ gRPC reflection working"
        echo "Available services:"
        echo "$GRPC_SERVICES"
    else
        echo "⚠️  gRPC reflection result: $GRPC_SERVICES"
    fi
else
    echo "⚠️  grpcurl not available, skipping gRPC tests"
    echo "💡 Install grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

# 显示服务状态总结
echo ""
echo "📊 Service Status Summary:"
echo "=========================="
if [ -n "$HTTP_PORT_CHECK" ]; then
    echo "✅ HTTP Service: RUNNING (port 8080)"
else
    echo "❌ HTTP Service: NOT RUNNING"
fi

if [ -n "$GRPC_PORT_CHECK" ]; then
    echo "✅ gRPC Service: RUNNING (port 9093)"
else
    echo "❌ gRPC Service: NOT RUNNING"
fi

# 清理
echo ""
echo "🧹 Stopping services..."
kill $SERVICE_PID 2>/dev/null
wait $SERVICE_PID 2>/dev/null

echo "✅ Test completed" 