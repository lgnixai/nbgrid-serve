/**
 * 记录客户端
 * 处理记录的 CRUD 操作、查询、批量操作等功能
 */

import { HttpClient } from '../core/http-client';
import { 
  Record,
  CreateRecordRequest,
  UpdateRecordRequest,
  BulkCreateRecordRequest,
  BulkUpdateRecordRequest,
  BulkDeleteRecordRequest,
  RecordQuery,
  FilterExpression,
  SortExpression,
  PaginatedResponse,
  PaginationParams,
  JsonObject,
  FilterOperator
} from '../types';

export class RecordClient {
  private httpClient: HttpClient;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
  }

  // ==================== 基础 CRUD 操作 ====================

  /**
   * 创建记录
   */
  public async create(recordData: CreateRecordRequest): Promise<Record> {
    return this.httpClient.post<Record>('/api/records', recordData);
  }

  /**
   * 获取记录列表
   */
  public async list(params?: PaginationParams & { 
    table_id?: string;
    filter?: FilterExpression;
    sort?: SortExpression[];
  }): Promise<PaginatedResponse<Record>> {
    const resp = await this.httpClient.get<any>('/api/records', params);
    // 适配后端统一分页结构 { data: { list, pagination } } → 已在 HttpClient 解包为 { list, pagination }
    if (resp && resp.list && resp.pagination) {
      const page = resp.pagination.page ?? 0;
      const limit = resp.pagination.limit ?? (params?.limit ?? 20);
      return {
        data: resp.list as Record[],
        total: resp.pagination.total ?? resp.list.length,
        limit,
        offset: resp.pagination.offset ?? ((page > 0 ? (page - 1) * limit : (params?.offset ?? 0)))
      } as PaginatedResponse<Record>;
    }
    return resp as PaginatedResponse<Record>;
  }

  /**
   * 获取数据表的记录列表
   */
  public async listTableRecords(tableId: string, params?: PaginationParams & {
    filter?: FilterExpression;
    sort?: SortExpression[];
  }): Promise<PaginatedResponse<Record>> {
    const resp = await this.httpClient.get<any>(`/api/records`, { table_id: tableId, ...(params || {}) });
    if (resp && resp.list && resp.pagination) {
      const limit = resp.pagination.limit ?? (params?.limit ?? 20);
      const page = resp.pagination.page ?? 0;
      return {
        data: resp.list as Record[],
        total: resp.pagination.total ?? resp.list.length,
        limit,
        offset: resp.pagination.offset ?? ((page > 0 ? (page - 1) * limit : (params?.offset ?? 0)))
      } as PaginatedResponse<Record>;
    }
    return resp as PaginatedResponse<Record>;
  }

  /**
   * 获取记录详情
   */
  public async get(recordId: string): Promise<Record> {
    return this.httpClient.get<Record>(`/api/records/${recordId}`);
  }

  /**
   * 更新记录
   */
  public async update(recordId: string, updates: UpdateRecordRequest): Promise<Record> {
    return this.httpClient.put<Record>(`/api/records/${recordId}`, updates);
  }

  /**
   * 删除记录
   */
  public async delete(recordId: string): Promise<void> {
    await this.httpClient.delete(`/api/records/${recordId}`);
  }

  // ==================== 批量操作 ====================

  /**
   * 批量创建记录
   */
  public async bulkCreate(bulkData: BulkCreateRecordRequest): Promise<Record[]> {
    // 后端期望请求体为 CreateRecordRequest[]
    const payload = bulkData.records.map((data) => ({
      table_id: bulkData.table_id,
      data
    }));
    return this.httpClient.post<Record[]>('/api/records/bulk', payload);
  }

  /**
   * 批量更新记录
   */
  public async bulkUpdate(bulkData: BulkUpdateRecordRequest): Promise<Record[]> {
    return this.httpClient.put<Record[]>('/api/records/bulk', bulkData);
  }

  /**
   * 批量删除记录
   */
  public async bulkDelete(bulkData: BulkDeleteRecordRequest): Promise<void> {
    await this.httpClient.post('/api/records/bulk-delete', bulkData);
  }

  // ==================== 查询操作 ====================

  /**
   * 复杂查询
   */
  public async query(query: RecordQuery): Promise<PaginatedResponse<Record>> {
    return this.httpClient.post<PaginatedResponse<Record>>('/api/records/query', query);
  }

  /**
   * 搜索记录
   */
  public async search(tableId: string, searchQuery: string, params?: PaginationParams): Promise<PaginatedResponse<Record>> {
    return this.httpClient.get<PaginatedResponse<Record>>(`/api/search`, {
      query: searchQuery,
      scope: 'records',
      table_id: tableId,
      ...(params || {})
    } as any);
  }

  /**
   * 高级搜索
   */
  public async advancedSearch(tableId: string, filters: FilterExpression[], params?: PaginationParams): Promise<PaginatedResponse<Record>> {
    return this.httpClient.post<PaginatedResponse<Record>>(`/api/search/advanced`, {
      scope: 'records',
      table_id: tableId,
      filters,
      ...(params || {})
    } as any);
  }

  // ==================== 统计和聚合 ====================

  /**
   * 获取记录统计信息
   */
  public async getStats(tableId: string): Promise<{
    total_records: number;
    created_today: number;
    updated_today: number;
    by_field: JsonObject;
    last_activity_at: string;
  }> {
    return this.httpClient.get(`/api/records/stats`, { table_id: tableId });
  }

  /**
   * 获取字段统计
   */
  public async getFieldStats(tableId: string, fieldId: string): Promise<{
    field_id: string;
    field_name: string;
    field_type: string;
    total_values: number;
    unique_values: number;
    null_values: number;
    distribution: globalThis.Record<string, number>;
  }> {
    return this.httpClient.get(`/api/tables/${tableId}/fields/${fieldId}/stats`);
  }

  /**
   * 聚合查询
   */
  public async aggregate(tableId: string, aggregation: {
    group_by?: string[];
    aggregations: Array<{
      field: string;
      function: 'count' | 'sum' | 'avg' | 'min' | 'max' | 'distinct';
      alias?: string;
    }>;
    filter?: FilterExpression;
  }): Promise<any[]> {
    return this.httpClient.post(`/api/tables/${tableId}/records/aggregate`, aggregation);
  }

  // ==================== 字段值操作 ====================

  /**
   * 更新单个字段值
   */
  public async updateFieldValue(recordId: string, fieldId: string, value: any): Promise<Record> {
    return this.httpClient.patch<Record>(`/api/records/${recordId}/fields/${fieldId}`, { value });
  }

  /**
   * 获取字段值
   */
  public async getFieldValue(recordId: string, fieldId: string): Promise<any> {
    return this.httpClient.get(`/api/records/${recordId}/fields/${fieldId}`);
  }

  /**
   * 批量更新字段值
   */
  public async bulkUpdateFieldValues(updates: Array<{
    record_id: string;
    field_id: string;
    value: any;
  }>): Promise<Record[]> {
    return this.httpClient.post<Record[]>('/api/records/bulk-update-fields', { updates });
  }

  // ==================== 记录关系操作 ====================

  /**
   * 获取关联记录
   */
  public async getLinkedRecords(recordId: string, linkFieldId: string, params?: PaginationParams): Promise<PaginatedResponse<Record>> {
    return this.httpClient.get<PaginatedResponse<Record>>(`/api/records/${recordId}/links/${linkFieldId}`, params);
  }

  /**
   * 添加关联记录
   */
  public async addLinkedRecord(recordId: string, linkFieldId: string, linkedRecordId: string): Promise<void> {
    await this.httpClient.post(`/api/records/${recordId}/links/${linkFieldId}`, { linked_record_id: linkedRecordId });
  }

  /**
   * 移除关联记录
   */
  public async removeLinkedRecord(recordId: string, linkFieldId: string, linkedRecordId: string): Promise<void> {
    await this.httpClient.delete(`/api/records/${recordId}/links/${linkFieldId}/${linkedRecordId}`);
  }

  /**
   * 批量添加关联记录
   */
  public async bulkAddLinkedRecords(recordId: string, linkFieldId: string, linkedRecordIds: string[]): Promise<void> {
    await this.httpClient.post(`/api/records/${recordId}/links/${linkFieldId}/bulk`, { linked_record_ids: linkedRecordIds });
  }

  // ==================== 记录版本管理 ====================

  /**
   * 获取记录版本历史
   */
  public async getVersionHistory(recordId: string, params?: PaginationParams): Promise<PaginatedResponse<{
    version: number;
    data: JsonObject;
    changes: JsonObject;
    updated_by: string;
    updated_at: string;
  }>> {
    return this.httpClient.get<PaginatedResponse<any>>(`/api/records/${recordId}/versions`, params);
  }

  /**
   * 获取特定版本的记录
   */
  public async getRecordVersion(recordId: string, version: number): Promise<Record> {
    return this.httpClient.get<Record>(`/api/records/${recordId}/versions/${version}`);
  }

  /**
   * 恢复到特定版本
   */
  public async restoreToVersion(recordId: string, version: number): Promise<Record> {
    return this.httpClient.post<Record>(`/api/records/${recordId}/restore/${version}`);
  }

  // ==================== 导入导出 ====================

  /**
   * 导出记录
   */
  public async exportRecords(tableId: string, format: 'json' | 'csv' | 'xlsx' = 'json', params?: {
    filter?: FilterExpression;
    fields?: string[];
  }): Promise<Blob> {
    // 通过查询字符串传递导出参数
    const query: any = { format };
    if (params?.filter) query.filter = JSON.stringify(params.filter);
    if (params?.fields) query.fields = params.fields.join(',');
    query.table_id = tableId;
    return this.httpClient.downloadFile(`/api/records/export`, { params: query } as any);
  }

  /**
   * 导入记录
   */
  public async importRecords(tableId: string, file: File | Buffer, options?: {
    update_existing?: boolean;
    skip_errors?: boolean;
    field_mapping?: globalThis.Record<string, string>;
  }): Promise<{
    imported: number;
    updated: number;
    errors: Array<{ row: number; error: string }>;
  }> {
    return this.httpClient.uploadFile(`/api/records/import`, file, 'file', { table_id: tableId, ...(options || {}) });
  }

  // ==================== 记录操作工具 ====================

  /**
   * 复制记录
   */
  public async duplicate(recordId: string, newData?: JsonObject): Promise<Record> {
    return this.httpClient.post<Record>(`/api/records/${recordId}/duplicate`, { data: newData });
  }

  /**
   * 移动记录到其他表
   */
  public async moveToTable(recordId: string, targetTableId: string, fieldMapping?: globalThis.Record<string, string>): Promise<Record> {
    return this.httpClient.post<Record>(`/api/records/${recordId}/move`, {
      target_table_id: targetTableId,
      field_mapping: fieldMapping
    });
  }

  /**
   * 验证记录数据
   */
  public async validate(recordData: JsonObject, tableId: string): Promise<{
    valid: boolean;
    errors: Array<{ field: string; error: string }>;
  }> {
    return this.httpClient.post(`/api/tables/${tableId}/records/validate`, recordData);
  }

  // ==================== 查询构建器 ====================

  /**
   * 创建查询构建器
   */
  public queryBuilder(tableId: string): RecordQueryBuilder {
    return new RecordQueryBuilder(this, tableId);
  }
}

