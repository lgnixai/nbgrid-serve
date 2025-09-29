# 认证与授权

## 概述

Teable Go Backend 使用JWT (JSON Web Token) 进行用户认证和授权。系统支持用户注册、登录、token刷新等完整的认证流程。

## 认证流程

### 1. 用户注册
用户首先需要注册账户，注册成功后会自动登录并返回认证token。

### 2. 用户登录
已注册用户可以通过邮箱和密码登录，获取访问token。

### 3. Token刷新
当access token即将过期时，可以使用refresh token获取新的access token。

### 4. 权限验证
每次API调用时，系统会验证JWT token的有效性和用户权限。

## API端点

### 用户注册

**端点**: `POST /api/auth/register`

**描述**: 创建新用户账户

**请求头**:
```http
Content-Type: application/json
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "张三",
  "avatar": "https://example.com/avatar.jpg"
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 用户邮箱，必须唯一 |
| password | string | 是 | 密码，至少8位，包含字母数字 |
| name | string | 是 | 用户姓名 |
| avatar | string | 否 | 头像URL |

**成功响应** (201):
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "张三",
      "avatar": "https://example.com/avatar.jpg",
      "is_admin": false,
      "is_active": true,
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 86400
    }
  },
  "message": "注册成功"
}
```

**错误响应**:
```json
{
  "error": "邮箱已被注册",
  "code": "BUSINESS_RESOURCE_CONFLICT",
  "details": "邮箱 user@example.com 已被使用"
}
```

### 用户登录

**端点**: `POST /api/auth/login`

**描述**: 用户登录认证

**请求头**:
```http
Content-Type: application/json
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "remember_me": true
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 用户邮箱 |
| password | string | 是 | 用户密码 |
| remember_me | boolean | 否 | 是否记住登录状态 |

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "张三",
      "avatar": "https://example.com/avatar.jpg",
      "is_admin": false,
      "is_active": true,
      "last_login_at": "2024-12-19T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 86400
    }
  },
  "message": "登录成功"
}
```

**错误响应**:
```json
{
  "error": "邮箱或密码错误",
  "code": "AUTH_INVALID_CREDENTIALS",
  "details": "请检查邮箱和密码是否正确"
}
```

### Token刷新

**端点**: `POST /api/auth/refresh`

**描述**: 使用refresh token获取新的access token

**请求头**:
```http
Content-Type: application/json
```

**请求体**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  },
  "message": "Token刷新成功"
}
```

**错误响应**:
```json
{
  "error": "Refresh token无效",
  "code": "AUTH_TOKEN_INVALID",
  "details": "请重新登录"
}
```

### 用户登出

**端点**: `POST /api/auth/logout`

**描述**: 用户登出，使token失效

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**成功响应** (200):
```json
{
  "success": true,
  "message": "登出成功"
}
```

## JWT Token结构

### Access Token Payload
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "张三",
  "is_admin": false,
  "is_system": false,
  "token_type": "access",
  "iat": 1640995200,
  "exp": 1641081600,
  "iss": "teable-api"
}
```

### Refresh Token Payload
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token_type": "refresh",
  "iat": 1640995200,
  "exp": 1641600000,
  "iss": "teable-api"
}
```

### Token字段说明
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | string | 用户唯一标识 |
| email | string | 用户邮箱 |
| name | string | 用户姓名 |
| is_admin | boolean | 是否为管理员 |
| is_system | boolean | 是否为系统用户 |
| token_type | string | Token类型 (access/refresh) |
| iat | number | 签发时间戳 |
| exp | number | 过期时间戳 |
| iss | string | 签发者 |

## 权限系统

### 用户角色
1. **普通用户** (`user`): 基本功能访问权限
2. **管理员** (`admin`): 系统管理权限
3. **系统用户** (`system`): 内部服务权限

### 权限级别
1. **公开** (`public`): 无需认证
2. **用户** (`user`): 需要有效token
3. **管理员** (`admin`): 需要管理员权限
4. **系统** (`system`): 仅系统内部使用

### 权限检查
系统通过中间件自动检查用户权限：

```go
// 用户权限中间件
router.Use(middleware.AuthMiddleware(authService))

// 管理员权限中间件
router.Use(middleware.AdminRequiredMiddleware())
```

## 安全特性

### 密码安全
- 密码必须至少8位字符
- 必须包含字母、数字和特殊字符
- 使用bcrypt进行密码哈希
- 密码强度检查

### Token安全
- 使用HS256算法签名
- Access token有效期24小时
- Refresh token有效期7天
- 支持token黑名单机制

### 登录安全
- 支持多设备登录管理
- 登录设备追踪
- 异常登录检测
- IP地址记录

### 会话管理
- 支持并发会话限制
- 会话超时自动清理
- 设备绑定功能
- 远程登出支持

## 错误码说明

### 认证相关错误
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| AUTH_TOKEN_INVALID | 401 | Token格式错误或无效 |
| AUTH_TOKEN_EXPIRED | 401 | Token已过期 |
| AUTH_TOKEN_MISSING | 401 | 缺少认证token |
| AUTH_INVALID_CREDENTIALS | 401 | 邮箱或密码错误 |
| AUTH_ACCOUNT_DISABLED | 401 | 账户已被禁用 |
| AUTH_ACCOUNT_LOCKED | 401 | 账户已被锁定 |
| AUTH_INSUFFICIENT_PERMISSIONS | 403 | 权限不足 |
| AUTH_REFRESH_TOKEN_INVALID | 401 | Refresh token无效 |

### 验证相关错误
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| VALIDATION_EMAIL_INVALID | 400 | 邮箱格式不正确 |
| VALIDATION_PASSWORD_WEAK | 400 | 密码强度不够 |
| VALIDATION_NAME_REQUIRED | 400 | 姓名为必填项 |
| VALIDATION_EMAIL_REQUIRED | 400 | 邮箱为必填项 |

## 使用示例

### JavaScript/TypeScript
```javascript
// 用户登录
const loginResponse = await fetch('/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePassword123!'
  })
});

const loginData = await loginResponse.json();
const accessToken = loginData.data.tokens.access_token;

// 使用token调用API
const apiResponse = await fetch('/api/users/profile', {
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json'
  }
});
```

### cURL
```bash
# 用户登录
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'

# 使用token调用API
curl -X GET http://localhost:3000/api/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Go
```go
// 用户登录
loginReq := map[string]string{
    "email":    "user@example.com",
    "password": "SecurePassword123!",
}

loginJSON, _ := json.Marshal(loginReq)
resp, _ := http.Post("http://localhost:3000/api/auth/login", 
    "application/json", bytes.NewBuffer(loginJSON))

// 使用token调用API
req, _ := http.NewRequest("GET", "http://localhost:3000/api/users/profile", nil)
req.Header.Set("Authorization", "Bearer "+accessToken)
client := &http.Client{}
resp, _ := client.Do(req)
```
