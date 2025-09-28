#!/bin/bash

# 测试运行脚本
# 用于运行Go后端的各种测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${2}${1}${NC}"
}

# 打印标题
print_title() {
    echo ""
    print_message "================================================" $BLUE
    print_message "$1" $BLUE
    print_message "================================================" $BLUE
    echo ""
}

# 检查Go环境
check_go_env() {
    print_title "检查Go环境"
    
    if ! command -v go &> /dev/null; then
        print_message "错误: Go未安装或不在PATH中" $RED
        exit 1
    fi
    
    go_version=$(go version)
    print_message "Go版本: $go_version" $GREEN
    
    # 检查Go模块
    if [ ! -f "go.mod" ]; then
        print_message "错误: 当前目录不是Go模块" $RED
        exit 1
    fi
    
    print_message "Go模块检查通过" $GREEN
}

# 安装测试依赖
install_test_deps() {
    print_title "安装测试依赖"
    
    # 安装必要的测试包
    go mod download
    
    # 安装测试工具
    if ! command -v testify &> /dev/null; then
        print_message "安装testify..." $YELLOW
        go install github.com/stretchr/testify@latest
    fi
    
    print_message "测试依赖安装完成" $GREEN
}

# 运行单元测试
run_unit_tests() {
    print_title "运行单元测试"
    
    # 运行所有单元测试
    print_message "运行所有单元测试..." $YELLOW
    go test -v ./internal/domain/... -race -coverprofile=coverage.out
    
    # 生成覆盖率报告
    if [ -f "coverage.out" ]; then
        print_message "生成覆盖率报告..." $YELLOW
        go tool cover -html=coverage.out -o coverage.html
        go tool cover -func=coverage.out
        
        # 计算覆盖率百分比
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        print_message "总覆盖率: $coverage" $GREEN
        
        # 如果覆盖率低于80%，显示警告
        coverage_num=$(echo $coverage | sed 's/%//')
        if (( $(echo "$coverage_num < 80" | bc -l) )); then
            print_message "警告: 测试覆盖率低于80%" $YELLOW
        fi
    fi
}

# 运行集成测试
run_integration_tests() {
    print_title "运行集成测试"
    
    # 检查测试数据库配置
    if [ -z "$TEST_DB_HOST" ]; then
        print_message "设置测试数据库配置..." $YELLOW
        export TEST_DB_HOST=${TEST_DB_HOST:-localhost}
        export TEST_DB_USER=${TEST_DB_USER:-postgres}
        export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-password}
        export TEST_DB_NAME=${TEST_DB_NAME:-teable_test}
        export TEST_DB_SSL_MODE=${TEST_DB_SSL_MODE:-disable}
    fi
    
    # 创建测试数据库
    print_message "创建测试数据库..." $YELLOW
    go run test_setup.go -create-db
    
    # 运行集成测试
    print_message "运行集成测试..." $YELLOW
    go test -v ./internal/infrastructure/... -tags=integration -race
    
    # 清理测试数据库
    print_message "清理测试数据库..." $YELLOW
    go run test_setup.go -drop-db
}

# 运行性能测试
run_performance_tests() {
    print_title "运行性能测试"
    
    print_message "运行基准测试..." $YELLOW
    go test -v ./internal/domain/... -bench=. -benchmem
    
    print_message "运行压力测试..." $YELLOW
    go test -v ./internal/domain/... -bench=BenchmarkStress -benchtime=30s
}

# 运行代码质量检查
run_code_quality() {
    print_title "运行代码质量检查"
    
    # 格式化检查
    print_message "检查代码格式..." $YELLOW
    if ! go fmt ./...; then
        print_message "代码格式检查失败" $RED
        exit 1
    fi
    print_message "代码格式检查通过" $GREEN
    
    # 静态分析
    if command -v golangci-lint &> /dev/null; then
        print_message "运行静态分析..." $YELLOW
        golangci-lint run
        print_message "静态分析完成" $GREEN
    else
        print_message "golangci-lint未安装，跳过静态分析" $YELLOW
    fi
    
    # 安全检查
    if command -v gosec &> /dev/null; then
        print_message "运行安全检查..." $YELLOW
        gosec ./...
        print_message "安全检查完成" $GREEN
    else
        print_message "gosec未安装，跳过安全检查" $YELLOW
    fi
}

