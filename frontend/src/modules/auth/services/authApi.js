// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient';

export async function login({ email, password }) {
  const { data } = await request({ method: 'post', url: '/admin/login', data: { email, password } });
  return data; // { access_token, expires_at }
}

export async function getCurrentPermissions() {
  const { data } = await request({ method: 'get', url: '/admin/permissions' });
  return data ?? [];
}
