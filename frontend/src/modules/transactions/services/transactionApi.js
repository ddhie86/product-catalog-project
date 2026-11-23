// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient';

const clean = (obj) => Object.fromEntries(Object.entries(obj).filter(([, v]) => v !== '' && v != null));

/** params: { status, page, limit } -> { data: Transaction[], meta } */
export const fetchAdminTransactions = (params = {}) => request({ method: 'get', url: '/admin/transactions', params: clean(params) });

export async function fetchTransactionById(id) {
  const { data } = await request({ method: 'get', url: `/admin/transactions/${id}` });
  return data; // termasuk items[]
}

/** -> { total_revenue, total_paid_transactions, pending_count, low_stock_alerts } */
export async function fetchDashboardSummary() {
  const { data } = await request({ method: 'get', url: '/admin/reports/summary' });
  return data;
}
