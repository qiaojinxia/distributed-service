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

# 部署模式选择
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎯 请选择部署模式："
echo "   1️⃣  仅基础设施 - 启动数据库、缓存、监控等基础服务 (用于本地调试)"
echo "   2️⃣  完整部署 - 启动所有服务包括应用程序"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
# shellcheck disable=SC2162
read -t 15 -p "🤔 请选择部署模式 (1/2，15秒后默认选择完整部署): " deploy_mode
echo ""

# 设置默认值
if [[ -z "$deploy_mode" ]]; then
    deploy_mode="2"
fi

# 根据模式设置相关变量
case $deploy_mode in
    "1")
        echo "🛠️  选择模式: 仅基础设施部署 (本地调试模式)"
        COMPOSE_PROFILES="infrastructure"
        MODE_DESC="基础设施"
        DEPLOY_TYPE="infrastructure"
        ;;
    "2")
        echo "🚀 选择模式: 完整部署 (生产模式)"
        COMPOSE_PROFILES="full"
        MODE_DESC="完整部署"
        DEPLOY_TYPE="full"
        ;;
    *)
        echo "❌ 无效选择，使用默认完整部署模式"
        COMPOSE_PROFILES="full"
        MODE_DESC="完整部署"
        DEPLOY_TYPE="full"
        ;;
esac

echo "📋 当前部署模式: $MODE_DESC"
echo ""

# 停止并删除现有容器
echo "🧹 清理现有容器..."
docker-compose down --remove-orphans

# 删除现有镜像（可选，默认不删除）
if [[ "$DEPLOY_TYPE" == "full" ]]; then
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
fi

# 构建并启动服务
if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "🏗️  启动基础设施服务..."
    # 仅启动基础设施服务（不包括应用程序）
    docker-compose up -d mysql redis rabbitmq consul prometheus grafana jaeger
else
    echo "🏗️  构建并启动所有服务..."
docker-compose up --build -d
fi

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
if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    health_checks=(
        "Consul:http://localhost:8500/v1/status/leader"
        "RabbitMQ:http://localhost:15672"
        "Grafana:http://localhost:3000"
        "Prometheus:http://localhost:9091"
        "Jaeger:http://localhost:16686"
    )
else
health_checks=(
    "应用服务:http://localhost:8080/health"
    "Consul:http://localhost:8500/v1/status/leader"
    "RabbitMQ:http://localhost:15672"
    "Grafana:http://localhost:3000"
    "Prometheus:http://localhost:9091"
    "Jaeger:http://localhost:16686"
)
fi

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
echo "🎉 $MODE_DESC 完成！服务访问地址："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "🛠️  基础设施服务:"
    echo "🗃️  MySQL 数据库:       localhost:3306 (testuser/testpass)"
    echo "🚀 Redis 缓存:         localhost:6379"
    echo "🐰 RabbitMQ 管理:      http://localhost:15672 (guest/guest)"
    echo "🗂️  Consul 服务发现:   http://localhost:8500"
    echo "📊 Prometheus 监控:    http://localhost:9091"
    echo "📈 Grafana 面板:       http://localhost:3000 (admin/admin123)"
    echo "🔍 Jaeger 链路追踪:    http://localhost:16686"
    echo ""
    echo "💻 本地开发启动命令:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "# 在项目根目录执行:"
    echo "go run main.go"
    echo ""
    echo "# 应用启动后可访问:"
    echo "📱 主应用:              http://localhost:8080"
    echo "🚀 gRPC 服务:           grpc://localhost:9090"
    echo "📚 API 文档:            http://localhost:8080/swagger/index.html"
    echo "🏥 健康检查:            http://localhost:8080/health"
    echo "📊 应用指标:            http://localhost:9090/metrics"
