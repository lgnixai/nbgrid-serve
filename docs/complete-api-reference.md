# 完整API参考文档

## 概述

本文档提供了Teable Go Backend的完整API端点列表，包括之前遗漏的删除操作和复杂的视图管理功能。

## 完整的API端点列表

### 1. 系统健康检查 (无需认证)

| 方法 | 端点 | 描述 | 状态码 |
|------|------|------|--------|
| GET | `/health` | 完整健康检查 | 200/503 |
| GET | `/ready` | 就绪检查 | 200/503 |
| GET | `/alive` | 存活检查 | 200 |
| GET | `/ping` | 简单ping检查 | 200 |

### 2. 认证相关

| 方法 | 端点 | 描述 | 认证要求 |
|------|------|------|----------|
| POST | `/api/auth/register` | 用户注册 | ❌ |
| POST | `/api/auth/login` | 用户登录 | ❌ |
| POST | `/api/auth/refresh` | 刷新Token | ❌ |
| POST | `/api/auth/logout` | 用户登出 | ✅ |

### 3. 用户管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/users/profile` | 获取用户资料 | 用户 |
| PUT | `/api/users/profile` | 更新用户资料 | 用户 |
| POST | `/api/users/change-password` | 修改密码 | 用户 |
| GET | `/api/users/:id/activity` | 获取用户活动 | 用户 |
| GET | `/api/users/preferences` | 获取用户偏好 | 用户 |
| PUT | `/api/users/preferences` | 更新用户偏好 | 用户 |

### 4. 空间管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/spaces` | 创建空间 | 用户 |
| GET | `/api/spaces` | 获取空间列表 | 用户 |
| GET | `/api/spaces/:id` | 获取空间详情 | 空间成员 |
| PUT | `/api/spaces/:id` | 更新空间 | 空间管理员 |
| **DELETE** | `/api/spaces/:id` | **删除空间** | 空间所有者 |
| POST | `/api/spaces/:id/collaborators` | 添加协作者 | 空间管理员 |
| GET | `/api/spaces/:id/collaborators` | 获取协作者列表 | 空间成员 |
| **DELETE** | `/api/spaces/:id/collaborators/:collab_id` | **移除协作者** | 空间管理员 |
| PUT | `/api/spaces/:id/collaborators/:collab_id/role` | 更新协作者角色 | 空间管理员 |

### 5. 基础表管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/bases` | 创建基础表 | 空间成员 |
| GET | `/api/bases` | 获取基础表列表 | 用户 |
| GET | `/api/bases/:id` | 获取基础表详情 | 基础表成员 |
| PUT | `/api/bases/:id` | 更新基础表 | 基础表管理员 |
| **DELETE** | `/api/bases/:id` | **删除基础表** | 基础表所有者 |
| GET | `/api/bases/:id/permissions` | 检查用户权限 | 基础表成员 |
| GET | `/api/bases/:id/stats` | 获取基础表统计 | 基础表成员 |
| GET | `/api/bases/space/:space_id/stats` | 获取空间基础表统计 | 空间成员 |
| POST | `/api/bases/bulk-update` | 批量更新基础表 | 空间管理员 |
| POST | `/api/bases/bulk-delete` | **批量删除基础表** | 空间管理员 |
| GET | `/api/bases/export` | 导出基础表 | 空间成员 |
| POST | `/api/bases/import` | 导入基础表 | 空间管理员 |
| POST | `/api/bases/:id/tables` | 创建数据表 | 基础表成员 |
| GET | `/api/bases/:id/tables` | 获取基础表下的数据表 | 基础表成员 |

### 6. 数据表管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/tables` | 创建数据表 | 基础表成员 |
| GET | `/api/tables` | 获取数据表列表 | 用户 |
| GET | `/api/tables/:id` | 获取数据表详情 | 数据表成员 |
| PUT | `/api/tables/:id` | 更新数据表 | 数据表管理员 |
| **DELETE** | `/api/tables/:id` | **删除数据表** | 数据表所有者 |

