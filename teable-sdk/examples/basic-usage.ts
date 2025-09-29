/**
 * Teable SDK 基础使用示例
 * 展示如何使用 SDK 进行基本的 CRUD 操作
 */

import Teable from '../src/index';

// 初始化 SDK
const teable = new Teable({
  baseUrl: process.env.TEABLE_BASE_URL || 'http://127.0.0.1:3000',
  debug: true,
  disableProxy: true
});

const PREFIX = 'SDK Demo';
const RUN_ID = Math.random().toString(36).slice(2, 8);

async function cleanupPreviousResources() {
  try {
    // 删除此前由示例创建的空间（前缀匹配）
    const pageSize = 50;
    let offset = 0;
    while (true) {
      const page = await teable.listSpaces({ limit: pageSize, offset });
      const items = (page as any)?.data ?? [];
      if (!items.length) break;
      for (const sp of items) {
        const name = sp.name ?? sp.data?.name;
        if (name && String(name).startsWith(PREFIX)) {
          try {
            await teable.deleteSpace(sp.id ?? sp.data?.id);
          } catch {}
        }
      }
      offset += items.length;
      if (items.length < pageSize) break;
    }
  } catch (e) {
    // 若清理失败不影响后续流程
    console.log('清理旧资源时发生问题，继续执行');
  }
}

