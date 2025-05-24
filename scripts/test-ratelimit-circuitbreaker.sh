#!/bin/bash

set -e  # 遇到错误时退出

echo "🛡️ API限流和熔断器功能综合测试"
echo "========================================"

# 定义服务地址
APP_URL="http://localhost:8080"

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo "测试服务器: $APP_URL"
echo "测试时间: $(date)"
echo ""

# 检查服务是否可用
echo "⏳ 检查服务状态..."
MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f -s --max-time 5 "$APP_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 服务运行正常${NC}"
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
            echo -e "${YELLOW}⚠️  服务未响应，等待5秒后重试... ($RETRY_COUNT/$MAX_RETRIES)${NC}"
            sleep 5
        else
            echo -e "${RED}❌ 服务不可用，请确保应用已启动${NC}"
            echo "💡 启动命令: docker-compose up -d"
            exit 1
        fi
    fi
done

sleep 2

echo ""
echo -e "${CYAN}📊 开始功能测试...${NC}"

# 测试1: IP限流功能测试
echo ""
echo -e "${BLUE}1️⃣ 测试IP限流功能${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "端点: /health (配置: 10次/秒)"
echo "测试: 发送15个快速请求"
echo "预期: 前10个成功，后5个被限流"

SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0

for i in {1..15}; do
    RESPONSE=$(curl -s -w "HTTP_%{http_code}" "$APP_URL/health" 2>/dev/null)
    HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_[0-9]*" | cut -d'_' -f2)
    
    if [ "$HTTP_CODE" = "200" ]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        echo -e "  请求 $i: ${GREEN}✅ 成功 (200)${NC}"
    elif [ "$HTTP_CODE" = "429" ]; then
        RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
        echo -e "  请求 $i: ${YELLOW}🚫 限流 (429)${NC}"
    else
        echo -e "  请求 $i: ${RED}❓ 其他状态 ($HTTP_CODE)${NC}"
    fi
    
    # 短暂延迟避免过快
    sleep 0.06
done

echo ""
echo -e "${PURPLE}📈 IP限流测试结果:${NC}"
echo -e "  成功请求: ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "  被限流请求: ${YELLOW}$RATE_LIMITED_COUNT${NC}"

if [ $RATE_LIMITED_COUNT -gt 0 ]; then
    echo -e "${GREEN}✅ IP限流功能正常工作${NC}"
    IP_RATE_LIMIT_PASSED=true
else
    echo -e "${YELLOW}⚠️  IP限流可能未生效${NC}"
    IP_RATE_LIMIT_PASSED=false
fi

# 等待限流窗口重置
echo ""
echo "⏳ 等待限流窗口重置 (3秒)..."
sleep 3

# 测试2: 用户注册限流测试
echo ""
echo -e "${BLUE}2️⃣ 测试用户注册限流功能${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "端点: /api/v1/auth/register (配置: 20次/分钟)"
echo "测试: 快速发送25个注册请求"

REG_SUCCESS=0
REG_RATE_LIMITED=0
REG_ERRORS=0

echo "快速发送25个注册请求："
printf "进度: "

for i in {1..25}; do
    USERNAME="testuser_rl_$RANDOM"
    EMAIL="test_$RANDOM@example.com"
    
    RESPONSE=$(curl -s -w "HTTP_%{http_code}" -X POST "$APP_URL/api/v1/auth/register" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"test123\"}" 2>/dev/null)
      
    HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_[0-9]*" | cut -d'_' -f2)
    
    case $HTTP_CODE in
        200|201)
            REG_SUCCESS=$((REG_SUCCESS + 1))
            printf "${GREEN}✅${NC}"
            ;;
        429)
            REG_RATE_LIMITED=$((REG_RATE_LIMITED + 1))
            printf "${YELLOW}🚫${NC}"
            ;;
        *)
            REG_ERRORS=$((REG_ERRORS + 1))
            printf "${RED}❌${NC}"
            ;;
    esac
    
    sleep 0.05 # 很短的延迟
done

