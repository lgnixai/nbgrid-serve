/**
 * ViewClient 端点对齐测试
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
import { ViewClient } from '../src/clients/view-client';

describe('ViewClient endpoints', () => {
  let http: HttpClient;
  let client: ViewClient;

  beforeEach(() => {
    http = new HttpClient({ baseUrl: 'https://api.test.com' });
    client = new ViewClient(http);
    (axios as any).__instance.request.mockReset();
    (axios as any).__instance.get.mockReset();
  });

  it('createShareLink 应发起 POST /api/view-shares/:view_id/share', async () => {
    (axios as any).__instance.request.mockResolvedValueOnce({ data: { success: true, data: { share_id: 's1', share_url: 'url' } } });
    await client.createShareLink('v_1', { allow_edit: false });
    expect((axios as any).__instance.request).toHaveBeenCalledWith(expect.objectContaining({ method: 'POST', url: '/api/view-shares/v_1/share' }));
  });
});


