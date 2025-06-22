#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 服务地址
HTTP_BASE_URL="http://localhost:8080"

echo -e "${CYAN}🧪 HTTP + gRPC 集成测试${NC}"
echo -e "${CYAN}==============================${NC}"
echo -e "${YELLOW}📋 测试说明: HTTP 接口会调用后端的 gRPC 服务${NC}"
echo ""

# 等待服务启动
echo -e "${BLUE}⏳ 等待服务启动...${NC}"
sleep 2

# 1. 健康检查
echo -e "${GREEN}1️⃣  测试 HTTP 健康检查${NC}"
curl -s -X GET "$HTTP_BASE_URL/health" | jq '.'
echo ""

# 2. gRPC 健康检查
echo -e "${GREEN}2️⃣  测试 gRPC 健康检查 (通过 HTTP)${NC}"
curl -s -X GET "$HTTP_BASE_URL/grpc/health" | jq '.'
echo ""

# 3. Ping 测试
echo -e "${GREEN}3️⃣  测试 Ping${NC}"
curl -s -X GET "$HTTP_BASE_URL/ping" | jq '.'
echo ""

# 4. 列出用户 (初始数据)
echo -e "${GREEN}4️⃣  列出所有用户 (初始数据)${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users" | jq '.'
echo ""

# 5. 获取特定用户
echo -e "${GREEN}5️⃣  获取用户 ID=1${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users/1" | jq '.'
echo ""

# 6. 创建新用户
echo -e "${GREEN}6️⃣  创建新用户${NC}"
NEW_USER=$(curl -s -X POST "$HTTP_BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "phone": "+1-555-0199"
  }')
echo "$NEW_USER" | jq '.'

# 提取新创建用户的ID
NEW_USER_ID=$(echo "$NEW_USER" | jq -r '.user.id')
echo -e "${YELLOW}📝 新用户ID: $NEW_USER_ID${NC}"
echo ""

# 7. 更新用户
echo -e "${GREEN}7️⃣  更新用户 ID=$NEW_USER_ID${NC}"
curl -s -X PUT "$HTTP_BASE_URL/api/users/$NEW_USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Test User",
    "email": "updated@example.com",
    "phone": "+1-555-0200"
  }' | jq '.'
echo ""

# 8. 验证更新
echo -e "${GREEN}8️⃣  验证用户更新${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users/$NEW_USER_ID" | jq '.'
echo ""

# 9. 分页测试
echo -e "${GREEN}9️⃣  测试分页 (page=1, page_size=2)${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users?page=1&page_size=2" | jq '.'
echo ""

# 10. 搜索测试
echo -e "${GREEN}🔟 测试搜索 (search=Alice)${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users?search=Alice" | jq '.'
echo ""

# 11. 错误测试 - 获取不存在的用户
echo -e "${GREEN}1️⃣1️⃣ 错误测试: 获取不存在的用户${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users/999" | jq '.'
echo ""

# 12. 错误测试 - 创建重复邮箱的用户
echo -e "${GREEN}1️⃣2️⃣ 错误测试: 创建重复邮箱的用户${NC}"
curl -s -X POST "$HTTP_BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Duplicate Email User",
    "email": "alice@example.com"
  }' | jq '.'
echo ""

# 13. 删除测试用户
echo -e "${GREEN}1️⃣3️⃣ 删除测试用户 ID=$NEW_USER_ID${NC}"
curl -s -X DELETE "$HTTP_BASE_URL/api/users/$NEW_USER_ID" | jq '.'
echo ""

# 14. 验证删除
echo -e "${GREEN}1️⃣4️⃣ 验证用户已删除${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users/$NEW_USER_ID" | jq '.'
echo ""

# 15. 最终用户列表
echo -e "${GREEN}1️⃣5️⃣ 最终用户列表${NC}"
curl -s -X GET "$HTTP_BASE_URL/api/users" | jq '.'
echo ""

echo -e "${CYAN}✅ 测试完成!${NC}"
echo -e "${PURPLE}💡 提示: 所有 HTTP 接口都在后台调用 gRPC 服务${NC}"
echo -e "${PURPLE}🔍 查看服务日志可以看到 gRPC 调用的详细信息${NC}" 