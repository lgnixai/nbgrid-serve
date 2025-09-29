# Teable TypeScript SDK

一个功能强大的 TypeScript SDK，用于与 Teable 协作数据库平台进行交互。该 SDK 提供了类似 Airtable SDK 的 API 设计，支持完整的 CRUD 操作、实时协作、高级查询等功能。

## 特性

- 🚀 **完整的 API 覆盖** - 支持所有 Teable 平台功能
- 🔄 **实时协作** - WebSocket 支持，实时数据同步
- 📊 **多种视图类型** - 网格、表单、看板、日历、画廊视图
- 🔍 **高级查询** - 复杂查询、聚合、搜索功能
- 🛡️ **类型安全** - 完整的 TypeScript 类型定义
- 🔧 **易于使用** - 类似 Airtable SDK 的 API 设计
- 📦 **模块化** - 按功能模块组织，按需使用
- 🎯 **错误处理** - 完善的错误处理和重试机制

## 安装

```bash
npm install @teable/sdk
```

## 快速开始

### 基本使用

```typescript
import Teable from '@teable/sdk';

// 初始化 SDK
const teable = new Teable({
  baseUrl: 'https://api.teable.ai',
  debug: true
});

// 用户登录
const authResponse = await teable.login({
  email: 'user@example.com',
  password: 'password123'
});

// 创建空间
const space = await teable.createSpace({
  name: '我的工作空间',
  description: '用于项目管理的空间'
});

// 创建基础表
const base = await teable.createBase({
  space_id: space.id,
  name: '项目管理',
  description: '项目管理和任务跟踪'
});

// 创建数据表
const table = await teable.createTable({
  base_id: base.id,
  name: '任务列表',
  description: '项目任务管理表'
});
```

### 字段管理

```typescript
// 创建文本字段
const titleField = await teable.createField({
  table_id: table.id,
  name: '任务标题',
  type: 'text',
  required: true,
  field_order: 1
});

// 创建单选字段
const statusField = await teable.createField({
  table_id: table.id,
  name: '状态',
  type: 'single_select',
  required: true,
  options: {
    choices: [
      { id: 'todo', name: '待办', color: '#FF6B6B' },
      { id: 'doing', name: '进行中', color: '#4ECDC4' },
      { id: 'done', name: '已完成', color: '#45B7D1' }
    ]
  },
  field_order: 2
});

// 创建日期字段
const dueDateField = await teable.createField({
  table_id: table.id,
  name: '截止日期',
  type: 'date',
  field_order: 3
});
```

### 记录操作

```typescript
// 创建记录
const record = await teable.createRecord({
  table_id: table.id,
  data: {
    '任务标题': '设计用户界面',
    '状态': 'doing',
    '截止日期': '2024-12-31'
  }
});

// 查询记录
const records = await teable.listRecords({
  table_id: table.id,
  limit: 20
});

// 更新记录
const updatedRecord = await teable.updateRecord(record.id, {
  '状态': 'done'
});

// 批量创建记录
const bulkRecords = await teable.bulkCreateRecords(table.id, [
  {
    '任务标题': '编写API文档',
    '状态': 'todo',
    '截止日期': '2024-12-25'
  },
  {
    '任务标题': '单元测试',
    '状态': 'todo',
    '截止日期': '2024-12-28'
  }
]);
```

### 高级查询

```typescript
// 使用查询构建器
const highPriorityTasks = await teable.records.queryBuilder(table.id)
  .where('状态', 'equals', '进行中')
  .where('优先级', 'equals', '高')
  .orderBy('创建时间', 'desc')
  .limit(10)
  .execute();

// 复杂查询
const urgentTasks = await teable.records.queryBuilder(table.id)
  .where('截止日期', 'less_than_or_equal', '2024-12-31')
  .where('状态', 'not_equals', '已完成')
  .orderBy('截止日期', 'asc')
  .execute();

// 聚合查询
const statusStats = await teable.records.aggregate(table.id, {
  group_by: ['状态'],
  aggregations: [
    { field: 'id', function: 'count', alias: '任务数量' }
  ]
});

// 全文搜索
const searchResults = await teable.records.search(table.id, '用户界面 设计');
```

### 视图管理

```typescript
// 创建网格视图
const gridView = await teable.createView({
  table_id: table.id,
  name: '网格视图',
  type: 'grid',
  is_default: true
});

// 创建看板视图
const kanbanView = await teable.createView({
  table_id: table.id,
  name: '看板视图',
  type: 'kanban',
  config: {
    kanban: {
      group_field_id: statusField.id,
      card_fields: [titleField.id, dueDateField.id]
    }
  }
});

// 创建日历视图
const calendarView = await teable.createView({
  table_id: table.id,
  name: '日历视图',
  type: 'calendar',
  config: {
    calendar: {
      date_field_id: dueDateField.id,
      title_field_id: titleField.id
    }
  }
});

// 获取视图数据
const gridData = await teable.views.getGridData(gridView.id);
const kanbanData = await teable.views.getKanbanData(kanbanView.id);
const calendarData = await teable.views.getCalendarData(calendarView.id);
```

### 实时协作

```typescript
// 设置事件监听器
teable.onRecordChange((message) => {
  console.log('记录变更:', message.data);
});

teable.onCollaboration((message) => {
  console.log('协作事件:', message.data);
});

teable.onPresenceUpdate((message) => {
  console.log('在线状态更新:', message.data);
});

// 订阅表格的实时更新
teable.subscribeToTable(table.id);

// 更新在线状态
await teable.updatePresence('table', table.id, {
  x: 100,
  y: 200
});

// 更新光标位置
await teable.updateCursor('table', table.id, {
  x: 150,
  y: 250
}, titleField.id, record.id);
```

## API 参考

