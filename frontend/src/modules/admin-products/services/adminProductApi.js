// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient';

const clean = (obj) => Object.fromEntries(Object.entries(obj).filter(([, v]) => v !== '' && v != null));

/** params: { q, page, limit } -> { data: Product[], meta } */
export const getAdminProducts = (params = {}) => request({ method: 'get', url: '/admin/products', params: clean(params) });

/** data: { name, sku, description, price, stock, category_id, image_url } (JSON) */
export async function createProduct(data) {
  const res = await request({ method: 'post', url: '/admin/products', data });
  return res.data;
}

/** data: { name, description, price, stock, category_id, image_url } — SKU tidak bisa diubah. */
export async function updateProduct(id, data) {
  const res = await request({ method: 'put', url: `/admin/products/${id}`, data });
  return res.data;
}

export function deleteProduct(id) {
  return request({ method: 'delete', url: `/admin/products/${id}` });
}

export async function uploadProductImage(file, onUploadProgress) {
  const formData = new FormData();
  formData.append('file', file);
  const res = await request({
    method: 'post',
    url: '/admin/product-images',
    data: formData,
    onUploadProgress,
  });
  return res.data;
}
