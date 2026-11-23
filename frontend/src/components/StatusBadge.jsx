// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
const styles = {
  PENDING: { label: 'Menunggu pembayaran', cls: 'bg-signal-tint text-signal' },
  PAID: { label: 'Lunas', cls: 'bg-ok-tint text-ok' },
  EXPIRED: { label: 'Kedaluwarsa', cls: 'bg-paper text-ink-soft border border-line' },
  CANCELLED: { label: 'Dibatalkan', cls: 'bg-danger-tint text-danger' },
};

export function StatusBadge({ status }) {
  const s = styles[status] ?? { label: status, cls: 'bg-paper text-ink-soft' };
  return <span className={`inline-block rounded px-2 py-0.5 text-xs font-semibold ${s.cls}`}>{s.label}</span>;
}
