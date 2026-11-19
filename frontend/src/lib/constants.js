// Harus sama dengan LOW_STOCK_THRESHOLD di backend (.env).
export const LOW_STOCK_THRESHOLD = 5;

export const SORT_OPTIONS = [
  { value: 'newest', label: 'Terbaru' },
  { value: 'price_asc', label: 'Harga terendah' },
  { value: 'price_desc', label: 'Harga tertinggi' },
  { value: 'popular', label: 'Terpopuler' },
];

export const TRANSACTION_STATUSES = [
  { value: '', label: 'Semua' },
  { value: 'PENDING', label: 'Menunggu pembayaran' },
  { value: 'PAID', label: 'Lunas' },
  { value: 'EXPIRED', label: 'Kedaluwarsa' },
  { value: 'CANCELLED', label: 'Dibatalkan' },
];