async function basicUsageExample() {
  try {
    // 0. 预清理，确保幂等
   // await cleanupPreviousResources();

    // 1. 用户登录（如用户不存在则自动注册后重试登录）
    console.log('=== 用户登录 ===');
    const credentials = {
      "email": "test@example.com",
      "password": "TestPassword123!"
    };
    let authResponse;
    try {
      authResponse = await teable.login(credentials);
      console.log('登录成功:', authResponse);
    } catch (err: any) {
      if ((err?.status === 404) || (err?.code === 'NOT_FOUND')) {
        console.log('用户不存在，尝试自动注册...');
        await teable.register({ name: 'Example User', ...credentials });
        authResponse = await teable.login(credentials);
      } else {
        throw err;
      }
    }
    console.log('登录成功:', authResponse.user.name);
 
    // 2. 获取当前用户信息
    console.log('\n=== 获取用户信息 ===');
    const currentUser = await teable.getCurrentUser();
    console.log('当前用户:', currentUser);

    // 3. 创建空间（带随机后缀）
    console.log('\n=== 创建空间 ===');
    const space = await teable.createSpace({
      name: `${PREFIX} 工作空间 ${RUN_ID}`,
      description: '用于演示 SDK 使用的工作空间'
    });
    const spaceId = (space as any).id ?? (space as any).data?.id;
    const spaceName = (space as any).name ?? (space as any).data?.name;
    console.log('创建空间成功:', spaceName);
     
    // 4. 创建基础表（带随机后缀）
    console.log('\n=== 创建基础表 ===');
    const base = await teable.createBase({
      space_id: spaceId,
      name: `项目管理 ${RUN_ID}`,
      description: '项目管理和任务跟踪'
    });
    const baseId = (base as any).id ?? (base as any).data?.id;
    console.log('创建基础表成功:', (base as any).name ?? (base as any).data?.name);

    // 5. 创建数据表（带随机后缀，冲突时重试）
    console.log('\n=== 创建数据表 ===');
    const baseTableName = `任务列表 ${RUN_ID}`;
    async function safeCreateTable(name: string): Promise<any> {
      try {
        const t = await teable.createTable({ base_id: baseId, name, description: '项目任务管理表' });
        return t;
      } catch (err: any) {
        if (err?.status === 409 || err?.code === 'RESOURCE_EXISTS') {
          const retryName = `${name}-${Math.random().toString(36).slice(2, 6)}`;
          return safeCreateTable(retryName);
        }
        throw err;
      }
    }
    const table = await safeCreateTable(baseTableName);
    const tableId = (table as any).id ?? (table as any).data?.id;
    console.log('创建数据表成功:', (table as any).name ?? (table as any).data?.name);
    
    // 6. 创建字段
    console.log('\n=== 创建字段 ===');
    const titleField = await teable.createField({
      table_id: tableId,
      name: '任务标题',
      type: 'text',
      required: true,
      is_primary: true,
      field_order: 1
    });
    console.log('创建标题字段成功:', (titleField as any).name ?? (titleField as any).data?.name);

    const statusField = await teable.createField({
      table_id: tableId,
      name: '状态',
      type: 'single_select',
      required: true,
      options: JSON.stringify({
        choices: [
          { id: 'todo', name: '待办', color: '#FF6B6B' },
          { id: 'doing', name: '进行中', color: '#4ECDC4' },
          { id: 'done', name: '已完成', color: '#45B7D1' }
        ]
      }),
      field_order: 2
    });
    console.log('创建状态字段成功:', (statusField as any).name ?? (statusField as any).data?.name);
    
    const priorityField = await teable.createField({
      table_id: tableId,
      name: '优先级',
      type: 'single_select',
      options: JSON.stringify({
        choices: [
          { id: 'high', name: '高', color: '#FF6B6B' },
          { id: 'medium', name: '中', color: '#FFA726' },
          { id: 'low', name: '低', color: '#66BB6A' }
        ]
      }),
      field_order: 3
    });
    console.log('创建优先级字段成功:', (priorityField as any).name ?? (priorityField as any).data?.name);

    // 7. 创建记录
    console.log('\n=== 创建记录 ===');
    const record1 = await teable.createRecord({
      table_id: tableId,
      data: {
        '任务标题': `设计用户界面 ${RUN_ID}`,
        '状态': 'doing',
        '优先级': 'high'
      }
    });
    console.log('创建记录1成功:', (record1 as any).id ?? (record1 as any).data?.id);
   
    const record2 = await teable.createRecord({
      table_id: tableId,
      data: {
        '任务标题': `编写API文档 ${RUN_ID}`,
        '状态': 'todo',
        '优先级': 'medium'
      }
    });
    console.log('创建记录2成功:', (record2 as any).id ?? (record2 as any).data?.id);
  
    // 8. 批量创建记录
    console.log('\n=== 批量创建记录 ===');
    const records = await teable.bulkCreateRecords(tableId, [
      {
        '任务标题': `数据库设计 ${RUN_ID}`,
        '状态': 'done',
        '优先级': 'high'
      },
      {
        '任务标题': `单元测试 ${RUN_ID}`,
        '状态': 'todo',
        '优先级': 'low'
      },
      {
        '任务标题': `部署上线 ${RUN_ID}`,
        '状态': 'todo',
        '优先级': 'medium'
      }
    ]);
    console.log('批量创建记录成功，共创建:', (records as any)?.length ?? (records as any)?.data?.length, '条记录');

    // 9. 查询记录
    console.log('\n=== 查询记录 ===');
    const allRecords = await teable.listRecords({
      table_id: tableId,
      limit: 10
    });
    console.log('查询到记录总数:', allRecords.total);
    console.log('记录列表:', allRecords.data.map(r => ({
      id: r.id,
      title: r.data['任务标题'],
      status: r.data['状态'],
      priority: r.data['优先级']
    })));

    // 10. 更新记录
    console.log('\n=== 更新记录 ===');
    const updatedRecord = await teable.updateRecord((record1 as any).id ?? (record1 as any).data?.id, {
      '状态': 'done'
    });
    console.log('更新记录成功:', updatedRecord.data['状态']);

    // 11. 创建视图（带随机后缀）
    console.log('\n=== 创建视图 ===');
    const gridView = await teable.createView({
      table_id: tableId,
      name: `网格视图 ${RUN_ID}`,
      type: 'grid',
      is_default: true
    });
    console.log('创建网格视图成功:', (gridView as any).name ?? (gridView as any).data?.name);

    const kanbanView = await teable.createView({
      table_id: tableId,
      name: `看板视图 ${RUN_ID}`,
      type: 'kanban',
      config: {
        kanban: {
          group_field_id: (statusField as any).id ?? (statusField as any).data?.id,
          card_fields: [((titleField as any).id ?? (titleField as any).data?.id), ((priorityField as any).id ?? (priorityField as any).data?.id)]
        }
      }
    });
    console.log('创建看板视图成功:', (kanbanView as any).name ?? (kanbanView as any).data?.name);

    // 12. 获取视图数据（若未实现可跳过）
    console.log('\n=== 获取视图数据 ===');
    try {
      const gridData = await teable.views.getGridData(((gridView as any).id ?? (gridView as any).data?.id));
      console.log('网格视图数据:', {
        records: gridData.records?.length,
        fields: gridData.fields?.length,
        total: gridData.total
      });
    } catch { console.log('网格数据接口暂不可用，已跳过'); }

    try {
      const kanbanData = await teable.views.getKanbanData(((kanbanView as any).id ?? (kanbanView as any).data?.id));
      console.log('看板视图数据:', {
        groups: kanbanData.groups?.length,
        totalRecords: kanbanData.groups?.reduce((sum: number, g: any) => sum + (g.count ?? 0), 0)
      });
    } catch { console.log('看板数据接口暂不可用，已跳过'); }

    // 13. 搜索记录
    console.log('\n=== 搜索记录 ===');
    try {
      const searchResults = await teable.records.search(tableId, '设计');
      console.log('搜索结果:', searchResults.data.length, '条记录');
    } catch { console.log('搜索接口暂不可用，已跳过'); }

    // 14. 获取统计信息（如报错则跳过）
    console.log('\n=== 获取统计信息 ===');
    try {
      const tableStats = await teable.tables.getTableStats(tableId);
      console.log('表格统计:', tableStats);
      const baseStats = await teable.tables.getBaseStats(baseId);
      console.log('基础表统计:', baseStats);
    } catch (e) {
      console.log('统计接口暂不可用，已跳过');
    }

    console.log('\n=== 基础使用示例完成 ===');

  } catch (error) {
    console.error('示例执行出错:', error);
  }
}

// 运行示例
if (require.main === module) {
  basicUsageExample();
}

export { basicUsageExample };
