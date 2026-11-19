# Product Catalog — Backend (Go)

A Modular Monolith with Clean Architecture (lightweight DDD), aligned with the
*Detailed Project Proposal: Modular E-Commerce Product Catalog Platform*.

## Structure

```text
cmd/
  api/main.go            # Entry point and dependency-injection wiring
  seed/main.go           # Creates the initial admin account
internal/
  config/                # Environment configuration
  platform/              # Infrastructure adapters (database, Xendit, SMTP)
  modules/
    auth/                # Admin authentication (JWT)
    product/             # Admin product CMS, public catalog, ProductServiceContract
    transaction/         # Checkout, webhooks, transaction list, reports
    payment/             # Payment gateway contract (domain only)
    notification/        # Async email worker (goroutine and channel)
      └─ Each module: delivery/http · domain · repository · usecase
frontend/                # React + Vite (see frontend/README.md)
pkg/                     # Shared errors, responses, middleware, JWT, logging
migrations/              # SQL migrations (golang-migrate)
```

Dependency rule: `delivery → usecase → domain ← repository`. The **transaction**
module accesses products only through `product/domain.ProductServiceContract`; it
does not directly access the product repository or tables.

## Run locally

```bash
cp .env.example .env            # Set JWT_SECRET, SEED_ADMIN_PASSWORD, and XENDIT_* before use
make up                         # Start PostgreSQL and Mailpit
make tidy                       # Update Go module dependencies and go.sum
make migrate-up                 # Requires the golang-migrate CLI
make seed                       # Admin credentials: SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD
make run                        # http://localhost:8080/healthz
make test
```

## API endpoints

| Method | Path | Auth |
| --- | --- | --- |
| POST | `/api/v1/admin/login` | – |
| GET/POST | `/api/v1/admin/products` | JWT |
| PUT/DELETE | `/api/v1/admin/products/:id` | JWT |
| GET | `/api/v1/admin/transactions` (`?status=`) | JWT |
| GET | `/api/v1/admin/transactions/:id` *(additional)* | JWT |
| GET | `/api/v1/admin/reports/summary` | JWT |
| GET | `/api/v1/products` (`q, category, sort, page, limit`) | – |
| GET | `/api/v1/products/:id` | – |
| GET | `/api/v1/categories` *(additional; used by the frontend)* | – |
| POST | `/api/v1/checkout` | – |
| POST | `/api/v1/payments/webhook` (`X-Callback-Token`) | Gateway token |

`sort` options: `newest` (default), `price_asc`, `price_desc`, and `popular`.

## Design decisions beyond the proposal

1. **Database schema** — The ERD in section 3.5 omits `sku`, `category_id`, and
   `deleted_at` (used in section 6.1), as well as `customer_phone` and
   `shipping_address` (required by the checkout form in section 2.2.B). These
   fields and a `categories` table are included in the migration.
2. **`GET /admin/transactions/:id`** — Added because the transaction-detail
   requirement did not specify an endpoint.
3. **Router: Gin** — The proposal allows either Gin or Echo.
4. **Gateway: Xendit** — The proposal allows either Midtrans or Xendit; the
   integration is behind a `Gateway` interface.
5. **Stock reservation** — Stock is deducted at checkout and restored if payment
   creation fails or the payment expires.
6. **Idempotent webhooks** — Status changes use `UpdateStatusIf(PENDING → target)`.
7. **Popular sort** — Currently falls back to `newest`; sales data is needed to
   implement popularity sorting (see the repository TODO).
8. **Email delivery** — If the queue is full, the event is dropped and logged so
   the webhook is not blocked. There is no persistent outbox yet.