### 7. 字段管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/fields` | 创建字段 | 数据表成员 |
| GET | `/api/fields` | 获取字段列表 | 数据表成员 |
| GET | `/api/fields/:id` | 获取字段详情 | 数据表成员 |
| PUT | `/api/fields/:id` | 更新字段 | 数据表管理员 |
| **DELETE** | `/api/fields/:id` | **删除字段** | 数据表管理员 |
| GET | `/api/fields/types` | 获取字段类型 | 用户 |
| GET | `/api/fields/types/:type` | 获取字段类型信息 | 用户 |
| POST | `/api/fields/:field_id/validate` | 验证字段值 | 数据表成员 |

### 8. 记录管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/records` | 创建记录 | 数据表成员 |
| GET | `/api/records` | 获取记录列表 | 数据表成员 |
| GET | `/api/records/:id` | 获取记录详情 | 数据表成员 |
| PUT | `/api/records/:id` | 更新记录 | 数据表成员 |
| **DELETE** | `/api/records/:id` | **删除记录** | 数据表成员 |
| POST | `/api/records/bulk` | 批量创建记录 | 数据表成员 |
| PUT | `/api/records/bulk` | 批量更新记录 | 数据表成员 |
| **DELETE** | `/api/records/bulk` | **批量删除记录** | 数据表成员 |
| POST | `/api/records/query` | 复杂查询 | 数据表成员 |
| GET | `/api/records/stats` | 获取记录统计 | 数据表成员 |
| POST | `/api/records/export` | 导出记录 | 数据表成员 |
| POST | `/api/records/import` | 导入记录 | 数据表管理员 |

### 9. 视图管理 (完整功能)

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/views` | 创建视图 | 数据表成员 |
| GET | `/api/views` | 获取视图列表 | 数据表成员 |
| GET | `/api/views/:id` | 获取视图详情 | 数据表成员 |
| PUT | `/api/views/:id` | 更新视图 | 数据表成员 |
| **DELETE** | `/api/views/:id` | **删除视图** | 数据表管理员 |

#### 视图配置
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/views/:id/config` | 获取视图配置 | 数据表成员 |
| PUT | `/api/views/:id/config` | 更新视图配置 | 数据表成员 |

#### 网格视图功能
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/views/:id/grid/data` | 获取网格视图数据 | 数据表成员 |
| PUT | `/api/views/:id/grid/config` | 更新网格视图配置 | 数据表成员 |
| POST | `/api/views/:id/grid/columns` | 添加网格列 | 数据表成员 |
| PUT | `/api/views/:id/grid/columns/:field_id` | 更新网格列 | 数据表成员 |
| **DELETE** | `/api/views/:id/grid/columns/:field_id` | **移除网格列** | 数据表成员 |
| PUT | `/api/views/:id/grid/columns/reorder` | 重新排序网格列 | 数据表成员 |

#### 表单视图功能
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/views/:id/form/data` | 获取表单视图数据 | 数据表成员 |
| PUT | `/api/views/:id/form/config` | 更新表单视图配置 | 数据表成员 |
| POST | `/api/views/:id/form/fields` | 添加表单字段 | 数据表成员 |
| PUT | `/api/views/:id/form/fields/:field_id` | 更新表单字段 | 数据表成员 |
| **DELETE** | `/api/views/:id/form/fields/:field_id` | **移除表单字段** | 数据表成员 |
| PUT | `/api/views/:id/form/fields/reorder` | 重新排序表单字段 | 数据表成员 |

#### 看板视图功能
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/views/:id/kanban/data` | 获取看板视图数据 | 数据表成员 |
| PUT | `/api/views/:id/kanban/config` | 更新看板视图配置 | 数据表成员 |
| POST | `/api/views/:id/kanban/move` | 移动看板卡片 | 数据表成员 |

#### 日历视图功能
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/views/:id/calendar/data` | 获取日历视图数据 | 数据表成员 |
| PUT | `/api/views/:id/calendar/config` | 更新日历视图配置 | 数据表成员 |

#### 画廊视图功能
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/views/:id/gallery/data` | 获取画廊视图数据 | 数据表成员 |
| PUT | `/api/views/:id/gallery/config` | 更新画廊视图配置 | 数据表成员 |

### 10. Pin管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/pin/list` | 获取Pin列表 | 用户 |

