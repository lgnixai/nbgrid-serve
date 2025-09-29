/**
 * 表格客户端
 * 处理基础表、数据表、字段管理等功能
 */

import { HttpClient } from '../core/http-client';
import { 
  Base,
  Table,
  Field,
  CreateBaseRequest,
  UpdateBaseRequest,
  CreateTableRequest,
  UpdateTableRequest,
  CreateFieldRequest,
  UpdateFieldRequest,
  FieldType,
  FieldOptions,
  PaginatedResponse,
  PaginationParams
} from '../types';

export class TableClient {
  private httpClient: HttpClient;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
  }

  // ==================== 基础表管理 ====================

  /**
   * 创建基础表
   */
  public async createBase(baseData: CreateBaseRequest): Promise<Base> {
    return this.httpClient.post<Base>('/api/bases', baseData);
  }

  /**
   * 获取基础表列表
   */
  public async listBases(params?: PaginationParams & { space_id?: string }): Promise<PaginatedResponse<Base>> {
    return this.httpClient.get<PaginatedResponse<Base>>('/api/bases', params);
  }

  /**
   * 获取基础表详情
   */
  public async getBase(baseId: string): Promise<Base> {
    return this.httpClient.get<Base>(`/api/bases/${baseId}`);
  }

  /**
   * 更新基础表
   */
  public async updateBase(baseId: string, updates: UpdateBaseRequest): Promise<Base> {
    return this.httpClient.put<Base>(`/api/bases/${baseId}`, updates);
  }

  /**
   * 删除基础表
   */
  public async deleteBase(baseId: string): Promise<void> {
    await this.httpClient.delete(`/api/bases/${baseId}`);
  }

  /**
   * 获取基础表统计信息
   */
  public async getBaseStats(baseId: string): Promise<{
    total_tables: number;
    total_fields: number;
    total_records: number;
    total_views: number;
    last_activity_at: string;
  }> {
    return this.httpClient.get(`/api/bases/${baseId}/stats`);
  }

  /**
   * 获取空间下的基础表统计
   */
  public async getSpaceBaseStats(spaceId: string): Promise<{
    total_bases: number;
    total_tables: number;
    total_fields: number;
    total_records: number;
  }> {
    return this.httpClient.get(`/api/bases/space/${spaceId}/stats`);
  }

  // ==================== 数据表管理 ====================

  /**
   * 创建数据表
   */
  public async createTable(tableData: CreateTableRequest): Promise<Table> {
    return this.httpClient.post<Table>('/api/tables', tableData);
  }

  /**
   * 在基础表下创建数据表
   */
  public async createTableInBase(baseId: string, tableData: Omit<CreateTableRequest, 'base_id'>): Promise<Table> {
    return this.httpClient.post<Table>(`/api/bases/${baseId}/tables`, { ...tableData, base_id: baseId });
  }

  /**
   * 获取数据表列表
   */
  public async listTables(params?: PaginationParams & { base_id?: string }): Promise<PaginatedResponse<Table>> {
    return this.httpClient.get<PaginatedResponse<Table>>('/api/tables', params);
  }

  /**
   * 获取基础表下的数据表列表
   */
  public async listTablesInBase(baseId: string, params?: PaginationParams): Promise<PaginatedResponse<Table>> {
    return this.httpClient.get<PaginatedResponse<Table>>(`/api/bases/${baseId}/tables`, params);
  }

  /**
   * 获取数据表详情
   */
  public async getTable(tableId: string): Promise<Table> {
    return this.httpClient.get<Table>(`/api/tables/${tableId}`);
  }

  /**
   * 更新数据表
   */
  public async updateTable(tableId: string, updates: UpdateTableRequest): Promise<Table> {
    return this.httpClient.put<Table>(`/api/tables/${tableId}`, updates);
  }

  /**
   * 删除数据表
   */
  public async deleteTable(tableId: string): Promise<void> {
    await this.httpClient.delete(`/api/tables/${tableId}`);
  }

  /**
   * 复制数据表
   */
  public async duplicateTable(tableId: string, newName: string): Promise<Table> {
    return this.httpClient.post<Table>(`/api/tables/${tableId}/duplicate`, { name: newName });
  }

  // ==================== 字段管理 ====================

  /**
   * 创建字段
   */
  public async createField(fieldData: CreateFieldRequest): Promise<Field> {
    return this.httpClient.post<Field>('/api/fields', fieldData);
  }

  /**
   * 获取字段列表
   */
  public async listFields(params?: PaginationParams & { table_id?: string }): Promise<PaginatedResponse<Field>> {
    return this.httpClient.get<PaginatedResponse<Field>>('/api/fields', params);
  }

  /**
   * 获取数据表的字段列表
   */
  public async listTableFields(tableId: string, params?: PaginationParams): Promise<PaginatedResponse<Field>> {
    return this.httpClient.get<PaginatedResponse<Field>>(`/api/tables/${tableId}/fields`, params);
  }

  /**
   * 获取字段详情
   */
  public async getField(fieldId: string): Promise<Field> {
    return this.httpClient.get<Field>(`/api/fields/${fieldId}`);
  }

  /**
   * 更新字段
   */
  public async updateField(fieldId: string, updates: UpdateFieldRequest): Promise<Field> {
    return this.httpClient.put<Field>(`/api/fields/${fieldId}`, updates);
  }

  /**
   * 删除字段
   */
  public async deleteField(fieldId: string): Promise<void> {
    await this.httpClient.delete(`/api/fields/${fieldId}`);
  }

  /**
   * 获取支持的字段类型
   */
  public async getFieldTypes(): Promise<Array<{
    type: FieldType;
    name: string;
    description: string;
    icon: string;
    supported_options: string[];
  }>> {
    return this.httpClient.get('/api/fields/types');
  }

  /**
   * 获取特定字段类型的信息
   */
  public async getFieldTypeInfo(fieldType: FieldType): Promise<{
    type: FieldType;
    name: string;
    description: string;
    icon: string;
    supported_options: string[];
    default_options: FieldOptions;
    validation_rules: string[];
  }> {
    return this.httpClient.get(`/api/fields/types/${fieldType}`);
  }

  /**
   * 验证字段值
   */
  public async validateFieldValue(fieldId: string, value: any): Promise<{
    valid: boolean;
    errors: string[];
  }> {
    return this.httpClient.post(`/api/fields/${fieldId}/validate`, { value });
  }

  /**
   * 重新排序字段
   */
  public async reorderFields(tableId: string, fieldIds: string[]): Promise<Field[]> {
    return this.httpClient.put<Field[]>(`/api/tables/${tableId}/fields/reorder`, { field_ids: fieldIds });
  }

  /**
   * 批量创建字段
   */
  public async bulkCreateFields(fields: CreateFieldRequest[]): Promise<Field[]> {
    return this.httpClient.post<Field[]>('/api/fields/bulk', { fields });
  }

  /**
   * 批量更新字段
   */
  public async bulkUpdateFields(updates: Array<{ field_id: string; updates: UpdateFieldRequest }>): Promise<Field[]> {
    return this.httpClient.put<Field[]>('/api/fields/bulk', { updates });
  }

  /**
   * 批量删除字段
   */
  public async bulkDeleteFields(fieldIds: string[]): Promise<void> {
    await this.httpClient.post('/api/fields/bulk-delete', { field_ids: fieldIds });
  }

  // ==================== 表格 Schema 管理 ====================

  /**
   * 获取表格 Schema
   */
  public async getTableSchema(tableId: string): Promise<{
    table: Table;
    fields: Field[];
    schema_version: number;
  }> {
    return this.httpClient.get(`/api/tables/${tableId}`);
  }

  /**
   * 更新表格 Schema
   */
  public async updateTableSchema(tableId: string, schema: {
    fields?: Field[];
    schema_version?: number;
  }): Promise<{
    table: Table;
    fields: Field[];
    schema_version: number;
  }> {
    return this.httpClient.put(`/api/tables/${tableId}`, schema);
  }

  /**
   * 验证表格 Schema
   */
  public async validateTableSchema(tableId: string, schema: {
    fields: Field[];
  }): Promise<{
    valid: boolean;
    errors: string[];
    warnings: string[];
  }> {
    return this.httpClient.post(`/api/tables/${tableId}/schema/validate`, schema);
  }

  // ==================== 批量操作 ====================

  /**
   * 批量更新基础表
   */
  public async bulkUpdateBases(updates: Array<{ base_id: string; updates: UpdateBaseRequest }>): Promise<Base[]> {
    return this.httpClient.post<Base[]>('/api/bases/bulk-update', { updates });
  }

  /**
   * 批量删除基础表
   */
  public async bulkDeleteBases(baseIds: string[]): Promise<void> {
    await this.httpClient.post('/api/bases/bulk-delete', { base_ids: baseIds });
  }

  /**
   * 批量更新数据表
   */
  public async bulkUpdateTables(updates: Array<{ table_id: string; updates: UpdateTableRequest }>): Promise<Table[]> {
    return this.httpClient.post<Table[]>('/api/tables/bulk-update', { updates });
  }

  /**
   * 批量删除数据表
   */
  public async bulkDeleteTables(tableIds: string[]): Promise<void> {
    await this.httpClient.post('/api/tables/bulk-delete', { table_ids: tableIds });
  }

  // ==================== 导入导出 ====================

  /**
   * 导出基础表
   */
  public async exportBase(baseId: string, format: 'json' | 'csv' = 'json'): Promise<Blob> {
    return this.httpClient.downloadFile(`/api/bases/${baseId}/export?format=${format}`);
  }

  /**
   * 导入基础表
   */
  public async importBase(file: File | Buffer, options?: { name?: string; description?: string }): Promise<Base> {
    return this.httpClient.uploadFile<Base>('/api/bases/import', file, 'file', options);
  }

  /**
   * 导出数据表
   */
  public async exportTable(tableId: string, format: 'json' | 'csv' = 'json'): Promise<Blob> {
    return this.httpClient.downloadFile(`/api/tables/${tableId}/export?format=${format}`);
  }

  /**
   * 导入数据表
   */
  public async importTable(file: File | Buffer, baseId: string, options?: { name?: string; description?: string }): Promise<Table> {
    return this.httpClient.uploadFile<Table>('/api/tables/import', file, 'file', { base_id: baseId, ...options });
  }
}
