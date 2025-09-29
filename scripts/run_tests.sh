#!/bin/bash

# 测试运行脚本
# 用于运行项目的所有测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
check_go_environment() {
    print_message "检查Go环境..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go未安装或不在PATH中"
        exit 1
    fi
    
    go_version=$(go version)
    print_success "Go环境检查通过: $go_version"
}

# 安装测试依赖
install_test_dependencies() {
    print_message "安装测试依赖..."
    
    # 安装测试工具
    go install github.com/stretchr/testify/assert@latest
    go install github.com/stretchr/testify/require@latest
    go install github.com/stretchr/testify/mock@latest
    go install github.com/stretchr/testify/suite@latest
    
    # 安装测试覆盖率工具
    go install github.com/axw/gocov/gocov@latest
    go install github.com/AlekSi/gocov-xml@latest
    
    # 安装性能测试工具
    go install github.com/tsenart/vegeta@latest
    
    print_success "测试依赖安装完成"
}

# 运行单元测试
run_unit_tests() {
    print_message "运行单元测试..."
    
    # 设置测试环境变量
    export GO_ENV=test
    export DB_DRIVER=sqlite3
    export DB_DATABASE=:memory:
    export CACHE_TYPE=memory
    export JWT_SECRET=test-secret-key-for-testing-only
    
    # 运行单元测试
    go test -v -race -coverprofile=coverage.out ./internal/domain/... ./internal/application/...
    
    if [ $? -eq 0 ]; then
        print_success "单元测试通过"
    else
        print_error "单元测试失败"
        exit 1
    fi
}

# 运行集成测试
run_integration_tests() {
    print_message "运行集成测试..."
    
    # 设置测试环境变量
    export GO_ENV=test
    export DB_DRIVER=sqlite3
    export DB_DATABASE=:memory:
    export CACHE_TYPE=memory
    export JWT_SECRET=test-secret-key-for-testing-only
    
    # 运行集成测试
    go test -v -race -tags=integration ./internal/testing/...
    
    if [ $? -eq 0 ]; then
        print_success "集成测试通过"
    else
        print_error "集成测试失败"
        exit 1
    fi
}

# 运行API测试
run_api_tests() {
    print_message "运行API测试..."
    
    # 设置测试环境变量
    export GO_ENV=test
    export DB_DRIVER=sqlite3
    export DB_DATABASE=:memory:
    export CACHE_TYPE=memory
    export JWT_SECRET=test-secret-key-for-testing-only
    
    # 运行API测试
    go test -v -race ./internal/interfaces/http/...
    
    if [ $? -eq 0 ]; then
        print_success "API测试通过"
    else
        print_error "API测试失败"
        exit 1
    fi
}

# 运行性能测试
run_benchmark_tests() {
    print_message "运行性能测试..."
    
    # 设置测试环境变量
    export GO_ENV=test
    export DB_DRIVER=sqlite3
    export DB_DATABASE=:memory:
    export CACHE_TYPE=memory
    export JWT_SECRET=test-secret-key-for-testing-only
    
    # 运行性能测试
    go test -v -bench=. -benchmem ./internal/domain/... ./internal/application/...
    
    if [ $? -eq 0 ]; then
        print_success "性能测试完成"
    else
        print_warning "性能测试失败，但继续执行"
    fi
}

# 生成测试覆盖率报告
generate_coverage_report() {
    print_message "生成测试覆盖率报告..."
    
    if [ -f "coverage.out" ]; then
        # 生成HTML覆盖率报告
        go tool cover -html=coverage.out -o coverage.html
        
        # 生成XML覆盖率报告
        gocov convert coverage.out | gocov-xml > coverage.xml
        
        # 显示覆盖率统计
        go tool cover -func=coverage.out
        
        print_success "测试覆盖率报告生成完成"
        print_message "HTML报告: coverage.html"
        print_message "XML报告: coverage.xml"
    else
        print_warning "未找到覆盖率文件，跳过报告生成"
    fi
}

# 运行所有测试
run_all_tests() {
    print_message "开始运行所有测试..."
    
    # 检查Go环境
    check_go_environment
    
    # 安装测试依赖
    install_test_dependencies
    
    # 运行各种测试
    run_unit_tests
    run_integration_tests
    run_api_tests
    run_benchmark_tests
    
    # 生成覆盖率报告
    generate_coverage_report
    
    print_success "所有测试完成！"
}

# 清理测试文件
cleanup() {
    print_message "清理测试文件..."
    
    # 删除覆盖率文件
    rm -f coverage.out coverage.html coverage.xml
    
    # 删除测试数据库文件
    rm -f test.db test.db-shm test.db-wal
    
    # 删除临时文件
    rm -f *.tmp
    
    print_success "清理完成"
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -u, --unit        只运行单元测试"
    echo "  -i, --integration 只运行集成测试"
    echo "  -a, --api         只运行API测试"
    echo "  -b, --benchmark   只运行性能测试"
    echo "  -c, --coverage    生成测试覆盖率报告"
    echo "  -l, --clean       清理测试文件"
    echo "  -h, --help        显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                # 运行所有测试"
    echo "  $0 -u             # 只运行单元测试"
    echo "  $0 -c             # 生成覆盖率报告"
    echo "  $0 -l             # 清理测试文件"
}

# 主函数
main() {
    case "${1:-}" in
        -u|--unit)
            check_go_environment
            install_test_dependencies
            run_unit_tests
            generate_coverage_report
            ;;
        -i|--integration)
            check_go_environment
            install_test_dependencies
            run_integration_tests
            ;;
        -a|--api)
            check_go_environment
            install_test_dependencies
            run_api_tests
            ;;
        -b|--benchmark)
            check_go_environment
            install_test_dependencies
            run_benchmark_tests
            ;;
        -c|--coverage)
            generate_coverage_report
            ;;
        -l|--clean)
            cleanup
            ;;
        -h|--help)
            show_help
            ;;
        "")
            run_all_tests
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