### 11. 管理员功能

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/admin/users` | 获取用户列表 | 管理员 |
| GET | `/api/admin/users/:id` | 获取用户详情 | 管理员 |
| PUT | `/api/admin/users/:id` | 更新用户 | 管理员 |
| **DELETE** | `/api/admin/users/:id` | **删除用户** | 管理员 |
| POST | `/api/admin/users/:id/promote` | 提升为管理员 | 管理员 |
| POST | `/api/admin/users/:id/demote` | 降级为普通用户 | 管理员 |
| POST | `/api/admin/users/:id/activate` | 激活用户 | 管理员 |
| POST | `/api/admin/users/:id/deactivate` | 停用用户 | 管理员 |
| POST | `/api/admin/users/bulk-update` | 批量更新用户 | 管理员 |
| POST | `/api/admin/users/bulk-delete` | **批量删除用户** | 管理员 |
| GET | `/api/admin/users/export` | 导出用户 | 管理员 |
| POST | `/api/admin/users/import` | 导入用户 | 管理员 |
| GET | `/api/admin/users/stats` | 获取用户统计 | 管理员 |

### 12. 协作管理

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

#### 在线状态管理
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/collaboration/presence` | 更新在线状态 | 用户 |
| **DELETE** | `/api/collaboration/presence` | **移除在线状态** | 用户 |
| GET | `/api/collaboration/presence` | 获取在线状态 | 用户 |

#### 光标位置管理
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/collaboration/cursor` | 更新光标位置 | 用户 |
| **DELETE** | `/api/collaboration/cursor` | **移除光标位置** | 用户 |
| GET | `/api/collaboration/cursor` | 获取光标位置 | 用户 |

### 13. WebSocket实时通信

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

### 14. ShareDB协作

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

### 15. 附件管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/attachments` | 上传附件 | 用户 |
| GET | `/api/attachments` | 获取附件列表 | 用户 |
| GET | `/api/attachments/:id` | 获取附件详情 | 附件所有者 |
| GET | `/api/attachments/:id/download` | 下载附件 | 附件访问者 |
| PUT | `/api/attachments/:id` | 更新附件信息 | 附件所有者 |
| **DELETE** | `/api/attachments/:id` | **删除附件** | 附件所有者 |

### 16. 通知系统

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/notifications` | 获取通知列表 | 用户 |
| GET | `/api/notifications/:id` | 获取通知详情 | 用户 |
| PUT | `/api/notifications/:id` | 标记通知已读 | 用户 |
| **DELETE** | `/api/notifications/:id` | **删除通知** | 用户 |
| POST | `/api/notifications/mark-all-read` | 标记所有通知已读 | 用户 |

#### 通知模板管理
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/notifications/templates` | 创建通知模板 | 管理员 |
| GET | `/api/notifications/templates` | 获取模板列表 | 管理员 |
| GET | `/api/notifications/templates/:id` | 获取模板详情 | 管理员 |
| PUT | `/api/notifications/templates/:id` | 更新模板 | 管理员 |
| **DELETE** | `/api/notifications/templates/:id` | **删除模板** | 管理员 |
| GET | `/api/notifications/templates/type/:type` | 按类型获取模板 | 管理员 |

#### 订阅管理
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/notifications/subscriptions` | 创建订阅 | 用户 |
| GET | `/api/notifications/subscriptions/:id` | 获取订阅详情 | 用户 |
| PUT | `/api/notifications/subscriptions/:id` | 更新订阅 | 用户 |
| **DELETE** | `/api/notifications/subscriptions/:id` | **删除订阅** | 用户 |
| GET | `/api/notifications/subscriptions/user/:user_id` | 获取用户订阅 | 用户 |
| **DELETE** | `/api/notifications/subscriptions/user/:user_id` | **删除用户所有订阅** | 用户 |

### 17. 搜索功能

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/search` | 基础搜索 | 用户 |
| POST | `/api/search/advanced` | 高级搜索 | 用户 |
| GET | `/api/search/suggestions` | 搜索建议 | 用户 |
| GET | `/api/search/popular` | 热门搜索 | 用户 |
| GET | `/api/search/stats` | 搜索统计 | 用户 |

