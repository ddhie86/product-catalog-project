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
  components/            # Button, fields, modal, pagination, layouts, permission guards
  lib/                   # Currency/date formatting and constants
  services/apiClient.js  # Axios client and interceptors (JWT, 401 handling, error normalization)
  routes/index.jsx       # Application routes (admin pages are lazy-loaded)
  modules/
    auth/                # Admin login, current permissions and AuthContext
    catalog/             # CatalogPage, ProductCard, ProductDetailPage
    checkout/            # CheckoutPage, PaymentRedirect, PaymentSuccessPage
    admin-products/      # AdminProductList, ProductFormModal
    admin-categories/    # Category CRUD
    admin-users/         # Admin account CRUD and role assignment
    admin-roles/         # Role CRUD and permission matrix
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
| `/admin` | Protected dashboard or first menu available to the signed-in role |
| `/admin/products`, `/admin/categories` | Product and category management |
| `/admin/users`, `/admin/roles` | Admin accounts, role assignment, and permission management |
| `/admin/transactions` | Transaction list and details |

Admin navigation and create/edit/delete controls are shown according to the
current role's permissions. The backend independently enforces the same
permissions for each protected API action. Super Admins can configure custom
roles from **Role dan permission** and assign them from **Pengguna admin**.

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
