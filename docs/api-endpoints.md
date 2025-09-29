# API端点汇总

## 概述

本文档汇总了Teable Go Backend的所有API端点，按照功能模块分类整理。

## 基础信息

- **基础URL**: `http://localhost:3000`
- **API版本**: v1
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON

## 端点分类

### 1. 系统健康检查 (无需认证)

| 方法 | 端点 | 描述 | 状态码 |
|------|------|------|--------|
| GET | `/health` | 完整健康检查 | 200/503 |
| GET | `/ready` | 就绪检查 | 200/503 |
| GET | `/alive` | 存活检查 | 200 |
| GET | `/ping` | 简单ping检查 | 200 |

### 2. 认证相关 (部分无需认证)

| 方法 | 端点 | 描述 | 认证要求 |
|------|------|------|----------|
| POST | `/api/auth/register` | 用户注册 | ❌ |
| POST | `/api/auth/login` | 用户登录 | ❌ |
| POST | `/api/auth/refresh` | 刷新Token | ❌ |
| POST | `/api/auth/logout` | 用户登出 | ✅ |

### 3. 用户管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/users/profile` | 获取用户资料 | 用户 |
| PUT | `/api/users/profile` | 更新用户资料 | 用户 |
| POST | `/api/users/change-password` | 修改密码 | 用户 |
| GET | `/api/users/:id/activity` | 获取用户活动 | 用户 |
| GET | `/api/users/preferences` | 获取用户偏好 | 用户 |
| PUT | `/api/users/preferences` | 更新用户偏好 | 用户 |

### 4. 空间管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/spaces` | 创建空间 | 用户 |
| GET | `/api/spaces` | 获取空间列表 | 用户 |
| GET | `/api/spaces/:id` | 获取空间详情 | 空间成员 |
| PUT | `/api/spaces/:id` | 更新空间 | 空间管理员 |
| DELETE | `/api/spaces/:id` | 删除空间 | 空间所有者 |
| POST | `/api/spaces/:id/members` | 邀请成员 | 空间管理员 |
| GET | `/api/spaces/:id/members` | 获取成员列表 | 空间成员 |
| PUT | `/api/spaces/:id/members/:member_id` | 更新成员权限 | 空间管理员 |
| DELETE | `/api/spaces/:id/members/:member_id` | 移除成员 | 空间管理员 |
| POST | `/api/spaces/:id/leave` | 离开空间 | 空间成员 |

### 5. 基础表管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/bases` | 创建基础表 | 空间成员 |
| GET | `/api/bases` | 获取基础表列表 | 用户 |
| GET | `/api/bases/:id` | 获取基础表详情 | 基础表成员 |
| PUT | `/api/bases/:id` | 更新基础表 | 基础表管理员 |
| DELETE | `/api/bases/:id` | 删除基础表 | 基础表所有者 |
| GET | `/api/bases/:id/permissions` | 检查用户权限 | 基础表成员 |
| GET | `/api/bases/:id/stats` | 获取基础表统计 | 基础表成员 |
| GET | `/api/bases/space/:space_id/stats` | 获取空间基础表统计 | 空间成员 |
| POST | `/api/bases/bulk-update` | 批量更新基础表 | 空间管理员 |
| POST | `/api/bases/bulk-delete` | 批量删除基础表 | 空间管理员 |
| GET | `/api/bases/export` | 导出基础表 | 空间成员 |
| POST | `/api/bases/import` | 导入基础表 | 空间管理员 |
| POST | `/api/bases/:id/tables` | 创建数据表 | 基础表成员 |
| GET | `/api/bases/:id/tables` | 获取基础表下的数据表 | 基础表成员 |

### 6. 数据表管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/tables` | 创建数据表 | 基础表成员 |
| GET | `/api/tables` | 获取数据表列表 | 用户 |
| GET | `/api/tables/:id` | 获取数据表详情 | 数据表成员 |
| PUT | `/api/tables/:id` | 更新数据表 | 数据表管理员 |
| DELETE | `/api/tables/:id` | 删除数据表 | 数据表所有者 |