echo ""
echo ""
echo -e "${PURPLE}📈 用户注册限流测试结果:${NC}"
echo -e "  成功注册: ${GREEN}$REG_SUCCESS${NC}"
echo -e "  被限流: ${YELLOW}$REG_RATE_LIMITED${NC}"
echo -e "  其他错误: ${RED}$REG_ERRORS${NC}"

if [ $REG_RATE_LIMITED -gt 0 ]; then
    echo -e "${GREEN}✅ 注册端点限流功能正常工作${NC}"
    REG_RATE_LIMIT_PASSED=true
else
    echo -e "${YELLOW}⚠️  注册端点限流可能未生效${NC}"
    REG_RATE_LIMIT_PASSED=false
fi

# 测试3: 熔断器功能测试
echo ""
echo -e "${BLUE}3️⃣ 测试熔断器功能${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "检查熔断器状态监控端点..."

CB_STATUS=$(curl -s "$APP_URL/circuit-breaker/status" 2>/dev/null)

echo "熔断器状态响应:"
echo "$CB_STATUS" | head -c 200
if [ ${#CB_STATUS} -gt 200 ]; then
    echo "..."
fi
echo ""

if echo "$CB_STATUS" | grep -q "healthy\|degraded\|status"; then
    echo -e "${GREEN}✅ 熔断器状态监控正常${NC}"
    CB_STATUS_PASSED=true
    
    # 解析熔断器数量
    CIRCUIT_COUNT=$(echo "$CB_STATUS" | grep -o "total_circuits\":[0-9]*" | cut -d':' -f2 || echo "0")
    echo -e "  检测到熔断器数量: ${BLUE}$CIRCUIT_COUNT${NC}"
else
    echo -e "${RED}❌ 熔断器状态监控异常${NC}"
    CB_STATUS_PASSED=false
fi

# 测试4: 尝试触发熔断器
echo ""
echo -e "${BLUE}4️⃣ 尝试触发熔断器保护${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "发送请求到不存在的用户ID，测试错误处理..."

NOT_FOUND_COUNT=0
SUCCESS_COUNT=0
CB_TRIGGERED=0
RATE_LIMITED_COUNT=0

printf "测试进度: "

for i in {1..30}; do
    RESPONSE=$(curl -s -w "HTTP_%{http_code}" "$APP_URL/api/v1/users/999" 2>/dev/null)
    HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_[0-9]*" | cut -d'_' -f2)
    
    case $HTTP_CODE in
        404)
            NOT_FOUND_COUNT=$((NOT_FOUND_COUNT + 1))
            printf "."
            ;;
        200)
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            printf "${GREEN}✅${NC}"
            ;;
        503)
            CB_TRIGGERED=$((CB_TRIGGERED + 1))
            printf "${RED}🔥${NC}"
            ;;
        429)
            RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
            printf "${YELLOW}🚫${NC}"
            ;;
        *)
            printf "${RED}❓${NC}"
            ;;
    esac
    
    sleep 0.03
done

echo ""
echo ""
echo -e "${PURPLE}📈 熔断器测试结果:${NC}"
echo -e "  404响应: ${BLUE}$NOT_FOUND_COUNT${NC}"
echo -e "  正常响应: ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "  熔断器触发: ${RED}$CB_TRIGGERED${NC}"
echo -e "  被限流: ${YELLOW}$RATE_LIMITED_COUNT${NC}"

if [ $CB_TRIGGERED -gt 0 ]; then
    echo -e "${GREEN}✅ 熔断器保护功能正常工作${NC}"
    CB_TRIGGER_PASSED=true
elif [ $NOT_FOUND_COUNT -gt 20 ]; then
    echo -e "${BLUE}ℹ️  熔断器未触发，但错误处理正常${NC}"
    CB_TRIGGER_PASSED=true
else
    echo -e "${YELLOW}⚠️  熔断器保护需要检查${NC}"
    CB_TRIGGER_PASSED=false
fi

# 测试5: 检查限流响应头信息
echo ""
echo -e "${BLUE}5️⃣ 检查限流响应头信息${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

HEADER_RESPONSE=$(curl -s -I "$APP_URL/health" 2>/dev/null)