else
echo "📱 主应用:           http://localhost:8080"
echo "🚀 gRPC 服务:        grpc://localhost:9090"
echo "📚 API 文档:         http://localhost:8080/swagger/index.html"
echo "🏥 健康检查:         http://localhost:8080/health"
echo "🏥 gRPC 健康检查:    grpc://localhost:9090/grpc.health.v1.Health/Check"
echo "📊 指标监控:         http://localhost:9090/metrics"
echo "🔍 链路追踪:         http://localhost:16686"
echo "🗂️  服务注册中心:     http://localhost:8500"
echo "🐰 RabbitMQ 管理:    http://localhost:15672 (guest/guest)"
echo "📈 Prometheus:      http://localhost:9091"
echo "📊 Grafana:         http://localhost:3000 (admin/admin123)"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo ""
    echo "🔧 本地调试配置说明:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "1. 确认配置文件:"
    echo "   - 使用 config/config.yaml (开发环境配置)"
    echo "   - 数据库连接: localhost:3306"
    echo "   - Redis连接: localhost:6379"
    echo "   - Consul连接: localhost:8500"
    echo ""
    echo "2. 启动应用:"
    echo "   go run main.go"
    echo ""
    echo "3. 验证连接:"
    echo "   - 健康检查: curl http://localhost:8080/health"
    echo "   - API文档: http://localhost:8080/swagger/index.html"
    echo ""
    echo "4. 开发工具:"
    echo "   - 热重载: 推荐使用 air (go install github.com/cosmtrek/air@latest)"
    echo "   - 调试器: 使用 VS Code 或 GoLand 的调试功能"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

echo ""

# 显示测试命令（仅完整部署模式）
if [[ "$DEPLOY_TYPE" == "full" ]]; then
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
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
fi

# 数据库指标监控测试
echo "📊 数据库指标监控测试："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "# 本地调试模式下的指标测试 (启动应用后执行):"
    echo "curl http://localhost:9090/metrics | grep database_query_duration_seconds"
else
echo "# 1. 运行数据库指标测试脚本"
echo "./scripts/test-metrics.sh"
echo ""
echo "# 2. 查看 Prometheus 指标"
echo "curl http://localhost:9090/metrics | grep database_query_duration_seconds"
fi
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
if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "# 本地调试模式下的追踪测试 (启动应用后执行):"
    echo "curl -X POST http://localhost:8080/api/v1/auth/register \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -H 'X-Request-ID: local-debug-trace-\$(date +%s)' \\"
    echo "  -d '{\"username\":\"debuguser\",\"email\":\"debug@example.com\",\"password\":\"password123\"}'"
else
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
fi
echo ""
echo "📊 查看追踪数据："
echo "  1. 访问 Jaeger UI: http://localhost:16686"
echo "  2. 在 Service 下拉框选择 'distributed-service'"
echo "  3. 点击 'Find Traces' 查看追踪链路"
echo "  4. 点击具体 trace 查看详细信息"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "🔐 本地调试用测试账号："
    echo "  用户名: admin    密码: password123"
    echo "  用户名: testuser 密码: password123"
    echo ""
    echo "📖 开发文档："
    echo "  - 项目文档: README.md"
    echo "  - Docker部署: docs/README-Docker.md"
    echo "  - 分布式追踪: docs/TRACING.md"
    echo ""
    echo "💡 开发提示："
    echo "  - 修改代码后应用会自动重启"
    echo "  - 数据库数据保存在 Docker 卷中，重启不会丢失"
    echo "  - 监控数据可在 Grafana 中查看: http://localhost:3000"
    echo ""
else
echo "🔐 默认测试账号："
echo "  用户名: admin    密码: password123"
echo "  用户名: testuser 密码: password123"
echo ""
echo "📖 详细文档："
echo "  - 分布式追踪: docs/TRACING.md"
    echo "  - 部署文档: docs/README-Docker.md"
echo "  - 项目文档: README.md"
    echo "  - API保护测试: test/README_API_Protection_Test.md"
echo ""
fi

# 可选的功能验证测试（仅完整部署模式）
if [[ "$DEPLOY_TYPE" == "full" ]]; then
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 功能测试选项："
echo "   - 输入 '1': 运行数据库指标测试"
echo "   - 输入 '2': 运行链路追踪测试"
    echo "   - 输入 '3': 运行Go API保护测试"
    echo "   - 输入 '4': 运行所有基础测试"
