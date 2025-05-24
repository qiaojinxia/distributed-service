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

# 删除现有镜像（可选，默认不删除）
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "💡 选项：是否删除现有镜像重新构建？"
echo "   - 输入 'y' 或 'Y': 删除现有镜像，完全重新构建"
echo "   - 输入 'n' 或直接回车: 保留现有镜像，快速启动"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
# shellcheck disable=SC2162
read -t 10 -p "🤔 请选择 (y/N，10秒后自动选择N): " rebuild
echo ""

if [[ $rebuild =~ ^[Yy]$ ]]; then
    echo "🗑️  删除现有镜像..."
    docker rmi distributed-service_app 2>/dev/null || true
else
    echo "📦 保留现有镜像，进行快速部署..."
fi

# 构建并启动服务
echo "🏗️  构建并启动服务..."
docker-compose up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
for i in {1..6}; do
    # shellcheck disable=SC2003
    echo "   等待中... ($i/6) - $(expr "$i" \* 5)秒"
    sleep 5
done

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose ps

# 健康检查
echo "🔍 执行健康检查..."
health_checks=(
    "应用服务:http://localhost:8080/health"
    "Consul:http://localhost:8500/v1/status/leader"
    "RabbitMQ:http://localhost:15672"
    "Grafana:http://localhost:3000"
    "Prometheus:http://localhost:9091"
)

for check in "${health_checks[@]}"; do
    # shellcheck disable=SC2086
    name=$(echo $check | cut -d: -f1)
    # shellcheck disable=SC2086
    url=$(echo $check | cut -d: -f2-)
    echo -n "检查 $name ... "
    if curl -f -s "$url" > /dev/null 2>&1; then
        echo "✅ 正常"
    else
        echo "❌ 异常"
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
echo "🔍 链路追踪:         http://localhost:16686"
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

# 分布式追踪测试
echo "🔍 分布式链路追踪测试："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "# 1. 运行自动化追踪测试脚本"
echo "./scripts/test-tracing.sh"
echo ""
echo "# 2. 快速验证追踪功能"
echo "./scripts/verify-tracing.sh"
echo ""
echo "# 3. 手动测试追踪功能（带请求ID）"
echo "curl -X POST http://localhost:8080/api/v1/auth/register \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'X-Request-ID: trace-test-register-\$(date +%s)' \\"
echo "  -d '{\"username\":\"traceuser\",\"email\":\"trace@example.com\",\"password\":\"password123\"}'"
echo ""
echo "curl -X POST http://localhost:8080/api/v1/auth/login \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'X-Request-ID: trace-test-login-\$(date +%s)' \\"
echo "  -d '{\"username\":\"traceuser\",\"password\":\"password123\"}'"
echo ""
echo "📊 查看追踪数据："
echo "  1. 访问 Jaeger UI: http://localhost:16686"
echo "  2. 在 Service 下拉框选择 'distributed-service'"
echo "  3. 点击 'Find Traces' 查看追踪链路"
echo "  4. 点击具体 trace 查看详细信息"
echo ""
echo "🎯 追踪验证要点："
echo "  ✓ HTTP 请求层追踪 (路由、状态码、响应时间)"
echo "  ✓ Service 业务层追踪 (用户操作、执行时间)"
echo "  ✓ Repository 数据层追踪 (数据库操作、SQL 时间)"
echo "  ✓ 错误追踪 (异常信息和错误堆栈)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🔐 默认测试账号："
echo "  用户名: admin    密码: password123"
echo "  用户名: testuser 密码: password123"
echo ""
echo "📖 详细文档："
echo "  - 分布式追踪: docs/TRACING.md"
echo "  - 部署文档: README-Docker.md"
echo "  - 项目文档: README.md"
echo ""

# 可选的追踪功能验证（增加超时和更明显的提示）
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 链路追踪功能验证选项："
echo "   - 输入 'y' 或 'Y': 立即运行追踪验证测试"
echo "   - 输入 'n' 或直接回车: 跳过测试，稍后手动运行"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
read -t 15 -p "🤔 是否立即运行链路追踪测试？(y/N，15秒后自动跳过): " run_tracing_test
echo ""

if [[ $run_tracing_test =~ ^[Yy]$ ]]; then
    echo "🚀 开始运行链路追踪测试..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # 等待应用完全启动
    echo "⏳ 等待应用完全启动 (额外等待 10 秒)..."
    sleep 10
    
    # 检查验证脚本是否存在，优先使用快速验证脚本
    if [ -f "./scripts/verify-tracing.sh" ]; then
        echo "🎯 执行快速追踪验证..."
        chmod +x ./scripts/verify-tracing.sh
        ./scripts/verify-tracing.sh || echo "⚠️  验证脚本执行遇到问题，请检查服务状态"
    elif [ -f "./scripts/test-tracing.sh" ]; then
        echo "🎯 执行完整追踪测试..."
        chmod +x ./scripts/test-tracing.sh
        ./scripts/test-tracing.sh || echo "⚠️  测试脚本执行遇到问题，请检查服务状态"
    else
        echo "❌ 追踪测试脚本不存在"
        echo "💡 请手动运行以下命令测试追踪功能："
        echo "   curl -X GET http://localhost:8080/health -H 'X-Request-ID: manual-test'"
    fi
    
    echo ""
    echo "✅ 追踪测试完成！"
    echo "🔍 现在可以访问 Jaeger UI 查看追踪数据: http://localhost:16686"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
else
    echo "⏭️  跳过自动测试，您可以稍后手动运行:"
    echo "   ./scripts/verify-tracing.sh  # 快速验证"
    echo "   ./scripts/test-tracing.sh    # 完整测试"
fi

echo ""
echo "🎉 部署和配置完成！享受您的分布式微服务体验！"
echo ""
echo "🚨 故障排除提示："
echo "  - 如果服务无法访问，请运行: docker-compose ps"
echo "  - 查看应用日志: docker-compose logs -f app"
echo "  - 查看所有日志: docker-compose logs -f"
echo "  - 重启服务: docker-compose restart" 