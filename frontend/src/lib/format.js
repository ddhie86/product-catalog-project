const idr = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
});
const dateTime = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' });

export const formatRupiah = (n) => idr.format(Number(n) || 0);

export const formatDateTime = (iso) => (iso ? dateTime.format(new Date(iso)) : '-');

/** Hanya izinkan https (atau http di localhost) sebelum mengarahkan browser ke URL eksternal. */
export function isSafeRedirectUrl(raw) {
  try {
    const u = new URL(raw);
    return u.protocol === 'https:' || (u.protocol === 'http:' && u.hostname === 'localhost');
  } catch {
    return false;
  }
}