/**
 * 记录查询构建器
 * 提供链式 API 来构建复杂查询
 */
export class RecordQueryBuilder {
  private client: RecordClient;
  private query: RecordQuery;

  constructor(client: RecordClient, tableId: string) {
    this.client = client;
    this.query = {
      table_id: tableId,
      filter: undefined,
      sort: undefined,
      limit: 20,
      offset: 0
    };
  }

  public where(field: string, operator: FilterOperator, value: any): RecordQueryBuilder {
    if (!this.query.filter) {
      this.query.filter = { field, operator, value } as any;
    } else {
      // 构建复合条件
      this.query.filter = {
        logic: 'and',
        conditions: [this.query.filter as any, { field, operator, value } as any]
      } as any;
    }
    return this;
  }

  public orWhere(field: string, operator: FilterOperator, value: any): RecordQueryBuilder {
    if (!this.query.filter) {
      this.query.filter = { field, operator, value } as any;
    } else {
      this.query.filter = {
        logic: 'or',
        conditions: [this.query.filter as any, { field, operator, value } as any]
      } as any;
    }
    return this;
  }

  public orderBy(field: string, direction: 'asc' | 'desc' = 'asc'): RecordQueryBuilder {
    this.query.sort!.push({ field, direction });
    return this;
  }

  public limit(count: number): RecordQueryBuilder {
    this.query.limit = count;
    return this;
  }

  public offset(count: number): RecordQueryBuilder {
    this.query.offset = count;
    return this;
  }

  public async execute(): Promise<PaginatedResponse<Record>> {
    return this.client.query(this.query);
  }

  public async first(): Promise<Record | null> {
    const result = await this.limit(1).execute();
    return (result.data.length > 0 ? (result.data[0] as Record) : null);
  }

  public async count(): Promise<number> {
    const result = await this.limit(1).execute();
    return result.total;
  }
}
