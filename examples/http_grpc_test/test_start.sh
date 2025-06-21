#!/bin/bash

echo "🚀 Starting HTTP service test..."

# 启动服务
./http-grpc-test &
SERVICE_PID=$!

echo "📍 Service PID: $SERVICE_PID"

# 等待3秒让服务启动
echo "⏳ Waiting 3 seconds for service to start..."
sleep 3

# 检查端口
echo "🔍 Checking port 8080..."
lsof -i :8080

# 测试HTTP端点
echo "🧪 Testing HTTP endpoint..."
curl -s http://localhost:8080/health || echo "❌ HTTP request failed"

# 清理
echo "🧹 Stopping service..."
kill $SERVICE_PID 2>/dev/null
wait $SERVICE_PID 2>/dev/null

echo "✅ Test completed" 