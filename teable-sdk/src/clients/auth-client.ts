/**
 * 认证客户端
 * 处理用户登录、注册、token 管理等功能
 */

import { HttpClient } from '../core/http-client';
import { 
  User, 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  UserPreferences
} from '../types';

export class AuthClient {
  private httpClient: HttpClient;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
  }

  /**
   * 用户登录
   */
  public async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.httpClient.post<AuthResponse>('/api/auth/login', credentials);
    
    // 自动设置 token
    this.httpClient.setAccessToken(response.access_token);
    this.httpClient.setRefreshToken(response.refresh_token);
    
    return response;
  }

  /**
   * 用户注册
   */
  public async register(userData: RegisterRequest): Promise<AuthResponse> {
    const response = await this.httpClient.post<AuthResponse>('/api/auth/register', userData);
    
    // 自动设置 token
    this.httpClient.setAccessToken(response.access_token);
    this.httpClient.setRefreshToken(response.refresh_token);
    
    return response;
  }

  /**
   * 刷新访问令牌
   */
  public async refreshToken(): Promise<{ access_token: string; refresh_token: string }> {
    return this.httpClient.post('/api/auth/refresh');
  }

  /**
   * 用户登出
   */
  public async logout(): Promise<void> {
    await this.httpClient.post('/api/auth/logout');
    
    // 清除本地 token
    this.httpClient.clearTokens();
  }

  /**
   * 获取当前用户信息
   */
  public async getCurrentUser(): Promise<User> {
    return this.httpClient.get<User>('/api/users/profile');
  }

  /**
   * 更新用户资料
   */
  public async updateProfile(updates: Partial<User>): Promise<User> {
    return this.httpClient.put<User>('/api/users/profile', updates);
  }

  /**
   * 修改密码
   */
  public async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await this.httpClient.post('/api/users/change-password', {
      current_password: currentPassword,
      new_password: newPassword
    });
  }

  /**
   * 获取用户活动记录
   */
  public async getUserActivity(userId: string, limit?: number, offset?: number): Promise<any[]> {
    return this.httpClient.get(`/api/users/${userId}/activity`, { limit, offset });
  }

  /**
   * 获取用户偏好设置
   */
  public async getUserPreferences(): Promise<UserPreferences> {
    return this.httpClient.get<UserPreferences>('/api/users/preferences');
  }

  /**
   * 更新用户偏好设置
   */
  public async updateUserPreferences(preferences: Partial<UserPreferences>): Promise<UserPreferences> {
    return this.httpClient.put<UserPreferences>('/api/users/preferences', preferences);
  }

  /**
   * 检查是否已登录
   */
  public isAuthenticated(): boolean {
    // 这里可以添加更复杂的 token 验证逻辑
    return !!(this.httpClient as any)['accessToken'];
  }

  /**
   * 获取当前访问令牌
   */
  public getAccessToken(): string | undefined {
    return (this.httpClient as any)['accessToken'];
  }

  /**
   * 获取当前刷新令牌
   */
  public getRefreshToken(): string | undefined {
    return (this.httpClient as any)['refreshToken'];
  }
}