#### 搜索索引管理
| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| POST | `/api/search/indexes` | 创建搜索索引 | 管理员 |
| GET | `/api/search/indexes` | 获取索引列表 | 管理员 |
| GET | `/api/search/indexes/:id` | 获取索引详情 | 管理员 |
| PUT | `/api/search/indexes/:id` | 更新索引 | 管理员 |
| **DELETE** | `/api/search/indexes/:id` | **删除索引** | 管理员 |
| **DELETE** | `/api/search/indexes/by-source` | **按来源删除索引** | 管理员 |
| POST | `/api/search/indexes/rebuild` | 重建索引 | 管理员 |
| POST | `/api/search/indexes/optimize` | 优化索引 | 管理员 |
| GET | `/api/search/indexes/stats` | 获取索引统计 | 管理员 |

### 18. 系统管理

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/system/metrics` | 获取系统指标 | 管理员 |

### 19. 系统信息

| 方法 | 端点 | 描述 | 权限要求 |
|------|------|------|----------|
| GET | `/api/info` | 获取系统信息 | ❌ |

## 删除操作汇总

以下是所有DELETE操作的汇总：

### 核心资源删除
- `DELETE /api/spaces/:id` - 删除空间
- `DELETE /api/bases/:id` - 删除基础表
- `DELETE /api/tables/:id` - 删除数据表
- `DELETE /api/fields/:id` - 删除字段
- `DELETE /api/records/:id` - 删除记录
- `DELETE /api/views/:id` - 删除视图

### 批量删除操作
- `DELETE /api/bases/bulk-delete` - 批量删除基础表
- `DELETE /api/records/bulk` - 批量删除记录
- `DELETE /api/admin/users/bulk-delete` - 批量删除用户

### 关系删除操作
- `DELETE /api/spaces/:id/collaborators/:collab_id` - 移除协作者
- `DELETE /api/views/:id/grid/columns/:field_id` - 移除网格列
- `DELETE /api/views/:id/form/fields/:field_id` - 移除表单字段

### 管理功能删除
- `DELETE /api/admin/users/:id` - 删除用户
- `DELETE /api/attachments/:id` - 删除附件
- `DELETE /api/notifications/:id` - 删除通知
- `DELETE /api/notifications/templates/:id` - 删除通知模板
- `DELETE /api/notifications/subscriptions/:id` - 删除订阅
- `DELETE /api/notifications/subscriptions/user/:user_id` - 删除用户所有订阅
- `DELETE /api/search/indexes/:id` - 删除搜索索引
- `DELETE /api/search/indexes/by-source` - 按来源删除索引

### 协作功能删除
- `DELETE /api/collaboration/presence` - 移除在线状态
- `DELETE /api/collaboration/cursor` - 移除光标位置

## 视图管理功能详解

视图系统是Teable的核心功能之一，支持多种视图类型：

### 支持的视图类型
1. **网格视图 (Grid View)** - 传统的表格视图
2. **表单视图 (Form View)** - 表单式数据录入
3. **看板视图 (Kanban View)** - 看板式项目管理
4. **日历视图 (Calendar View)** - 日历式时间管理
5. **画廊视图 (Gallery View)** - 卡片式展示

### 视图配置功能
- 视图基本配置管理
- 视图特定配置 (每种视图类型有独特的配置选项)
- 列/字段的动态添加、更新、删除和重排序
- 数据获取和展示优化

## 权限层级说明

### 权限级别
1. **❌ 无需认证**: 公开访问
2. **✅ 用户权限**: 需要有效的JWT token
3. **👥 空间成员**: 需要是空间的成员
4. **🔧 空间管理员**: 需要空间的管理权限
5. **👑 空间所有者**: 需要是空间的所有者
6. **🛡️ 管理员**: 需要系统管理员权限

### 删除权限要求
- **资源所有者**: 可以删除自己创建的资源
- **空间管理员**: 可以删除空间内的资源
- **系统管理员**: 可以删除任何资源
- **特殊权限**: 某些删除操作需要特定权限

## 总结

通过这次完整分析，我们发现Teable Go Backend API实际上包含了：

- **总计约150+个API端点**
- **20+个DELETE操作**
- **5种不同的视图类型**
- **完整的协作功能**
- **丰富的管理功能**

这比我之前整理的文档要全面得多。建议更新所有相关文档以反映这个完整的API集合。
