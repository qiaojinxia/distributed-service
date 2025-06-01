#!/bin/bash

# API保护演示脚本 - 快速展示主要功能
# 此脚本运行关键测试用例，展示API保护功能

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                    API保护功能演示                              ║"
echo "║          基于 config.yaml 的 HTTP API 限流 & 熔断测试           ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${YELLOW}📋 演示测试用例:${NC}"
echo "  1. 健康检查接口限流 (2 QPS)"
echo "  2. 认证接口限流 (10 req/min)" 
echo "  3. 用户接口限流 (30 req/min)"
echo "  4. 通配符优先级匹配"
echo "  5. 并发请求处理"
echo ""

echo -e "${BLUE}🚀 开始运行API保护演示测试...${NC}"
echo "════════════════════════════════════════════════════════════════"

# 运行关键测试用例
echo -e "${GREEN}测试 1: 健康检查接口限流 (2 QPS)${NC}"
go test ./test -run TestAPIProtectionWithRealConfig/TestHealthCheckRateLimit -v

echo ""
echo -e "${GREEN}测试 2: 认证接口限流 (10 req/min)${NC}"
go test ./test -run TestAPIProtectionWithRealConfig/TestAuthAPIRateLimit -v

echo ""
echo -e "${GREEN}测试 3: 优先级匹配测试${NC}"
go test ./test -run TestAPIProtectionWithRealConfig/TestPriorityMatching -v

echo ""
echo -e "${GREEN}测试 4: 并发请求处理${NC}"
go test ./test -run TestAPIProtectionWithRealConfig/TestConcurrentRequests -v

echo ""
echo "════════════════════════════════════════════════════════════════"
echo -e "${GREEN}🎉 API保护演示完成！${NC}"
echo ""
echo -e "${YELLOW}💡 配置要点:${NC}"
echo "  • 健康检查: /health → 2 QPS"
echo "  • 认证接口: /api/v1/auth/* → 10 req/min"
echo "  • 用户接口: /api/v1/users/* → 30 req/min" 
echo "  • 通用API: /api/* → 100 req/min (兜底)"
echo ""
echo -e "${BLUE}📚 完整测试: ./test/run_api_test.sh all${NC}"
echo -e "${BLUE}📖 文档参考: test/README_API_Protection_Test.md${NC}" 