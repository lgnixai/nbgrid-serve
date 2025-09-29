/**
 * Teable SDK 主入口文件
 * 提供统一的 API 接口，类似 Airtable SDK 的设计模式
 */

import { HttpClient } from './core/http-client';
import { WebSocketClient } from './core/websocket-client';
import { AuthClient } from './clients/auth-client';
import { SpaceClient } from './clients/space-client';
import { TableClient } from './clients/table-client';
import { RecordClient } from './clients/record-client';
import { ViewClient } from './clients/view-client';
import { CollaborationClient } from './clients/collaboration-client';

import { 
  TeableConfig, 
  User,
  Space,
  Base,
  Table,
  Field,
  Record,
  View,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  CreateSpaceRequest,
  CreateBaseRequest,
  CreateTableRequest,
  CreateFieldRequest,
  CreateRecordRequest,
  CreateViewRequest,
  PaginatedResponse,
  PaginationParams,
  FilterExpression,
  SortExpression,
  ViewType,
  ViewConfig,
  FieldType,
  FieldOptions,
  CollaborationSession,
  Presence,
  CursorPosition,
  WebSocketMessage,
  CollaborationMessage,
  RecordChangeMessage,
  TeableError,
  AuthenticationError,
  AuthorizationError,
  NotFoundError,
  ValidationError,
  RateLimitError,
  ServerError,
  JsonObject
} from './types';

/**
 * Teable SDK 主类
 * 提供类似 Airtable SDK 的 API 设计
 */
export class Teable {
  private httpClient: HttpClient;
  private wsClient?: WebSocketClient;
  private authClient: AuthClient;
  private spaceClient: SpaceClient;
  private tableClient: TableClient;
  private recordClient: RecordClient;
  private viewClient: ViewClient;
  private collaborationClient: CollaborationClient;

  constructor(config: TeableConfig) {
    // 初始化 HTTP 客户端
    this.httpClient = new HttpClient(config);

    // 初始化各个功能客户端
    this.authClient = new AuthClient(this.httpClient);
    this.spaceClient = new SpaceClient(this.httpClient);
    this.tableClient = new TableClient(this.httpClient);
    this.recordClient = new RecordClient(this.httpClient);
    this.viewClient = new ViewClient(this.httpClient);
    this.collaborationClient = new CollaborationClient(this.httpClient);

    // 如果配置了 WebSocket，初始化 WebSocket 客户端
    if (config.accessToken) {
      this.initializeWebSocket(config);
    }
  }

  private async initializeWebSocket(config: TeableConfig): Promise<void> {
    const wsOptions = {
      debug: config.debug || false,
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000
    } as any;

    this.wsClient = new WebSocketClient(config, wsOptions);
    this.collaborationClient.setWebSocketClient(this.wsClient);
  }

  // ==================== 认证相关方法 ====================

  /**
   * 用户登录
   */
  public async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.authClient.login(credentials);
    console.log('登录成功:', response);
    // 登录成功后初始化 WebSocket 连接
    if (this.wsClient) {
      await this.wsClient.connect();
    }
    
