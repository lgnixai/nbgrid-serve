/**
 * RecordClient 端点对齐测试
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
import { RecordClient } from '../src/clients/record-client';

describe('RecordClient endpoints', () => {
  let http: HttpClient;
  let client: RecordClient;

  beforeEach(() => {
    http = new HttpClient({ baseUrl: 'https://api.test.com' });
    client = new RecordClient(http);
    (axios as any).__instance.request.mockReset();
    (axios as any).__instance.get.mockReset();
  });

  it('listTableRecords 应发起 GET /api/records?table_id=...', async () => {
    (axios as any).__instance.request.mockResolvedValueOnce({ data: { success: true, data: { data: [], total: 0, limit: 20, offset: 0 } } });
    await client.listTableRecords('tbl_1', { limit: 10 });
    expect((axios as any).__instance.request).toHaveBeenCalledWith(expect.objectContaining({ method: 'GET', url: '/api/records', params: expect.objectContaining({ table_id: 'tbl_1', limit: 10 }) }));
  });

  it('getStats 应发起 GET /api/records/stats?table_id=...', async () => {
    (axios as any).__instance.request.mockResolvedValueOnce({ data: { success: true, data: { total_records: 0, created_today: 0, updated_today: 0, by_field: {}, last_activity_at: '' } } });
    await client.getStats('tbl_2');
    expect((axios as any).__instance.request).toHaveBeenCalledWith(expect.objectContaining({ method: 'GET', url: '/api/records/stats', params: expect.objectContaining({ table_id: 'tbl_2' }) }));
  });

  it('exportRecords 应发起 GET /api/records/export 并携带 table_id', async () => {
    (axios as any).__instance.get.mockResolvedValueOnce({ data: new Blob() });
    await client.exportRecords('tbl_3', 'json');
    expect((axios as any).__instance.get).toHaveBeenCalledWith('/api/records/export', expect.objectContaining({ params: expect.objectContaining({ table_id: 'tbl_3', format: 'json' }) }));
  });
});


