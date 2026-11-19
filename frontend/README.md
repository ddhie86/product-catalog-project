# Product Catalog — Frontend (React)

React 18 + Vite + Tailwind CSS, struktur **feature-based** sesuai TDD bagian 3.3.
Data fetching memakai **TanStack Query** (opsi React Query di TDD); state yang tersisa
hanya autentikasi admin (Context), jadi Redux Toolkit tidak diperlukan.

## Menjalankan

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173  (proxy /api -> http://localhost:8080)
npm test
npm run build
```

Backend harus berjalan di `:8080` (lihat README di root). Untuk production, set
`VITE_API_BASE_URL` ke URL API.

## Struktur

```
src/
  assets/
  components/            # Button, FormFields, Modal, Pagination, StatusBadge, layout, ProtectedRoute
  lib/                   # format (Rupiah, tanggal), konstanta
  services/apiClient.js  # axios + interceptor (JWT, error 401, normalisasi error)
  routes/index.jsx       # semua route (halaman admin di-lazy-load)
  modules/
    auth/                # login admin + AuthContext            (tambahan di luar TDD)
    catalog/             # CatalogPage, ProductCard, ProductDetailPage
    checkout/            # CheckoutPage, PaymentRedirect, PaymentSuccessPage
    admin-products/      # AdminProductList, ProductFormModal
    admin-dashboard/     # DashboardPage
    transactions/        # TransactionListPage
```

## Route

| Path | Halaman |
| --- | --- |
| `/` | Katalog (filter di URL: `q`, `category`, `sort`, `page`) |
| `/products/:id` | Detail produk |
| `/checkout?product=:id&qty=:n` | Form checkout |
| `/payment/redirect` | Instruksi + pengalihan ke payment gateway |
| `/payment/success` | Tujuan `success_redirect_url` gateway |
| `/admin/login` | Login admin |
| `/admin`, `/admin/products`, `/admin/transactions` | Panel admin (butuh login) |

## Catatan desain

- Font Plus Jakarta Sans (paket `@fontsource`, tanpa CDN). Warna ada di `tailwind.config.js`:
  biru tua = merek, amber = perlu perhatian (pending, stok menipis).
- Status transaksi selalu memakai label teks, tidak hanya warna.
- Token JWT disimpan di `localStorage` (mudah, tapi rentan XSS). Untuk hardening,
  pindah ke httpOnly cookie dari backend.
- `PaymentRedirect` hanya mengarahkan ke URL `https` (atau `http://localhost`).
- Checkout saat ini 1 produk per pesanan ("Beli sekarang"), sesuai TDD dan skema API yang
  sudah menerima `items[]`; keranjang bisa ditambah tanpa perubahan backend.
