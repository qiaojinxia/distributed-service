#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}📦 生成 protobuf 文件...${NC}"

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}❌ protoc 未找到，请先安装 Protocol Buffers 编译器${NC}"
    echo -e "${YELLOW}💡 安装方法:${NC}"
    echo "   macOS: brew install protobuf"
    echo "   Linux: apt-get install protobuf-compiler"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}⚠️  protoc-gen-go 未找到，正在安装...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# 检查 protoc-gen-go-grpc 是否安装
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}⚠️  protoc-gen-go-grpc 未找到，正在安装...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# 创建输出目录
mkdir -p proto/user

# 生成 Go 代码
echo -e "${GREEN}🔨 生成用户服务 protobuf 文件...${NC}"

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/user.proto

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ protobuf 文件生成成功!${NC}"
    echo -e "${GREEN}📁 生成的文件:${NC}"
    echo "   - proto/user/user.pb.go"
    echo "   - proto/user/user_grpc.pb.go"
else
    echo -e "${RED}❌ protobuf 文件生成失败${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 完成!${NC}" 