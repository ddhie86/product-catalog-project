// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient';

const clean = (obj) => Object.fromEntries(Object.entries(obj).filter(([, v]) => v !== '' && v != null));

/** params: { q, category, sort, page, limit } -> { data: Product[], meta } */
export function fetchCatalogProducts(params = {}) {
  return request({ method: 'get', url: '/products', params: clean(params) });
}

export async function fetchProductById(id) {
  const { data } = await request({ method: 'get', url: `/products/${id}` });
  return data;
}

export async function fetchCategories() {
  const { data } = await request({ method: 'get', url: '/categories' });
  return data ?? [];
}
