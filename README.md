# Product Catalog Platform

A Go modular-monolith backend and React admin/catalog frontend, organized with
Clean Architecture (lightweight DDD) and aligned with the *Detailed Project
Proposal: Modular E-Commerce Product Catalog Platform*.

## Structure

```text
cmd/
  api/main.go            # Entry point and dependency-injection wiring
  seed/main.go           # Creates the initial admin account
internal/
  config/                # Environment configuration
  platform/              # Infrastructure adapters (database, Xendit, SMTP)
  modules/
    auth/                # Admin authentication, users, roles, permissions (RBAC)
    product/             # Product/category management, public catalog, ProductServiceContract
    transaction/         # Checkout, webhooks, transaction list, reports
    contact/             # Public contact form, message storage, support email forwarding
    activity/            # Admin activity snapshots and audit log
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
cp .env.example .env            # Set DB_PASSWORD, JWT_SECRET, SUPPORT_EMAIL, SEED_ADMIN_PASSWORD, and XENDIT_* before use
make up                         # Start PostgreSQL and Mailpit
make tidy                       # Update Go module dependencies and go.sum
make migrate-up                 # Requires the golang-migrate CLI
make seed                       # Creates Super Admin from SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD
make run                        # http://localhost:8080/healthz
make test
```

Run `make migrate-up` after pulling schema changes. Migration `000002_rbac`
creates the permission and role tables, installs the default roles, and assigns
existing admin accounts to **Super Admin**. The seed command also assigns the
new initial account to **Super Admin**.

## API endpoints and permissions

All `/admin` routes except login require a valid admin JWT. Management routes
also enforce their permission on the backend; hiding a menu or button in the
frontend is not the security boundary.

| Method | Path | Required permission |
| --- | --- | --- |
| POST | `/api/v1/admin/login` | – |
| GET | `/api/v1/admin/permissions` | JWT |
| GET/POST | `/api/v1/admin/products` | `products:view` / `products:create` |
| PUT/DELETE | `/api/v1/admin/products/:id` | `products:edit` / `products:delete` |
| GET/POST | `/api/v1/admin/categories` | `categories:view` / `categories:create` |
| PUT/DELETE | `/api/v1/admin/categories/:id` | `categories:edit` / `categories:delete` |
| GET/POST | `/api/v1/admin/users` | `users:view` / `users:create` |
| PUT/DELETE | `/api/v1/admin/users/:id` | `users:edit` / `users:delete` |
| GET | `/api/v1/admin/roles` | `roles:view` or `users:assign_roles` |
| POST | `/api/v1/admin/roles` | `roles:create` |
| PUT/DELETE | `/api/v1/admin/roles/:id` | `roles:edit` / `roles:delete` |
| GET | `/api/v1/admin/transactions` (`?status=`) | `transactions:view` |
| GET | `/api/v1/admin/transactions/:id` *(additional)* | `transactions:view` |
| GET | `/api/v1/admin/reports/summary` | `dashboard:view` |
| GET | `/api/v1/products` (`q, category, sort, page, limit`) | – |
| GET | `/api/v1/products/:id` | – |
| GET | `/api/v1/categories` | – |
| POST | `/api/v1/checkout` | – |
| POST | `/api/v1/payments/webhook` (`X-Callback-Token`) | Gateway token |
| POST | `/api/v1/contact-us` | – |
| GET | `/api/v1/admin/contact-messages` | `contact_messages:view` |
| POST | `/api/v1/admin/contact-messages/:id/reply` | `contact_messages:reply` |
| GET | `/api/v1/admin/activity-logs` | `activity_logs:view` |

`sort` options: `newest` (default), `price_asc`, `price_desc`, and `popular`.
Contact messages (`name`, `email`, `type`, and `message`) are saved to
`contact_message` and forwarded through SMTP to `SUPPORT_EMAIL`.
Successful admin changes to products, categories, users, roles, and contact-message
replies are recorded in `activity_log` with before/after snapshots.

### Admin roles

The default roles are **Super Admin** (all permissions), **Catalog Manager**
(product and category management), and **Viewer** (read-only dashboard, catalog,
and transaction access). Super Admins can create custom roles and choose menu
and action permissions from the role-management page. Permissions from multiple
assigned roles are combined. Assigning roles to admin users requires
`users:assign_roles`; creating an account without that permission assigns the
default role. The system Super Admin role cannot be edited or deleted, and the
last Super Admin cannot be removed or demoted.
Including `role_ids` in a user create/update request also requires
`users:assign_roles`; otherwise role assignments are preserved on update and
new accounts receive the default role.

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
9. **Admin RBAC** — Menu visibility and management actions are defined by
   database-backed roles and permissions. The role-management UI can customize
   roles while the system Super Admin role remains protected.

## Database security

- Database settings are supplied as separate `DB_HOST`, `DB_PORT`, `DB_NAME`,
  `DB_USER`, `DB_PASSWORD`, `DB_SSLMODE`, and optional `DB_SSLROOTCERT`
  environment variables. The committed `.env.example` contains no database
  password; `.env` is ignored by Git.
- Generate a local password with `openssl rand -hex 32`. Docker Compose maps the
  PostgreSQL port to `127.0.0.1` only and requires the database settings.
- `DB_SSLMODE=disable` is intended only for local development. Production
  configuration rejects non-TLS modes. Prefer `verify-full`, setting
  `DB_SSLROOTCERT` to the provider's CA certificate when needed. `require` is
  accepted when certificate verification is unavailable.
- Store production credentials in your deployment secret manager, not in source
  control or container images.

## Copyright and usage restrictions

Copyright (c) 2026 Dedy Setiawan. All Rights Reserved. This code is published
solely for portfolio evaluation and may not be used, copied, or redistributed
without prior written permission.
