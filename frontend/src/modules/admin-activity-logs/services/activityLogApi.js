// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient.js';

export async function fetchActivityLogs(params = {}) {
  const { data, meta } = await request({ method: 'get', url: '/admin/activity-logs', params });
  return { data, meta };
}