### 7. 字段管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/fields` | 创建字段 | 数据表成员 |
| GET | `/api/fields` | 获取字段列表 | 数据表成员 |
| GET | `/api/fields/:id` | 获取字段详情 | 数据表成员 |
| PUT | `/api/fields/:id` | 更新字段 | 数据表管理员 |
| DELETE | `/api/fields/:id` | 删除字段 | 数据表管理员 |
| GET | `/api/fields/types` | 获取字段类型 | 用户 |
| GET | `/api/fields/types/:type` | 获取字段类型信息 | 用户 |
| POST | `/api/fields/:field_id/validate` | 验证字段值 | 数据表成员 |

### 8. 记录管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/records` | 创建记录 | 数据表成员 |
| GET | `/api/records` | 获取记录列表 | 数据表成员 |
| GET | `/api/records/:id` | 获取记录详情 | 数据表成员 |
| PUT | `/api/records/:id` | 更新记录 | 数据表成员 |
| DELETE | `/api/records/:id` | 删除记录 | 数据表成员 |
| POST | `/api/records/bulk-create` | 批量创建记录 | 数据表成员 |
| PUT | `/api/records/bulk-update` | 批量更新记录 | 数据表成员 |
| DELETE | `/api/records/bulk-delete` | 批量删除记录 | 数据表成员 |
| GET | `/api/records/export` | 导出记录 | 数据表成员 |
| POST | `/api/records/import` | 导入记录 | 数据表管理员 |

### 9. 视图管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/views` | 创建视图 | 数据表成员 |
| GET | `/api/views` | 获取视图列表 | 数据表成员 |
| GET | `/api/views/:id` | 获取视图详情 | 数据表成员 |
| PUT | `/api/views/:id` | 更新视图 | 数据表成员 |
| DELETE | `/api/views/:id` | 删除视图 | 数据表管理员 |
| POST | `/api/views/:id/duplicate` | 复制视图 | 数据表成员 |
| POST | `/api/views/:id/export` | 导出视图 | 数据表成员 |

### 10. 权限管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/permissions` | 获取权限列表 | 用户 |
| GET | `/api/permissions/:id` | 获取权限详情 | 权限所有者 |
| PUT | `/api/permissions/:id` | 更新权限 | 权限管理员 |
| DELETE | `/api/permissions/:id` | 删除权限 | 权限管理员 |
| POST | `/api/permissions/bulk-update` | 批量更新权限 | 空间管理员 |

### 11. 分享管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/shares` | 创建分享 | 数据表成员 |
| GET | `/api/shares` | 获取分享列表 | 用户 |
| GET | `/api/shares/:id` | 获取分享详情 | 分享所有者 |
| PUT | `/api/shares/:id` | 更新分享 | 分享所有者 |
| DELETE | `/api/shares/:id` | 删除分享 | 分享所有者 |
| GET | `/api/shares/public/:token` | 公开访问分享 | ❌ |

### 12. 附件管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/attachments` | 上传附件 | 用户 |
| GET | `/api/attachments` | 获取附件列表 | 用户 |
| GET | `/api/attachments/:id` | 获取附件详情 | 附件所有者 |
| GET | `/api/attachments/:id/download` | 下载附件 | 附件访问者 |
| PUT | `/api/attachments/:id` | 更新附件信息 | 附件所有者 |
| DELETE | `/api/attachments/:id` | 删除附件 | 附件所有者 |