### 主要类

- `Teable` - 主 SDK 类
- `HttpClient` - HTTP 客户端
- `WebSocketClient` - WebSocket 客户端
- `AuthClient` - 认证客户端
- `SpaceClient` - 空间管理客户端
- `TableClient` - 表格管理客户端
- `RecordClient` - 记录操作客户端
- `ViewClient` - 视图管理客户端
- `CollaborationClient` - 协作功能客户端

### 支持的操作

#### 认证
- `login(credentials)` - 用户登录
- `register(userData)` - 用户注册
- `logout()` - 用户登出
- `getCurrentUser()` - 获取当前用户信息

#### 空间管理
- `createSpace(data)` - 创建空间
- `listSpaces(params)` - 获取空间列表
- `getSpace(id)` - 获取空间详情
- `updateSpace(id, updates)` - 更新空间
- `deleteSpace(id)` - 删除空间

#### 基础表管理
- `createBase(data)` - 创建基础表
- `listBases(params)` - 获取基础表列表
- `getBase(id)` - 获取基础表详情
- `updateBase(id, updates)` - 更新基础表
- `deleteBase(id)` - 删除基础表

#### 数据表管理
- `createTable(data)` - 创建数据表
- `listTables(params)` - 获取数据表列表
- `getTable(id)` - 获取数据表详情
- `updateTable(id, updates)` - 更新数据表
- `deleteTable(id)` - 删除数据表

#### 字段管理
- `createField(data)` - 创建字段
- `listFields(params)` - 获取字段列表
- `getField(id)` - 获取字段详情
- `updateField(id, updates)` - 更新字段
- `deleteField(id)` - 删除字段

#### 记录操作
- `createRecord(data)` - 创建记录
- `listRecords(params)` - 获取记录列表
- `getRecord(id)` - 获取记录详情
- `updateRecord(id, updates)` - 更新记录
- `deleteRecord(id)` - 删除记录
- `bulkCreateRecords(tableId, records)` - 批量创建记录
- `bulkUpdateRecords(updates)` - 批量更新记录
- `bulkDeleteRecords(ids)` - 批量删除记录

#### 查询功能
- `queryBuilder(tableId)` - 创建查询构建器
- `search(tableId, query)` - 全文搜索
- `advancedSearch(tableId, filters)` - 高级搜索
- `aggregate(tableId, config)` - 聚合查询

#### 视图管理
- `createView(data)` - 创建视图
- `listViews(params)` - 获取视图列表
- `getView(id)` - 获取视图详情
- `updateView(id, updates)` - 更新视图
- `deleteView(id)` - 删除视图

#### 协作功能
- `createCollaborationSession(data)` - 创建协作会话
- `updatePresence(resourceType, resourceId, cursor)` - 更新在线状态
- `updateCursor(resourceType, resourceId, cursor, fieldId, recordId)` - 更新光标位置
- `subscribeToTable(tableId)` - 订阅表格更新
- `subscribeToRecord(tableId, recordId)` - 订阅记录更新
- `subscribeToView(viewId)` - 订阅视图更新

### 字段类型

SDK 支持以下字段类型：

- `text` - 文本
- `number` - 数字
- `single_select` - 单选
- `multi_select` - 多选
- `date` - 日期
- `time` - 时间
- `datetime` - 日期时间
- `checkbox` - 复选框
- `url` - 链接
- `email` - 邮箱
- `phone` - 电话
- `currency` - 货币
- `percent` - 百分比
- `duration` - 时长
- `rating` - 评分
- `slider` - 滑块
- `long_text` - 长文本
- `attachment` - 附件
- `link` - 关联
- `lookup` - 查找
- `formula` - 公式
- `rollup` - 汇总
- `count` - 计数
- `created_time` - 创建时间
- `last_modified_time` - 最后修改时间
- `created_by` - 创建者
- `last_modified_by` - 最后修改者
- `auto_number` - 自动编号

### 视图类型

SDK 支持以下视图类型：

- `grid` - 网格视图
- `form` - 表单视图
- `kanban` - 看板视图
- `calendar` - 日历视图
- `gallery` - 画廊视图

## 错误处理

SDK 提供了完善的错误处理机制：

```typescript
import { 
  TeableError,
  AuthenticationError,
  AuthorizationError,
  NotFoundError,
  ValidationError,
  RateLimitError,
  ServerError
} from '@teable/sdk';

try {
  const record = await teable.createRecord(data);
} catch (error) {
  if (error instanceof AuthenticationError) {
    console.log('认证失败，请重新登录');
  } else if (error instanceof ValidationError) {
    console.log('数据验证失败:', error.details);
  } else if (error instanceof RateLimitError) {
    console.log('请求频率超限，请稍后重试');
  } else {
    console.log('未知错误:', error.message);
  }
}
```

## 配置选项

```typescript
const teable = new Teable({
  baseUrl: 'https://api.teable.ai',     // API 基础 URL
  apiKey: 'your-api-key',               // API 密钥（可选）
  accessToken: 'your-access-token',     // 访问令牌（可选）
  refreshToken: 'your-refresh-token',   // 刷新令牌（可选）
  timeout: 30000,                       // 请求超时时间（毫秒）
  retries: 3,                          // 重试次数
  retryDelay: 1000,                    // 重试延迟（毫秒）
  userAgent: 'MyApp/1.0.0',            // 用户代理
  debug: false                         // 调试模式
});
```

## 示例项目

查看 `examples/` 目录中的完整示例：

- `basic-usage.ts` - 基础使用示例
- `collaboration-example.ts` - 协作功能示例
- `advanced-queries.ts` - 高级查询示例

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 支持

如有问题，请访问 [GitHub Issues](https://github.com/teable/teable-sdk/issues) 或联系我们的支持团队。
