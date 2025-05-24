#!/bin/bash

# 限流测试脚本
# 测试不同端点的限流配置

set -e  # 遇到错误时退出

BASE_URL="http://localhost:8080"
HEALTH_ENDPOINT="$BASE_URL/health"
AUTH_REGISTER_ENDPOINT="$BASE_URL/api/v1/auth/register"

echo "🛡️ API限流功能测试脚本"
echo "========================================"
echo "测试服务器: $BASE_URL"
echo "测试时间: $(date)"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查服务是否可用
echo "⏳ 检查服务状态..."
if ! curl -f -s --max-time 5 "$HEALTH_ENDPOINT" > /dev/null 2>&1; then
    echo -e "${RED}❌ 服务不可用，请确保应用已启动${NC}"
    echo "💡 启动命令: docker-compose up -d 或 go run main.go"
    exit 1
fi
echo -e "${GREEN}✅ 服务运行正常${NC}"
echo ""

# 测试1: 健康检查端点限流 (10-S，每秒10次)
echo -e "${BLUE}1️⃣ 测试健康检查端点限流${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "配置: 10次/秒 (10-S)"
echo "端点: $HEALTH_ENDPOINT"
echo "测试: 发送15个快速请求"
echo "预期: 前10个成功(200)，后5个被限流(429)"
echo ""

SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0
ERROR_COUNT=0

echo "发送请求："
for i in {1..15}; do
    response=$(curl -s -w "%{http_code}" -o /dev/null "$HEALTH_ENDPOINT" 2>/dev/null)
    
    case $response in
        "200")
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            echo -e "请求 $i: ${GREEN}✅ 成功 (200)${NC}"
            ;;
        "429")
            RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
            echo -e "请求 $i: ${YELLOW}🚫 限流 (429)${NC}"
            ;;
        *)
            ERROR_COUNT=$((ERROR_COUNT + 1))
            echo -e "请求 $i: ${RED}❌ 错误 ($response)${NC}"
            ;;
    esac
    
    # 每秒内发送请求
    if [ $i -lt 15 ]; then
        sleep 0.06  # 稍微调整以适应网络延迟
    fi
done

echo ""
echo "📊 健康检查端点测试结果："
echo -e "  成功请求: ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "  被限流请求: ${YELLOW}$RATE_LIMITED_COUNT${NC}"
echo -e "  错误请求: ${RED}$ERROR_COUNT${NC}"

if [ $RATE_LIMITED_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ 限流功能正常工作${NC}"
else
    echo -e "${YELLOW}⚠️  限流可能未生效，请检查配置${NC}"
fi

echo ""
echo "⏳ 等待限流窗口重置 (2秒)..."
sleep 2

# 测试2: 认证注册端点限流 (20-M，每分钟20次)
echo ""
echo -e "${BLUE}2️⃣ 测试认证注册端点限流${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "配置: 20次/分钟 (20-M)"
echo "端点: $AUTH_REGISTER_ENDPOINT"
echo "测试: 快速发送25个注册请求"
echo "预期: 前20个处理(200/400/409)，后5个被限流(429)"
echo ""

REG_SUCCESS=0
REG_BUSINESS_ERROR=0
REG_RATE_LIMITED=0
REG_ERROR=0

echo "发送注册请求："
for i in {1..25}; do
    USERNAME="test_rl_$RANDOM"
    EMAIL="test_$RANDOM@example.com"
    
    response=$(curl -s -w "%{http_code}" -o /dev/null -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"password123\"}" \
        "$AUTH_REGISTER_ENDPOINT" 2>/dev/null)
    
    case $response in
        200|201)
            REG_SUCCESS=$((REG_SUCCESS + 1))
            echo -e "请求 $i: ${GREEN}✅ 成功 ($response)${NC}"
            ;;
        400|409)
            REG_BUSINESS_ERROR=$((REG_BUSINESS_ERROR + 1))
            echo -e "请求 $i: ${BLUE}ℹ️  业务错误 ($response)${NC}"
            ;;
        429)
            REG_RATE_LIMITED=$((REG_RATE_LIMITED + 1))
            echo -e "请求 $i: ${YELLOW}🚫 限流 ($response)${NC}"
            ;;
        *)
            REG_ERROR=$((REG_ERROR + 1))
            echo -e "请求 $i: ${RED}❌ 其他错误 ($response)${NC}"
            ;;
    esac
    
    # 快速发送以触发限流
    sleep 0.02
done

echo ""
echo "📊 注册端点测试结果："
echo -e "  成功注册: ${GREEN}$REG_SUCCESS${NC}"
echo -e "  业务错误: ${BLUE}$REG_BUSINESS_ERROR${NC}"
echo -e "  被限流: ${YELLOW}$REG_RATE_LIMITED${NC}"
echo -e "  其他错误: ${RED}$REG_ERROR${NC}"

if [ $REG_RATE_LIMITED -gt 0 ]; then
    echo -e "${GREEN}✅ 注册端点限流功能正常工作${NC}"
else
    echo -e "${YELLOW}⚠️  注册端点限流可能未生效${NC}"
fi

