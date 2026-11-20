import { request } from '@/services/apiClient';

export async function getAdminRoles() {
  const { data } = await request({ method: 'get', url: '/admin/roles' });
  return data ?? { roles: [], permissions: [] };
}

export async function createRole(data) {
  const res = await request({ method: 'post', url: '/admin/roles', data });
  return res.data;
}

export async function updateRole(id, data) {
  const res = await request({ method: 'put', url: `/admin/roles/${id}`, data });
  return res.data;
}

export function deleteRole(id) {
  return request({ method: 'delete', url: `/admin/roles/${id}` });
}
