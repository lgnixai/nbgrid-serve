/**
 * 视图客户端
 * 处理各种视图类型的管理和配置
 */

import { HttpClient } from '../core/http-client';
import { 
  View,
  ViewType,
  ViewConfig,
  GridViewConfig,
  FormViewConfig,
  KanbanViewConfig,
  CalendarViewConfig,
  GalleryViewConfig,
  CreateViewRequest,
  UpdateViewRequest,
  PaginatedResponse,
  PaginationParams,
  JsonObject
} from '../types';

export class ViewClient {
  private httpClient: HttpClient;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
  }

  // ==================== 基础视图管理 ====================

  /**
   * 创建视图
   */
  public async create(viewData: CreateViewRequest): Promise<View> {
    return this.httpClient.post<View>('/api/views', viewData);
  }

  /**
   * 获取视图列表
   */
  public async list(params?: PaginationParams & { table_id?: string }): Promise<PaginatedResponse<View>> {
    return this.httpClient.get<PaginatedResponse<View>>('/api/views', params);
  }

  /**
   * 获取数据表的视图列表
   */
  public async listTableViews(tableId: string, params?: PaginationParams): Promise<PaginatedResponse<View>> {
    return this.httpClient.get<PaginatedResponse<View>>(`/api/tables/${tableId}/views`, params);
  }

  /**
   * 获取视图详情
   */
  public async get(viewId: string): Promise<View> {
    return this.httpClient.get<View>(`/api/views/${viewId}`);
  }

  /**
   * 更新视图
   */
  public async update(viewId: string, updates: UpdateViewRequest): Promise<View> {
    return this.httpClient.put<View>(`/api/views/${viewId}`, updates);
  }

  /**
   * 删除视图
   */
  public async delete(viewId: string): Promise<void> {
    await this.httpClient.delete(`/api/views/${viewId}`);
  }

  /**
   * 复制视图
   */
  public async duplicate(viewId: string, newName: string): Promise<View> {
    return this.httpClient.post<View>(`/api/views/${viewId}/duplicate`, { name: newName });
  }

  // ==================== 视图配置管理 ====================

  /**
   * 获取视图配置
   */
  public async getConfig(viewId: string): Promise<ViewConfig> {
    return this.httpClient.get<ViewConfig>(`/api/views/${viewId}/config`);
  }

  /**
   * 更新视图配置
   */
  public async updateConfig(viewId: string, config: ViewConfig): Promise<ViewConfig> {
    return this.httpClient.put<ViewConfig>(`/api/views/${viewId}/config`, config);
  }

  // ==================== 网格视图 ====================

  /**
   * 获取网格视图数据
   */
  public async getGridData(viewId: string, params?: PaginationParams): Promise<{
    records: any[];
    fields: any[];
    total: number;
    limit: number;
    offset: number;
  }> {
    return this.httpClient.get(`/api/views/${viewId}/grid/data`, params);
  }

  /**
   * 更新网格视图配置
   */
  public async updateGridConfig(viewId: string, config: GridViewConfig): Promise<GridViewConfig> {
    return this.httpClient.put<GridViewConfig>(`/api/views/${viewId}/grid/config`, config);
  }

  /**
   * 添加网格列
   */
  public async addGridColumn(viewId: string, fieldId: string, config?: {
    width?: number;
    visible?: boolean;
    frozen?: boolean;
  }): Promise<void> {
    await this.httpClient.post(`/api/views/${viewId}/grid/columns`, {
      field_id: fieldId,
      ...config
    });
  }

  /**
   * 更新网格列
   */
  public async updateGridColumn(viewId: string, fieldId: string, config: {
    width?: number;
    visible?: boolean;
    frozen?: boolean;
  }): Promise<void> {
    await this.httpClient.put(`/api/views/${viewId}/grid/columns/${fieldId}`, config);
  }

  /**
   * 移除网格列
   */
  public async removeGridColumn(viewId: string, fieldId: string): Promise<void> {
    await this.httpClient.delete(`/api/views/${viewId}/grid/columns/${fieldId}`);
  }

  /**
   * 重新排序网格列
   */
  public async reorderGridColumns(viewId: string, fieldIds: string[]): Promise<void> {
    await this.httpClient.put(`/api/views/${viewId}/grid/columns/reorder`, { field_ids: fieldIds });
  }

  // ==================== 表单视图 ====================

  /**
   * 获取表单视图数据
   */
  public async getFormData(viewId: string): Promise<{
    fields: any[];
    config: FormViewConfig;
  }> {
    return this.httpClient.get(`/api/views/${viewId}/form/data`);
  }

  /**
   * 更新表单视图配置
   */
  public async updateFormConfig(viewId: string, config: FormViewConfig): Promise<FormViewConfig> {
    return this.httpClient.put<FormViewConfig>(`/api/views/${viewId}/form/config`, config);
  }

  /**
   * 添加表单字段
   */
  public async addFormField(viewId: string, fieldId: string, config?: {
    required?: boolean;
    visible?: boolean;
    order?: number;
  }): Promise<void> {
    await this.httpClient.post(`/api/views/${viewId}/form/fields`, {
      field_id: fieldId,
      ...config
    });
  }

  /**
   * 更新表单字段
   */
  public async updateFormField(viewId: string, fieldId: string, config: {
    required?: boolean;
    visible?: boolean;
    order?: number;
  }): Promise<void> {
    await this.httpClient.put(`/api/views/${viewId}/form/fields/${fieldId}`, config);
  }

  /**
   * 移除表单字段
   */
  public async removeFormField(viewId: string, fieldId: string): Promise<void> {
    await this.httpClient.delete(`/api/views/${viewId}/form/fields/${fieldId}`);
  }

  /**
   * 重新排序表单字段
   */
  public async reorderFormFields(viewId: string, fieldIds: string[]): Promise<void> {
    await this.httpClient.put(`/api/views/${viewId}/form/fields/reorder`, { field_ids: fieldIds });
  }

  // ==================== 看板视图 ====================

  /**
   * 获取看板视图数据
   */
  public async getKanbanData(viewId: string): Promise<{
    groups: Array<{
      id: string;
      name: string;
      records: any[];
      count: number;
    }>;
    config: KanbanViewConfig;
  }> {
    return this.httpClient.get(`/api/views/${viewId}/kanban/data`);
  }

  /**
   * 更新看板视图配置
   */
  public async updateKanbanConfig(viewId: string, config: KanbanViewConfig): Promise<KanbanViewConfig> {
    return this.httpClient.put<KanbanViewConfig>(`/api/views/${viewId}/kanban/config`, config);
  }

  /**
   * 移动看板卡片
   */
  public async moveKanbanCard(viewId: string, recordId: string, newGroupValue: string): Promise<void> {
    await this.httpClient.post(`/api/views/${viewId}/kanban/move`, {
      record_id: recordId,
      new_group_value: newGroupValue
    });
  }

  // ==================== 日历视图 ====================

  /**
   * 获取日历视图数据
   */
  public async getCalendarData(viewId: string, params?: {
    start_date?: string;
    end_date?: string;
    view_mode?: 'month' | 'week' | 'day';
  }): Promise<{
    events: Array<{
      id: string;
      title: string;
      start_date: string;
      end_date?: string;
      color?: string;
      record: any;
    }>;
    config: CalendarViewConfig;
  }> {
    return this.httpClient.get(`/api/views/${viewId}/calendar/data`, params);
  }

  /**
   * 更新日历视图配置
   */
  public async updateCalendarConfig(viewId: string, config: CalendarViewConfig): Promise<CalendarViewConfig> {
    return this.httpClient.put<CalendarViewConfig>(`/api/views/${viewId}/calendar/config`, config);
  }

  // ==================== 画廊视图 ====================

  /**
   * 获取画廊视图数据
   */
  public async getGalleryData(viewId: string, params?: PaginationParams): Promise<{
    records: any[];
    config: GalleryViewConfig;
    total: number;
    limit: number;
    offset: number;
  }> {
    return this.httpClient.get(`/api/views/${viewId}/gallery/data`, params);
  }

  /**
   * 更新画廊视图配置
   */
  public async updateGalleryConfig(viewId: string, config: GalleryViewConfig): Promise<GalleryViewConfig> {
    return this.httpClient.put<GalleryViewConfig>(`/api/views/${viewId}/gallery/config`, config);
  }

  // ==================== 视图数据操作 ====================

  /**
   * 在视图中创建记录
   */
  public async createRecordInView(viewId: string, data: JsonObject): Promise<any> {
    return this.httpClient.post(`/api/views/${viewId}/records`, data);
  }

  /**
   * 在视图中更新记录
   */
  public async updateRecordInView(viewId: string, recordId: string, data: JsonObject): Promise<any> {
    return this.httpClient.put(`/api/views/${viewId}/records/${recordId}`, data);
  }

  /**
   * 在视图中删除记录
   */
  public async deleteRecordInView(viewId: string, recordId: string): Promise<void> {
    await this.httpClient.delete(`/api/views/${viewId}/records/${recordId}`);
  }

  // ==================== 视图共享 ====================

  /**
   * 创建视图共享链接
   */
  public async createShareLink(viewId: string, options?: {
    expires_at?: string;
    password?: string;
    allow_edit?: boolean;
  }): Promise<{
    share_id: string;
    share_url: string;
    expires_at?: string;
  }> {
    return this.httpClient.post(`/api/view-shares/${viewId}/share`, options);
  }

  // 后端未提供获取/删除视图分享详情的公开路由，此处移除

  // ==================== 视图模板 ====================

  /**
   * 获取视图模板列表
   */
  public async getTemplates(viewType?: ViewType): Promise<Array<{
    id: string;
    name: string;
    type: ViewType;
    description: string;
    config: ViewConfig;
    preview_image?: string;
  }>> {
    return this.httpClient.get('/api/views/templates', viewType ? { type: viewType } : undefined);
  }

  /**
   * 从模板创建视图
   */
  public async createFromTemplate(tableId: string, templateId: string, name: string): Promise<View> {
    return this.httpClient.post<View>('/api/views/from-template', {
      table_id: tableId,
      template_id: templateId,
      name
    });
  }

  // ==================== 视图统计 ====================

  /**
   * 获取视图使用统计
   */
  public async getViewStats(viewId: string): Promise<{
    view_id: string;
    total_views: number;
    unique_viewers: number;
    last_viewed_at: string;
    popular_times: Array<{ hour: number; count: number }>;
  }> {
    return this.httpClient.get(`/api/views/${viewId}/stats`);
  }

  /**
   * 获取表格的视图统计
   */
  public async getTableViewStats(tableId: string): Promise<{
    total_views: number;
    views_by_type: Record<ViewType, number>;
    most_popular_view: string;
    last_activity_at: string;
  }> {
    return this.httpClient.get(`/api/tables/${tableId}/views/stats`);
  }
}
