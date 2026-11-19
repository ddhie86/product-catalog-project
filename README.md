# Product Catalog — Backend (Go)

Modular Monolith + Clean Architecture (lightweight DDD) sesuai dokumen
*Proposal Detail Proyek: Platform E-Commerce Product Catalog Modular*.

## Struktur

```
cmd/
  api/main.go            # entry point + dependency injection wiring
  seed/main.go           # buat admin awal
internal/
  config/                # env config
  platform/              # adapter infrastruktur (database, payment/xendit, mailer/smtp)
  modules/
    auth/                # login admin (JWT)
    product/             # CMS produk (admin) + katalog publik + ProductServiceContract
    transaction/         # checkout, webhook, daftar transaksi, laporan
    payment/             # kontrak Gateway (domain saja)
    notification/        # worker email async (goroutine + channel)
      └─ tiap modul: delivery/http · domain · repository · usecase
frontend/                # React + Vite (lihat frontend/README.md)
pkg/                     # apperror, response, middleware, token(JWT), logger
migrations/              # SQL (golang-migrate)
```

Aturan dependensi: `delivery → usecase → domain ← repository`. Modul **transaction**
hanya mengakses produk lewat `product/domain.ProductServiceContract`, tidak pernah
menyentuh repository/tabel produk secara langsung.

## Menjalankan

```bash
cp .env.example .env            # isi JWT_SECRET, SEED_ADMIN_PASSWORD, dan XENDIT_* sebelum dipakai
make up                         # Postgres + Mailpit
make tidy                       # generate go.sum
make migrate-up                 # brew install golang-migrate
make seed                       # admin: SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD
make run                        # http://localhost:8080/healthz
make test
```

## Endpoint

| Method | Path | Auth |
| --- | --- | --- |
| POST | `/api/v1/admin/login` | – |
| GET/POST | `/api/v1/admin/products` | JWT |
| PUT/DELETE | `/api/v1/admin/products/:id` | JWT |
| GET | `/api/v1/admin/transactions` (`?status=`) | JWT |
| GET | `/api/v1/admin/transactions/:id` *(tambahan)* | JWT |
| GET | `/api/v1/admin/reports/summary` | JWT |
| GET | `/api/v1/products` (`q, category, sort, page, limit`) | – |
| GET | `/api/v1/products/:id` | – |
| GET | `/api/v1/categories` *(tambahan, dipakai frontend)* | – |
| POST | `/api/v1/checkout` | – |
| POST | `/api/v1/payments/webhook` (`X-Callback-Token`) | token gateway |

`sort`: `newest` (default) · `price_asc` · `price_desc` · `popular`.

## Keputusan desain di luar dokumen (mohon dikonfirmasi)

1. **Skema DB** — ERD 3.5 belum memuat `sku`, `category_id`, `deleted_at` (dipakai di 6.1)
   maupun `customer_phone`/`shipping_address` (form checkout 2.2.B). Semuanya ditambahkan
   di migrasi, plus tabel `categories`.
2. **`GET /admin/transactions/:id`** ditambahkan karena requirement "Detail Transaksi"
   belum punya endpoint.
3. **Router: Gin** (dokumen membolehkan Gin atau Echo).
4. **Gateway: Xendit** (dokumen membolehkan Midtrans/Xendit) di balik interface `Gateway`.
5. **Reservasi stok**: stok dikurangi saat checkout, dikembalikan bila gagal bayar/`EXPIRED`.
6. **Webhook idempoten** lewat `UpdateStatusIf(PENDING → target)`.
7. **Sort `popular`** masih fallback ke `newest` (butuh data penjualan; lihat TODO di repository).
8. **Email**: antrian penuh → event di-drop & di-log (tidak memblokir webhook). Belum ada outbox/persisten.
