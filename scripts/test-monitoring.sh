#!/bin/bash

echo "=== 监控系统测试脚本 ==="

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
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local endpoint=$1
    local description=$2
    
    echo -n "测试 $description ... "
    
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$BASE_URL$endpoint")
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ 成功${NC}"
        if [ "$3" = "show" ]; then
            echo "响应内容:"
            cat /tmp/response.json | jq . 2>/dev/null || cat /tmp/response.json
            echo ""
        fi
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        if [ -f /tmp/response.json ]; then
            cat /tmp/response.json
        fi
    fi
    echo ""
}

# 检查服务是否启动
echo -e "${BLUE}1. 检查服务状态${NC}"
test_endpoint "/health" "基础健康检查"

echo -e "${BLUE}2. 测试监控API接口${NC}"
test_endpoint "/api/v1/monitor/system" "系统资源监控" show
test_endpoint "/api/v1/monitor/services" "服务状态监控" show
test_endpoint "/api/v1/monitor/process" "进程监控"
test_endpoint "/api/v1/monitor/stats" "综合监控统计"
test_endpoint "/api/v1/monitor/health" "详细健康检查"
test_endpoint "/api/v1/monitor/metrics/history" "历史指标数据"

echo -e "${BLUE}3. Web界面访问${NC}"
dashboard_response=$(curl -s -w "%{http_code}" -o /tmp/dashboard.html "$BASE_URL/monitor")
dashboard_code="${dashboard_response: -3}"

if [ "$dashboard_code" = "200" ]; then
    echo -e "${GREEN}✓ 监控Dashboard可访问${NC}"
    echo -e "${YELLOW}📊 打开浏览器访问: $BASE_URL/monitor${NC}"
else
    echo -e "${RED}✗ Dashboard访问失败 (HTTP $dashboard_code)${NC}"
fi

echo ""
echo -e "${BLUE}4. 监控功能说明${NC}"
echo "🖥️  系统监控: CPU、内存、磁盘、网络使用情况"
echo "🔧 服务监控: MySQL、Redis、RabbitMQ、Consul、gRPC状态"
echo "⚙️  进程监控: 当前进程的资源使用情况"
echo "📊 实时界面: 自动刷新的美观Web界面"
echo ""

echo -e "${BLUE}5. API接口列表${NC}"
echo "GET $BASE_URL/monitor                    - 监控Dashboard"
echo "GET $BASE_URL/api/v1/monitor/system     - 系统资源统计"
echo "GET $BASE_URL/api/v1/monitor/services   - 服务健康状态"
echo "GET $BASE_URL/api/v1/monitor/process    - 进程统计信息"
echo "GET $BASE_URL/api/v1/monitor/stats      - 综合监控数据"
echo "GET $BASE_URL/api/v1/monitor/health     - 详细健康检查"
echo ""

echo -e "${GREEN}测试完成！${NC}"
echo -e "${YELLOW}💡 提示: 如果某些服务显示不健康，请确保MySQL、Redis、RabbitMQ、Consul等服务已启动${NC}"

# 清理临时文件
rm -f /tmp/response.json /tmp/dashboard.html 