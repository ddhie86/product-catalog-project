// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchDashboardSummary } from '@/modules/transactions/services/transactionApi';
import { PageHeader } from '@/components/PageHeader.jsx';
import { ErrorState, LoadingState } from '@/components/States.jsx';
import { formatRupiah } from '@/lib/format';

function Metric({ label, value, tone = 'text-ink', note }) {
  return (
    <div className="rounded-lg border border-line bg-white p-5">
      <dt className="text-sm text-ink-soft">{label}</dt>
      <dd className={`mt-1 text-2xl font-semibold ${tone}`}>{value}</dd>
      {note && <p className="mt-1 text-xs text-ink-soft">{note}</p>}
    </div>
  );
}

export default function DashboardPage() {
  const { data, isPending, isError, error, refetch } = useQuery({
    queryKey: ['admin-summary'],
    queryFn: fetchDashboardSummary,
  });

  return (
    <>
      <PageHeader title="Dasbor" description="Ringkasan penjualan dan stok." />
      {isPending && <LoadingState label="Memuat ringkasan" />}
      {isError && <ErrorState error={error} onRetry={refetch} />}
      {data && (
        <dl className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <Metric label="Total omzet" value={formatRupiah(data.total_revenue)} note="Dari transaksi lunas" />
          <Metric label="Transaksi lunas" value={data.total_paid_transactions} />
          <Metric label="Menunggu pembayaran" value={data.pending_count} tone={data.pending_count > 0 ? 'text-signal' : 'text-ink'} />
          <Metric label="Produk stok menipis" value={data.low_stock_alerts} tone={data.low_stock_alerts > 0 ? 'text-signal' : 'text-ink'} note="Stok kurang dari batas minimum" />
        </dl>
      )}
      {data?.low_stock_alerts > 0 && (
        <p className="mt-6 text-sm">
          Ada {data.low_stock_alerts} produk yang hampir habis.{' '}
          <Link to="/admin/products" className="font-medium text-brand">
            Periksa stok produk
          </Link>
        </p>
      )}
    </>
  );
}