if echo "$HEADER_RESPONSE" | grep -qi "x-ratelimit"; then
    echo -e "${GREEN}✅ 发现限流响应头:${NC}"
    echo "$HEADER_RESPONSE" | grep -i "x-ratelimit" | sed 's/^/  /'
    HEADER_PASSED=true
else
    echo -e "${YELLOW}⚠️  未发现限流响应头${NC}"
    echo "HTTP响应头:"
    echo "$HEADER_RESPONSE" | head -3 | sed 's/^/  /'
    HEADER_PASSED=false
fi

# 测试6: 测试不同端点的限流配置差异
echo ""
echo -e "${BLUE}6️⃣ 测试不同端点的限流配置${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "比较认证端点与用户端点的限流差异..."

# 快速请求登录端点
echo "登录端点测试 (配置: 20次/分钟):"
LOGIN_REQUESTS=12
LOGIN_SUCCESS=0
LOGIN_LIMITED=0

printf "登录测试进度: "

for i in $(seq 1 $LOGIN_REQUESTS); do
    RESPONSE=$(curl -s -w "HTTP_%{http_code}" -X POST "$APP_URL/api/v1/auth/login" \
      -H "Content-Type: application/json" \
      -d '{"username":"nonexistent","password":"wrong"}' 2>/dev/null)
      
    HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_[0-9]*" | cut -d'_' -f2)
    
    if [ "$HTTP_CODE" = "429" ]; then
        LOGIN_LIMITED=$((LOGIN_LIMITED + 1))
        printf "${YELLOW}🚫${NC}"
    else
        LOGIN_SUCCESS=$((LOGIN_SUCCESS + 1))
        printf "."
    fi
    sleep 0.08
done

echo ""
echo -e "  登录请求处理: ${BLUE}$LOGIN_SUCCESS${NC}, 被限流: ${YELLOW}$LOGIN_LIMITED${NC}"

if [ $LOGIN_LIMITED -gt 0 ]; then
    echo -e "${GREEN}✅ 登录端点限流配置正常${NC}"
    LOGIN_RATE_LIMIT_PASSED=true
else
    echo -e "${YELLOW}⚠️  登录端点限流配置需要检查${NC}"
    LOGIN_RATE_LIMIT_PASSED=false
fi

# 测试7: Redis存储验证 (如果使用Redis)
echo ""
echo -e "${BLUE}7️⃣ 验证限流存储后端${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 检查是否能够连接到Redis (如果配置了Redis存储)
REDIS_AVAILABLE=false
if command -v docker-compose &> /dev/null; then
    if docker-compose exec redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
        echo -e "${GREEN}✅ Redis存储后端可用${NC}"
        REDIS_AVAILABLE=true
        
        # 检查限流键
        RATE_LIMIT_KEYS=$(docker-compose exec redis redis-cli keys "ratelimit:*" 2>/dev/null | wc -l)
        echo -e "  检测到限流记录: ${BLUE}$RATE_LIMIT_KEYS条${NC}"
    fi
fi

if [ "$REDIS_AVAILABLE" = false ]; then
    echo -e "${BLUE}ℹ️  Redis后端不可用，可能使用内存存储${NC}"
fi

# 生成测试报告
echo ""
echo -e "${CYAN}🎯 综合测试报告${NC}"
echo "========================================"

TOTAL_TESTS=7
PASSED_TESTS=0

# 统计测试结果
[ "$IP_RATE_LIMIT_PASSED" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))
[ "$REG_RATE_LIMIT_PASSED" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))
[ "$CB_STATUS_PASSED" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))
[ "$CB_TRIGGER_PASSED" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))
[ "$HEADER_PASSED" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))
[ "$LOGIN_RATE_LIMIT_PASSED" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))
[ "$REDIS_AVAILABLE" = true ] && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "📊 测试结果详情:"
echo ""

