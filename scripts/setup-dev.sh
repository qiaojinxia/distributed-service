#!/bin/bash

# 开发环境设置脚本
echo "🛠️  设置分布式微服务开发环境..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.19+"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 设置Go bin路径
export PATH=$PATH:$(go env GOPATH)/bin

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

# 生成Swagger文档
echo "📚 生成 API 文档..."
swag init -g main.go --output ./docs

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
git add docs/
EOF
    
    chmod +x .git/hooks/pre-commit
    echo "✅ Git pre-commit hook 已设置"
fi

echo ""
echo "🎉 开发环境设置完成！"
echo ""
echo "📖 可用命令："
echo "  swag init           - 生成 Swagger 文档"
echo "  golangci-lint run   - 运行代码检查"
echo "  goimports -w .      - 格式化代码"
echo ""
echo "🚀 启动开发服务器："
echo "  go run main.go                    # 默认配置"
echo "  CONFIG_FILE=config/config-local.yaml go run main.go  # 本地模式"
echo ""
echo "🐳 Docker 部署："
echo "  ./deploy.sh         # 完整部署"
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