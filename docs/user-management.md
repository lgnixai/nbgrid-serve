# 用户管理

## 概述

用户管理模块提供用户资料管理、偏好设置、活动记录等功能。所有用户相关的API都需要有效的JWT token认证。

## API端点

### 获取用户资料

**端点**: `GET /api/users/profile`

**描述**: 获取当前登录用户的详细资料

**请求头**:
```http
Authorization: Bearer <access_token>
```

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "张三",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "这是我的个人简介",
    "phone": "+86-138-0013-8000",
    "location": "北京市",
    "website": "https://example.com",
    "is_admin": false,
    "is_active": true,
    "email_verified": true,
    "phone_verified": false,
    "preferences": {
      "language": "zh-CN",
      "timezone": "Asia/Shanghai",
      "theme": "light",
      "notifications": {
        "email": true,
        "push": true,
        "desktop": false
      }
    },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z",
    "last_login_at": "2024-12-19T10:30:00Z"
  }
}
```

### 更新用户资料

**端点**: `PUT /api/users/profile`

**描述**: 更新当前登录用户的资料信息

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "张三",
  "avatar": "https://example.com/new-avatar.jpg",
  "bio": "更新后的个人简介",
  "phone": "+86-138-0013-8000",
  "location": "上海市",
  "website": "https://mywebsite.com"
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 用户姓名 |
| avatar | string | 否 | 头像URL |
| bio | string | 否 | 个人简介 |
| phone | string | 否 | 手机号码 |
| location | string | 否 | 所在位置 |
| website | string | 否 | 个人网站 |

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "张三",
    "avatar": "https://example.com/new-avatar.jpg",
    "bio": "更新后的个人简介",
    "phone": "+86-138-0013-8000",
    "location": "上海市",
    "website": "https://mywebsite.com",
    "updated_at": "2024-12-19T11:00:00Z"
  },
  "message": "资料更新成功"
}
```

### 修改密码

**端点**: `POST /api/users/change-password`

**描述**: 修改当前用户的登录密码

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**:
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword456!",
  "confirm_password": "NewPassword456!"
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| current_password | string | 是 | 当前密码 |
| new_password | string | 是 | 新密码 |
| confirm_password | string | 是 | 确认新密码 |

**成功响应** (200):
```json
{
  "success": true,
  "message": "密码修改成功"
}
```

**错误响应**:
```json
{
  "error": "当前密码错误",
  "code": "AUTH_INVALID_CREDENTIALS",
  "details": "请检查当前密码是否正确"
}
```

### 获取用户活动记录

**端点**: `GET /api/users/:id/activity`

**描述**: 获取指定用户的活动记录

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户ID |

**查询参数**:
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| limit | integer | 20 | 每页记录数 |
| offset | integer | 0 | 偏移量 |
| type | string | all | 活动类型筛选 |

**成功响应** (200):
```json
{
  "data": [
    {
      "id": "activity_001",
      "type": "login",
      "description": "用户登录",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
      "location": "北京市",
      "device": "Desktop",
      "created_at": "2024-12-19T10:30:00Z"
    },
    {
      "id": "activity_002",
      "type": "table_created",
      "description": "创建了数据表 '项目列表'",
      "metadata": {
        "table_id": "table_123",
        "table_name": "项目列表"
      },
      "created_at": "2024-12-19T09:15:00Z"
    }
  ],
  "total": 50,
  "limit": 20,
  "offset": 0
}
```

### 获取用户偏好设置

**端点**: `GET /api/users/preferences`

**描述**: 获取当前用户的偏好设置

**请求头**:
```http
Authorization: Bearer <access_token>
```

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "language": "zh-CN",
    "timezone": "Asia/Shanghai",
    "theme": "light",
    "date_format": "YYYY-MM-DD",
    "time_format": "24h",
    "currency": "CNY",
    "notifications": {
      "email": {
        "enabled": true,
        "types": ["system", "collaboration", "reminder"]
      },
      "push": {
        "enabled": true,
        "types": ["mention", "assignment", "due_date"]
      },
      "desktop": {
        "enabled": false,
        "types": []
      }
    },
    "privacy": {
      "profile_visibility": "team",
      "activity_visibility": "private",
      "search_visibility": "public"
    },
    "workspace": {
      "default_view": "table",
      "auto_save": true,
      "show_grid_lines": true,
      "compact_mode": false
    }
  }
}
```

### 更新用户偏好设置

**端点**: `PUT /api/users/preferences`

**描述**: 更新当前用户的偏好设置

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**:
```json
{
  "language": "en-US",
  "timezone": "America/New_York",
  "theme": "dark",
  "date_format": "MM/DD/YYYY",
  "time_format": "12h",
  "currency": "USD",
  "notifications": {
    "email": {
      "enabled": true,
      "types": ["system", "collaboration"]
    },
    "push": {
      "enabled": false,
      "types": []
    }
  },
  "privacy": {
    "profile_visibility": "public",
    "activity_visibility": "team"
  },
  "workspace": {
    "default_view": "kanban",
    "auto_save": false,
    "show_grid_lines": false,
    "compact_mode": true
  }
}
```

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "language": "en-US",
    "timezone": "America/New_York",
    "theme": "dark",
    "updated_at": "2024-12-19T11:30:00Z"
  },
  "message": "偏好设置更新成功"
}
```