# 测试3: 检查限流响应头
echo ""
echo -e "${BLUE}3️⃣ 检查限流响应头信息${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

HEADER_RESPONSE=$(curl -s -i "$HEALTH_ENDPOINT" 2>/dev/null)

echo "健康检查端点响应头："
if echo "$HEADER_RESPONSE" | grep -qi "x-ratelimit"; then
    echo -e "${GREEN}✅ 发现限流响应头:${NC}"
    echo "$HEADER_RESPONSE" | grep -i "x-ratelimit" | while read -r line; do
        echo "  $line"
    done
    
    # 解析限流信息
    LIMIT=$(echo "$HEADER_RESPONSE" | grep -i "x-ratelimit-limit" | cut -d' ' -f2 | tr -d '\r\n')
    REMAINING=$(echo "$HEADER_RESPONSE" | grep -i "x-ratelimit-remaining" | cut -d' ' -f2 | tr -d '\r\n')
    RESET=$(echo "$HEADER_RESPONSE" | grep -i "x-ratelimit-reset" | cut -d' ' -f2 | tr -d '\r\n')
    
    echo ""
    echo "📊 限流状态解析："
    [ -n "$LIMIT" ] && echo -e "  限制数量: ${BLUE}$LIMIT${NC}"
    [ -n "$REMAINING" ] && echo -e "  剩余次数: ${GREEN}$REMAINING${NC}"
    [ -n "$RESET" ] && echo -e "  重置时间: ${BLUE}$(date -r "$RESET" 2>/dev/null || echo "$RESET")${NC}"
    HEADER_PASSED=true
else
    echo -e "${YELLOW}⚠️  未发现限流响应头${NC}"
    echo "HTTP状态行:"
    echo "$HEADER_RESPONSE" | head -1
    echo ""
    echo "完整响应头:"
    echo "$HEADER_RESPONSE" | head -5 | sed 's/^/  /'
    HEADER_PASSED=false
fi

# 测试4: 快速触发限流验证
echo ""
echo -e "${BLUE}4️⃣ 快速限流触发验证${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "连续快速请求健康检查端点..."

QUICK_SUCCESS=0
QUICK_LIMITED=0

for i in {1..12}; do
    response=$(curl -s -w "%{http_code}" -o /dev/null "$HEALTH_ENDPOINT" 2>/dev/null)
    
    if [ "$response" = "200" ]; then
        QUICK_SUCCESS=$((QUICK_SUCCESS + 1))
        echo -n "✅"
    elif [ "$response" = "429" ]; then
        QUICK_LIMITED=$((QUICK_LIMITED + 1))
        echo -n "🚫"
    else
        echo -n "❌"
    fi
    
    sleep 0.05
done

echo ""
echo ""
echo "📊 快速测试结果:"
echo -e "  成功: ${GREEN}$QUICK_SUCCESS${NC}, 限流: ${YELLOW}$QUICK_LIMITED${NC}"

# 测试总结
echo ""
echo -e "${BLUE}🎯 测试总结${NC}"
echo "========================================"

TOTAL_TESTS=4
PASSED_TESTS=0

# 评估测试结果
if [ $RATE_LIMITED_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ 健康检查限流测试: 通过${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}⚠️  健康检查限流测试: 需要检查${NC}"
fi

if [ $REG_RATE_LIMITED -gt 0 ]; then
    echo -e "${GREEN}✅ 注册端点限流测试: 通过${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}⚠️  注册端点限流测试: 需要检查${NC}"
fi

if [ "$HEADER_PASSED" = true ]; then
    echo -e "${GREEN}✅ 限流响应头测试: 通过${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}⚠️  限流响应头测试: 需要检查${NC}"
fi

if [ $QUICK_LIMITED -gt 0 ]; then
    echo -e "${GREEN}✅ 快速限流触发测试: 通过${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${YELLOW}⚠️  快速限流触发测试: 需要检查${NC}"
fi

echo ""
echo -e "📊 测试通过率: ${GREEN}$PASSED_TESTS/$TOTAL_TESTS${NC}"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "${GREEN}🎉 所有限流测试通过！${NC}"
    exit 0
elif [ $PASSED_TESTS -gt 2 ]; then
    echo -e "${YELLOW}⚠️  大部分测试通过，部分功能需要检查${NC}"
    exit 0
else
    echo -e "${RED}❌ 多个测试失败，请检查限流配置${NC}"
    exit 1
fi

echo ""
echo "💡 说明:"
echo "- ✅ 成功: 请求正常处理"
echo "- 🚫 限流: 触发限流保护 (429状态码)"
echo "- ℹ️  业务错误: 应用层错误 (如用户已存在)"
echo "- ❌ 错误: 其他HTTP错误"
echo ""
echo "📖 限流配置:"
echo "- 健康检查: 10次/秒"
echo "- 认证端点: 20次/分钟"
echo "- 用户公开API: 30次/分钟"
echo "- 用户保护API: 50次/分钟"
echo ""
echo "🔍 监控端点:"
echo "- Prometheus指标: http://localhost:9090/metrics"
echo "- 应用状态: http://localhost:8080/health" 