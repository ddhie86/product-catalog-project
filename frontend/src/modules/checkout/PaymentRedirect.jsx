// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useEffect, useState } from 'react';
import { Link, Navigate, useLocation } from 'react-router-dom';
import { Button } from '@/components/Button.jsx';
import { formatRupiah, isSafeRedirectUrl } from '@/lib/format';

const COUNTDOWN_SECONDS = 5;

export default function PaymentRedirect() {
  const { state } = useLocation();
  const [seconds, setSeconds] = useState(COUNTDOWN_SECONDS);
  const [cancelled, setCancelled] = useState(false);

  const url = state?.paymentUrl;
  const safe = isSafeRedirectUrl(url);

  useEffect(() => {
    if (!safe || cancelled) return undefined;
    if (seconds <= 0) {
      window.location.assign(url);
      return undefined;
    }
    const t = setTimeout(() => setSeconds((s) => s - 1), 1000);
    return () => clearTimeout(t);
  }, [safe, cancelled, seconds, url]);

  // Halaman dibuka langsung (tanpa melalui checkout) -> tidak ada data pesanan.
  if (!state) return <Navigate to="/" replace />;

  return (
    <div className="mx-auto max-w-md rounded-lg border border-line bg-white p-8 text-center">
      <h1 className="text-xl font-semibold">Pesanan dibuat</h1>
      <p className="mt-2 text-sm text-ink-soft">
        Nomor pesanan <span className="font-semibold text-ink">{state.invoiceNumber}</span>
        <br />
        Total {formatRupiah(state.totalAmount)}
      </p>

      {safe ? (
        <>
          <p className="mt-6 text-sm" aria-live="polite">
            {cancelled ? 'Pengalihan dibatalkan.' : `Anda akan diarahkan ke halaman pembayaran dalam ${seconds} detik.`}
          </p>
          <div className="mt-4 flex justify-center gap-2">
            <Button onClick={() => window.location.assign(url)}>Bayar sekarang</Button>
            {!cancelled && (
              <Button variant="secondary" onClick={() => setCancelled(true)}>
                Tunda
              </Button>
            )}
          </div>
          <p className="mt-6 text-xs text-ink-soft">Tautan pembayaran juga dikirim ke email Anda.</p>
        </>
      ) : (
        <p role="alert" className="mt-6 text-sm text-danger">
          Tautan pembayaran tidak valid. Cek email Anda atau hubungi penjual dengan nomor pesanan di atas.
        </p>
      )}
      <Link to="/" className="mt-6 inline-block text-sm text-brand">
        Kembali ke katalog
      </Link>
    </div>
  );
}
