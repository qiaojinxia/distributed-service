#!/bin/bash

# 开发环境设置脚本
echo "🛠️  设置分布式微服务开发环境..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.23+"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 设置Go bin路径
# shellcheck disable=SC2155
export PATH=$PATH:$(go env GOPATH)/bin

# 检查 protoc 是否安装
check_protoc() {
    if ! command -v protoc &> /dev/null; then
        echo "⚠️  protoc 未安装，请先安装 Protocol Buffers 编译器"
        echo ""
        echo "📦 安装方式："
        echo "  macOS:   brew install protobuf"
        echo "  Ubuntu:  sudo apt-get install protobuf-compiler"
        echo "  CentOS:  sudo yum install protobuf-compiler"
        echo "  Windows: 从 https://github.com/protocolbuffers/protobuf/releases 下载"
        echo ""
        return 1
    else
        echo "✅ protoc 已安装: $(protoc --version)"
        return 0
    fi
}

# 安装必要的开发工具
echo "📦 安装开发工具..."

# 安装swag (Swagger文档生成)
if ! command -v swag &> /dev/null; then
    echo "🔧 安装 swag (Swagger文档生成器)..."
    go install github.com/swaggo/swag/cmd/swag@latest
else
    echo "✅ swag 已安装: $(swag --version)"
fi

# 安装golangci-lint (代码检查)
if ! command -v golangci-lint &> /dev/null; then
    echo "🔧 安装 golangci-lint (代码检查工具)..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
else
    echo "✅ golangci-lint 已安装"
fi

# 安装goimports (代码格式化)
if ! command -v goimports &> /dev/null; then
    echo "🔧 安装 goimports (代码格式化工具)..."
    go install golang.org/x/tools/cmd/goimports@latest
else
    echo "✅ goimports 已安装"
fi

# 检查 protoc 并提示安装
echo "🚀 检查 gRPC 开发环境..."
check_protoc
PROTOC_AVAILABLE=$?

# 安装 protoc-gen-go (Protocol Buffers Go 生成器)
if ! command -v protoc-gen-go &> /dev/null; then
    echo "🔧 安装 protoc-gen-go (Protocol Buffers Go 生成器)..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
else
    echo "✅ protoc-gen-go 已安装"
fi

# 安装 protoc-gen-go-grpc (gRPC Go 生成器)
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "🔧 安装 protoc-gen-go-grpc (gRPC Go 生成器)..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
else
    echo "✅ protoc-gen-go-grpc 已安装"
fi

# 安装 grpcurl (gRPC 命令行客户端)
if ! command -v grpcurl &> /dev/null; then
    echo "🔧 安装 grpcurl (gRPC 命令行客户端)..."
    go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
else
    echo "✅ grpcurl 已安装: $(grpcurl --version)"
fi

# 测试 protoc 插件是否正确安装
if [ $PROTOC_AVAILABLE -eq 0 ]; then
    echo "🧪 验证 protoc 插件..."
    if protoc --go_out=. --go_opt=paths=source_relative --version > /dev/null 2>&1; then
        echo "✅ protoc-gen-go 插件工作正常"
    else
        echo "⚠️  protoc-gen-go 插件可能未正确安装"
    fi
    
    if protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative --version > /dev/null 2>&1; then
        echo "✅ protoc-gen-go-grpc 插件工作正常"
    else
        echo "⚠️  protoc-gen-go-grpc 插件可能未正确安装"
    fi
fi

# 生成Swagger文档
echo "📚 生成 API 文档..."
swag init -g main.go --output ./docs

# 如果 protoc 可用，生成 gRPC 代码
if [ $PROTOC_AVAILABLE -eq 0 ]; then
    echo "🚀 生成 gRPC 代码..."
    if [ -f "Makefile" ]; then
        make proto
    else
        echo "⚠️  Makefile 不存在，跳过 proto 文件生成"
        echo "💡 您可以手动运行以下命令生成 proto 代码："
        echo "   make proto"
    fi
fi

# 设置Git hooks (可选)
if [ -d ".git" ]; then
    echo "🔗 设置 Git hooks..."
    mkdir -p .git/hooks
    
    # 创建pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "🔍 运行代码检查..."
golangci-lint run
if [ $? -ne 0 ]; then
    echo "❌ 代码检查失败，请修复后再提交"
    exit 1
fi

echo "📚 更新 API 文档..."
swag init -g main.go --output ./docs

echo "🚀 生成 gRPC 代码..."
if [ -f "Makefile" ]; then
    make proto
fi

git add docs/ api/proto/
EOF
    
    chmod +x .git/hooks/pre-commit
    echo "✅ Git pre-commit hook 已设置"
fi

echo ""
echo "🎉 开发环境设置完成！"
echo ""
echo "📖 可用命令："
echo "  swag init              - 生成 Swagger 文档"
echo "  golangci-lint run      - 运行代码检查"
echo "  goimports -w .         - 格式化代码"
echo "  make proto             - 生成 Protocol Buffers 代码"
echo "  grpcurl --help         - gRPC 命令行客户端帮助"
echo ""
echo "🚀 启动开发服务器："
echo "  go run main.go                    # 默认配置"
echo "  CONFIG_FILE=config/config-local.yaml go run main.go  # 本地模式"
echo ""
echo "🐳 Docker 部署："
echo "  ./deploy.sh            - 完整部署"
echo ""
echo "🧪 gRPC 测试："
echo "  cd examples/grpc-client && go run main.go  # 运行示例客户端"
echo "  grpcurl -plaintext localhost:9090 list     # 列出可用服务"
echo ""

# 检查环境变量
echo "🔧 环境变量检查："
echo "  GOPATH: $(go env GOPATH)"
echo "  GOROOT: $(go env GOROOT)"
echo "  PATH包含Go bin: $(echo $PATH | grep -q "$(go env GOPATH)/bin" && echo "✅ 是" || echo "❌ 否")"

# 如果PATH不包含Go bin，提示用户
if ! echo $PATH | grep -q "$(go env GOPATH)/bin"; then
    echo ""
    echo "⚠️  请将以下命令添加到您的 shell 配置文件中："
    echo "export PATH=\$PATH:\$(go env GOPATH)/bin"
    echo ""
    echo "对于 zsh: ~/.zshrc"
    echo "对于 bash: ~/.bashrc"
fi

# gRPC 开发环境状态总结
echo ""
echo "🚀 gRPC 开发环境状态："
if [ $PROTOC_AVAILABLE -eq 0 ]; then
    echo "  ✅ protoc (Protocol Buffers 编译器)"
else
    echo "  ❌ protoc (需要手动安装)"
fi
echo "  ✅ protoc-gen-go (Go protobuf 生成器)"
echo "  ✅ protoc-gen-go-grpc (Go gRPC 生成器)"
echo "  ✅ grpcurl (gRPC 客户端工具)"

if [ $PROTOC_AVAILABLE -ne 0 ]; then
    echo ""
    echo "⚠️  protoc 未安装，gRPC 开发功能将受限"
    echo "   请按照上述提示安装 protoc 后重新运行此脚本"
fi

echo ""
echo "📚 相关文档："
echo "  README.md         - 项目概览和开发指南"
echo "  README-gRPC.md    - gRPC 服务使用指南"
echo "  README-Docker.md  - 容器化部署指南" 