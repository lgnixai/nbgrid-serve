/**
 * Teable SDK 核心类型定义
 * 基于 Go 后端的数据模型设计
 */

// ==================== 基础类型 ====================

export interface BaseEntity {
  id: string;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by?: string;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data: T;
  message?: string;
}

export interface ApiError {
  error: string;
  code: string;
  details?: string;
  trace_id?: string;
}

// ==================== 用户相关类型 ====================

export interface User extends BaseEntity {
  name: string;
  email: string;
  phone?: string;
  avatar?: string;
  is_system: boolean;
  is_admin: boolean;
  is_trial_used: boolean;
  notify_meta?: string;
  last_sign_time?: string;
  deactivated_time?: string;
  permanent_deleted_time?: string;
  ref_meta?: string;
}

export interface UserPreferences {
  theme?: 'light' | 'dark' | 'auto';
  language?: string;
  timezone?: string;
  notifications?: NotificationSettings;
}

export interface NotificationSettings {
  email_notifications: boolean;
  push_notifications: boolean;
  collaboration_notifications: boolean;
  system_notifications: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type?: string;
  session_id?: string;
}

// ==================== 空间相关类型 ====================

export type SpaceStatus = 'active' | 'archived' | 'deleted';

export interface Space extends BaseEntity {
  name: string;
  description?: string;
  icon?: string;
  status: SpaceStatus;
  member_count: number;
}

export type CollaboratorRole = 'owner' | 'admin' | 'editor' | 'viewer';

export type CollaboratorStatus = 'pending' | 'accepted' | 'rejected' | 'revoked';

export interface SpaceCollaborator extends BaseEntity {
  space_id: string;
  user_id: string;
  role: CollaboratorRole;
  invited_by: string;
  accepted_at?: string;
  revoked_at?: string;
  status: CollaboratorStatus;
}

export interface CreateSpaceRequest {
  name: string;
  description?: string;
  icon?: string;
}

export interface UpdateSpaceRequest {
  name?: string;
  description?: string;
  icon?: string;
}

export interface AddCollaboratorRequest {
  user_id: string;
  role: CollaboratorRole;
}

// ==================== 基础表相关类型 ====================

export type BaseStatus = 'active' | 'archived' | 'deleted';

export interface Base extends BaseEntity {
  space_id: string;
  name: string;
  description?: string;
  icon?: string;
  is_system: boolean;
  status: BaseStatus;
  table_count: number;
}

export interface CreateBaseRequest {
  space_id: string;
  name: string;
  description?: string;
  icon?: string;
}

export interface UpdateBaseRequest {
  name?: string;
  description?: string;
  icon?: string;
}

// ==================== 数据表相关类型 ====================

export interface Table extends BaseEntity {
  base_id: string;
  name: string;
  description?: string;
  icon?: string;
  is_system: boolean;
  schema_version: number;
  fields?: Field[];
}

export interface CreateTableRequest {
  base_id: string;
  name: string;
  description?: string;
  icon?: string;
}

export interface UpdateTableRequest {
  name?: string;
  description?: string;
  icon?: string;
}

// ==================== 字段相关类型 ====================

export type FieldType = 
  | 'text'
  | 'number'
  | 'single_select'
  | 'multi_select'
  | 'date'
  | 'time'
  | 'datetime'
  | 'checkbox'
  | 'url'
  | 'email'
  | 'phone'
  | 'currency'
  | 'percent'
  | 'duration'
  | 'rating'
  | 'slider'
  | 'long_text'
  | 'attachment'
  | 'link'
  | 'lookup'
  | 'formula'
  | 'rollup'
  | 'count'
  | 'created_time'
  | 'last_modified_time'
  | 'created_by'
  | 'last_modified_by'
  | 'auto_number';

export interface FieldOptions {
  placeholder?: string;
  help_text?: string;
  choices?: SelectOption[];
  min_value?: number;
  max_value?: number;
  decimal?: number;
  min_length?: number;
  max_length?: number;
  pattern?: string;
  date_format?: string;
  time_format?: string;
  max_file_size?: number;
  allowed_types?: string[];
  link_table_id?: string;
  link_field_id?: string;
  formula?: string;
  validation_rules?: ValidationRule[];
}

export interface SelectOption {
  id: string;
  name: string;
  color?: string;
}

export interface ValidationRule {
  type: string;
  value: any;
  message?: string;
}

export interface Field extends BaseEntity {
  table_id: string;
  name: string;
  type: FieldType;
  description?: string;
  required: boolean;
  is_unique: boolean;
  is_primary: boolean;
  is_computed: boolean;
  is_lookup: boolean;
  default_value?: string;
  options?: FieldOptions;
  field_order: number;
  version: number;
}

