/**
 * 协作客户端
 * 处理实时协作、在线状态、光标位置等功能
 */

import { HttpClient } from '../core/http-client';
import { WebSocketClient } from '../core/websocket-client';
import { 
  CollaborationSession,
  CollaborationParticipant,
  Presence,
  CursorPosition,
  WebSocketMessage,
  CollaborationMessage,
  RecordChangeMessage
} from '../types';

export class CollaborationClient {
  private httpClient: HttpClient;
  private wsClient: WebSocketClient | undefined;

  constructor(httpClient: HttpClient, wsClient?: WebSocketClient) {
    this.httpClient = httpClient;
    this.wsClient = wsClient;
  }

  // ==================== 协作会话管理 ====================

  /**
   * 创建协作会话
   */
  public async createSession(sessionData: {
    name: string;
    description?: string;
    resource_type: string;
    resource_id: string;
  }): Promise<CollaborationSession> {
    return this.httpClient.post<CollaborationSession>('/api/collaboration/sessions', sessionData);
  }

  /**
   * 获取协作会话列表
   */
  public async listSessions(params?: {
    limit?: number;
    offset?: number;
    resource_type?: string;
    resource_id?: string;
  }): Promise<{
    data: CollaborationSession[];
    total: number;
    limit: number;
    offset: number;
  }> {
    return this.httpClient.get('/api/collaboration/sessions', params);
  }

  /**
   * 获取协作会话详情
   */
  public async getSession(sessionId: string): Promise<CollaborationSession> {
    return this.httpClient.get<CollaborationSession>(`/api/collaboration/sessions/${sessionId}`);
  }

  /**
   * 更新协作会话
   */
  public async updateSession(sessionId: string, updates: {
    name?: string;
    description?: string;
  }): Promise<CollaborationSession> {
    return this.httpClient.put<CollaborationSession>(`/api/collaboration/sessions/${sessionId}`, updates);
  }

  /**
   * 结束协作会话
   */
  public async endSession(sessionId: string): Promise<void> {
    await this.httpClient.delete(`/api/collaboration/sessions/${sessionId}`);
  }

  /**
   * 加入协作会话
   */
  public async joinSession(sessionId: string): Promise<CollaborationParticipant> {
    return this.httpClient.post<CollaborationParticipant>(`/api/collaboration/sessions/${sessionId}/join`);
  }

  /**
   * 离开协作会话
   */
  public async leaveSession(sessionId: string): Promise<void> {
    await this.httpClient.post(`/api/collaboration/sessions/${sessionId}/leave`);
  }

  /**
   * 获取参与者列表
   */
  public async getParticipants(sessionId: string): Promise<CollaborationParticipant[]> {
    return this.httpClient.get<CollaborationParticipant[]>(`/api/collaboration/sessions/${sessionId}/participants`);
  }

  /**
   * 邀请参与协作
   */
  public async inviteToSession(sessionId: string, userIds: string[]): Promise<void> {
    await this.httpClient.post(`/api/collaboration/sessions/${sessionId}/invite`, { user_ids: userIds });
  }

  /**
   * 移除参与者
   */
  public async removeParticipant(sessionId: string, userId: string): Promise<void> {
    await this.httpClient.post(`/api/collaboration/sessions/${sessionId}/kick`, { user_id: userId });
  }

  // ==================== 在线状态管理 ====================

  /**
   * 更新在线状态
   */
  public async updatePresence(resourceType: string, resourceId: string, cursorPosition?: CursorPosition): Promise<Presence> {
    const presence = await this.httpClient.post<Presence>('/api/collaboration/presence', {
      resource_type: resourceType,
      resource_id: resourceId,
      cursor_position: cursorPosition
    });

    // 同时通过 WebSocket 发送状态更新
    if (this.wsClient) {
      this.wsClient.updatePresence(resourceType, resourceId, cursorPosition);
    }

    return presence;
  }

  /**
   * 移除在线状态
   */
  public async removePresence(): Promise<void> {
    await this.httpClient.delete('/api/collaboration/presence');
  }

  /**
   * 获取在线状态列表
   */
  public async getPresenceList(resourceType?: string, resourceId?: string): Promise<Presence[]> {
    return this.httpClient.get<Presence[]>('/api/collaboration/presence', {
      resource_type: resourceType,
      resource_id: resourceId
    });
  }

  // ==================== 光标位置管理 ====================

