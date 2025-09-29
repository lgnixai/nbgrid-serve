#!/bin/bash

# Teable Go Backend 开发环境启动脚本
# 自动配置代理和启动服务

echo "🚀 启动 Teable Go Backend 开发环境..."

# 设置Go环境变量
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

echo "✅ Go环境变量已设置:"
echo "   GO111MODULE=$GO111MODULE"
echo "   GOPROXY=$GOPROXY"

# 启动代理 15236 (如果尚未设置)
if [ -z "$HTTP_PROXY" ]; then
    echo "🌐 启动代理 15236..."
    export HTTP_PROXY=http://127.0.0.1:15236
    export HTTPS_PROXY=http://127.0.0.1:15236
    export http_proxy=http://127.0.0.1:15236
    export https_proxy=http://127.0.0.1:15236
    echo "✅ 代理已设置: $HTTP_PROXY"
else
    echo "🌐 当前代理: $HTTP_PROXY"
fi

# 检查依赖服务
echo "🔍 检查依赖服务..."

# 检查PostgreSQL
if ! pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
    echo "⚠️  PostgreSQL 未运行，请先启动 PostgreSQL"
    echo "   可以使用: brew services start postgresql@17"
fi

# 检查Redis
if ! redis-cli ping >/dev/null 2>&1; then
    echo "⚠️  Redis 未运行，请先启动 Redis"
    echo "   可以使用: brew services start redis"
fi

echo ""
echo "🎯 启动后端服务..."
echo "   服务地址: http://localhost:3000"
echo "   Swagger文档: http://localhost:3000/swagger/index.html"
echo ""

# 启动Go服务
go run cmd/server/main.go