### 13. 通知系统 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/notifications` | 获取通知列表 | 用户 |
| GET | `/api/notifications/:id` | 获取通知详情 | 用户 |
| PUT | `/api/notifications/:id` | 标记通知已读 | 用户 |
| DELETE | `/api/notifications/:id` | 删除通知 | 用户 |
| POST | `/api/notifications/mark-all-read` | 标记所有通知已读 | 用户 |
| POST | `/api/notifications/subscriptions` | 创建订阅 | 用户 |
| GET | `/api/notifications/subscriptions/:id` | 获取订阅详情 | 用户 |
| PUT | `/api/notifications/subscriptions/:id` | 更新订阅 | 用户 |
| DELETE | `/api/notifications/subscriptions/:id` | 删除订阅 | 用户 |
| GET | `/api/notifications/subscriptions/user/:user_id` | 获取用户订阅 | 用户 |
| DELETE | `/api/notifications/subscriptions/user/:user_id` | 删除用户所有订阅 | 用户 |

### 14. 搜索功能 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/search` | 基础搜索 | 用户 |
| POST | `/api/search/advanced` | 高级搜索 | 用户 |
| GET | `/api/search/suggestions` | 搜索建议 | 用户 |
| GET | `/api/search/popular` | 热门搜索 | 用户 |
| GET | `/api/search/stats` | 搜索统计 | 用户 |
| POST | `/api/search/indexes` | 创建搜索索引 | 管理员 |
| GET | `/api/search/indexes` | 获取索引列表 | 管理员 |
| GET | `/api/search/indexes/:id` | 获取索引详情 | 管理员 |
| PUT | `/api/search/indexes/:id` | 更新索引 | 管理员 |
| DELETE | `/api/search/indexes/:id` | 删除索引 | 管理员 |
| DELETE | `/api/search/indexes/by-source` | 按来源删除索引 | 管理员 |
| POST | `/api/search/indexes/rebuild` | 重建索引 | 管理员 |
| POST | `/api/search/indexes/optimize` | 优化索引 | 管理员 |
| GET | `/api/search/indexes/stats` | 获取索引统计 | 管理员 |

### 15. WebSocket实时通信 (需要认证)

| 方法 | 端点 | 描述 | 认证要求 |
|------|------|------|----------|
| GET | `/api/ws/socket` | WebSocket连接 | 连接时验证 |
| GET | `/api/ws/stats` | 获取WebSocket统计 | 管理员 |
| POST | `/api/ws/broadcast/channel` | 频道广播 | 管理员 |
| POST | `/api/ws/broadcast/user` | 用户广播 | 管理员 |
| POST | `/api/ws/publish/document` | 发布文档操作 | 用户 |
| POST | `/api/ws/publish/record` | 发布记录操作 | 用户 |
| POST | `/api/ws/publish/view` | 发布视图操作 | 用户 |
| POST | `/api/ws/publish/field` | 发布字段操作 | 用户 |

### 16. ShareDB协作 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/sharedb/connect` | 建立协作连接 | 用户 |
| POST | `/api/sharedb/disconnect` | 断开协作连接 | 用户 |
| GET | `/api/sharedb/sessions` | 获取协作会话 | 用户 |
| POST | `/api/sharedb/sessions/:id/join` | 加入协作会话 | 用户 |
| POST | `/api/sharedb/sessions/:id/leave` | 离开协作会话 | 用户 |
| GET | `/api/sharedb/sessions/:id/participants` | 获取参与者列表 | 用户 |
| POST | `/api/sharedb/operations` | 提交操作 | 用户 |
| GET | `/api/sharedb/operations` | 获取操作历史 | 用户 |

### 17. 协作管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/collaboration/sessions` | 创建协作会话 | 用户 |
| GET | `/api/collaboration/sessions` | 获取协作会话列表 | 用户 |
| GET | `/api/collaboration/sessions/:id` | 获取协作会话详情 | 用户 |
| PUT | `/api/collaboration/sessions/:id` | 更新协作会话 | 会话管理员 |
| DELETE | `/api/collaboration/sessions/:id` | 结束协作会话 | 会话管理员 |
| POST | `/api/collaboration/sessions/:id/join` | 加入协作会话 | 用户 |
| POST | `/api/collaboration/sessions/:id/leave` | 离开协作会话 | 用户 |
| GET | `/api/collaboration/sessions/:id/participants` | 获取参与者列表 | 用户 |
| POST | `/api/collaboration/sessions/:id/invite` | 邀请参与协作 | 会话管理员 |
| POST | `/api/collaboration/sessions/:id/kick` | 移除参与者 | 会话管理员 |

