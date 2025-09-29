# 空间管理

## 概述

空间管理模块提供工作空间的创建、管理、成员邀请等功能。空间是Teable中的顶级组织单位，用于管理数据表、用户权限和协作。

## API端点

### 创建空间

**端点**: `POST /api/spaces`

**描述**: 创建一个新的工作空间

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "我的工作空间",
  "description": "这是一个用于项目管理的工作空间",
  "icon": "🏢"
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 空间名称，最多100字符 |
| description | string | 否 | 空间描述，最多500字符 |
| icon | string | 否 | 空间图标，支持emoji或URL |

**成功响应** (201):
```json
{
  "success": true,
  "data": {
    "id": "space_550e8400-e29b-41d4-a716-446655440000",
    "name": "我的工作空间",
    "description": "这是一个用于项目管理的工作空间",
    "icon": "🏢",
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "is_public": false,
    "member_count": 1,
    "table_count": 0,
    "permissions": {
      "can_edit": true,
      "can_delete": true,
      "can_invite": true,
      "can_manage_permissions": true
    },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  },
  "message": "空间创建成功"
}
```

### 获取空间列表

**端点**: `GET /api/spaces`

**描述**: 获取用户有权限访问的空间列表

**请求头**:
```http
Authorization: Bearer <access_token>
```

**查询参数**:
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| limit | integer | 20 | 每页记录数，最大100 |
| offset | integer | 0 | 偏移量 |
| sort | string | updated_at | 排序字段 |
| order | string | desc | 排序方向 |
| search | string | - | 搜索关键词 |
| type | string | all | 空间类型 (owned/joined/all) |