echo "   - 输入 'n' 或直接回车: 跳过测试，稍后手动运行"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
# shellcheck disable=SC2162
    read -t 15 -p "🤔 请选择要运行的测试 (1/2/3/4/N，15秒后自动跳过): " test_choice
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
            echo "🛡️ 开始运行Go API保护测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "🎯 执行API保护功能测试..."
            echo "cd test && go test -v -run TestAPIProtectionWithRealConfig"
            echo ""
            echo "💡 手动运行命令："
            echo "   cd test"
            echo "   go test -v -run TestAPIProtectionWithRealConfig"
            echo "   ./run_api_test.sh"
            echo "   ./demo_api_test.sh"
        echo ""
            echo "✅ API保护测试提示完成！"
        echo "🛡️ 现在可以访问以下地址查看状态:"
        echo "   - 应用健康检查: http://localhost:8080/health"
        echo "   - Prometheus指标: http://localhost:9090/metrics"
            echo "   - API文档: http://localhost:8080/swagger/index.html"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    "4")
        echo "🚀 开始运行所有基础测试..."
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        # 运行数据库指标测试
            echo "📊 1/3 执行数据库指标测试..."
        if [ -f "./scripts/test-metrics.sh" ]; then
            chmod +x ./scripts/test-metrics.sh
            ./scripts/test-metrics.sh || echo "⚠️  指标测试遇到问题"
        else
            echo "❌ 数据库指标测试脚本不存在"
        fi
        
        echo ""
            echo "🔍 2/3 执行链路追踪测试..."
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
            echo "🛡️ 3/3 API保护测试提示..."
            echo "💡 请手动运行API保护测试："
            echo "   cd test && go test -v -run TestAPIProtectionWithRealConfig"
        
        echo ""
        echo "✅ 所有基础测试完成！"
        echo "📊 监控地址: http://localhost:9091 (Prometheus), http://localhost:3000 (Grafana)"
        echo "🔍 追踪地址: http://localhost:16686 (Jaeger)"
            echo "🛡️ API保护测试: cd test && ./run_api_test.sh"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        ;;
    *)
            echo "⏭️  跳过自动测试，您可以稍后手动运行相关测试脚本"
        echo "   ./scripts/test-metrics.sh                     # 数据库指标测试"
        echo "   ./scripts/verify-tracing.sh                   # 链路追踪快速验证"
        echo "   ./scripts/test-tracing.sh                     # HTTP 链路追踪完整测试"
            echo "   cd test && ./run_api_test.sh                  # API保护机制测试"
        ;;
esac
else
    echo "⏭️  基础设施模式：请先启动应用 (go run main.go) 再运行相关测试"
fi

echo ""
if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "🎉 基础设施部署完成！现在可以开始本地开发调试了！"
    echo ""
    echo "📝 下一步操作："
    echo "  1. 在新终端中执行: go run main.go"
    echo "  2. 等待应用启动完成"
    echo "  3. 访问 http://localhost:8080/health 验证应用状态"
    echo "  4. 开始开发和调试"
else
    echo "🎉 完整部署完成！享受您的分布式微服务体验！"
fi

echo ""
echo "🚨 故障排除提示："
echo "  - 如果服务无法访问，请运行: docker-compose ps"
if [[ "$DEPLOY_TYPE" == "infrastructure" ]]; then
    echo "  - 查看基础设施日志: docker-compose logs -f"
    echo "  - 本地应用问题: 检查 go run main.go 输出"
else
echo "  - 查看应用日志: docker-compose logs -f app"
echo "  - 查看所有日志: docker-compose logs -f"
fi
echo "  - 重启服务: docker-compose restart"
echo ""
echo "📚 更多帮助："
echo "  - 项目文档: README.md"
echo "  - Docker 部署: docs/README-Docker.md"
echo "  - gRPC 使用指南: docs/README-gRPC.md"
echo "  - 分布式追踪: docs/TRACING.md"
echo "  - API保护测试: test/README_API_Protection_Test.md" 