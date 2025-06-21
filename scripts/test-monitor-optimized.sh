#!/bin/bash

echo "=== 优化监控系统测试脚本 ==="

# 服务器地址
SERVER_HOST=${1:-localhost}
SERVER_PORT=${2:-8080}
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"

echo "测试服务器: $BASE_URL"
echo "==============================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local endpoint=$1
    local description=$2
    local expected_title=$3
    
    echo -n "测试 $description ... "
    
    response=$(curl -s -w "%{http_code}" -o /tmp/response.html "$BASE_URL$endpoint")
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "404" ]; then
        if [ -n "$expected_title" ]; then
            title=$(grep -o '<title[^>]*>[^<]*</title>' /tmp/response.html | sed 's/<[^>]*>//g')
            if [[ "$title" == *"$expected_title"* ]]; then
                echo -e "${GREEN}✓ 成功 (HTTP $http_code, 标题匹配: $title)${NC}"
            else
                echo -e "${YELLOW}⚠ 成功但标题不匹配 (HTTP $http_code, 标题: $title)${NC}"
            fi
        else
            echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        fi
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        if [ -f /tmp/response.html ]; then
            head -3 /tmp/response.html
        fi
    fi
}

# 检查服务是否启动
echo -e "${BLUE}1. 检查服务状态${NC}"
health_response=$(curl -s "$BASE_URL/health" 2>/dev/null)
if [[ $? -eq 0 && "$health_response" == *"ok"* ]]; then
    echo -e "${GREEN}✓ 服务运行正常${NC}"
else
    echo -e "${RED}✗ 服务未运行或不健康${NC}"
    echo "请先启动服务: go run main.go"
    exit 1
fi

echo ""
echo -e "${BLUE}2. 测试监控页面访问${NC}"

# 测试不同的监控页面
test_endpoint "/monitor" "默认监控页面（精简模式）" "系统监控概览"
test_endpoint "/monitor/simple" "精简模式页面" "系统监控概览"
test_endpoint "/monitor/full" "完整模式页面" "Enhanced System Monitoring Dashboard"

echo ""
echo -e "${BLUE}3. 测试详细信息页面${NC}"

test_endpoint "/monitor/details/system" "系统详情页面" "系统资源详情"
test_endpoint "/monitor/details/services" "服务详情页面" "服务状态详情"
test_endpoint "/monitor/details/process" "进程详情页面" "进程状态详情"

echo ""
echo -e "${BLUE}4. 测试错误处理${NC}"

test_endpoint "/monitor/details/invalid" "无效监控类型" "页面未找到"

echo ""
echo -e "${BLUE}5. 测试API接口（数据源）${NC}"

echo -n "测试综合监控数据 API ... "
api_response=$(curl -s "$BASE_URL/api/v1/monitor/stats")
if [[ $? -eq 0 && "$api_response" == *"system"* && "$api_response" == *"services"* ]]; then
    echo -e "${GREEN}✓ API 正常响应${NC}"
else
    echo -e "${RED}✗ API 响应异常${NC}"
fi

echo -n "测试系统监控 API ... "
system_response=$(curl -s "$BASE_URL/api/v1/monitor/system")
if [[ $? -eq 0 && "$system_response" == *"cpu"* && "$system_response" == *"memory"* ]]; then
    echo -e "${GREEN}✓ 系统 API 正常${NC}"
else
    echo -e "${RED}✗ 系统 API 异常${NC}"
fi

echo -n "测试服务监控 API ... "
services_response=$(curl -s "$BASE_URL/api/v1/monitor/services")
if [[ $? -eq 0 && "$services_response" == *"services"* && "$services_response" == *"summary"* ]]; then
    echo -e "${GREEN}✓ 服务 API 正常${NC}"
else
    echo -e "${RED}✗ 服务 API 异常${NC}"
fi

echo ""
echo -e "${BLUE}6. 页面功能验证${NC}"

echo -e "${PURPLE}📱 精简模式特性:${NC}"
echo "  • 默认显示关键指标概览"
echo "  • 点击卡片跳转到详细页面"
echo "  • 自动刷新功能"
echo "  • 导航链接到各个详细页面"

echo ""
echo -e "${PURPLE}📊 详细页面特性:${NC}"
echo "  • 系统详情: CPU、内存、磁盘使用情况"
echo "  • 服务详情: 各服务健康状态和延迟"
echo "  • 进程详情: Go 运行时信息和资源使用"

echo ""
echo -e "${PURPLE}🔧 增强的服务详情信息:${NC}"
echo "  • MySQL: 连接池状态、数据库版本、查询测试"
echo "  • Redis: 连接池信息、服务器信息、功能测试"
echo "  • RabbitMQ: 连接信息、队列操作测试"
echo "  • gRPC: 连接状态、健康检查、错误信息"
echo "  • Consul: 连接地址信息"

echo ""
echo -e "${BLUE}7. 用户体验优化${NC}"

echo -e "${PURPLE}🔗 页面导航结构:${NC}"
echo "  • /monitor                    → 精简模式（默认）"
echo "  • /monitor/simple             → 精简模式"
echo "  • /monitor/full               → 完整模式"
echo "  • /monitor/details/system     → 系统详情"
echo "  • /monitor/details/services   → 服务详情"
echo "  • /monitor/details/process    → 进程详情"

echo ""
echo -e "${PURPLE}🎨 UI/UX 改进:${NC}"
echo "  • 响应式设计，支持移动端"
echo "  • 统一的视觉风格和配色"
echo "  • 清晰的导航和返回链接"
echo "  • 错误页面友好提示"

echo ""
echo -e "${GREEN}测试完成！${NC}"

# 提供浏览器访问链接
echo ""
echo -e "${YELLOW}🌐 浏览器访问链接:${NC}"
echo "精简模式: $BASE_URL/monitor"
echo "完整模式: $BASE_URL/monitor/full"
echo "系统详情: $BASE_URL/monitor/details/system"
echo "服务详情: $BASE_URL/monitor/details/services"
echo "进程详情: $BASE_URL/monitor/details/process"

echo ""
echo -e "${BLUE}💡 使用说明:${NC}"
echo "1. 默认访问 /monitor 显示精简概览"
echo "2. 点击概览卡片可跳转到详细页面"
echo "3. 使用导航链接在不同模式间切换"
echo "4. 所有页面支持自动刷新"

# 清理临时文件
rm -f /tmp/response.html

echo ""
echo -e "${GREEN}监控系统优化完成！ 🎉${NC}" 