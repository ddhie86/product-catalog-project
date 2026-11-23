// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient';

export async function getAdminUsers() {
  const { data } = await request({ method: 'get', url: '/admin/users' });
  return data ?? [];
}

export async function createAdminUser(data) {
  const res = await request({ method: 'post', url: '/admin/users', data });
  return res.data;
}

export async function updateAdminUser(id, data) {
  const res = await request({ method: 'put', url: `/admin/users/${id}`, data });
  return res.data;
}

export function deleteAdminUser(id) {
  return request({ method: 'delete', url: `/admin/users/${id}` });
}
