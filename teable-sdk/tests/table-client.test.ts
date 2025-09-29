/**
 * TableClient 端点对齐测试
 */

jest.mock('axios', () => {
  const axiosInstanceMock = {
    interceptors: { request: { use: jest.fn() }, response: { use: jest.fn() } },
    defaults: { headers: {} },
    request: jest.fn(),
    get: jest.fn()
  } as any;
  const axiosMock = { create: jest.fn(() => axiosInstanceMock) } as any;
  (axiosMock as any).__instance = axiosInstanceMock;
  return axiosMock;
});

import axios from 'axios';
import { HttpClient } from '../src/core/http-client';
import { TableClient } from '../src/clients/table-client';

describe('TableClient endpoints', () => {
  let http: HttpClient;
  let client: TableClient;

  beforeEach(() => {
    http = new HttpClient({ baseUrl: 'https://api.test.com' });
    client = new TableClient(http);
    (axios as any).__instance.request.mockReset();
    (axios as any).__instance.get.mockReset();
  });

  it('getTableSchema 对齐到 GET /api/tables/:id', async () => {
    (axios as any).__instance.request.mockResolvedValueOnce({ data: { success: true, data: {} } });
    await client.getTableSchema('t1');
    expect((axios as any).__instance.request).toHaveBeenCalledWith(expect.objectContaining({ method: 'GET', url: '/api/tables/t1' }));
  });

  it('updateTableSchema 对齐到 PUT /api/tables/:id', async () => {
    (axios as any).__instance.request.mockResolvedValueOnce({ data: { success: true, data: {} } });
    await client.updateTableSchema('t2', { schema_version: 2 });
    expect((axios as any).__instance.request).toHaveBeenCalledWith(expect.objectContaining({ method: 'PUT', url: '/api/tables/t2' }));
  });
});


