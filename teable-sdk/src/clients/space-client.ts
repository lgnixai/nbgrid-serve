/**
 * 空间客户端
 * 处理空间管理、协作者管理等功能
 */

import { HttpClient } from '../core/http-client';
import { 
  Space, 
  SpaceCollaborator,
  CreateSpaceRequest, 
  UpdateSpaceRequest,
  AddCollaboratorRequest,
  CollaboratorRole,
  PaginatedResponse,
  PaginationParams
} from '../types';

export class SpaceClient {
  private httpClient: HttpClient;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
  }

  /**
   * 创建空间
   */
  public async create(spaceData: CreateSpaceRequest): Promise<Space> {
    return this.httpClient.post<Space>('/api/spaces', spaceData);
  }

  /**
   * 获取空间列表
   */
  public async list(params?: PaginationParams & { search?: string }): Promise<PaginatedResponse<Space>> {
    return this.httpClient.get<PaginatedResponse<Space>>('/api/spaces', params);
  }

  /**
   * 获取空间详情
   */
  public async get(spaceId: string): Promise<Space> {
    return this.httpClient.get<Space>(`/api/spaces/${spaceId}`);
  }

  /**
   * 更新空间
   */
  public async update(spaceId: string, updates: UpdateSpaceRequest): Promise<Space> {
    return this.httpClient.put<Space>(`/api/spaces/${spaceId}`, updates);
  }

  /**
   * 删除空间
   */
  public async delete(spaceId: string): Promise<void> {
    await this.httpClient.delete(`/api/spaces/${spaceId}`);
  }

  /**
   * 归档空间
   */
  // 后端当前未提供对应路由，移除以对齐 API

  /**
   * 恢复空间
   */
  // 后端当前未提供对应路由，移除以对齐 API

  /**
   * 添加协作者
   */
  public async addCollaborator(spaceId: string, collaboratorData: AddCollaboratorRequest): Promise<SpaceCollaborator> {
    return this.httpClient.post<SpaceCollaborator>(`/api/spaces/${spaceId}/collaborators`, collaboratorData);
  }

  /**
   * 获取协作者列表
   */
  public async getCollaborators(spaceId: string, params?: PaginationParams): Promise<PaginatedResponse<SpaceCollaborator>> {
    return this.httpClient.get<PaginatedResponse<SpaceCollaborator>>(`/api/spaces/${spaceId}/collaborators`, params);
  }

  /**
   * 移除协作者
   */
  public async removeCollaborator(spaceId: string, collaboratorId: string): Promise<void> {
    await this.httpClient.delete(`/api/spaces/${spaceId}/collaborators/${collaboratorId}`);
  }

  /**
   * 更新协作者角色
   */
  public async updateCollaboratorRole(spaceId: string, collaboratorId: string, role: CollaboratorRole): Promise<SpaceCollaborator> {
    return this.httpClient.put<SpaceCollaborator>(`/api/spaces/${spaceId}/collaborators/${collaboratorId}/role`, { role });
  }

  /**
   * 接受空间邀请
   */
  // 后端当前未提供对应路由，移除以对齐 API

  /**
   * 拒绝空间邀请
   */
  // 后端当前未提供对应路由，移除以对齐 API

  /**
   * 离开空间
   */
  // 后端当前未提供对应路由，移除以对齐 API

  /**
   * 获取空间统计信息
   */
  public async getStats(spaceId: string): Promise<{
    total_bases: number;
    total_tables: number;
    total_records: number;
    total_collaborators: number;
    storage_used: number;
    last_activity_at: string;
  }> {
    return this.httpClient.get(`/api/spaces/${spaceId}/stats`);
  }

  /**
   * 批量更新空间
   */
  public async bulkUpdate(updates: Array<{ space_id: string; updates: UpdateSpaceRequest }>): Promise<Space[]> {
    return this.httpClient.post<Space[]>('/api/spaces/bulk-update', { updates });
  }

  /**
   * 批量删除空间
   */
  public async bulkDelete(spaceIds: string[]): Promise<void> {
    await this.httpClient.post('/api/spaces/bulk-delete', { space_ids: spaceIds });
  }

  /**
   * 导出空间
   */
  public async export(spaceId: string, format: 'json' | 'csv' = 'json'): Promise<Blob> {
    return this.httpClient.downloadFile(`/api/spaces/${spaceId}/export?format=${format}`);
  }

  /**
   * 导入空间
   */
  public async import(file: File | Buffer, options?: { name?: string; description?: string }): Promise<Space> {
    return this.httpClient.uploadFile<Space>('/api/spaces/import', file, 'file', options);
  }

  /**
   * 搜索空间
   */
  // 后端当前未提供对应路由，移除以对齐 API（应使用 /api/search）

  /**
   * 获取用户可访问的空间列表
   */
  // 后端当前未提供对应路由，移除以对齐 API

  /**
   * 检查用户对空间的权限
   */
  // 后端接口为 GET /api/spaces/:id/permissions（返回权限详情），如需检查请在应用侧解析

  /**
   * 获取空间活动日志
   */
  // 后端当前未提供对应路由，移除以对齐 API
}
