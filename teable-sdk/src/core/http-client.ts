/**
 * HTTP 客户端核心实现
 * 提供统一的 HTTP 请求处理、错误处理、重试机制等
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { 
  TeableConfig, 
  RequestOptions, 
  ApiResponse, 
  // ApiError,
  TeableError,
  AuthenticationError,
  AuthorizationError,
  NotFoundError,
  ValidationError,
  RateLimitError,
  ServerError
} from '../types';

export class HttpClient {
  private axiosInstance: AxiosInstance;
  private config: TeableConfig;
  private accessToken: string | undefined;
  private refreshToken: string | undefined;

  constructor(config: TeableConfig) {
    this.config = config;
    this.accessToken = config.accessToken;
    this.refreshToken = config.refreshToken;
    
    const baseConfig: any = {
      baseURL: config.baseUrl,
      timeout: config.timeout || 30000,
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': config.userAgent || 'Teable-SDK/1.0.0',
        ...(config.apiKey && { 'X-API-Key': config.apiKey }),
        ...(this.accessToken && { 'Authorization': `Bearer ${this.accessToken}` })
      }
    };

    // 可选禁用代理，避免本机代理劫持 localhost
    if (config.disableProxy) {
      baseConfig.proxy = false;
      // axios 在 Node 环境也会读取 HTTP(S)_PROXY 环境变量；这里显式禁用
      baseConfig.transport = undefined;
    }

    this.axiosInstance = axios.create(baseConfig);

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    // 请求拦截器
    this.axiosInstance.interceptors.request.use(
      (config: any) => {
        if (this.config.debug) {
          console.log(`[Teable SDK] ${config.method?.toUpperCase()} ${config.url}`, {
            params: config.params,
            data: config.data
          });
        }
        return config;
      },
      (error: any) => {
        if (this.config.debug) {
          console.error('[Teable SDK] Request error:', error);
        }
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.axiosInstance.interceptors.response.use(
      (response: any) => {
        if (this.config.debug) {
          console.log(`[Teable SDK] Response:`, {
            status: response.status,
            data: response.data
          });
        }
        return response;
      },
      async (error: AxiosError) => {
        if (this.config.debug) {
          console.error('[Teable SDK] Response error:', error);
        }

        // 处理 401 错误，尝试刷新 token
        if (error.response?.status === 401 && this.refreshToken) {
          try {
            await this.refreshAccessToken();
            // 重试原始请求
            const originalRequest = error.config;
            if (originalRequest) {
              originalRequest.headers = originalRequest.headers || {};
              originalRequest.headers['Authorization'] = `Bearer ${this.accessToken}`;
              return this.axiosInstance.request(originalRequest);
            }
        } catch (refreshError) {
          // 刷新失败，清除 token
          this.clearTokensInternal();
          throw new AuthenticationError('Token refresh failed');
        }
        }

        throw this.handleError(error);
      }
    );
  }

  private handleError(error: AxiosError): TeableError {
    const response = error.response;
    const status = response?.status;
    const data = response?.data as any;

    if (!status) {
      return new TeableError(
        error.message || 'Network error',
        'NETWORK_ERROR',
        undefined,
        error
      );
    }

    const message = (data?.error || data?.message || error.message || 'Unknown error');
    const code = (data?.code !== undefined ? String(data.code) : 'UNKNOWN_ERROR');

    switch (status) {
      case 401:
        return new AuthenticationError(message);
      case 403:
        return new AuthorizationError(message);
      case 404:
        return new NotFoundError(message);
      case 422:
        return new ValidationError(message, data?.details);
      case 429:
        return new RateLimitError(message);
      case 500:
      case 502:
      case 503:
      case 504:
        return new ServerError(message);
      default:
        return new TeableError(message, code, status, data);
    }
  }

  private async refreshAccessToken(): Promise<void> {
    if (!this.refreshToken) {
      throw new AuthenticationError('No refresh token available');
    }

    try {
      const response = await axios.post(`${this.config.baseUrl}/api/auth/refresh`, {
        refresh_token: this.refreshToken
      });

      const data = response.data as ApiResponse<{ access_token: string; refresh_token: string }>;
      this.accessToken = data.data.access_token;
      this.refreshToken = data.data.refresh_token;

      // 更新默认请求头
      this.axiosInstance.defaults.headers['Authorization'] = `Bearer ${this.accessToken}`;
    } catch (error) {
      this.clearTokensInternal();
      throw new AuthenticationError('Failed to refresh access token');
    }
  }

  private clearTokensInternal(): void {
    this.accessToken = undefined;
    this.refreshToken = undefined;
    delete this.axiosInstance.defaults.headers['Authorization'];
  }

  public setAccessToken(token: string): void {
    this.accessToken = token;
    this.axiosInstance.defaults.headers['Authorization'] = `Bearer ${token}`;
  }

  public setRefreshToken(token: string): void {
    this.refreshToken = token;
  }

  public clearTokens(): void {
    this.clearTokensInternal();
  }

  public async get<T = any>(
    url: string, 
    params?: globalThis.Record<string, any>, 
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>('GET', url, { params, ...options });
  }

  public async post<T = any>(
    url: string, 
    data?: any, 
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>('POST', url, { data, ...options });
  }

  public async put<T = any>(
    url: string, 
    data?: any, 
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>('PUT', url, { data, ...options });
  }

  public async patch<T = any>(
    url: string, 
    data?: any, 
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>('PATCH', url, { data, ...options });
  }

  public async delete<T = any>(
    url: string, 
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>('DELETE', url, options);
  }

  private async request<T = any>(
    method: string,
    url: string,
    config: AxiosRequestConfig & RequestOptions = {}
  ): Promise<T> {
    const { retries = this.config.retries || 0, retryDelay = this.config.retryDelay || 1000, ...axiosConfig } = config;

    let lastError: Error;

    for (let attempt = 0; attempt <= retries; attempt++) {
      try {
        const response: AxiosResponse<any> = await this.axiosInstance.request({
          method,
          url,
          ...axiosConfig
        });

        // 检查响应格式
        if (response.data && typeof response.data === 'object') {
          if ('success' in response.data) {
            // 标准 API 响应格式
            if (response.data.success) {
              return response.data.data;
            } else {
              throw new TeableError(
                response.data.message || 'API request failed',
                'API_ERROR',
                response.status,
                response.data
              );
            }
          } else if ('code' in response.data && 'data' in response.data) {
            // 后端统一返回格式 { code, data, ... }
            return response.data.data as T;
          } else if ('data' in response.data) {
            // 分页响应格式
            return response.data as T;
          } else {
            // 直接返回数据
            return response.data as T;
          }
        }

        return response.data as T;
      } catch (error) {
        lastError = error as Error;

        // 如果是最后一次尝试，或者错误不应该重试，直接抛出
        if (attempt === retries || !this.shouldRetry(error as TeableError)) {
          throw error;
        }

        // 等待后重试
        if (retryDelay > 0) {
          await this.delay(retryDelay * Math.pow(2, attempt)); // 指数退避
        }
      }
    }

    throw lastError!;
  }

  private shouldRetry(error: TeableError): boolean {
    // 网络错误和服务器错误可以重试
    if (error.code === 'NETWORK_ERROR') {
      return true;
    }

    if (error.status && error.status >= 500) {
      return true;
    }

    // 429 错误可以重试
    if (error.status === 429) {
      return true;
    }

    return false;
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // 文件上传方法
  public async uploadFile<T = any>(
    url: string,
    file: File | any,
    fieldName: string = 'file',
    additionalData?: globalThis.Record<string, any>,
    options?: RequestOptions
  ): Promise<T> {
    const formData = new FormData();
    
    if (file instanceof File) {
      formData.append(fieldName, file);
    } else {
      formData.append(fieldName, new Blob([file]));
    }

    if (additionalData) {
      Object.entries(additionalData).forEach(([key, value]) => {
        formData.append(key, String(value));
      });
    }

    return this.request<T>('POST', url, {
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      ...options
    });
  }

  // 流式下载方法
  public async downloadFile(
    url: string,
    options?: RequestOptions
  ): Promise<Blob> {
    const response = await this.axiosInstance.get(url, {
      responseType: 'blob',
      ...options
    });

    return response.data;
  }

  // 健康检查
  public async healthCheck(): Promise<{ status: string; timestamp: string; version: string }> {
    return this.get('/health');
  }

  // 获取系统信息
  public async getSystemInfo(): Promise<any> {
    return this.get('/api/info');
  }
}
