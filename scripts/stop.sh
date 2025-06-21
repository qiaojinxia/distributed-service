#!/bin/bash

echo "🛑 停止分布式服务..."

if [ "$1" = "docker" ]; then
    echo "🐳 停止Docker容器..."
    docker-compose down
    echo "✅ Docker容器已停止"
    
elif [ "$1" = "k8s" ]; then
    echo "☸️  停止Kubernetes服务..."
    kubectl delete -f k8s/
    echo "✅ Kubernetes服务已停止"
    
else
    echo "💻 停止本地进程..."
    pkill -f "./build/app" 2>/dev/null || echo "⚠️  未找到运行的进程"
    echo "✅ 本地进程已停止"
fi

echo "🎉 服务已停止!"