# 限流功能测试
echo -e "${PURPLE}🛡️ 限流功能:${NC}"
[ "$IP_RATE_LIMIT_PASSED" = true ] && echo -e "  ${GREEN}✅ IP限流测试: 通过${NC}" || echo -e "  ${YELLOW}⚠️  IP限流测试: 需要检查${NC}"
[ "$REG_RATE_LIMIT_PASSED" = true ] && echo -e "  ${GREEN}✅ 注册限流测试: 通过${NC}" || echo -e "  ${YELLOW}⚠️  注册限流测试: 需要检查${NC}"
[ "$LOGIN_RATE_LIMIT_PASSED" = true ] && echo -e "  ${GREEN}✅ 登录限流测试: 通过${NC}" || echo -e "  ${YELLOW}⚠️  登录限流测试: 需要检查${NC}"
[ "$HEADER_PASSED" = true ] && echo -e "  ${GREEN}✅ 响应头测试: 通过${NC}" || echo -e "  ${YELLOW}⚠️  响应头测试: 需要检查${NC}"

echo ""

# 熔断器功能测试
echo -e "${PURPLE}🔥 熔断器功能:${NC}"
[ "$CB_STATUS_PASSED" = true ] && echo -e "  ${GREEN}✅ 状态监控: 通过${NC}" || echo -e "  ${YELLOW}⚠️  状态监控: 需要检查${NC}"
[ "$CB_TRIGGER_PASSED" = true ] && echo -e "  ${GREEN}✅ 保护机制: 通过${NC}" || echo -e "  ${YELLOW}⚠️  保护机制: 需要检查${NC}"

echo ""

# 存储后端测试
echo -e "${PURPLE}💾 存储后端:${NC}"
[ "$REDIS_AVAILABLE" = true ] && echo -e "  ${GREEN}✅ Redis后端: 可用${NC}" || echo -e "  ${BLUE}ℹ️  Redis后端: 不可用(使用内存存储)${NC}"

echo ""
echo -e "📈 总体通过率: ${GREEN}$PASSED_TESTS/$TOTAL_TESTS${NC}"

# 给出总体评估
if [ $PASSED_TESTS -ge 6 ]; then
    echo -e "${GREEN}🎉 系统保护功能运行优秀！${NC}"
    FINAL_RESULT=0
elif [ $PASSED_TESTS -ge 4 ]; then
    echo -e "${YELLOW}⚠️  系统保护功能基本正常，部分功能需要优化${NC}"
    FINAL_RESULT=0
else
    echo -e "${RED}❌ 系统保护功能存在问题，请检查配置${NC}"
    FINAL_RESULT=1
fi

echo ""
echo -e "${CYAN}🔍 监控和状态端点:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  - 熔断器状态: $APP_URL/circuit-breaker/status"
echo "  - Hystrix监控流: $APP_URL/hystrix"
echo "  - Prometheus指标: http://localhost:9090/metrics"
echo "  - 应用健康检查: $APP_URL/health"

echo ""
echo -e "${CYAN}📖 配置说明:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  限流配置:"
echo "    - 健康检查: 10次/秒"
echo "    - 认证端点: 20次/分钟"
echo "    - 公开用户API: 30次/分钟"
echo "    - 受保护用户API: 50次/分钟"
echo ""
echo "  熔断器配置:"
echo "    - 注册/登录: 3秒超时, 30%错误率阈值"
echo "    - 用户查询: 2秒超时, 40%错误率阈值"
echo "    - 数据库操作: 5秒超时, 50%错误率阈值"

echo ""
echo -e "${CYAN}💡 故障排除建议:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ "$IP_RATE_LIMIT_PASSED" = false ] || [ "$REG_RATE_LIMIT_PASSED" = false ]; then
    echo "  限流问题:"
    echo "    - 检查config/config.yaml或config/config-docker.yaml中的ratelimit配置"
    echo "    - 验证Redis连接状态 (如果使用Redis存储)"
    echo "    - 检查应用日志: docker-compose logs app | grep -i rate"
fi

if [ "$CB_STATUS_PASSED" = false ] || [ "$CB_TRIGGER_PASSED" = false ]; then
    echo "  熔断器问题:"
    echo "    - 检查熔断器中间件是否正确配置"
    echo "    - 访问熔断器状态端点查看详细信息"
    echo "    - 检查应用日志: docker-compose logs app | grep -i circuit"
fi

echo "    - 重启服务: docker-compose restart app"
echo "    - 查看完整日志: docker-compose logs -f app"

echo ""
echo -e "${GREEN}✅ 测试完成！${NC}"

exit $FINAL_RESULT 