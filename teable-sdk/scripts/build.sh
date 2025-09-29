#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$DIR"

echo "Building teable-sdk..."
npm run -s build
echo "Build complete."
#!/bin/bash

# Teable SDK 构建脚本

set -e

echo "🚀 开始构建 Teable SDK..."

# 清理之前的构建
echo "🧹 清理之前的构建..."
rm -rf dist/
rm -rf node_modules/.cache/

# 安装依赖
echo "📦 安装依赖..."
npm install

# 运行类型检查
echo "🔍 运行类型检查..."
npx tsc --noEmit

# 运行 ESLint 检查
echo "🔧 运行 ESLint 检查..."
npx eslint src/**/*.ts --fix

# 构建项目
echo "🏗️ 构建项目..."
npx tsc

# 运行测试
echo "🧪 运行测试..."
npm test

# 生成文档
echo "📚 生成文档..."
npx typedoc src/index.ts --out docs

echo "✅ 构建完成！"
echo "📁 构建文件位于: dist/"
echo "📚 文档位于: docs/"