**成功响应** (200):
```json
{
  "data": [
    {
      "id": "space_550e8400-e29b-41d4-a716-446655440000",
      "name": "我的工作空间",
      "description": "这是一个用于项目管理的工作空间",
      "icon": "🏢",
      "owner_id": "550e8400-e29b-41d4-a716-446655440000",
      "is_public": false,
      "member_count": 5,
      "table_count": 12,
      "permissions": {
        "can_edit": true,
        "can_delete": true,
        "can_invite": true,
        "can_manage_permissions": true
      },
      "last_activity_at": "2024-12-19T10:30:00Z",
      "created_at": "2024-12-19T09:00:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### 获取空间详情

**端点**: `GET /api/spaces/:id`

**描述**: 获取指定空间的详细信息

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "id": "space_550e8400-e29b-41d4-a716-446655440000",
    "name": "我的工作空间",
    "description": "这是一个用于项目管理的工作空间",
    "icon": "🏢",
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "is_public": false,
    "settings": {
      "default_permissions": "read",
      "allow_guest_access": false,
      "auto_archive_inactive": true,
      "retention_days": 90
    },
    "statistics": {
      "member_count": 5,
      "table_count": 12,
      "record_count": 1250,
      "storage_used": 52428800
    },
    "permissions": {
      "can_edit": true,
      "can_delete": true,
      "can_invite": true,
      "can_manage_permissions": true,
      "can_export": true,
      "can_import": true
    },
    "members": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "owner@example.com",
        "name": "张三",
        "avatar": "https://example.com/avatar.jpg",
        "role": "owner",
        "joined_at": "2024-12-19T09:00:00Z"
      }
    ],
    "created_at": "2024-12-19T09:00:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### 更新空间

**端点**: `PUT /api/spaces/:id`

**描述**: 更新空间的基本信息

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |

**请求体**:
```json
{
  "name": "更新后的工作空间",
  "description": "更新后的描述信息",
  "icon": "🏠",
  "is_public": false
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 空间名称 |
| description | string | 否 | 空间描述 |
| icon | string | 否 | 空间图标 |
| is_public | boolean | 否 | 是否公开 |

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "id": "space_550e8400-e29b-41d4-a716-446655440000",
    "name": "更新后的工作空间",
    "description": "更新后的描述信息",
    "icon": "🏠",
    "updated_at": "2024-12-19T11:00:00Z"
  },
  "message": "空间更新成功"
}
```

### 删除空间

**端点**: `DELETE /api/spaces/:id`

**描述**: 删除指定的空间（仅空间所有者可操作）

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |

**成功响应** (200):
```json
{
  "success": true,
  "message": "空间删除成功"
}
```

### 邀请成员

**端点**: `POST /api/spaces/:id/members`

**描述**: 邀请新成员加入空间

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |

**请求体**:
```json
{
  "email": "newmember@example.com",
  "role": "member",
  "message": "欢迎加入我们的工作空间！"
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 被邀请用户的邮箱 |
| role | string | 是 | 角色 (member/admin) |
| message | string | 否 | 邀请消息 |

**成功响应** (201):
```json
{
  "success": true,
  "data": {
    "invitation_id": "inv_550e8400-e29b-41d4-a716-446655440000",
    "email": "newmember@example.com",
    "role": "member",
    "status": "pending",
    "expires_at": "2024-12-26T10:30:00Z",
    "created_at": "2024-12-19T10:30:00Z"
  },
  "message": "邀请发送成功"
}
```

### 获取成员列表

**端点**: `GET /api/spaces/:id/members`

**描述**: 获取空间成员列表

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |

**查询参数**:
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| limit | integer | 20 | 每页记录数 |
| offset | integer | 0 | 偏移量 |
| role | string | all | 角色筛选 |

**成功响应** (200):
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "owner@example.com",
      "name": "张三",
      "avatar": "https://example.com/avatar.jpg",
      "role": "owner",
      "status": "active",
      "permissions": {
        "can_edit": true,
        "can_delete": true,
        "can_invite": true,
        "can_manage_permissions": true
      },
      "joined_at": "2024-12-19T09:00:00Z",
      "last_active_at": "2024-12-19T10:30:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "email": "member@example.com",
      "name": "李四",
      "avatar": "https://example.com/avatar2.jpg",
      "role": "member",
      "status": "active",
      "permissions": {
        "can_edit": true,
        "can_delete": false,
        "can_invite": false,
        "can_manage_permissions": false
      },
      "joined_at": "2024-12-19T09:30:00Z",
      "last_active_at": "2024-12-19T10:15:00Z"
    }
  ],
  "total": 2,
  "limit": 20,
  "offset": 0
}
```

### 更新成员权限

**端点**: `PUT /api/spaces/:id/members/:member_id`

**描述**: 更新指定成员的权限和角色

**请求头**:
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |
| member_id | string | 成员ID |

**请求体**:
```json
{
  "role": "admin",
  "permissions": {
    "can_edit": true,
    "can_delete": true,
    "can_invite": true,
    "can_manage_permissions": false
  }
}
```

**成功响应** (200):
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "role": "admin",
    "permissions": {
      "can_edit": true,
      "can_delete": true,
      "can_invite": true,
      "can_manage_permissions": false
    },
    "updated_at": "2024-12-19T11:00:00Z"
  },
  "message": "成员权限更新成功"
}
```

### 移除成员

**端点**: `DELETE /api/spaces/:id/members/:member_id`

**描述**: 从空间中移除指定成员

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |
| member_id | string | 成员ID |

**成功响应** (200):
```json
{
  "success": true,
  "message": "成员移除成功"
}
```

### 离开空间

**端点**: `POST /api/spaces/:id/leave`

**描述**: 当前用户离开指定的空间

**请求头**:
```http
Authorization: Bearer <access_token>
```

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 空间ID |

**成功响应** (200):
```json
{
  "success": true,
  "message": "已成功离开空间"
}
```

## 数据模型

### 空间实体 (Space)
```json
{
  "id": "string",              // 空间唯一标识
  "name": "string",            // 空间名称
  "description": "string",     // 空间描述
  "icon": "string",            // 空间图标
  "owner_id": "string",        // 所有者用户ID
  "is_public": "boolean",      // 是否公开
  "settings": "object",        // 空间设置
  "statistics": "object",      // 统计信息
  "permissions": "object",     // 当前用户权限
  "created_at": "datetime",    // 创建时间
  "updated_at": "datetime"     // 更新时间
}
```

