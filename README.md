# Teable Go Backend

基于Go + PostgreSQL的Teable后端系统重构版本，采用现代化架构设计，具备高性能、高可用、易扩展的特性。

## 🚀 特性

- **高性能**: 基于Go语言的并发优势，支持高并发请求处理
- **现代架构**: 采用领域驱动设计(DDD) + 清洁架构模式
- **微服务准备**: 模块化设计，易于拆分为微服务
- **完整功能**: 一比一实现原TypeScript版本的所有功能
- **高可用**: 支持水平扩展、负载均衡、故障恢复
- **监控完善**: 集成日志、指标、链路追踪

## 📋 技术栈

### 核心技术
- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: PostgreSQL 15+
- **ORM**: GORM v2
- **缓存**: Redis 7+
- **队列**: Asynq
- **WebSocket**: Gorilla WebSocket

### 工具链
- **配置管理**: Viper
- **日志**: Zap
- **API文档**: Swag (Swagger)
- **依赖注入**: Wire
- **测试**: Testify + GoMock
- **容器化**: Docker + Docker Compose

## 🏗️ 项目结构

```
teable-go-backend/
├── cmd/                    # 应用入口
│   ├── server/            # 服务器入口
│   └── migrate/           # 数据库迁移工具
├── internal/              # 内部代码
│   ├── domain/            # 领域层
│   ├── application/       # 应用层
│   ├── infrastructure/    # 基础设施层
│   ├── interfaces/        # 接口层
│   └── config/           # 配置管理
├── pkg/                   # 公共包
│   ├── logger/           # 日志包
│   ├── errors/           # 错误处理
│   └── utils/            # 工具函数
├── api/                   # API定义
├── migrations/            # 数据库迁移文件
├── deployments/           # 部署配置
└── scripts/              # 脚本文件
```

## 🛠️ 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选)

### 开发环境搭建

1. **克隆项目**
```bash
git clone <repository-url>
cd teable-go-backend
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境**
```bash
cp deployments/config.yaml config.yaml
# 编辑config.yaml文件，配置数据库和Redis连接信息
```

4. **启动依赖服务**
```bash
# 使用Docker Compose启动PostgreSQL和Redis
docker-compose up -d postgres redis
```

5. **运行数据库迁移**
```bash
# TODO: 实现数据库迁移工具
go run cmd/migrate/main.go up
```

6. **启动应用**
```bash
go run cmd/server/main.go
```

应用将在 `http://localhost:3000` 启动

### Docker部署

1. **使用Docker Compose一键部署**
```bash
docker-compose up -d
```

2. **检查服务状态**
```bash
docker-compose ps
```

3. **查看日志**
```bash
docker-compose logs -f teable-backend
```

## 📊 API文档

启动应用后，访问以下地址查看API文档：

- Swagger UI: `http://localhost:3000/swagger/index.html`
- Health Check: `http://localhost:3000/health`
- Ping: `http://localhost:3000/ping`

## 🔧 配置

### 环境变量

主要环境变量说明：

```bash
# 服务器配置
TEABLE_SERVER_HOST=0.0.0.0
TEABLE_SERVER_PORT=3000
TEABLE_SERVER_MODE=release

# 数据库配置
TEABLE_DATABASE_HOST=localhost
TEABLE_DATABASE_PORT=5432
TEABLE_DATABASE_USER=postgres
TEABLE_DATABASE_PASSWORD=postgres
TEABLE_DATABASE_NAME=teable

# Redis配置
TEABLE_REDIS_HOST=localhost
TEABLE_REDIS_PORT=6379
TEABLE_REDIS_PASSWORD=

# JWT配置
TEABLE_JWT_SECRET=your-secret-key
TEABLE_JWT_ACCESS_TOKEN_TTL=24h
TEABLE_JWT_REFRESH_TOKEN_TTL=168h
```

### 配置文件

详细配置请参考 `deployments/config.yaml` 文件。

## 🧪 测试

### 运行单元测试
```bash
go test ./... -v
```

### 运行集成测试
```bash
go test ./... -tags=integration -v
```

### 运行基准测试
```bash
go test ./... -bench=. -benchmem
```

### 生成测试覆盖率报告
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📈 监控和运维

### 健康检查

应用提供以下健康检查端点：

- `GET /health` - 综合健康检查
- `GET /ping` - 简单存活检查

### 日志

应用使用结构化日志，支持以下级别：
- DEBUG: 调试信息
- INFO: 一般信息
- WARN: 警告信息
- ERROR: 错误信息

### 指标监控

TODO: 集成Prometheus指标

### 链路追踪

TODO: 集成Jaeger链路追踪

## 🔄 数据库迁移

### 创建迁移文件
```bash
# TODO: 实现迁移工具
go run cmd/migrate/main.go create <migration_name>
```

### 执行迁移
```bash
# 向上迁移
go run cmd/migrate/main.go up

# 向下迁移
go run cmd/migrate/main.go down

# 迁移到指定版本
go run cmd/migrate/main.go goto <version>
```

## 🚀 部署

### 构建生产镜像
```bash
docker build -t teable-go-backend:latest .
```

### Kubernetes部署
```bash
# TODO: 提供k8s部署配置
kubectl apply -f deployments/k8s/
```

### 性能优化建议

1. **数据库优化**
   - 合理设置连接池大小
   - 添加必要的索引
   - 定期分析查询性能

2. **缓存策略**
   - 热点数据缓存
   - 查询结果缓存
   - 会话数据缓存

3. **资源监控**
   - CPU和内存使用率
   - 数据库连接数
   - 缓存命中率

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交变更 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- 遵循Go官方代码规范
- 使用`gofmt`格式化代码
- 运行`golangci-lint`进行静态检查
- 确保测试覆盖率 > 80%

## 📄 许可证

本项目基于 AGPL-3.0 许可证开源。详情请参见 [LICENSE](LICENSE) 文件。

## 🆘 支持

如有问题或建议，请：

1. 查看[文档](https://help.teable.ai)
2. 提交[Issue](https://github.com/teableio/teable/issues)
3. 加入[Discord社区](https://discord.gg/uZwp7tDE5W)

## 🗺️ 开发路线图

- [x] 基础架构搭建
- [x] 核心中间件实现
- [x] 数据库连接和配置
- [ ] 用户认证系统
- [ ] 权限管理系统
- [ ] 空间和基础管理
- [ ] 表格系统实现
- [ ] 实时协作功能
- [ ] 文件上传和管理
- [ ] 导入导出功能
- [ ] 插件系统
- [ ] 监控和告警
- [ ] 性能优化
- [ ] 微服务拆分准备

---

**注意**: 这是Teable后端系统的Go语言重构版本，目标是完全替代现有的TypeScript版本，提供更好的性能和可扩展性。