export interface CreateFieldRequest {
  table_id: string;
  name: string;
  type: FieldType;
  description?: string;
  required?: boolean;
  is_unique?: boolean;
  is_primary?: boolean;
  default_value?: string;
  options?: FieldOptions;
  field_order?: number;
}

export interface UpdateFieldRequest {
  name?: string;
  type?: FieldType;
  description?: string;
  required?: boolean;
  is_unique?: boolean;
  is_primary?: boolean;
  default_value?: string;
  options?: FieldOptions;
  field_order?: number;
}

// ==================== 记录相关类型 ====================

export interface Record extends BaseEntity {
  table_id: string;
  data: JsonObject;
  version: number;
  hash: string;
}

export interface CreateRecordRequest {
  table_id: string;
  data: JsonObject;
}

export interface UpdateRecordRequest {
  data: JsonObject;
}

export interface BulkCreateRecordRequest {
  table_id: string;
  records: JsonObject[];
}

export interface BulkUpdateRecordRequest {
  records: Array<{
    id: string;
    data: JsonObject;
  }>;
}

export interface BulkDeleteRecordRequest {
  record_ids: string[];
}

export interface RecordQuery {
  table_id: string;
  filter?: FilterExpression | undefined;
  sort?: SortExpression[] | undefined;
  limit?: number | undefined;
  offset?: number | undefined;
}

export interface FilterExpression {
  field: string;
  operator: FilterOperator;
  value: any;
  logic?: 'and' | 'or';
  conditions?: FilterExpression[];
}

export type FilterOperator = 
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'ends_with'
  | 'greater_than'
  | 'greater_than_or_equal'
  | 'less_than'
  | 'less_than_or_equal'
  | 'is_empty'
  | 'is_not_empty'
  | 'in'
  | 'not_in'
  | 'has_any_of'
  | 'has_all_of';

export interface SortExpression {
  field: string;
  direction: 'asc' | 'desc';
}

// ==================== 视图相关类型 ====================

export type ViewType = 'grid' | 'form' | 'kanban' | 'calendar' | 'gallery';

export interface View extends BaseEntity {
  table_id: string;
  name: string;
  type: ViewType;
  description?: string;
  config: ViewConfig;
  is_default: boolean;
}

export interface ViewConfig {
  // 通用配置
  filter?: FilterExpression;
  sort?: SortExpression[];
  
  // 网格视图配置
  grid?: GridViewConfig;
  
  // 表单视图配置
  form?: FormViewConfig;
  
  // 看板视图配置
  kanban?: KanbanViewConfig;
  
  // 日历视图配置
  calendar?: CalendarViewConfig;
  
  // 画廊视图配置
  gallery?: GalleryViewConfig;
}

export interface GridViewConfig {
  columns: GridColumn[];
  row_height?: 'short' | 'medium' | 'tall' | undefined;
  show_row_numbers?: boolean | undefined;
  show_column_headers?: boolean | undefined;
}

export interface GridColumn {
  field_id: string;
  width?: number;
  visible?: boolean;
  frozen?: boolean;
}

export interface FormViewConfig {
  fields: FormField[];
  submit_button_text?: string;
  success_message?: string;
  redirect_url?: string;
}

export interface FormField {
  field_id: string;
  required?: boolean | undefined;
  visible?: boolean | undefined;
  order?: number | undefined;
}

export interface KanbanViewConfig {
  group_field_id: string;
  card_fields: string[];
  card_height?: 'short' | 'medium' | 'tall' | undefined;
  show_empty_groups?: boolean | undefined;
}

export interface CalendarViewConfig {
  date_field_id: string;
  title_field_id: string;
  color_field_id?: string | undefined;
  show_weekends?: boolean | undefined;
  start_day_of_week?: number | undefined; // 0-6, 0=Sunday
}

export interface GalleryViewConfig {
  card_fields: string[];
  card_size?: 'small' | 'medium' | 'large' | undefined;
  show_field_names?: boolean | undefined;
  cover_field_id?: string | undefined;
}

export interface CreateViewRequest {
  table_id: string;
  name: string;
  type: ViewType;
  description?: string;
  config?: ViewConfig;
  is_default?: boolean;
}

export interface UpdateViewRequest {
  name?: string;
  description?: string;
  config?: ViewConfig;
  is_default?: boolean;
}

// ==================== 协作相关类型 ====================

export interface CollaborationSession extends BaseEntity {
  name: string;
  description?: string;
  participants: CollaborationParticipant[];
  is_active: boolean;
}