    return response;
  }

  /**
   * 用户注册
   */
  public async register(userData: RegisterRequest): Promise<AuthResponse> {
    const response = await this.authClient.register(userData);
    
    // 注册成功后初始化 WebSocket 连接
    if (this.wsClient) {
      await this.wsClient.connect();
    }
    
    return response;
  }

  /**
   * 用户登出
   */
  public async logout(): Promise<void> {
    await this.authClient.logout();
    
    // 登出后断开 WebSocket 连接
    if (this.wsClient) {
      this.wsClient.disconnect();
    }
  }

  /**
   * 获取当前用户信息
   */
  public async getCurrentUser(): Promise<User> {
    return this.authClient.getCurrentUser();
  }

  /**
   * 检查是否已登录
   */
  public isAuthenticated(): boolean {
    return this.authClient.isAuthenticated();
  }

  // ==================== 空间管理方法 ====================

  /**
   * 创建空间
   */
  public async createSpace(spaceData: CreateSpaceRequest): Promise<Space> {
    return this.spaceClient.create(spaceData);
  }

  /**
   * 获取空间列表
   */
  public async listSpaces(params?: PaginationParams & { search?: string }): Promise<PaginatedResponse<Space>> {
    return this.spaceClient.list(params);
  }

  /**
   * 获取空间详情
   */
  public async getSpace(spaceId: string): Promise<Space> {
    return this.spaceClient.get(spaceId);
  }

  /**
   * 更新空间
   */
  public async updateSpace(spaceId: string, updates: Partial<Space>): Promise<Space> {
    return this.spaceClient.update(spaceId, updates);
  }

  /**
   * 删除空间
   */
  public async deleteSpace(spaceId: string): Promise<void> {
    await this.spaceClient.delete(spaceId);
  }

  // ==================== 基础表管理方法 ====================

  /**
   * 创建基础表
   */
  public async createBase(baseData: CreateBaseRequest): Promise<Base> {
    return this.tableClient.createBase(baseData);
  }

  /**
   * 获取基础表列表
   */
  public async listBases(params?: PaginationParams & { space_id?: string }): Promise<PaginatedResponse<Base>> {
    return this.tableClient.listBases(params);
  }

  /**
   * 获取基础表详情
   */
  public async getBase(baseId: string): Promise<Base> {
    return this.tableClient.getBase(baseId);
  }

  /**
   * 更新基础表
   */
  public async updateBase(baseId: string, updates: Partial<Base>): Promise<Base> {
    return this.tableClient.updateBase(baseId, updates);
  }

  /**
   * 删除基础表
   */
  public async deleteBase(baseId: string): Promise<void> {
    await this.tableClient.deleteBase(baseId);
  }

  // ==================== 数据表管理方法 ====================

  /**
   * 创建数据表
   */
  public async createTable(tableData: CreateTableRequest): Promise<Table> {
    return this.tableClient.createTable(tableData);
  }

  /**
   * 获取数据表列表
   */
  public async listTables(params?: PaginationParams & { base_id?: string }): Promise<PaginatedResponse<Table>> {
    return this.tableClient.listTables(params);
  }

  /**
   * 获取数据表详情
   */
  public async getTable(tableId: string): Promise<Table> {
    return this.tableClient.getTable(tableId);
  }

  /**
   * 更新数据表
   */
  public async updateTable(tableId: string, updates: Partial<Table>): Promise<Table> {
    return this.tableClient.updateTable(tableId, updates);
  }

  /**
   * 删除数据表
   */
  public async deleteTable(tableId: string): Promise<void> {
    await this.tableClient.deleteTable(tableId);
  }

  // ==================== 字段管理方法 ====================

  /**
   * 创建字段
   */
  public async createField(fieldData: CreateFieldRequest): Promise<Field> {
    return this.tableClient.createField(fieldData);
  }

  /**
   * 获取字段列表
   */
  public async listFields(params?: PaginationParams & { table_id?: string }): Promise<PaginatedResponse<Field>> {
    return this.tableClient.listFields(params);
  }

  /**
   * 获取字段详情
   */
  public async getField(fieldId: string): Promise<Field> {
    return this.tableClient.getField(fieldId);
  }

  /**
   * 更新字段
   */
  public async updateField(fieldId: string, updates: Partial<Field>): Promise<Field> {
    return this.tableClient.updateField(fieldId, updates);
  }

  /**
   * 删除字段
   */
  public async deleteField(fieldId: string): Promise<void> {
    await this.tableClient.deleteField(fieldId);
  }

  // ==================== 记录管理方法 ====================

  /**
   * 创建记录
   */
  public async createRecord(recordData: CreateRecordRequest): Promise<Record> {
    return this.recordClient.create(recordData);
  }

  /**
   * 获取记录列表
   */
  public async listRecords(params?: PaginationParams & { 
    table_id?: string;
    filter?: FilterExpression;
    sort?: SortExpression[];
  }): Promise<PaginatedResponse<Record>> {
    return this.recordClient.list(params);
  }

  /**
   * 获取记录详情
   */
  public async getRecord(recordId: string): Promise<Record> {
    return this.recordClient.get(recordId);
  }

  /**
   * 更新记录
   */
  public async updateRecord(recordId: string, updates: JsonObject): Promise<Record> {
    return this.recordClient.update(recordId, { data: updates });
  }

  /**
   * 删除记录
   */
  public async deleteRecord(recordId: string): Promise<void> {
    await this.recordClient.delete(recordId);
  }

  /**
   * 批量创建记录
   */
  public async bulkCreateRecords(tableId: string, records: JsonObject[]): Promise<Record[]> {
    return this.recordClient.bulkCreate({ table_id: tableId, records });
  }

  /**
   * 批量更新记录
   */
  public async bulkUpdateRecords(updates: Array<{ id: string; data: JsonObject }>): Promise<Record[]> {
    return this.recordClient.bulkUpdate({ records: updates });
  }

  /**
   * 批量删除记录
   */
  public async bulkDeleteRecords(recordIds: string[]): Promise<void> {
    await this.recordClient.bulkDelete({ record_ids: recordIds });
  }

  // ==================== 视图管理方法 ====================

  /**
   * 创建视图
   */
  public async createView(viewData: CreateViewRequest): Promise<View> {
    return this.viewClient.create(viewData);
  }

  /**
   * 获取视图列表
   */
  public async listViews(params?: PaginationParams & { table_id?: string }): Promise<PaginatedResponse<View>> {
    return this.viewClient.list(params);
  }

  /**
   * 获取视图详情
   */
  public async getView(viewId: string): Promise<View> {
    return this.viewClient.get(viewId);
  }

  /**
   * 更新视图
   */
  public async updateView(viewId: string, updates: Partial<View>): Promise<View> {
    return this.viewClient.update(viewId, updates);
  }

  /**
   * 删除视图
   */
  public async deleteView(viewId: string): Promise<void> {
    await this.viewClient.delete(viewId);
  }

  // ==================== 协作功能方法 ====================

  /**
   * 创建协作会话
   */
  public async createCollaborationSession(sessionData: {
    name: string;
    description?: string;
    resource_type: string;
    resource_id: string;
  }): Promise<CollaborationSession> {
    return this.collaborationClient.createSession(sessionData);
  }

  /**
   * 更新在线状态
   */
  public async updatePresence(resourceType: string, resourceId: string, cursorPosition?: CursorPosition): Promise<Presence> {
    return this.collaborationClient.updatePresence(resourceType, resourceId, cursorPosition);
  }

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
    return this.collaborationClient.updateCursor(resourceType, resourceId, cursorPosition, fieldId, recordId);
  }

  /**
   * 订阅表格的实时更新
   */
  public subscribeToTable(tableId: string): void {
    this.collaborationClient.subscribeToTable(tableId);
  }

  /**
   * 订阅记录的实时更新
   */
  public subscribeToRecord(tableId: string, recordId: string): void {
    this.collaborationClient.subscribeToRecord(tableId, recordId);
  }

  /**
   * 订阅视图的实时更新
   */
  public subscribeToView(viewId: string): void {
    this.collaborationClient.subscribeToView(viewId);
  }

  // ==================== 事件监听方法 ====================

  /**
   * 监听协作事件
   */
  public onCollaboration(callback: (message: CollaborationMessage) => void): void {
    this.collaborationClient.onCollaboration(callback);
  }

  /**
   * 监听记录变更事件
   */
  public onRecordChange(callback: (message: RecordChangeMessage) => void): void {
    this.collaborationClient.onRecordChange(callback);
  }

  /**
   * 监听在线状态更新事件
   */
  public onPresenceUpdate(callback: (message: WebSocketMessage) => void): void {
    this.collaborationClient.onPresenceUpdate(callback);
  }

  /**
   * 监听光标更新事件
   */
  public onCursorUpdate(callback: (message: WebSocketMessage) => void): void {
    this.collaborationClient.onCursorUpdate(callback);
  }

  /**
   * 监听通知事件
   */
  public onNotification(callback: (message: WebSocketMessage) => void): void {
    this.collaborationClient.onNotification(callback);
  }

  // ==================== 工具方法 ====================

  /**
   * 获取系统健康状态
   */
  public async healthCheck(): Promise<{ status: string; timestamp: string; version: string }> {
    return this.httpClient.healthCheck();
  }

  /**
   * 获取系统信息
   */
  public async getSystemInfo(): Promise<any> {
    return this.httpClient.getSystemInfo();
  }

  /**
   * 获取 WebSocket 连接状态
   */
  public getWebSocketState(): 'connecting' | 'connected' | 'disconnected' {
    return this.wsClient?.getConnectionState() || 'disconnected';
  }

  /**
   * 手动连接 WebSocket
   */
  public async connectWebSocket(): Promise<void> {
    if (this.wsClient) {
      await this.wsClient.connect();
    }
  }

  /**
   * 断开 WebSocket 连接
   */
  public disconnectWebSocket(): void {
    if (this.wsClient) {
      this.wsClient.disconnect();
    }
  }

  // ==================== 客户端访问器 ====================

  /**
   * 获取认证客户端
   */
  public get auth(): AuthClient {
    return this.authClient;
  }

  /**
   * 获取空间客户端
   */
  public get spaces(): SpaceClient {
    return this.spaceClient;
  }

  /**
   * 获取表格客户端
   */
  public get tables(): TableClient {
    return this.tableClient;
  }

  /**
   * 获取记录客户端
   */
  public get records(): RecordClient {
    return this.recordClient;
  }

  /**
   * 获取视图客户端
   */
  public get views(): ViewClient {
    return this.viewClient;
  }

  /**
   * 获取协作客户端
   */
  public get collaboration(): CollaborationClient {
    return this.collaborationClient;
  }
}

// ==================== 导出所有类型和类 ====================

export {
  // 核心类
  HttpClient,
  WebSocketClient,
  
  // 客户端类
  AuthClient,
  SpaceClient,
  TableClient,
  RecordClient,
  ViewClient,
  CollaborationClient,
  
  // 类型定义
  TeableConfig,
  User,
  Space,
  Base,
  Table,
  Field,
  Record,
  View,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  CreateSpaceRequest,
  CreateBaseRequest,
  CreateTableRequest,
  CreateFieldRequest,
  CreateRecordRequest,
  CreateViewRequest,
  PaginatedResponse,
  PaginationParams,
  FilterExpression,
  SortExpression,
  ViewType,
  ViewConfig,
  FieldType,
  FieldOptions,
  CollaborationSession,
  Presence,
  CursorPosition,
  WebSocketMessage,
  CollaborationMessage,
  RecordChangeMessage,
  
  // 错误类
  TeableError,
  AuthenticationError,
  AuthorizationError,
  NotFoundError,
  ValidationError,
  RateLimitError,
  ServerError,
  JsonObject
};

// 默认导出主类
export default Teable;
