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
    "Jaeger:http://localhost:16686"
)

for check in "${health_checks[@]}"; do
    # shellcheck disable=SC2086
    name=$(echo $check | cut -d: -f1)
    # shellcheck disable=SC2086
    url=$(echo $check | cut -d: -f2-)
    echo -n "检查 $name ... "
    if curl -f -s --max-time 5 --connect-timeout 3 "$url" > /dev/null 2>&1; then
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

# 数据库指标监控测试
echo "📊 数据库指标监控测试："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "# 1. 运行数据库指标测试脚本"
echo "./scripts/test-metrics.sh"
echo ""
echo "# 2. 查看 Prometheus 指标"
echo "curl http://localhost:9090/metrics | grep database_query_duration_seconds"
echo ""
echo "# 3. 在 Prometheus UI 中查询数据库指标"
echo "访问: http://localhost:9091"
echo "查询: database_query_duration_seconds"
echo "查询: rate(database_query_duration_seconds_count[5m])"
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

# 可选的功能验证测试
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 功能测试选项："
echo "   - 输入 '1': 运行数据库指标测试"
echo "   - 输入 '2': 运行链路追踪测试"
echo "   - 输入 '3': 运行限流测试"
echo "   - 输入 '4': 运行限流和熔断器综合测试"
echo "   - 输入 '5': 运行所有基础测试"
echo "   - 输入 '6': 运行所有测试(包括系统保护)"
echo "   - 输入 'n' 或直接回车: 跳过测试，稍后手动运行"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
# shellcheck disable=SC2162
read -t 15 -p "🤔 请选择要运行的测试 (1/2/3/4/5/6/N，15秒后自动跳过): " test_choice
echo ""

# 等待应用完全启动
echo "⏳ 等待应用完全启动 (额外等待 10 秒)..."
sleep 10