### 18. Pin管理 (需要认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/pin/list` | 获取Pin列表 | 用户 |

### 19. 管理员功能 (需要管理员权限)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/admin/users` | 获取用户列表 | 管理员 |
| GET | `/api/admin/users/:id` | 获取用户详情 | 管理员 |
| PUT | `/api/admin/users/:id` | 更新用户 | 管理员 |
| DELETE | `/api/admin/users/:id` | 删除用户 | 管理员 |
| POST | `/api/admin/users/:id/promote` | 提升为管理员 | 管理员 |
| POST | `/api/admin/users/:id/demote` | 降级为普通用户 | 管理员 |
| POST | `/api/admin/users/:id/activate` | 激活用户 | 管理员 |
| POST | `/api/admin/users/:id/deactivate` | 停用用户 | 管理员 |
| POST | `/api/admin/users/bulk-update` | 批量更新用户 | 管理员 |
| POST | `/api/admin/users/bulk-delete` | 批量删除用户 | 管理员 |
| GET | `/api/admin/users/export` | 导出用户 | 管理员 |
| POST | `/api/admin/users/import` | 导入用户 | 管理员 |
| GET | `/api/admin/users/stats` | 获取用户统计 | 管理员 |

### 20. 系统管理 (需要管理员权限)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/system/metrics` | 获取系统指标 | 管理员 |

### 21. 系统信息 (无需认证)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/info` | 获取系统信息 | ❌ |

## 权限说明

### 权限级别
1. **❌ 无需认证**: 公开访问
2. **✅ 用户权限**: 需要有效的JWT token
3. **👥 空间成员**: 需要是空间的成员
4. **🔧 空间管理员**: 需要空间的管理权限
5. **👑 空间所有者**: 需要是空间的所有者
6. **🛡️ 管理员**: 需要系统管理员权限

### 角色定义
- **用户**: 已注册的普通用户
- **空间成员**: 空间的参与者
- **空间管理员**: 可以管理空间成员和权限
- **空间所有者**: 空间的创建者，拥有最高权限
- **系统管理员**: 可以管理整个系统

## 状态码说明

### 成功状态码
- **200 OK**: 请求成功
- **201 Created**: 资源创建成功
- **204 No Content**: 请求成功，无返回内容

### 客户端错误
- **400 Bad Request**: 请求参数错误
- **401 Unauthorized**: 未认证或认证失败
- **403 Forbidden**: 无权限访问
- **404 Not Found**: 资源不存在
- **409 Conflict**: 资源冲突
- **422 Unprocessable Entity**: 数据验证失败

### 服务器错误
- **500 Internal Server Error**: 服务器内部错误
- **503 Service Unavailable**: 服务不可用

## 分页参数

大多数列表接口支持分页参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| limit | integer | 20 | 每页记录数，最大100 |
| offset | integer | 0 | 偏移量，从0开始 |

## 排序参数

支持的排序参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| sort | string | created_at | 排序字段 |
| order | string | desc | 排序方向 (asc/desc) |

## 搜索参数

支持的搜索参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| search | string | 搜索关键词 |
| filter | string | 筛选条件 |
| type | string | 类型筛选 |

## 响应格式

### 成功响应
```json
{
  "success": true,
  "data": { ... },
  "message": "操作成功"
}
```

### 分页响应
```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

### 错误响应
```json
{
  "error": "错误描述",
  "code": "ERROR_CODE",
  "details": "详细错误信息",
  "trace_id": "请求追踪ID"
}
```
