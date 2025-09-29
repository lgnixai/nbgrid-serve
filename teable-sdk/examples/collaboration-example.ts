/**
 * Teable SDK 协作功能示例
 * 展示如何使用 SDK 进行实时协作
 */

import Teable from '../src/index';

// 初始化 SDK
const teable = new Teable({
  baseUrl: 'http://localhost:3000',
  debug: true
});

async function collaborationExample() {
  try {
    // 1. 用户登录
    console.log('=== 用户登录 ===');
    const authResponse = await teable.login({
      email: 'user@example.com',
      password: 'password123'
    });
    console.log('登录成功:', authResponse.user.name);

    // 2. 检查 WebSocket 连接状态
    console.log('\n=== WebSocket 连接状态 ===');
    console.log('连接状态:', teable.getWebSocketState());

    // 3. 手动连接 WebSocket（如果需要）
    if (teable.getWebSocketState() === 'disconnected') {
      await teable.connectWebSocket();
      console.log('WebSocket 连接成功');
    }

    // 4. 创建协作会话
    console.log('\n=== 创建协作会话 ===');
    const session = await teable.createCollaborationSession({
      name: '项目协作会话',
      description: '用于项目开发的实时协作',
      resource_type: 'table',
      resource_id: 'table_123' // 假设的表格ID
    });
    console.log('创建协作会话成功:', session.name);

    // 5. 设置事件监听器
    console.log('\n=== 设置事件监听器 ===');
    
    // 监听协作事件
    teable.onCollaboration((message) => {
      console.log('收到协作消息:', {
        type: message.type,
        action: message.data.action,
        user_id: message.user_id,
        timestamp: message.timestamp
      });
    });

    // 监听记录变更事件
    teable.onRecordChange((message) => {
      console.log('收到记录变更:', {
        action: message.data.action,
        table_id: message.data.table_id,
        record_id: message.data.record_id,
        changes: message.data.changes
      });
    });

    // 监听在线状态更新
    teable.onPresenceUpdate((message) => {
      console.log('收到在线状态更新:', {
        user_id: message.user_id,
        data: message.data
      });
    });

    // 监听光标更新
    teable.onCursorUpdate((message) => {
      console.log('收到光标更新:', {
        user_id: message.user_id,
        data: message.data
      });
    });

    // 监听通知
    teable.onNotification((message) => {
      console.log('收到通知:', {
        type: message.type,
        data: message.data
      });
    });

    // 6. 订阅表格的实时更新
    console.log('\n=== 订阅表格更新 ===');
    teable.subscribeToTable('table_123');
    console.log('已订阅表格 table_123 的实时更新');

    // 7. 订阅特定记录的更新
    console.log('\n=== 订阅记录更新 ===');
    teable.subscribeToRecord('table_123', 'record_456');
    console.log('已订阅记录 record_456 的实时更新');

    // 8. 更新在线状态
    console.log('\n=== 更新在线状态 ===');
    const presence = await teable.updatePresence('table', 'table_123', {
      x: 100,
      y: 200
    });
    console.log('更新在线状态成功:', presence);

    // 9. 更新光标位置
    console.log('\n=== 更新光标位置 ===');
    await teable.updateCursor('table', 'table_123', {
      x: 150,
      y: 250
    }, 'field_789', 'record_456');
    console.log('更新光标位置成功');

    // 10. 获取在线用户列表
    console.log('\n=== 获取在线用户列表 ===');
    const presenceList = await teable.collaboration.getPresenceList('table', 'table_123');
    console.log('在线用户数量:', presenceList.length);
    presenceList.forEach(p => {
      console.log(`用户 ${p.user_id} 在位置 (${p.cursor_position?.x}, ${p.cursor_position?.y})`);
    });

    // 11. 获取光标位置列表
    console.log('\n=== 获取光标位置列表 ===');
    const cursorList = await teable.collaboration.getCursorList('table', 'table_123');
    console.log('光标位置数量:', cursorList.length);
    cursorList.forEach(c => {
      console.log(`用户 ${c.user_id} 的光标在字段 ${c.field_id}，记录 ${c.record_id}`);
    });

    // 12. 加入协作会话
    console.log('\n=== 加入协作会话 ===');
    const participant = await teable.collaboration.joinSession(session.id);
    console.log('加入协作会话成功，角色:', participant.role);

    // 13. 获取参与者列表
    console.log('\n=== 获取参与者列表 ===');
    const participants = await teable.collaboration.getParticipants(session.id);
    console.log('参与者数量:', participants.length);
    participants.forEach(p => {
      console.log(`参与者 ${p.user_id}，角色: ${p.role}，加入时间: ${p.joined_at}`);
    });

    // 14. 获取协作统计信息
    console.log('\n=== 获取协作统计信息 ===');
    const stats = await teable.collaboration.getCollaborationStats();
    console.log('协作统计:', {
      active_sessions: stats.active_sessions,
      total_participants: stats.total_participants,
      online_users: stats.online_users
    });

    // 15. 模拟一些协作活动
    console.log('\n=== 模拟协作活动 ===');
    
    // 模拟光标移动
    for (let i = 0; i < 5; i++) {
      await new Promise(resolve => setTimeout(resolve, 1000));
      await teable.updateCursor('table', 'table_123', {
        x: 100 + i * 10,
        y: 200 + i * 5
      }, 'field_789', 'record_456');
      console.log(`光标移动到 (${100 + i * 10}, ${200 + i * 5})`);
    }

    // 16. 离开协作会话
    console.log('\n=== 离开协作会话 ===');
    await teable.collaboration.leaveSession(session.id);
    console.log('已离开协作会话');

    // 17. 移除在线状态
    console.log('\n=== 移除在线状态 ===');
    await teable.collaboration.removePresence();
    console.log('已移除在线状态');

    // 18. 取消订阅
    console.log('\n=== 取消订阅 ===');
    teable.collaboration.unsubscribeFromTable('table_123');
    teable.collaboration.unsubscribeFromRecord('table_123', 'record_456');
    console.log('已取消所有订阅');

    console.log('\n=== 协作功能示例完成 ===');

  } catch (error) {
    console.error('协作示例执行出错:', error);
  }
}

// 运行示例
if (require.main === module) {
  collaborationExample();
}

export { collaborationExample };
