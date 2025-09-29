/**
 * Teable SDK 高级查询示例
 * 展示如何使用 SDK 进行复杂的数据查询和操作
 */

import Teable from '../src/index';

// 初始化 SDK（支持环境变量与禁用代理）
const teable = new Teable({
  baseUrl: process.env.TEABLE_BASE_URL || 'http://127.0.0.1:3000',
  debug: true,
  disableProxy: true
});

async function advancedQueriesExample() {
  try {
    // 1. 用户登录
    console.log('=== 用户登录 ===');
    const credentials = { email: 'test@example.com', password: 'TestPassword123!' };
    let authResponse;
    try {
      authResponse = await teable.login(credentials);
    } catch (err: any) {
      if (err?.status === 404 || err?.code === 'NOT_FOUND') {
        console.log('用户不存在，尝试自动注册...');
        await teable.register({ name: 'Example User', ...credentials });
        authResponse = await teable.login(credentials);
      } else {
        throw err;
      }
    }
    console.log('登录成功:', authResponse.user.name);

    // 可选：从环境变量提供 TABLE_ID，否则跳过依赖现有数据表的步骤
    const tableId = process.env.TABLE_ID;
    if (!tableId) {
      console.log('\n未设置 TABLE_ID，跳过依赖既有表的数据操作（查询、聚合、统计等）。');
      return;
    }

    // 2. 使用查询构建器进行复杂查询
    console.log('\n=== 使用查询构建器 ===');
    try {
      const highPriorityTasks = await teable.records.queryBuilder(tableId)
        .where('状态', 'equals', '进行中')
        .where('优先级', 'equals', '高')
        .orderBy('创建时间', 'desc')
        .limit(10)
        .execute();
      console.log('高优先级进行中任务:', highPriorityTasks.data.length, '条');
    } catch { console.log('查询构建器暂不可用，已跳过'); }

    try {
      const designTasks = await teable.records.queryBuilder(tableId)
        .where('任务标题', 'contains', '设计')
        .orderBy('优先级', 'desc')
        .execute();
      console.log('设计相关任务:', designTasks.data.length, '条');
    } catch { console.log('设计相关任务查询不可用，已跳过'); }

    const today = new Date().toISOString().split('T')[0];
    try {
      const todayTasks = await teable.records.queryBuilder(tableId)
        .where('创建时间', 'greater_than_or_equal', today)
        .execute();
      console.log('今天创建的任务:', todayTasks.data.length, '条');
    } catch { console.log('今天创建的任务查询不可用，已跳过'); }

    // 3. 高级搜索（替代 /api/search）
    console.log('\n=== 高级搜索 ===');
    try {
      const advancedSearchResults = await teable.records.advancedSearch(tableId, [
        { field: '任务标题', operator: 'contains', value: '用户界面' },
        { field: '任务标题', operator: 'contains', value: '设计' }
      ], { limit: 20 });
      console.log('高级搜索(contains)结果:', advancedSearchResults.data.length, '条');
    } catch { console.log('高级搜索不可用，已跳过'); }

    // 4. 聚合查询
    console.log('\n=== 聚合查询 ===');
    try {
      const statusStats = await teable.records.aggregate(tableId, {
        group_by: ['状态'],
        aggregations: [
          { field: 'id', function: 'count', alias: '任务数量' }
        ]
      });
      console.log('按状态统计:', statusStats);
    } catch { console.log('按状态统计不可用，已跳过'); }

    try {
      const priorityStatusStats = await teable.records.aggregate(tableId, {
        group_by: ['优先级', '状态'],
        aggregations: [
          { field: 'id', function: 'count', alias: '数量' }
        ]
      });
      console.log('按优先级和状态统计:', priorityStatusStats);
    } catch { console.log('按优先级和状态统计不可用，已跳过'); }

    try {
      const avgCompletionTime = await teable.records.aggregate(tableId, {
        aggregations: [
          { field: '完成时间', function: 'avg', alias: '平均完成时间' },
          { field: 'id', function: 'count', alias: '总任务数' }
        ],
        filter: { field: '状态', operator: 'equals', value: '已完成' }
      });
      console.log('平均完成时间统计:', avgCompletionTime);
    } catch { console.log('平均完成时间统计不可用，已跳过'); }

    // 5. 字段统计
    console.log('\n=== 字段统计 ===');
    try {
      const statusFieldStats = await teable.records.getFieldStats(tableId, 'status_field_id');
      console.log('状态字段统计:', {
        total_values: statusFieldStats.total_values,
        unique_values: statusFieldStats.unique_values,
        null_values: statusFieldStats.null_values,
        distribution: statusFieldStats.distribution
      });
    } catch { console.log('字段统计不可用，已跳过'); }

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
