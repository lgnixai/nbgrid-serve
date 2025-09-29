# 快速开始指南

## 概述

本指南将帮助您快速上手Teable Go Backend API，包括环境配置、认证流程和基本使用示例。

## 环境准备

### 1. 启动服务

确保后端服务已启动：

```bash
# 启动开发服务器
go run cmd/server/main.go

# 或使用预编译的二进制文件
./bin/teable-backend
```

服务启动后，默认监听 `http://localhost:3000`

### 2. 健康检查

首先验证服务是否正常运行：

```bash
curl http://localhost:3000/health
```

预期响应：
```json
{
  "status": "healthy",
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "services": {
    "database": {"status": "healthy"},
    "redis": {"status": "healthy"}
  }
}
```

## 认证流程

### 1. 用户注册

```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!",
    "name": "测试用户"
  }'
```

响应示例：
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "test@example.com",
      "name": "测试用户"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 86400
    }
  }
}
```

### 2. 保存Token

从注册响应中提取 `access_token` 并保存，后续请求需要用到：

```bash
# 设置环境变量（Linux/Mac）
export ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 或保存到文件
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." > token.txt
```

### 3. 用户登录（可选）

如果已有账户，可以直接登录：

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!"
  }'
```

## 基本使用流程

### 1. 获取用户资料

```bash
curl -X GET http://localhost:3000/api/users/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 2. 创建工作空间

```bash
curl -X POST http://localhost:3000/api/spaces \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的第一个工作空间",
    "description": "用于测试API的工作空间",
    "icon": "🏢"
  }'
```

保存返回的空间ID：
```bash
export SPACE_ID="space_550e8400-e29b-41d4-a716-446655440000"
```

### 3. 创建数据表

```bash
curl -X POST http://localhost:3000/api/tables \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "项目列表",
    "description": "管理项目信息的表格",
    "base_id": "'$SPACE_ID'"
  }'
```

保存返回的数据表ID：
```bash
export TABLE_ID="table_550e8400-e29b-41d4-a716-446655440000"
```

### 4. 添加字段

```bash
curl -X POST http://localhost:3000/api/fields \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "table_id": "'$TABLE_ID'",
    "name": "项目名称",
    "type": "text",
    "required": true
  }'
```

### 5. 创建记录

```bash
curl -X POST http://localhost:3000/api/records \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "table_id": "'$TABLE_ID'",
    "data": {
      "项目名称": "API测试项目",
      "状态": "进行中",
      "负责人": "测试用户"
    }
  }'
```

### 6. 查询记录

```bash
curl -X GET "http://localhost:3000/api/records?table_id=$TABLE_ID&limit=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## 使用Postman

### 1. 导入集合

1. 打开Postman
2. 点击 "Import" 按钮
3. 选择 `docs/postman-collection.json` 文件
4. 导入成功后会看到 "Teable Go Backend API" 集合

### 2. 配置环境变量

1. 点击集合右上角的 "..." 菜单
2. 选择 "Edit"
3. 在 "Variables" 标签页中设置：
   - `base_url`: `http://localhost:3000`
   - `access_token`: 从登录响应中获取的token

### 3. 执行测试流程

1. 首先执行 "认证" > "用户注册" 或 "用户登录"
2. 检查响应，token会自动保存到环境变量
3. 依次执行其他API测试

## 常见问题

### 1. 认证失败

**问题**: 收到401 Unauthorized错误

**解决方案**:
- 检查token是否正确
- 确认token是否过期（默认24小时）
- 使用refresh token获取新的access token

```bash
curl -X POST http://localhost:3000/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

### 2. 权限不足

**问题**: 收到403 Forbidden错误

**解决方案**:
- 确认用户是否有足够的权限
- 检查是否是空间的成员
- 联系空间管理员获取权限

### 3. 资源不存在

**问题**: 收到404 Not Found错误

**解决方案**:
- 检查资源ID是否正确
- 确认资源是否已被删除
- 验证用户是否有权限访问该资源

### 4. 请求参数错误

**问题**: 收到400 Bad Request错误

**解决方案**:
- 检查请求体格式是否正确
- 确认必填字段是否已提供
- 验证字段类型和格式

## 开发工具推荐

### 1. API测试工具
- **Postman**: 功能强大的API测试工具
- **Insomnia**: 轻量级的API客户端
- **curl**: 命令行工具，适合脚本化测试

### 2. 代码生成工具
- **OpenAPI Generator**: 根据API文档生成客户端代码
- **Swagger Codegen**: 生成多种语言的SDK

### 3. 监控工具
- **Postman Monitor**: 自动化API监控
- **New Relic**: 应用性能监控
- **DataDog**: 基础设施监控

## 进阶使用

### 1. 批量操作

```bash
# 批量创建记录
curl -X POST http://localhost:3000/api/records/bulk-create \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "table_id": "'$TABLE_ID'",
    "records": [
      {"data": {"项目名称": "项目1", "状态": "已完成"}},
      {"data": {"项目名称": "项目2", "状态": "进行中"}}
    ]
  }'
```

### 2. 高级搜索

```bash
# 高级搜索
curl -X POST http://localhost:3000/api/search/advanced \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "项目",
    "filters": {
      "status": "进行中",
      "created_after": "2024-01-01"
    },
    "sort": {"field": "created_at", "order": "desc"}
  }'
```

### 3. WebSocket实时通信

```javascript
// 建立WebSocket连接
const ws = new WebSocket('ws://localhost:3000/api/ws/socket?token=' + accessToken);

ws.onopen = function() {
  console.log('WebSocket连接已建立');
};

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};

// 发送协作操作
ws.send(JSON.stringify({
  type: 'record_update',
  table_id: 'table_123',
  record_id: 'record_456',
  operation: 'update',
  data: { name: '新名称' }
}));
```

## 获取帮助

### 1. 文档资源
- 完整API文档: `docs/README.md`
- 端点汇总: `docs/api-endpoints.md`
- 认证指南: `docs/authentication.md`

### 2. 技术支持
- 邮箱: support@teable.ai
- 文档: https://docs.teable.ai
- GitHub: https://github.com/teableio/teable

### 3. 社区支持
- 论坛: https://community.teable.ai
- Discord: https://discord.gg/teable
- Stack Overflow: 使用 `teable` 标签

---

*祝您使用愉快！如有问题，请随时联系我们的技术支持团队。*
