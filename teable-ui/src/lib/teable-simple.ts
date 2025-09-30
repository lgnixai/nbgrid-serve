// 简化的 Teable SDK 包装器，避免复杂的依赖问题
import axios from 'axios';

const BASE_URL = import.meta.env.VITE_TEABLE_BASE_URL || "http://127.0.0.1:3000";

interface LoginRequest {
  email: string;
  password: string;
}

interface AuthResponse {
  user: any;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface Space {
  id: string;
  name: string;
  description?: string;
  status: string;
}

interface Base {
  id: string;
  space_id: string;
  name: string;
  description?: string;
  status: string;
}

interface Table {
  id: string;
  base_id: string;
  name: string;
  description?: string;
}

interface Field {
  id: string;
  table_id: string;
  name: string;
  type: string;
  options?: any;
}

interface RecordItem {
  id: string;
  table_id: string;
  [key: string]: any;
}

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

class SimpleTeableClient {
  private accessToken: string | null = null;
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  // ========== Tables ==========
  async getTable(params: { table_id: string }): Promise<{ data: Table & { fields?: Field[] } }> {
    try {
      const response = await axios.get(`${this.baseURL}/api/tables/${params.table_id}`, {
        headers: this.getHeaders(),
      });
      return { data: response.data.data };
    } catch (error: any) {
      throw new Error(`获取数据表失败: ${error.response?.data?.message || error.message}`);
    }
  }

  // ========== Fields ==========
  async listFields(params: { table_id: string; limit?: number; offset?: number }): Promise<PaginatedResponse<Field>> {
    try {
      const response = await axios.get(`${this.baseURL}/api/fields`, {
        headers: this.getHeaders(),
        params,
      });
      const backendData = response.data.data;
      return {
        data: backendData.list,
        total: backendData.pagination.total,
        limit: backendData.pagination.limit,
        offset: backendData.pagination.page * backendData.pagination.limit,
      };
    } catch (error: any) {
      throw new Error(`获取字段列表失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async createField(body: { table_id: string; name: string; type: string; options?: any }): Promise<{ data: Field }> {
    try {
      const response = await axios.post(`${this.baseURL}/api/fields`, body, { headers: this.getHeaders() });
      return { data: response.data.data };
    } catch (error: any) {
      throw new Error(`创建字段失败: ${error.response?.data?.message || error.message}`);
    }
  }

  // ========== Records ==========
  async listRecords(params: { table_id: string; limit?: number; offset?: number }): Promise<PaginatedResponse<RecordItem>> {
    try {
      const response = await axios.get(`${this.baseURL}/api/records`, {
        headers: this.getHeaders(),
        params,
      });
      const backendData = response.data.data;
      // 后端常用结构 { list, pagination }
      if (backendData?.list && backendData?.pagination) {
        return {
          data: backendData.list,
          total: backendData.pagination.total,
          limit: backendData.pagination.limit,
          offset: backendData.pagination.page * backendData.pagination.limit,
        };
      }
      // 兜底：直接返回 data
      return backendData as PaginatedResponse<RecordItem>;
    } catch (error: any) {
      throw new Error(`获取记录列表失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async createRecord(body: { table_id: string; fields: Record<string, any> }): Promise<{ data: RecordItem }> {
    try {
      const response = await axios.post(`${this.baseURL}/api/records`, body, { headers: this.getHeaders() });
      return { data: response.data.data };
    } catch (error: any) {
      throw new Error(`创建记录失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async updateRecord(body: { table_id: string; record_id: string; fields: Record<string, any> }): Promise<{ data: RecordItem }> {
    try {
      const response = await axios.put(`${this.baseURL}/api/records/${body.record_id}`, body, { headers: this.getHeaders() });
      return { data: response.data.data };
    } catch (error: any) {
      throw new Error(`更新记录失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async deleteRecord(body: { table_id: string; record_id: string }): Promise<void> {
    try {
      await axios.delete(`${this.baseURL}/api/records/${body.record_id}`, { headers: this.getHeaders() });
    } catch (error: any) {
      throw new Error(`删除记录失败: ${error.response?.data?.message || error.message}`);
    }
  }

  private getHeaders() {
    const headers: any = {
      'Content-Type': 'application/json',
    };
    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }
    return headers;
  }

  async login(credentials: LoginRequest): Promise<AuthResponse> {
    try {
      const response = await axios.post(`${this.baseURL}/api/auth/login`, credentials);
      // 后端返回的数据结构是 { code, data: { user, access_token, ... } }
      const authData = response.data.data;
      this.accessToken = authData.access_token;
      return authData;
    } catch (error: any) {
      throw new Error(`登录失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async listSpaces(params?: { limit?: number; offset?: number }): Promise<PaginatedResponse<Space>> {
    try {
      const response = await axios.get(`${this.baseURL}/api/spaces`, {
        headers: this.getHeaders(),
        params: { limit: 50, ...params }
      });
      // 后端返回的数据结构是 { code, data: { list, pagination } }
      const backendData = response.data.data;
      return {
        data: backendData.list,
        total: backendData.pagination.total,
        limit: backendData.pagination.limit,
        offset: backendData.pagination.page * backendData.pagination.limit
      };
    } catch (error: any) {
      throw new Error(`获取空间列表失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async listBases(params?: { limit?: number; offset?: number; space_id?: string }): Promise<PaginatedResponse<Base>> {
    try {
      const response = await axios.get(`${this.baseURL}/api/bases`, {
        headers: this.getHeaders(),
        params: { limit: 100, ...params }
      });
      // 后端返回的数据结构是 { code, data: { list, pagination } }
      const backendData = response.data.data;
      return {
        data: backendData.list,
        total: backendData.pagination.total,
        limit: backendData.pagination.limit,
        offset: backendData.pagination.page * backendData.pagination.limit
      };
    } catch (error: any) {
      throw new Error(`获取基础表列表失败: ${error.response?.data?.message || error.message}`);
    }
  }

  async listTables(params?: { limit?: number; offset?: number; base_id?: string }): Promise<PaginatedResponse<Table>> {
    try {
      const response = await axios.get(`${this.baseURL}/api/tables`, {
        headers: this.getHeaders(),
        params: { limit: 200, ...params }
      });
      // 后端返回的数据结构是 { code, data: { list, pagination } }
      const backendData = response.data.data;
      return {
        data: backendData.list,
        total: backendData.pagination.total,
        limit: backendData.pagination.limit,
        offset: backendData.pagination.page * backendData.pagination.limit
      };
    } catch (error: any) {
      throw new Error(`获取数据表列表失败: ${error.response?.data?.message || error.message}`);
    }
  }

  isAuthenticated(): boolean {
    return !!this.accessToken;
  }

  async logout(): Promise<void> {
    this.accessToken = null;
  }
}

const teable = new SimpleTeableClient(BASE_URL);

let loginPromise: Promise<void> | null = null;

export const ensureLogin = (creds?: LoginRequest): Promise<void> => {
  if (teable.isAuthenticated()) return Promise.resolve();
  if (loginPromise) return loginPromise;

  const credentials: LoginRequest = creds ?? {
    email: "test@example.com",
    password: "TestPassword123!",
  };

  loginPromise = teable
    .login(credentials)
    .then(() => {})
    .finally(() => {
      loginPromise = null;
    });

  return loginPromise;
};

export default teable;
