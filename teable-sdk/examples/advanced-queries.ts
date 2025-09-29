/**
 * Teable SDK 高级查询示例
 * 展示如何使用 SDK 进行复杂的数据查询和操作
 */

import Teable from '../src/index';

// 初始化 SDK
const teable = new Teable({
  baseUrl: 'http://localhost:3000',
  debug: true
});

async function advancedQueriesExample() {
  try {
    // 1. 用户登录
    console.log('=== 用户登录 ===');
    const authResponse = await teable.login({
      email: 'user@example.com',
      password: 'password123'
    });
    console.log('登录成功:', authResponse.user.name);

    // 假设我们有一个任务管理表
    const tableId = 'table_123';

    // 2. 使用查询构建器进行复杂查询
    console.log('\n=== 使用查询构建器 ===');
    
    // 查询状态为"进行中"且优先级为"高"的任务
    const highPriorityTasks = await teable.records.queryBuilder(tableId)
      .where('状态', 'equals', '进行中')
      .where('优先级', 'equals', '高')
      .orderBy('创建时间', 'desc')
      .limit(10)
      .execute();
    
    console.log('高优先级进行中任务:', highPriorityTasks.data.length, '条');

    // 查询标题包含"设计"的任务
    const designTasks = await teable.records.queryBuilder(tableId)
      .where('任务标题', 'contains', '设计')
      .orderBy('优先级', 'desc')
      .execute();
    
    console.log('设计相关任务:', designTasks.data.length, '条');

    // 查询今天创建的任务
    const today = new Date().toISOString().split('T')[0];
    const todayTasks = await teable.records.queryBuilder(tableId)
      .where('创建时间', 'greater_than_or_equal', today)
      .execute();
    
    console.log('今天创建的任务:', todayTasks.data.length, '条');

    // 3. 高级搜索
    console.log('\n=== 高级搜索 ===');
    
    // 全文搜索
    const searchResults = await teable.records.search(tableId, '用户界面 设计');
    console.log('全文搜索结果:', searchResults.data.length, '条');

    // 复合条件搜索
    const advancedSearchResults = await teable.records.advancedSearch(tableId, [
      { field: '状态', operator: 'equals', value: '待办' },
      { field: '优先级', operator: 'in', value: ['高', '中'] },
      { field: '创建时间', operator: 'greater_than', value: '2024-01-01' }
    ], { limit: 20 });
    
    console.log('复合条件搜索结果:', advancedSearchResults.data.length, '条');

    // 4. 聚合查询
    console.log('\n=== 聚合查询 ===');
    
    // 按状态分组统计任务数量
    const statusStats = await teable.records.aggregate(tableId, {
      group_by: ['状态'],
      aggregations: [
        { field: 'id', function: 'count', alias: '任务数量' }
      ]
    });
    console.log('按状态统计:', statusStats);

    // 按优先级和状态分组统计
    const priorityStatusStats = await teable.records.aggregate(tableId, {
      group_by: ['优先级', '状态'],
      aggregations: [
        { field: 'id', function: 'count', alias: '数量' }
      ]
    });
    console.log('按优先级和状态统计:', priorityStatusStats);

    // 计算平均完成时间（假设有完成时间字段）
    const avgCompletionTime = await teable.records.aggregate(tableId, {
      aggregations: [
        { field: '完成时间', function: 'avg', alias: '平均完成时间' },
        { field: 'id', function: 'count', alias: '总任务数' }
      ],
      filter: { field: '状态', operator: 'equals', value: '已完成' }
    });
    console.log('平均完成时间统计:', avgCompletionTime);

    // 5. 字段统计
    console.log('\n=== 字段统计 ===');
    
    // 获取状态字段的统计信息
    const statusFieldStats = await teable.records.getFieldStats(tableId, 'status_field_id');
    console.log('状态字段统计:', {
      total_values: statusFieldStats.total_values,
      unique_values: statusFieldStats.unique_values,
      null_values: statusFieldStats.null_values,
      distribution: statusFieldStats.distribution
    });

    // 6. 批量操作
    console.log('\n=== 批量操作 ===');
    
    // 批量更新任务状态
    const tasksToUpdate = highPriorityTasks.data.slice(0, 3);
    const bulkUpdateData = tasksToUpdate.map(task => ({
      id: task.id,
      data: { '状态': '已完成' }
    }));
    
    const updatedTasks = await teable.bulkUpdateRecords(bulkUpdateData);
    console.log('批量更新任务状态成功:', updatedTasks.length, '条');

    // 批量删除已完成的任务（谨慎操作）
    const completedTasks = await teable.records.queryBuilder(tableId)
      .where('状态', 'equals', '已完成')
      .where('完成时间', 'less_than', '2024-01-01') // 删除很久以前完成的任务
      .limit(5)
      .execute();
    
    if (completedTasks.data.length > 0) {
      const taskIds = completedTasks.data.map(task => task.id);
      await teable.bulkDeleteRecords(taskIds);
      console.log('批量删除旧任务成功:', taskIds.length, '条');
    }

    // 7. 记录版本管理
    console.log('\n=== 记录版本管理 ===');
    
    if (tasksToUpdate.length > 0) {
      const recordId = tasksToUpdate[0].id;
      
      // 获取记录版本历史
      const versionHistory = await teable.records.getVersionHistory(recordId, { limit: 5 });
      console.log('记录版本历史:', versionHistory.data.length, '个版本');
      
      // 获取特定版本的记录
      if (versionHistory.data.length > 1) {
        const oldVersion = await teable.records.getRecordVersion(recordId, versionHistory.data[1].version);
        console.log('获取历史版本成功，版本号:', oldVersion.version);
      }
    }

    // 8. 记录关系操作
    console.log('\n=== 记录关系操作 ===');
    
    // 假设有项目表和任务表的关联关系
    const projectTableId = 'project_table_123';
    const taskTableId = 'task_table_123';
    const linkFieldId = 'project_link_field';
    
    // 获取项目的关联任务
    const projectTasks = await teable.records.getLinkedRecords('project_record_456', linkFieldId);
    console.log('项目关联任务数量:', projectTasks.data.length);

    // 添加任务到项目
    if (projectTasks.data.length > 0) {
      await teable.records.addLinkedRecord('project_record_456', linkFieldId, projectTasks.data[0].id);
      console.log('添加任务关联成功');
    }

    // 9. 数据验证
    console.log('\n=== 数据验证 ===');
    
    // 验证记录数据
    const testData = {
      '任务标题': '测试任务',
      '状态': '待办',
      '优先级': '高',
      '截止时间': '2024-12-31'
    };
    
    const validationResult = await teable.records.validate(testData, tableId);
    console.log('数据验证结果:', {
      valid: validationResult.valid,
      errors: validationResult.errors
    });

    // 10. 导入导出
    console.log('\n=== 导入导出 ===');
    
    // 导出记录为 JSON
    const exportData = await teable.records.exportRecords(tableId, 'json', {
      filter: { field: '状态', operator: 'not_equals', value: '已完成' },
      fields: ['任务标题', '状态', '优先级', '创建时间']
    });
    console.log('导出数据大小:', exportData.size, 'bytes');

    // 11. 复杂查询示例
    console.log('\n=== 复杂查询示例 ===');
    
    // 查询本周创建且未完成的高优先级任务
    const oneWeekAgo = new Date();
    oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
    const oneWeekAgoStr = oneWeekAgo.toISOString().split('T')[0];
    
    const thisWeekHighPriorityTasks = await teable.records.queryBuilder(tableId)
      .where('创建时间', 'greater_than_or_equal', oneWeekAgoStr)
      .where('状态', 'not_equals', '已完成')
      .where('优先级', 'equals', '高')
      .orderBy('创建时间', 'desc')
      .limit(50)
      .execute();
    
    console.log('本周高优先级未完成任务:', thisWeekHighPriorityTasks.data.length, '条');

    // 查询有截止时间且即将到期的任务
    const threeDaysLater = new Date();
    threeDaysLater.setDate(threeDaysLater.getDate() + 3);
    const threeDaysLaterStr = threeDaysLater.toISOString().split('T')[0];
    
    const urgentTasks = await teable.records.queryBuilder(tableId)
      .where('截止时间', 'is_not_empty', null)
      .where('截止时间', 'less_than_or_equal', threeDaysLaterStr)
      .where('状态', 'not_equals', '已完成')
      .orderBy('截止时间', 'asc')
      .execute();
    
    console.log('即将到期的任务:', urgentTasks.data.length, '条');

    console.log('\n=== 高级查询示例完成 ===');

  } catch (error) {
    console.error('高级查询示例执行出错:', error);
  }
}

// 运行示例
if (require.main === module) {
  advancedQueriesExample();
}

export { advancedQueriesExample };
