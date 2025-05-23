#!/bin/bash

# 一键部署脚本
echo "🚀 开始部署分布式微服务..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 停止并删除现有容器
echo "🧹 清理现有容器..."
docker-compose down --remove-orphans

# 删除现有镜像（可选）
read -p "是否删除现有镜像重新构建？(y/N): " rebuild
if [[ $rebuild =~ ^[Yy]$ ]]; then
    echo "🗑️  删除现有镜像..."
    docker rmi distributed-service_app 2>/dev/null || true
fi

# 构建并启动服务
echo "🏗️  构建并启动服务..."
docker-compose up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose ps

# 健康检查
echo "🔍 执行健康检查..."
health_checks=(
    "http://localhost:8080/health"
    "http://localhost:8500/v1/status/leader"
    "http://localhost:15672"
    "http://localhost:3000"
    "http://localhost:9091"
)

for url in "${health_checks[@]}"; do
    echo "检查 $url ..."
    if curl -f -s "$url" > /dev/null; then
        echo "✅ $url 响应正常"
    else
        echo "❌ $url 响应异常"
    fi
done

# 显示访问地址
echo ""
echo "🎉 部署完成！服务访问地址："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📱 主应用:           http://localhost:8080"
echo "📚 API 文档:         http://localhost:8080/swagger/index.html"
echo "🏥 健康检查:         http://localhost:8080/health"
echo "📊 指标监控:         http://localhost:9090/metrics"
echo "🗂️  服务注册中心:     http://localhost:8500"
echo "🐰 RabbitMQ 管理:    http://localhost:15672 (guest/guest)"
echo "📈 Prometheus:      http://localhost:9091"
echo "📊 Grafana:         http://localhost:3000 (admin/admin123)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 提示："
echo "  - 查看日志: docker-compose logs -f app"
echo "  - 停止服务: docker-compose down"
echo "  - 重启服务: docker-compose restart"
echo ""

# 显示测试命令
echo "🧪 JWT 认证 API 测试命令："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "# 1. 注册新用户"
echo "curl -X POST http://localhost:8080/api/v1/auth/register \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"username\":\"newuser\",\"email\":\"new@example.com\",\"password\":\"password123\"}'"
echo ""
echo "# 2. 用户登录 (使用测试账号: admin/password123)"
echo "curl -X POST http://localhost:8080/api/v1/auth/login \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"username\":\"admin\",\"password\":\"password123\"}'"
echo ""
echo "# 3. 使用 JWT Token 访问受保护的 API (替换 YOUR_JWT_TOKEN)"
echo "curl -X POST http://localhost:8080/api/v1/users \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \\"
echo "  -d '{\"username\":\"protecteduser\",\"email\":\"protected@example.com\",\"password\":\"password123\"}'"
echo ""
echo "# 4. 获取用户信息 (无需认证)"
echo "curl http://localhost:8080/api/v1/users/1"
echo ""
echo "# 5. 修改密码 (需要认证)"
echo "curl -X POST http://localhost:8080/api/v1/auth/change-password \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \\"
echo "  -d '{\"old_password\":\"password123\",\"new_password\":\"newpassword123\"}'"
echo ""
echo "# 6. 刷新 Token"
echo "curl -X POST http://localhost:8080/api/v1/auth/refresh \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"token\":\"YOUR_JWT_TOKEN\"}'"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🔐 默认测试账号："
echo "  用户名: admin    密码: password123"
echo "  用户名: testuser 密码: password123" 