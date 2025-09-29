# API 基础信息

## 服务信息

### 基础配置
- **服务名称**: Teable Go Backend API
- **版本**: v1.0.0
- **协议**: HTTP/HTTPS
- **数据格式**: JSON
- **字符编码**: UTF-8

### 服务地址
```
开发环境: http://localhost:3000
测试环境: https://api-test.teable.ai
生产环境: https://api.teable.ai
```

### API版本
- **当前版本**: v1
- **版本路径**: `/api/v1` 或 `/api`
- **向后兼容**: 支持至少2个主要版本

## 认证方式

### JWT Token认证
大部分API端点需要JWT token认证，token需要在请求头中携带：

```http
Authorization: Bearer <jwt-token>
```

### Token类型
1. **Access Token**: 短期有效，用于API访问
   - 有效期: 24小时
   - 用途: 常规API调用

2. **Refresh Token**: 长期有效，用于刷新access token
   - 有效期: 7天
   - 用途: 获取新的access token

### 无需认证的端点
以下端点无需认证：
- 健康检查: `GET /health`, `GET /ready`, `GET /alive`
- 用户注册: `POST /api/auth/register`
- 用户登录: `POST /api/auth/login`
- 系统信息: `GET /api/info`

## 请求格式

### Content-Type
所有POST/PUT/PATCH请求都需要设置正确的Content-Type：

```http
Content-Type: application/json
```

### 请求头示例
```http
POST /api/auth/login HTTP/1.1
Host: localhost:3000
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
User-Agent: Teable-Client/1.0
```

### 查询参数
- 分页参数: `limit`, `offset`
- 排序参数: `sort`, `order`
- 筛选参数: `filter`, `search`

## 响应格式

### 统一响应结构

#### 成功响应
```json
{
  "success": true,
  "data": {
    // 具体数据内容
  },
  "message": "操作成功"
}
```

#### 分页响应
```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

#### 错误响应
```json
{
  "error": "错误描述信息",
  "code": "ERROR_CODE",
  "details": "详细错误信息",
  "trace_id": "请求追踪ID"
}
```

### HTTP状态码

| 状态码 | 含义 | 说明 |
|--------|------|------|
| 200 | OK | 请求成功 |
| 201 | Created | 资源创建成功 |
| 204 | No Content | 请求成功，无返回内容 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未认证或认证失败 |
| 403 | Forbidden | 无权限访问 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突 |
| 422 | Unprocessable Entity | 数据验证失败 |
| 429 | Too Many Requests | 请求频率限制 |
| 500 | Internal Server Error | 服务器内部错误 |
| 503 | Service Unavailable | 服务不可用 |

## 数据模型

### 通用字段
大部分实体都包含以下通用字段：

```json
{
  "id": "string",           // 唯一标识符
  "created_at": "datetime", // 创建时间
  "updated_at": "datetime", // 更新时间
  "created_by": "string",   // 创建者ID
  "updated_by": "string"    // 更新者ID
}
```

### ID格式
- 所有ID都是字符串类型
- 使用UUID v4格式或自定义ID生成器
- 示例: `"550e8400-e29b-41d4-a716-446655440000"`

### 时间格式
- 使用ISO 8601格式
- 时区: UTC
- 示例: `"2024-12-19T10:30:00Z"`

## 分页和排序

### 分页参数
```http
GET /api/records?limit=20&offset=0
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| limit | integer | 20 | 每页记录数，最大100 |
| offset | integer | 0 | 偏移量，从0开始 |

### 排序参数
```http
GET /api/records?sort=created_at&order=desc
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| sort | string | created_at | 排序字段 |
| order | string | desc | 排序方向: asc/desc |

## 错误处理

### 错误码分类

#### 认证相关 (AUTH_*)
- `AUTH_TOKEN_INVALID`: Token无效
- `AUTH_TOKEN_EXPIRED`: Token过期
- `AUTH_INSUFFICIENT_PERMISSIONS`: 权限不足

#### 验证相关 (VALIDATION_*)
- `VALIDATION_REQUIRED_FIELD`: 必填字段缺失
- `VALIDATION_INVALID_FORMAT`: 格式不正确
- `VALIDATION_VALUE_TOO_LONG`: 值过长

#### 业务相关 (BUSINESS_*)
- `BUSINESS_RESOURCE_NOT_FOUND`: 资源不存在
- `BUSINESS_RESOURCE_CONFLICT`: 资源冲突
- `BUSINESS_OPERATION_NOT_ALLOWED`: 操作不允许

#### 系统相关 (SYSTEM_*)
- `SYSTEM_INTERNAL_ERROR`: 系统内部错误
- `SYSTEM_SERVICE_UNAVAILABLE`: 服务不可用
- `SYSTEM_RATE_LIMIT_EXCEEDED`: 请求频率超限

### 错误响应示例
```json
{
  "error": "用户不存在",
  "code": "BUSINESS_RESOURCE_NOT_FOUND",
  "details": "用户ID '123' 不存在",
  "trace_id": "req_1234567890abcdef"
}
```

## 速率限制

### 限制规则
- **认证端点**: 5次/分钟
- **常规API**: 1000次/小时
- **上传端点**: 10次/分钟
- **WebSocket连接**: 10个/用户

### 响应头
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## CORS配置

### 允许的域名
- 开发环境: `http://localhost:*`
- 生产环境: `https://*.teable.ai`

### 允许的请求头
```
Authorization, Content-Type, X-Requested-With, Accept
```

### 允许的请求方法
```
GET, POST, PUT, DELETE, PATCH, OPTIONS
```

## 健康检查

### 检查端点
- `GET /health` - 完整健康检查
- `GET /ready` - 就绪检查
- `GET /alive` - 存活检查

### 响应示例
```json
{
  "status": "healthy",
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "uptime": "72h30m15s",
  "services": {
    "database": {
      "status": "healthy",
      "response_time": "5ms"
    },
    "redis": {
      "status": "healthy",
      "response_time": "2ms"
    }
  }
}
```
