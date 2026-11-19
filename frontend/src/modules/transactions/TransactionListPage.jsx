import { useState } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { fetchAdminTransactions } from './services/transactionApi';
import { PageHeader } from '@/components/PageHeader.jsx';
import { Pagination } from '@/components/Pagination.jsx';
import { StatusBadge } from '@/components/StatusBadge.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { TRANSACTION_STATUSES } from '@/lib/constants';
import { formatDateTime, formatRupiah } from '@/lib/format';

export default function TransactionListPage() {
  const [status, setStatus] = useState('');
  const [page, setPage] = useState(1);

  const list = useQuery({
    queryKey: ['admin-transactions', { status, page }],
    queryFn: () => fetchAdminTransactions({ status, page, limit: 10 }),
    placeholderData: keepPreviousData,
    refetchInterval: 30_000, // status pembayaran mendekati real-time
  });
  const items = list.data?.data ?? [];

  return (
    <>
      <PageHeader title="Transaksi" description="Pantau pesanan dan status pembayaran." />

      <div role="group" aria-label="Filter status" className="mb-4 flex flex-wrap gap-2">
        {TRANSACTION_STATUSES.map((s) => (
          <button
            key={s.value}
            type="button"
            aria-pressed={status === s.value}
            onClick={() => {
              setStatus(s.value);
              setPage(1);
            }}
            className={`rounded-full border px-3 py-1 text-sm ${status === s.value ? 'border-brand bg-brand-tint font-semibold text-brand' : 'border-line bg-white text-ink-soft hover:bg-paper'}`}
          >
            {s.label}
          </button>
        ))}
      </div>

      {list.isPending && <LoadingState label="Memuat transaksi" />}
      {list.isError && <ErrorState error={list.error} onRetry={list.refetch} />}
      {list.isSuccess && items.length === 0 && <EmptyState title="Belum ada transaksi" description="Pesanan konsumen akan muncul di sini." />}

      {items.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-line bg-white">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">Daftar transaksi</caption>
            <thead className="border-b border-line bg-paper text-ink-soft">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Nomor pesanan</th>
                <th scope="col" className="px-4 py-3 font-medium">Pembeli</th>
                <th scope="col" className="px-4 py-3 text-right font-medium">Total</th>
                <th scope="col" className="px-4 py-3 font-medium">Status</th>
                <th scope="col" className="px-4 py-3 font-medium">Dibuat</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {items.map((t) => (
                <tr key={t.id}>
                  <td className="px-4 py-3 font-medium">{t.invoice_number}</td>
                  <td className="px-4 py-3">
                    {t.customer_name}
                    <div className="text-xs text-ink-soft">{t.customer_email}</div>
                  </td>
                  <td className="px-4 py-3 text-right">{formatRupiah(t.total_amount)}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={t.status} />
                  </td>
                  <td className="px-4 py-3 text-ink-soft">{formatDateTime(t.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <Pagination meta={list.data?.meta} onPageChange={setPage} />
    </>
  );
}
