// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { PageHeader } from '@/components/PageHeader.jsx';
import { Pagination } from '@/components/Pagination.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { formatDateTime } from '@/lib/format';
import { fetchActivityLogs } from './services/activityLogApi.js';

const ACTION_LABELS = {
  CREATE: 'Tambah',
  UPDATE: 'Ubah',
  DELETE: 'Hapus',
  REPLY: 'Balas pesan',
};

const ENTITY_LABELS = {
  products: 'Produk',
  categories: 'Kategori',
  users: 'Pengguna admin',
  roles: 'Role',
  contact_messages: 'Pesan kontak',
};

function Snapshot({ title, data }) {
  return (
    <div className="min-w-0">
      <p className="mb-1 text-xs font-semibold text-ink-soft">{title}</p>
      <pre className="max-h-64 overflow-auto rounded-md bg-paper p-3 text-xs leading-5 text-ink">
        {data ? JSON.stringify(data, null, 2) : '—'}
      </pre>
    </div>
  );
}

export default function ActivityLogPage() {
  const [page, setPage] = useState(1);
  const logs = useQuery({
    queryKey: ['admin-activity-logs', page],
    queryFn: () => fetchActivityLogs({ page, limit: 10 }),
    placeholderData: keepPreviousData,
  });
  const items = logs.data?.data ?? [];

  return (
    <>
      <PageHeader title="Log aktivitas" description="Riwayat perubahan yang dilakukan melalui dashboard admin." />

      {logs.isPending && <LoadingState label="Memuat log aktivitas" />}
      {logs.isError && <ErrorState error={logs.error} onRetry={logs.refetch} />}
      {logs.isSuccess && items.length === 0 && (
        <EmptyState title="Belum ada aktivitas" description="Perubahan produk, kategori, role, pengguna, dan balasan pesan akan tercatat di sini." />
      )}

      {items.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-line bg-white">
          <table className="w-full min-w-[760px] text-left text-sm">
            <caption className="sr-only">Riwayat aktivitas admin</caption>
            <thead className="border-b border-line bg-paper text-ink-soft">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Waktu</th>
                <th scope="col" className="px-4 py-3 font-medium">Admin</th>
                <th scope="col" className="px-4 py-3 font-medium">Aktivitas</th>
                <th scope="col" className="px-4 py-3 font-medium">Perubahan data</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {items.map((item) => (
                <tr key={item.id} className="align-top">
                  <td className="whitespace-nowrap px-4 py-3 text-ink-soft">{formatDateTime(item.datetime)}</td>
                  <td className="px-4 py-3">{item.user}</td>
                  <td className="px-4 py-3">
                    <span className="font-medium">{ACTION_LABELS[item.action] ?? item.action}</span>
                    <span className="mt-1 block text-xs text-ink-soft">{ENTITY_LABELS[item.entity] ?? item.entity}</span>
                  </td>
                  <td className="px-4 py-3">
                    <details>
                      <summary className="cursor-pointer font-medium text-brand">Lihat detail</summary>
                      <div className="mt-3 grid gap-3 lg:grid-cols-2">
                        <Snapshot title="Sebelum" data={item.data_before_updated} />
                        <Snapshot title="Sesudah" data={item.data_after_updated} />
                      </div>
                    </details>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <Pagination meta={logs.data?.meta} onPageChange={setPage} />
    </>
  );
}
