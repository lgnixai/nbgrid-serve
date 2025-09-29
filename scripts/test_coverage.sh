#!/bin/bash

# 测试覆盖率脚本
# 用于运行所有测试并生成覆盖率报告

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 创建必要的目录
mkdir -p coverage

print_message $GREEN "🚀 开始运行测试覆盖率分析..."

# 清理之前的覆盖率文件
rm -f coverage/*.out coverage/*.html

# 运行单元测试
print_message $YELLOW "\n📋 运行单元测试..."
go test -v -coverprofile=coverage/unit.out -covermode=atomic ./internal/domain/... ./internal/application/...

# 运行集成测试
print_message $YELLOW "\n🔗 运行集成测试..."
go test -v -coverprofile=coverage/integration.out -covermode=atomic ./internal/testing/integration/...

# 运行基准测试
print_message $YELLOW "\n⚡ 运行基准测试..."
go test -bench=. -benchmem -coverprofile=coverage/bench.out -covermode=atomic ./internal/domain/...

# 合并覆盖率文件
print_message $YELLOW "\n📊 合并覆盖率报告..."
echo "mode: atomic" > coverage/coverage.out
tail -q -n +2 coverage/*.out >> coverage/coverage.out

# 生成覆盖率报告
print_message $YELLOW "\n📈 生成覆盖率报告..."
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# 计算总覆盖率
COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}')

print_message $GREEN "\n✅ 测试完成！"
print_message $GREEN "📊 总覆盖率: $COVERAGE"
print_message $GREEN "📄 详细报告: coverage/coverage.html"

# 检查覆盖率阈值
THRESHOLD=70.0
COVERAGE_VALUE=$(echo $COVERAGE | sed 's/%//')

if (( $(echo "$COVERAGE_VALUE < $THRESHOLD" | bc -l) )); then
    print_message $RED "\n❌ 覆盖率低于阈值 ${THRESHOLD}%"
    exit 1
else
    print_message $GREEN "\n✅ 覆盖率满足要求 (>= ${THRESHOLD}%)"
fi

# 生成详细的包级别覆盖率报告
print_message $YELLOW "\n📦 包级别覆盖率："
go tool cover -func=coverage/coverage.out | grep -E "^teable-go-backend" | sort -k3 -nr

# 找出未覆盖的代码行
print_message $YELLOW "\n🔍 未覆盖的关键文件："
go tool cover -func=coverage/coverage.out | grep -E "0.0%" | head -10

# 生成测试报告
print_message $YELLOW "\n📝 生成测试报告..."
cat > coverage/test_report.md << EOF
# 测试覆盖率报告

生成时间: $(date)

## 总体覆盖率
- **覆盖率**: $COVERAGE
- **阈值**: ${THRESHOLD}%
- **状态**: $([ $(echo "$COVERAGE_VALUE >= $THRESHOLD" | bc -l) -eq 1 ] && echo "✅ 通过" || echo "❌ 未通过")

## 测试统计
- 单元测试: ✅ 完成
- 集成测试: ✅ 完成
- 基准测试: ✅ 完成

## 覆盖率详情

### 按包统计
\`\`\`
$(go tool cover -func=coverage/coverage.out | grep -E "^teable-go-backend" | sort -k3 -nr | head -20)
\`\`\`

### 未覆盖的文件
\`\`\`
$(go tool cover -func=coverage/coverage.out | grep -E "0.0%" | head -10)
\`\`\`

## 改进建议
1. 增加未覆盖文件的测试用例
2. 提高关键业务逻辑的测试覆盖率
3. 添加更多的边界条件测试
4. 完善错误处理路径的测试

## 查看详细报告
打开 \`coverage/coverage.html\` 查看详细的代码覆盖率报告
EOF

print_message $GREEN "\n📊 测试报告已生成: coverage/test_report.md"

# 如果在CI环境中，上传覆盖率报告
if [ ! -z "$CI" ]; then
    print_message $YELLOW "\n☁️  上传覆盖率报告到代码覆盖率服务..."
    # 这里可以添加上传到 Codecov 或其他服务的命令
fi