  /**
   * 更新光标位置
   */
  public async updateCursor(
    resourceType: string, 
    resourceId: string, 
    cursorPosition: CursorPosition, 
    fieldId?: string, 
    recordId?: string
  ): Promise<void> {
    await this.httpClient.post('/api/collaboration/cursor', {
      resource_type: resourceType,
      resource_id: resourceId,
      cursor_position: cursorPosition,
      field_id: fieldId,
      record_id: recordId
    });

    // 同时通过 WebSocket 发送光标更新
    if (this.wsClient) {
      this.wsClient.updateCursor(resourceType, resourceId, cursorPosition, fieldId, recordId);
    }
  }

  /**
   * 移除光标位置
   */
  public async removeCursor(): Promise<void> {
    await this.httpClient.delete('/api/collaboration/cursor');
  }

  /**
   * 获取光标位置列表
   */
  public async getCursorList(resourceType?: string, resourceId?: string): Promise<Array<{
    user_id: string;
    resource_type: string;
    resource_id: string;
    cursor_position: CursorPosition;
    field_id?: string;
    record_id?: string;
    last_seen: string;
  }>> {
    return this.httpClient.get('/api/collaboration/cursor', {
      resource_type: resourceType,
      resource_id: resourceId
    });
  }

  // ==================== WebSocket 事件处理 ====================

  /**
   * 设置 WebSocket 客户端
   */
  public setWebSocketClient(wsClient: WebSocketClient): void {
    this.wsClient = wsClient;
    this.setupWebSocketEventHandlers();
  }

  private setupWebSocketEventHandlers(): void {
    if (!this.wsClient) return;

    // 这里不再重新发射事件，保留给上层通过 wsClient 自行监听
  }

  // ==================== 实时协作方法 ====================

  /**
   * 订阅表格的实时更新
   */
  public subscribeToTable(tableId: string): void {
    if (this.wsClient) {
      this.wsClient.subscribeToTable(tableId);
    }
  }

  /**
   * 取消订阅表格
   */
  public unsubscribeFromTable(tableId: string): void {
    if (this.wsClient) {
      this.wsClient.unsubscribeFromTable(tableId);
    }
  }

  /**
   * 订阅记录的实时更新
   */
  public subscribeToRecord(tableId: string, recordId: string): void {
    if (this.wsClient) {
      this.wsClient.subscribeToRecord(tableId, recordId);
    }
  }

  /**
   * 取消订阅记录
   */
  public unsubscribeFromRecord(tableId: string, recordId: string): void {
    if (this.wsClient) {
      this.wsClient.unsubscribeFromRecord(tableId, recordId);
    }
  }

  /**
   * 订阅视图的实时更新
   */
  public subscribeToView(viewId: string): void {
    if (this.wsClient) {
      this.wsClient.subscribeToView(viewId);
    }
  }

  /**
   * 取消订阅视图
   */
  public unsubscribeFromView(viewId: string): void {
    if (this.wsClient) {
      this.wsClient.unsubscribeFromView(viewId);
    }
  }

  // ==================== 协作统计 ====================

  /**
   * 获取协作统计信息
   */
  public async getCollaborationStats(): Promise<{
    active_sessions: number;
    total_participants: number;
    online_users: number;
    recent_activity: Array<{
      type: string;
      user_id: string;
      resource_type: string;
      resource_id: string;
      timestamp: string;
    }>;
  }> {
    return this.httpClient.get('/api/collaboration/stats');
  }

  /**
   * 获取用户协作活动
   */
  public async getUserCollaborationActivity(userId: string, params?: {
    limit?: number;
    offset?: number;
    start_date?: string;
    end_date?: string;
  }): Promise<{
    data: Array<{
      session_id: string;
      resource_type: string;
      resource_id: string;
      action: string;
      timestamp: string;
    }>;
    total: number;
  }> {
    return this.httpClient.get(`/api/collaboration/users/${userId}/activity`, params);
  }

  // ==================== 事件监听简化 ====================
  public onCollaboration(listener: (message: CollaborationMessage) => void): void {
    this.wsClient?.on('collaboration', listener);
  }
  public onRecordChange(listener: (message: RecordChangeMessage) => void): void {
    this.wsClient?.on('record_change', listener);
  }
  public onPresenceUpdate(listener: (message: WebSocketMessage) => void): void {
    this.wsClient?.on('presence_update', listener);
  }
  public onCursorUpdate(listener: (message: WebSocketMessage) => void): void {
    this.wsClient?.on('cursor_update', listener);
  }
  public onNotification(listener: (message: WebSocketMessage) => void): void {
    this.wsClient?.on('notification', listener);
  }
}
