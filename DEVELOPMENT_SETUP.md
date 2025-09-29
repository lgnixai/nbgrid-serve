# Teable Go Backend 开发环境配置

## 环境配置

### Fish Shell 配置

已为您配置了以下Fish shell环境变量：

```fish
# Go 环境变量
set -gx GO111MODULE on
set -gx GOPROXY https://goproxy.cn,direct
```

### 代理配置

已配置代理管理函数，支持以下命令：

```fish
# 查看当前代理状态
proxy

# 开启代理（指定端口）
proxy on 15236

# 关闭代理
proxy off
```

### 自动启动配置

Fish shell启动时会自动：
1. 设置Go环境变量
2. 自动启动代理15236（如果尚未设置）
3. 显示当前配置状态

## 启动开发环境

### 方法1: 使用启动脚本（推荐）

```bash
# Bash版本
./start_dev.sh

# Fish版本
./start_dev.fish
```

### 方法2: 手动启动

```bash
# 设置环境变量
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

# 启动代理
proxy on 15236

# 启动服务
go run cmd/server/main.go
```

## 服务地址

- 后端API: http://localhost:3000
- Swagger文档: http://localhost:3000/swagger/index.html

## 依赖服务

确保以下服务正在运行：

```bash
# PostgreSQL
brew services start postgresql@17

# Redis
brew services start redis
```

## 代理说明

代理15236用于：
- 加速Go模块下载
- 访问被墙的依赖包
- 提高开发效率

如果不需要代理，可以使用 `proxy off` 关闭。