### 空间成员 (SpaceMember)
```json
{
  "id": "string",              // 用户ID
  "email": "string",           // 用户邮箱
  "name": "string",            // 用户姓名
  "avatar": "string",          // 用户头像
  "role": "string",            // 角色 (owner/admin/member)
  "status": "string",          // 状态 (active/inactive/pending)
  "permissions": "object",     // 权限设置
  "joined_at": "datetime",     // 加入时间
  "last_active_at": "datetime" // 最后活跃时间
}
```

### 空间邀请 (SpaceInvitation)
```json
{
  "id": "string",              // 邀请ID
  "space_id": "string",        // 空间ID
  "email": "string",           // 被邀请邮箱
  "role": "string",            // 邀请角色
  "message": "string",         // 邀请消息
  "status": "string",          // 状态 (pending/accepted/declined/expired)
  "invited_by": "string",      // 邀请者ID
  "expires_at": "datetime",    // 过期时间
  "created_at": "datetime"     // 创建时间
}
```

## 角色和权限

### 角色定义
1. **所有者 (owner)**: 空间创建者，拥有所有权限
2. **管理员 (admin)**: 可以管理成员和权限，但不能删除空间
3. **成员 (member)**: 基本的编辑权限

### 权限矩阵
| 权限 | 所有者 | 管理员 | 成员 |
|------|--------|--------|------|
| 查看空间 | ✅ | ✅ | ✅ |
| 编辑空间信息 | ✅ | ✅ | ❌ |
| 删除空间 | ✅ | ❌ | ❌ |
| 邀请成员 | ✅ | ✅ | ❌ |
| 管理成员权限 | ✅ | ✅ | ❌ |
| 移除成员 | ✅ | ✅ | ❌ |
| 创建数据表 | ✅ | ✅ | ✅ |
| 编辑数据表 | ✅ | ✅ | ✅ |
| 删除数据表 | ✅ | ✅ | ❌ |
| 导出数据 | ✅ | ✅ | ✅ |

## 错误处理

### 常见错误码
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| SPACE_NOT_FOUND | 404 | 空间不存在 |
| SPACE_ACCESS_DENIED | 403 | 无权访问空间 |
| SPACE_NAME_DUPLICATE | 409 | 空间名称已存在 |
| MEMBER_NOT_FOUND | 404 | 成员不存在 |
| MEMBER_ALREADY_EXISTS | 409 | 成员已存在 |
| INVITATION_EXPIRED | 400 | 邀请已过期 |
| CANNOT_REMOVE_OWNER | 400 | 无法移除所有者 |
| INSUFFICIENT_PERMISSIONS | 403 | 权限不足 |

## 使用示例

### JavaScript/TypeScript
```javascript
// 创建空间
const createSpace = async (spaceData) => {
  const response = await fetch('/api/spaces', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(spaceData)
  });
  return response.json();
};

// 邀请成员
const inviteMember = async (spaceId, memberData) => {
  const response = await fetch(`/api/spaces/${spaceId}/members`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(memberData)
  });
  return response.json();
};

// 获取空间列表
const getSpaces = async (params = {}) => {
  const queryString = new URLSearchParams(params).toString();
  const response = await fetch(`/api/spaces?${queryString}`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  return response.json();
};
```

### cURL
```bash
# 创建空间
curl -X POST http://localhost:3000/api/spaces \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的工作空间",
    "description": "用于项目管理",
    "icon": "🏢"
  }'

# 邀请成员
curl -X POST http://localhost:3000/api/spaces/SPACE_ID/members \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newmember@example.com",
    "role": "member",
    "message": "欢迎加入！"
  }'

# 获取空间列表
curl -X GET "http://localhost:3000/api/spaces?limit=10&offset=0" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 最佳实践

### 1. 空间组织
- 根据项目或团队创建独立空间
- 使用清晰的命名和描述
- 合理设置空间权限

### 2. 成员管理
- 定期审查成员权限
- 及时移除不再需要的成员
- 为新成员提供适当的权限

### 3. 安全考虑
- 避免将敏感数据放在公开空间
- 定期备份重要数据
- 监控空间访问活动

### 4. 性能优化
- 合理控制空间大小
- 定期清理无用数据
- 使用搜索和筛选功能