## 数据模型

### 用户实体 (User)
```json
{
  "id": "string",              // 用户唯一标识
  "email": "string",           // 邮箱地址
  "name": "string",            // 用户姓名
  "avatar": "string",          // 头像URL
  "bio": "string",             // 个人简介
  "phone": "string",           // 手机号码
  "location": "string",        // 所在位置
  "website": "string",         // 个人网站
  "is_admin": "boolean",       // 是否为管理员
  "is_active": "boolean",      // 账户是否激活
  "email_verified": "boolean", // 邮箱是否验证
  "phone_verified": "boolean", // 手机是否验证
  "preferences": "object",     // 用户偏好设置
  "created_at": "datetime",    // 创建时间
  "updated_at": "datetime",    // 更新时间
  "last_login_at": "datetime"  // 最后登录时间
}
```

### 用户偏好设置 (UserPreferences)
```json
{
  "language": "string",        // 语言设置
  "timezone": "string",        // 时区设置
  "theme": "string",           // 主题设置 (light/dark/auto)
  "date_format": "string",     // 日期格式
  "time_format": "string",     // 时间格式 (12h/24h)
  "currency": "string",        // 货币设置
  "notifications": "object",   // 通知设置
  "privacy": "object",         // 隐私设置
  "workspace": "object"        // 工作区设置
}
```

### 用户活动记录 (UserActivity)
```json
{
  "id": "string",              // 活动记录ID
  "user_id": "string",         // 用户ID
  "type": "string",            // 活动类型
  "description": "string",     // 活动描述
  "ip_address": "string",      // IP地址
  "user_agent": "string",      // 用户代理
  "location": "string",        // 地理位置
  "device": "string",          // 设备类型
  "metadata": "object",        // 额外元数据
  "created_at": "datetime"     // 创建时间
}
```

## 活动类型

### 认证相关
- `login`: 用户登录
- `logout`: 用户登出
- `password_change`: 密码修改
- `profile_update`: 资料更新

### 工作区相关
- `space_created`: 创建空间
- `space_joined`: 加入空间
- `space_left`: 离开空间
- `space_deleted`: 删除空间

### 数据表相关
- `table_created`: 创建数据表
- `table_updated`: 更新数据表
- `table_deleted`: 删除数据表
- `table_shared`: 分享数据表

### 记录相关
- `record_created`: 创建记录
- `record_updated`: 更新记录
- `record_deleted`: 删除记录
- `record_exported`: 导出记录

### 协作相关
- `collaboration_started`: 开始协作
- `collaboration_ended`: 结束协作
- `comment_added`: 添加评论
- `mention_received`: 收到提及

## 错误处理

### 常见错误码
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| USER_NOT_FOUND | 404 | 用户不存在 |
| PROFILE_UPDATE_FAILED | 400 | 资料更新失败 |
| PASSWORD_CHANGE_FAILED | 400 | 密码修改失败 |
| INVALID_PASSWORD_FORMAT | 400 | 密码格式不正确 |
| PREFERENCES_UPDATE_FAILED | 400 | 偏好设置更新失败 |
| ACTIVITY_ACCESS_DENIED | 403 | 无权访问活动记录 |

### 错误响应示例
```json
{
  "error": "用户不存在",
  "code": "USER_NOT_FOUND",
  "details": "用户ID '123' 不存在",
  "trace_id": "req_1234567890abcdef"
}
```

## 使用示例

### JavaScript/TypeScript
```javascript
// 获取用户资料
const getUserProfile = async () => {
  const response = await fetch('/api/users/profile', {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  return response.json();
};

// 更新用户资料
const updateProfile = async (profileData) => {
  const response = await fetch('/api/users/profile', {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(profileData)
  });
  return response.json();
};

// 修改密码
const changePassword = async (passwordData) => {
  const response = await fetch('/api/users/change-password', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(passwordData)
  });
  return response.json();
};
```

### cURL
```bash
# 获取用户资料
curl -X GET http://localhost:3000/api/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# 更新用户资料
curl -X PUT http://localhost:3000/api/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新姓名",
    "bio": "新的个人简介",
    "location": "新城市"
  }'

# 修改密码
curl -X POST http://localhost:3000/api/users/change-password \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "旧密码",
    "new_password": "新密码",
    "confirm_password": "新密码"
  }'
```

## 最佳实践

### 1. 密码安全
- 定期更换密码
- 使用强密码策略
- 避免在多个平台使用相同密码

### 2. 资料管理
- 保持头像和简介的更新
- 设置合适的隐私级别
- 定期检查账户活动

### 3. 偏好设置
- 根据使用习惯调整界面设置
- 合理配置通知偏好
- 选择合适的工作区布局

### 4. 安全建议
- 启用双因素认证（如果支持）
- 定期检查登录设备
- 及时更新联系方式
