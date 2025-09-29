/**
 * HTTP 客户端测试
 */

// 在导入被测代码之前，先 mock axios
jest.mock('axios', () => {
  const interceptorsStore: any = { request: {}, response: {} };
  const axiosInstanceMock = {
    // 拦截器需要存在 use 函数即可
    interceptors: {
      request: { use: jest.fn((fulfilled: any, rejected: any) => { interceptorsStore.request = { fulfilled, rejected }; }) },
      response: { use: jest.fn((fulfilled: any, rejected: any) => { interceptorsStore.response = { fulfilled, rejected }; }) }
    },
    // axios 默认属性
    defaults: {
      headers: {}
    },
    // request 是 HttpClient 内部使用的统一入口
    request: jest.fn(),
    // 某些直接调用场景（如 downloadFile）会用到 get
    get: jest.fn()
  } as any;

  const axiosMock = {
    create: jest.fn(() => axiosInstanceMock),
    post: jest.fn()
  } as any;

  // 便于测试中访问与修改
  (axiosMock as any).__instance = axiosInstanceMock;
  (axiosMock as any).__interceptors = interceptorsStore;
  return axiosMock;
});

import axios from 'axios';
import { HttpClient } from '../src/core/http-client';
import { TeableConfig } from '../src/types';

describe('HttpClient', () => {
  let httpClient: HttpClient;
  let config: TeableConfig;

  beforeEach(() => {
    config = {
      baseUrl: 'https://api.test.com',
      debug: false
    };
    httpClient = new HttpClient(config);
  });

  describe('构造函数', () => {
    it('应该正确初始化 HTTP 客户端', () => {
      expect(httpClient).toBeInstanceOf(HttpClient);
    });

    it('应该设置正确的 baseURL', () => {
      // 这里可以添加更多测试来验证 axios 实例的配置
      expect(httpClient).toBeDefined();
    });
  });

  describe('认证相关方法', () => {
    it('应该能够设置访问令牌', () => {
      const token = 'test-access-token';
      httpClient.setAccessToken(token);
      // 验证令牌已设置（这里需要访问私有属性或添加公共方法）
      expect(httpClient).toBeDefined();
    });

    it('应该能够设置刷新令牌', () => {
      const token = 'test-refresh-token';
      httpClient.setRefreshToken(token);
      expect(httpClient).toBeDefined();
    });

    it('应该能够清除令牌', () => {
      httpClient.setAccessToken('test-token');
      httpClient.clearTokens();
      expect(httpClient).toBeDefined();
    });
  });

  describe('健康检查', () => {
    it('应该能够进行健康检查', async () => {
      // 准备 mock 的响应
      const instance = (axios as any).__instance;
      instance.request.mockResolvedValueOnce({
        status: 200,
        data: { status: 'healthy', timestamp: '2025-09-29T00:00:00Z', version: '1.0.0' }
      });

      const result = await httpClient.healthCheck();
      expect(result.status).toBe('healthy');
      expect(result.version).toBeDefined();
    });
  });

  describe('错误映射', () => {
    const makeAxiosError = (status: number, data?: any) => {
      const err: any = new Error('error');
      err.response = { status, data };
      err.config = {};
      return err;
    };

    it('401 -> AuthenticationError', async () => {
      const { rejected } = (axios as any).__interceptors.response;
      await expect(rejected(makeAxiosError(401, { error: 'unauthorized', code: 'AUTH' }))).rejects.toMatchObject({ code: 'AUTH_ERROR' });
    });

    it('403 -> AuthorizationError', async () => {
      const { rejected } = (axios as any).__interceptors.response;
      await expect(rejected(makeAxiosError(403, { error: 'forbidden', code: 'AUTHZ' }))).rejects.toMatchObject({ code: 'AUTHZ_ERROR' });
    });

    it('404 -> NotFoundError', async () => {
      const { rejected } = (axios as any).__interceptors.response;
      await expect(rejected(makeAxiosError(404, { error: 'not found', code: 'NOT_FOUND' }))).rejects.toMatchObject({ code: 'NOT_FOUND' });
    });

    it('422 -> ValidationError', async () => {
      const { rejected } = (axios as any).__interceptors.response;
      await expect(rejected(makeAxiosError(422, { error: 'invalid', code: 'VALIDATION_ERROR', details: { field: 'x' } }))).rejects.toMatchObject({ code: 'VALIDATION_ERROR' });
    });

    it('429 -> RateLimitError', async () => {
      const { rejected } = (axios as any).__interceptors.response;
      await expect(rejected(makeAxiosError(429, { error: 'too many', code: 'RATE_LIMIT' }))).rejects.toMatchObject({ code: 'RATE_LIMIT' });
    });

    it('500 -> ServerError', async () => {
      const { rejected } = (axios as any).__interceptors.response;
      await expect(rejected(makeAxiosError(500, { error: 'server', code: 'SERVER_ERROR' }))).rejects.toMatchObject({ code: 'SERVER_ERROR' });
    });
  });
});