export interface CollaborationParticipant {
  user_id: string;
  role: 'owner' | 'admin' | 'editor' | 'viewer';
  joined_at: string;
  last_activity_at?: string;
}

export interface Presence {
  user_id: string;
  resource_type: 'table' | 'view' | 'record';
  resource_id: string;
  cursor_position?: CursorPosition;
  last_seen: string;
}

export interface CursorPosition {
  x: number;
  y: number;
  field_id?: string;
  record_id?: string;
}

// ==================== 搜索相关类型 ====================

export interface SearchRequest {
  query: string;
  scope?: 'all' | 'spaces' | 'bases' | 'tables' | 'records';
  filters?: SearchFilter[];
  limit?: number;
  offset?: number;
}

export interface SearchFilter {
  field: string;
  operator: FilterOperator;
  value: any;
}

export interface SearchResult<T = any> {
  item: T;
  score: number;
  highlights: string[];
}

export interface SearchResponse<T = any> {
  results: SearchResult<T>[];
  total: number;
  query: string;
  took: number;
}

// ==================== 通知相关类型 ====================

export interface Notification extends BaseEntity {
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  data?: JsonObject;
  is_read: boolean;
  read_at?: string;
  expires_at?: string;
}

export type NotificationType = 
  | 'collaboration_invite'
  | 'collaboration_join'
  | 'collaboration_leave'
  | 'record_created'
  | 'record_updated'
  | 'record_deleted'
  | 'comment_added'
  | 'mention'
  | 'system_announcement';

export interface NotificationSubscription extends BaseEntity {
  user_id: string;
  type: NotificationType;
  resource_type: string;
  resource_id: string;
  is_enabled: boolean;
  channels: NotificationChannel[];
}

export type NotificationChannel = 'email' | 'push' | 'in_app';

// ==================== 附件相关类型 ====================

export interface Attachment extends BaseEntity {
  filename: string;
  original_name: string;
  mime_type: string;
  size: number;
  url: string;
  thumbnail_url?: string;
  metadata?: JsonObject;
}

export interface UploadAttachmentRequest {
  file: File | any;
  filename?: string;
  metadata?: JsonObject;
}

// ==================== WebSocket 相关类型 ====================

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
  user_id?: string;
}

export interface CollaborationMessage extends WebSocketMessage {
  type: 'collaboration';
  data: {
    action: 'cursor_move' | 'selection_change' | 'presence_update';
    resource_type: string;
    resource_id: string;
    payload: any;
  };
}

export interface RecordChangeMessage extends WebSocketMessage {
  type: 'record_change';
  data: {
    action: 'create' | 'update' | 'delete';
    table_id: string;
    record_id: string;
    changes?: JsonObject;
  };
}

// ==================== 配置相关类型 ====================

export interface TeableConfig {
  baseUrl: string;
  apiKey?: string;
  accessToken?: string;
  refreshToken?: string;
  timeout?: number;
  retries?: number;
  retryDelay?: number;
  userAgent?: string;
  debug?: boolean;
  disableProxy?: boolean;
}

export type JsonObject = { [key: string]: any };

export interface RequestOptions {
  timeout?: number;
  retries?: number;
  retryDelay?: number;
  headers?: globalThis.Record<string, string>;
}

// ==================== 错误相关类型 ====================

export class TeableError extends Error {
  public readonly code: string;
  public readonly status: number | undefined;
  public readonly details: any | undefined;

  constructor(message: string, code: string, status?: number, details?: any) {
    super(message);
    this.name = 'TeableError';
    this.code = code;
    this.status = status;
    this.details = details;
  }
}

export class AuthenticationError extends TeableError {
  constructor(message: string = 'Authentication failed') {
    super(message, 'AUTH_ERROR', 401);
  }
}

export class AuthorizationError extends TeableError {
  constructor(message: string = 'Insufficient permissions') {
    super(message, 'AUTHZ_ERROR', 403);
  }
}

export class NotFoundError extends TeableError {
  constructor(message: string = 'Resource not found') {
    super(message, 'NOT_FOUND', 404);
  }
}

export class ValidationError extends TeableError {
  constructor(message: string, details?: any) {
    super(message, 'VALIDATION_ERROR', 422, details);
  }
}

export class RateLimitError extends TeableError {
  constructor(message: string = 'Rate limit exceeded') {
    super(message, 'RATE_LIMIT', 429);
  }
}

export class ServerError extends TeableError {
  constructor(message: string = 'Internal server error') {
    super(message, 'SERVER_ERROR', 500);
  }
}

export interface TableStats {
  table_id: string;
  total_fields: number;
  total_records: number;
  total_views: number;
  last_activity_at?: string | undefined;
}
