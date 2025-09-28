# Teable Go Backend WebSocket功能

## 概述

本文档描述了Teable Go后端服务中WebSocket功能的实现，该功能参考了旧版NestJS服务端的WebSocket实现，提供了实时协作和消息推送能力。

## 架构设计

### 核心组件

1. **WebSocket管理器 (Manager)**
   - 管理所有WebSocket连接
   - 处理连接的注册、注销和心跳检测
   - 支持频道订阅和消息广播

2. **WebSocket处理器 (Handler)**
   - 处理WebSocket连接的升级
   - 管理消息的读取和写入
   - 实现各种消息类型的处理逻辑

3. **WebSocket服务 (Service)**
   - 提供高级API接口
   - 支持文档操作发布
   - 实现在线状态管理

### 消息类型

```go
// 连接相关
MessageTypeConnect    // 连接
MessageTypeConnected  // 已连接
MessageTypeDisconnect // 断开连接
MessageTypeError      // 错误

// 文档操作相关
MessageTypeSubscribe   // 订阅
MessageTypeUnsubscribe // 取消订阅
MessageTypeQuery       // 查询
MessageTypeSubmit      // 提交
MessageTypeOp          // 操作

// 心跳相关
MessageTypePing // 心跳
MessageTypePong // 心跳响应
```

## 功能特性

### 1. 连接管理
- 支持多用户同时连接
- 自动心跳检测和连接清理
- 连接状态统计

### 2. 频道订阅
- 支持集合级别订阅 (`collection`)
- 支持文档级别订阅 (`collection.document`)
- 动态订阅和取消订阅

### 3. 消息广播
- 向指定频道广播消息
- 向指定用户的所有连接广播
- 支持排除特定连接

### 4. 文档操作
- 记录操作发布 (`record_tableId`)
- 视图操作发布 (`view_tableId`)
- 字段操作发布 (`field_tableId`)
- 表元数据更新

### 5. 在线状态
- 用户在线状态管理
- 会话信息跟踪
- 在线用户查询

## API接口

### WebSocket连接
```
GET /api/ws/socket?user_id={user_id}&session_id={session_id}
```

### 管理接口

#### 获取统计信息
```
GET /api/ws/stats
```

#### 向频道广播消息
```
POST /api/ws/broadcast/channel
{
  "channel": "record_table_001",
  "message": {...},
  "exclude": ["conn_id_1", "conn_id_2"]
}
```

#### 向用户广播消息
```
POST /api/ws/broadcast/user
{
  "user_id": "user_001",
  "message": {...}
}
```

#### 发布文档操作
```
POST /api/ws/publish/document
{
  "collection": "record_table_001",
  "document": "record_001",
  "operation": [...]
}
```

#### 发布记录操作
```
POST /api/ws/publish/record
{
  "table_id": "table_001",
  "record_id": "record_001",
  "operation": [...]
}
```

#### 发布视图操作
```
POST /api/ws/publish/view
{
  "table_id": "table_001",
  "view_id": "view_001",
  "operation": [...]
}
```

#### 发布字段操作
```
POST /api/ws/publish/field
{
  "table_id": "table_001",
  "field_id": "field_001",
  "operation": [...]
}
```

## 消息格式

### 客户端消息
```json
{
  "type": "subscribe",
  "id": "msg_001",
  "collection": "record_table_001",
  "document": "record_001",
  "data": {...},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 服务器消息
```json
{
  "type": "op",
  "id": "msg_001",
  "collection": "record_table_001",
  "document": "record_001",
  "data": {
    "op": [...],
    "source": "server"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 错误消息
```json
{
  "type": "error",
  "id": "msg_001",
  "error": {
    "code": 400,
    "message": "Invalid request"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 使用示例

### 1. 连接WebSocket
```javascript
const ws = new WebSocket('ws://localhost:3000/api/ws/socket?user_id=user_001&session_id=session_001');

ws.onopen = function() {
    console.log('WebSocket连接已建立');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('收到消息:', message);
};
```

### 2. 订阅频道
```javascript
// 订阅记录频道
ws.send(JSON.stringify({
    type: 'subscribe',
    collection: 'record_table_001'
}));

// 订阅特定文档
ws.send(JSON.stringify({
    type: 'subscribe',
    collection: 'record_table_001',
    document: 'record_001'
}));
```

### 3. 发送心跳
```javascript
setInterval(() => {
    ws.send(JSON.stringify({
        type: 'ping',
        timestamp: Date.now()
    }));
}, 30000);
```

### 4. 提交操作
```javascript
ws.send(JSON.stringify({
    type: 'submit',
    collection: 'record_table_001',
    document: 'record_001',
    operation: [
        { p: ['name'], t: 'string', o: '新记录名称' }
    ]
}));
```

## 测试

### 运行测试
```bash
# 启动测试脚本
./start_websocket_test.sh

# 或手动运行测试
python3 test_websocket.py
```

### 测试内容
- 基本连接功能
- 订阅和取消订阅
- 消息发送和接收
- 多客户端连接
- 心跳机制

## 配置

### 环境变量
```bash
# WebSocket端口 (默认使用HTTP端口)
SOCKET_PORT=3000

# 心跳间隔 (秒)
HEARTBEAT_INTERVAL=30

# 连接超时 (秒)
CONNECTION_TIMEOUT=60
```

### 配置参数
```go
const (
    writeWait = 10 * time.Second      // 写等待时间
    pongWait = 60 * time.Second       // pong等待时间
    pingPeriod = 54 * time.Second     // ping间隔
    maxMessageSize = 512              // 最大消息大小
)
```

## 性能优化

### 1. 连接池管理
- 使用goroutine池处理连接
- 限制最大连接数
- 自动清理无效连接

### 2. 消息队列
- 使用缓冲通道处理消息
- 批量处理消息发送
- 异步消息处理

### 3. 内存管理
- 定期清理过期数据
- 限制消息历史长度
- 优化数据结构

## 监控和日志

### 日志级别
- INFO: 连接建立/断开
- WARN: 异常情况
- ERROR: 错误信息
- DEBUG: 详细调试信息

### 监控指标
- 连接数量
- 消息吞吐量
- 错误率
- 响应时间

## 安全考虑

### 1. 认证授权
- 连接时验证用户身份
- 基于权限的频道访问控制
- 防止未授权访问

### 2. 输入验证
- 消息格式验证
- 参数类型检查
- 防止恶意输入

### 3. 限流保护
- 连接频率限制
- 消息发送频率限制
- 防止DoS攻击

## 扩展功能

### 1. Redis集成
- 跨实例消息同步
- 分布式连接管理
- 消息持久化

### 2. 集群支持
- 多节点部署
- 负载均衡
- 故障转移

### 3. 消息队列
- 异步消息处理
- 消息重试机制
- 死信队列

## 故障排除

### 常见问题

1. **连接失败**
   - 检查服务器是否启动
   - 验证端口是否正确
   - 检查防火墙设置

2. **消息丢失**
   - 检查网络连接
   - 验证消息格式
   - 查看服务器日志

3. **性能问题**
   - 监控连接数量
   - 检查消息频率
   - 优化消息大小

### 调试工具
```bash
# 查看连接状态
curl http://localhost:3000/api/ws/stats

# 测试连接
wscat -c ws://localhost:3000/api/ws/socket?user_id=test

# 查看日志
tail -f logs/websocket.log
```

## 版本历史

- v1.0.0: 初始版本，基本WebSocket功能
- v1.1.0: 添加频道订阅功能
- v1.2.0: 实现在线状态管理
- v1.3.0: 添加文档操作发布

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

AGPL-3.0 License

