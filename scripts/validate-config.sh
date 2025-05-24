#!/bin/bash

# 配置验证脚本
# 验证限流配置是否正确

echo "=== 配置验证脚本 ==="
echo ""

# 检查配置文件是否存在
CONFIG_FILES=("config/config.yaml" "config/config-docker.yaml")

for config_file in "${CONFIG_FILES[@]}"; do
    echo "检查配置文件: $config_file"
    
    if [ ! -f "$config_file" ]; then
        echo "❌ 配置文件不存在: $config_file"
        continue
    fi
    
    echo "✅ 配置文件存在"
    
    # 检查限流配置是否存在
    if grep -q "ratelimit:" "$config_file"; then
        echo "✅ 包含限流配置"
        
        # 检查必要的配置项
        required_fields=("enabled" "store_type" "default_config" "endpoints")
        
        for field in "${required_fields[@]}"; do
            if grep -A 20 "ratelimit:" "$config_file" | grep -q "$field:"; then
                echo "  ✅ $field 配置存在"
            else
                echo "  ❌ $field 配置缺失"
            fi
        done
        
        # 检查限流格式
        echo "  检查限流格式:"
        grep -A 20 "ratelimit:" "$config_file" | grep -E ':\s*"[0-9]+-[SMHD]"' | while read -r line; do
            echo "    ✅ $line"
        done
        
    else
        echo "❌ 缺少限流配置"
    fi
    
    echo ""
done

# 验证编译
echo "验证代码编译:"
if go build -o /tmp/test-build . >/dev/null 2>&1; then
    echo "✅ 代码编译成功"
    rm -f /tmp/test-build
else
    echo "❌ 代码编译失败"
    echo "运行 'go build .' 查看详细错误"
fi

echo ""
echo "=== 验证完成 ===" 