# Product Catalog — Frontend (React)

React 18, Vite, and Tailwind CSS with the **feature-based** structure described in
section 3.3 of the technical requirements. Data fetching uses **TanStack Query**
(the proposal's React Query option). Admin authentication is managed with
Context, so Redux Toolkit is not needed.

## Run locally

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173 (proxies /api to http://localhost:8080)
npm test
npm run build
```

The backend must be running on port `8080` (see the root README). For production,
set `VITE_API_BASE_URL` to the API URL.

## Structure

```text
src/
  assets/
  components/            # Button, fields, modal, pagination, status badge, layouts, protected route
  lib/                   # Currency/date formatting and constants
  services/apiClient.js  # Axios client and interceptors (JWT, 401 handling, error normalization)
  routes/index.jsx       # Application routes (admin pages are lazy-loaded)
  modules/
    auth/                # Admin login and AuthContext (additional to the proposal)
    catalog/             # CatalogPage, ProductCard, ProductDetailPage
    checkout/            # CheckoutPage, PaymentRedirect, PaymentSuccessPage
    admin-products/      # AdminProductList, ProductFormModal
    admin-dashboard/     # DashboardPage
    transactions/        # TransactionListPage
```

## Routes

| Path | Page |
| --- | --- |
| `/` | Product catalog (URL filters: `q`, `category`, `sort`, `page`) |
| `/products/:id` | Product details |
| `/checkout?product=:id&qty=:n` | Checkout form |
| `/payment/redirect` | Payment instructions and gateway redirect |
| `/payment/success` | Gateway `success_redirect_url` destination |
| `/admin/login` | Admin login |
| `/admin`, `/admin/products`, `/admin/transactions` | Protected admin panel |

## Design notes

- Plus Jakarta Sans is bundled with `@fontsource` (no CDN). Colors are defined in
  `tailwind.config.js`: dark blue is the brand color, and amber indicates attention
  (pending payments or low stock).
- Transaction statuses always include text labels, not just colors.
- The JWT is stored in `localStorage`, which is convenient but vulnerable to XSS.
  For stronger protection, use an `httpOnly` cookie issued by the backend.
- `PaymentRedirect` only allows `https` URLs (or `http://localhost`).
- Checkout currently supports one product per order ("Buy now"). The API already
  accepts `items[]`, so a cart can be added without backend changes.