case $test_choice in
    "1")
        echo "📊 开始运行数据库指标测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        if [ -f "./scripts/test-metrics.sh" ]; then
            echo "🎯 执行数据库指标测试..."
            chmod +x ./scripts/test-metrics.sh
            ./scripts/test-metrics.sh || echo "⚠️  指标测试脚本执行遇到问题，请检查服务状态"
        else
            echo "❌ 数据库指标测试脚本不存在"
            echo "💡 请手动运行以下命令测试指标功能："
            echo "   curl http://localhost:9090/metrics | grep database_query_duration_seconds"
        fi
        
        echo ""
        echo "✅ 数据库指标测试完成！"
        echo "📊 现在可以访问以下地址查看指标数据:"
        echo "   - Prometheus: http://localhost:9091"
        echo "   - Grafana: http://localhost:3000 (admin/admin123)"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    "2")
        echo "🔍 开始运行链路追踪测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
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
        echo "✅ 链路追踪测试完成！"
        echo "🔍 现在可以访问 Jaeger UI 查看追踪数据: http://localhost:16686"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    "3")
        echo "🛡️ 开始运行限流测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        if [ -f "./scripts/test-ratelimit.sh" ]; then
            echo "🎯 执行限流功能测试..."
            chmod +x ./scripts/test-ratelimit.sh
            ./scripts/test-ratelimit.sh || echo "⚠️  限流测试遇到问题"
        else
            echo "❌ 限流测试脚本不存在"
            echo "💡 请手动运行以下命令测试限流功能："
            echo "   for i in {1..15}; do curl -w 'HTTP_%{http_code}\\n' -o /dev/null http://localhost:8080/health; sleep 0.1; done"
        fi
        
        echo ""
        echo "✅ 限流测试完成！"
        echo "🛡️ 现在可以访问以下地址查看状态:"
        echo "   - 应用健康检查: http://localhost:8080/health"
        echo "   - Prometheus指标: http://localhost:9090/metrics"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    "4")
        echo "🛡️ 开始运行限流和熔断器综合测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        if [ -f "./scripts/test-ratelimit-circuitbreaker.sh" ]; then
            echo "🎯 执行限流和熔断器综合测试..."
            chmod +x ./scripts/test-ratelimit-circuitbreaker.sh
            ./scripts/test-ratelimit-circuitbreaker.sh || echo "⚠️  综合测试遇到问题"
        else
            echo "❌ 限流熔断器测试脚本不存在"
            echo "💡 请手动运行以下命令测试功能："
            echo "   curl http://localhost:8080/circuit-breaker/status"
            echo "   curl http://localhost:8080/health (多次快速请求测试限流)"
        fi
        
        echo ""
        echo "✅ 限流和熔断器测试完成！"
        echo "🛡️ 现在可以访问以下地址查看状态:"
        echo "   - 熔断器状态: http://localhost:8080/circuit-breaker/status"
        echo "   - Hystrix流: http://localhost:8080/hystrix"
        echo "   - 限流验证: 快速访问 http://localhost:8080/health 测试限流"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    "5")
        echo "🚀 开始运行所有基础测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        # 运行数据库指标测试
        echo "📊 1/2 执行数据库指标测试..."
        if [ -f "./scripts/test-metrics.sh" ]; then
            chmod +x ./scripts/test-metrics.sh
            ./scripts/test-metrics.sh || echo "⚠️  指标测试遇到问题"
        else
            echo "❌ 数据库指标测试脚本不存在"
        fi
        
        echo ""
        echo "🔍 2/2 执行链路追踪测试..."
        if [ -f "./scripts/verify-tracing.sh" ]; then
            chmod +x ./scripts/verify-tracing.sh
            ./scripts/verify-tracing.sh || echo "⚠️  追踪测试遇到问题"
        elif [ -f "./scripts/test-tracing.sh" ]; then
            chmod +x ./scripts/test-tracing.sh
            ./scripts/test-tracing.sh || echo "⚠️  追踪测试遇到问题"
        else
            echo "❌ 追踪测试脚本不存在"
        fi
        
        echo ""
        echo "✅ 所有基础测试完成！"
        echo "📊 监控地址: http://localhost:9091 (Prometheus), http://localhost:3000 (Grafana)"
        echo "🔍 追踪地址: http://localhost:16686 (Jaeger)"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    "6")
        echo "🚀 开始运行所有测试(包括系统保护)..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        # 运行数据库指标测试
        echo "📊 1/4 执行数据库指标测试..."
        if [ -f "./scripts/test-metrics.sh" ]; then
            chmod +x ./scripts/test-metrics.sh
            ./scripts/test-metrics.sh || echo "⚠️  指标测试遇到问题"
        else
            echo "❌ 数据库指标测试脚本不存在"
        fi
        
        echo ""
        echo "🔍 2/4 执行链路追踪测试..."
        if [ -f "./scripts/verify-tracing.sh" ]; then
            chmod +x ./scripts/verify-tracing.sh
            ./scripts/verify-tracing.sh || echo "⚠️  追踪测试遇到问题"
        elif [ -f "./scripts/test-tracing.sh" ]; then
            chmod +x ./scripts/test-tracing.sh
            ./scripts/test-tracing.sh || echo "⚠️  追踪测试遇到问题"
        else
            echo "❌ 追踪测试脚本不存在"
        fi
        
        echo ""
        echo "🛡️ 3/4 执行限流测试..."
        if [ -f "./scripts/test-ratelimit.sh" ]; then
            chmod +x ./scripts/test-ratelimit.sh
            ./scripts/test-ratelimit.sh || echo "⚠️  限流测试遇到问题"
        else
            echo "❌ 限流测试脚本不存在"
        fi
        
        echo ""
        echo "🛡️ 4/4 执行限流和熔断器综合测试..."
        if [ -f "./scripts/test-ratelimit-circuitbreaker.sh" ]; then
            chmod +x ./scripts/test-ratelimit-circuitbreaker.sh
            ./scripts/test-ratelimit-circuitbreaker.sh || echo "⚠️  综合测试遇到问题"
        else
            echo "❌ 限流熔断器测试脚本不存在"
        fi
        
        echo ""
        echo "✅ 所有测试完成！"
        echo "📊 指标监控: http://localhost:9091 (Prometheus), http://localhost:3000 (Grafana)"
        echo "🔍 链路追踪: http://localhost:16686 (Jaeger)"
        echo "🛡️ 系统保护: http://localhost:8080/circuit-breaker/status (熔断器), http://localhost:8080/hystrix (流)"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    *)
        echo "⏭️  跳过自动测试，您可以稍后手动运行:"
        echo "   ./scripts/test-metrics.sh                     # 数据库指标测试"
        echo "   ./scripts/verify-tracing.sh                   # 链路追踪快速验证"
        echo "   ./scripts/test-tracing.sh                     # 链路追踪完整测试"
        echo "   ./scripts/test-ratelimit.sh                   # 限流功能测试"
        echo "   ./scripts/test-ratelimit-circuitbreaker.sh    # 限流和熔断器综合测试"
        ;;
esac

echo ""
echo "🎉 部署和配置完成！享受您的分布式微服务体验！"
echo ""
echo "🚨 故障排除提示："
echo "  - 如果服务无法访问，请运行: docker-compose ps"
echo "  - 查看应用日志: docker-compose logs -f app"
echo "  - 查看所有日志: docker-compose logs -f"
echo "  - 重启服务: docker-compose restart" 