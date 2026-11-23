// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Link } from 'react-router-dom';

// Tujuan success_redirect_url dari payment gateway (PAYMENT_SUCCESS_REDIRECT_URL di backend).
// Status final tetap ditentukan oleh webhook, bukan halaman ini.
export default function PaymentSuccessPage() {
  return (
    <div className="mx-auto max-w-md rounded-lg border border-line bg-white p-8 text-center">
      <h1 className="text-xl font-semibold">Terima kasih</h1>
      <p className="mt-2 text-sm text-ink-soft">
        Pembayaran Anda sedang kami verifikasi. Tanda terima akan dikirim ke email setelah pembayaran terkonfirmasi.
      </p>
      <Link to="/" className="mt-6 inline-block text-sm font-medium text-brand">
        Kembali ke katalog
      </Link>
    </div>
  );
}
