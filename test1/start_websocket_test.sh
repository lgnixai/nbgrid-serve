#!/bin/bash

# WebSocket测试启动脚本

echo "🚀 启动Teable Go Backend WebSocket测试"
echo "=================================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go"
    exit 1
fi

# 检查Python环境
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3未安装，请先安装Python3"
    exit 1
fi

# 检查websockets库
if ! python3 -c "import websockets" &> /dev/null; then
    echo "📦 安装websockets库..."
    pip3 install websockets
fi

# 进入项目目录
cd "$(dirname "$0")"

# 下载依赖
echo "📦 下载Go依赖..."
go mod tidy

# 构建项目
echo "🔨 构建项目..."
go build -o teable-go-backend cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "❌ 构建失败"
    exit 1
fi

echo "✅ 构建成功"

# 启动服务器
echo "🌐 启动服务器..."
./teable-go-backend &
SERVER_PID=$!

# 等待服务器启动
echo "⏳ 等待服务器启动..."
sleep 3

# 检查服务器是否启动成功
if ! curl -s http://localhost:3000/health > /dev/null; then
    echo "❌ 服务器启动失败"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

echo "✅ 服务器启动成功"

# 运行WebSocket测试
echo "🧪 运行WebSocket测试..."
python3 test_websocket.py

# 停止服务器
echo "🛑 停止服务器..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "✅ 测试完成"