# 生成测试报告
generate_test_report() {
    print_title "生成测试报告"
    
    # 创建报告目录
    mkdir -p test_reports
    
    # 生成测试结果报告
    print_message "生成测试结果报告..." $YELLOW
    go test -v ./... -json > test_reports/test_results.json 2>&1 || true
    
    # 生成覆盖率报告
    if [ -f "coverage.out" ]; then
        print_message "生成覆盖率报告..." $YELLOW
        go tool cover -html=coverage.out -o test_reports/coverage.html
        go tool cover -func=coverage.out > test_reports/coverage.txt
    fi
    
    # 生成基准测试报告
    print_message "生成基准测试报告..." $YELLOW
    go test -v ./... -bench=. -benchmem > test_reports/benchmark.txt 2>&1 || true
    
    print_message "测试报告已生成到 test_reports/ 目录" $GREEN
}

# 清理测试文件
cleanup() {
    print_title "清理测试文件"
    
    # 删除覆盖率文件
    rm -f coverage.out coverage.html
    
    # 删除测试数据库
    if [ -f "test_setup.go" ]; then
        print_message "清理测试数据库..." $YELLOW
        go run test_setup.go -drop-db 2>/dev/null || true
    fi
    
    print_message "清理完成" $GREEN
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示此帮助信息"
    echo "  -u, --unit              仅运行单元测试"
    echo "  -i, --integration       仅运行集成测试"
    echo "  -p, --performance       仅运行性能测试"
    echo "  -q, --quality           仅运行代码质量检查"
    echo "  -a, --all               运行所有测试（默认）"
    echo "  -r, --report            生成测试报告"
    echo "  -c, --cleanup           清理测试文件"
    echo "  -v, --verbose           详细输出"
    echo ""
    echo "环境变量:"
    echo "  TEST_DB_HOST            测试数据库主机（默认: localhost）"
    echo "  TEST_DB_USER            测试数据库用户（默认: postgres）"
    echo "  TEST_DB_PASSWORD        测试数据库密码（默认: password）"
    echo "  TEST_DB_NAME            测试数据库名称（默认: teable_test）"
    echo "  TEST_DB_SSL_MODE        测试数据库SSL模式（默认: disable）"
    echo ""
    echo "示例:"
    echo "  $0                      # 运行所有测试"
    echo "  $0 -u                   # 仅运行单元测试"
    echo "  $0 -i -r                # 运行集成测试并生成报告"
    echo "  $0 -a -r -v             # 运行所有测试，生成报告，详细输出"
}

# 主函数
main() {
    local run_unit=false
    local run_integration=false
    local run_performance=false
    local run_quality=false
    local run_all=true
    local generate_report=false
    local cleanup_files=false
    local verbose=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -u|--unit)
                run_unit=true
                run_all=false
                shift
                ;;
            -i|--integration)
                run_integration=true
                run_all=false
                shift
                ;;
            -p|--performance)
                run_performance=true
                run_all=false
                shift
                ;;
            -q|--quality)
                run_quality=true
                run_all=false
                shift
                ;;
            -a|--all)
                run_all=true
                shift
                ;;
            -r|--report)
                generate_report=true
                shift
                ;;
            -c|--cleanup)
                cleanup_files=true
                shift
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            *)
                print_message "未知选项: $1" $RED
                show_help
                exit 1
                ;;
        esac
    done
    
    # 设置详细输出
    if [ "$verbose" = true ]; then
        set -x
    fi
    
    # 检查Go环境
    check_go_env
    
    # 安装测试依赖
    install_test_deps
    
    # 运行测试
    if [ "$run_all" = true ]; then
        run_unit_tests
        run_integration_tests
        run_performance_tests
        run_code_quality
    else
        if [ "$run_unit" = true ]; then
            run_unit_tests
        fi
        if [ "$run_integration" = true ]; then
            run_integration_tests
        fi
        if [ "$run_performance" = true ]; then
            run_performance_tests
        fi
        if [ "$run_quality" = true ]; then
            run_code_quality
        fi
    fi
    
    # 生成报告
    if [ "$generate_report" = true ]; then
        generate_test_report
    fi
    
    # 清理文件
    if [ "$cleanup_files" = true ]; then
        cleanup
    fi
    
    print_title "测试完成"
    print_message "所有测试已成功完成！" $GREEN
}

# 运行主函数
main "$@"
