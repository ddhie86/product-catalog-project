import { request } from '@/services/apiClient';

export async function getAdminCategories() {
  const { data } = await request({ method: 'get', url: '/admin/categories' });
  return data ?? [];
}

export async function createCategory(data) {
  const res = await request({ method: 'post', url: '/admin/categories', data });
  return res.data;
}

export async function updateCategory(id, data) {
  const res = await request({ method: 'put', url: `/admin/categories/${id}`, data });
  return res.data;
}

export function deleteCategory(id) {
  return request({ method: 'delete', url: `/admin/categories/${id}` });
}
