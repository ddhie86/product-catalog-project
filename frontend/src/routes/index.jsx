import { lazy, Suspense } from 'react';
import { Route, Routes } from 'react-router-dom';

import ConsumerLayout from '@/components/ConsumerLayout.jsx';
import AdminLayout from '@/components/AdminLayout.jsx';
import ProtectedRoute from '@/components/ProtectedRoute.jsx';
import { LoadingState } from '@/components/States.jsx';

import CatalogPage from '@/modules/catalog/CatalogPage.jsx';
import ProductDetailPage from '@/modules/catalog/ProductDetailPage.jsx';
import CheckoutPage from '@/modules/checkout/CheckoutPage.jsx';
import PaymentRedirect from '@/modules/checkout/PaymentRedirect.jsx';
import PaymentSuccessPage from '@/modules/checkout/PaymentSuccessPage.jsx';
import LoginPage from '@/modules/auth/pages/LoginPage.jsx';

// Halaman admin dipisah ke chunk sendiri; konsumen tidak ikut mengunduhnya.
const DashboardPage = lazy(() => import('@/modules/admin-dashboard/DashboardPage.jsx'));
const AdminProductList = lazy(() => import('@/modules/admin-products/AdminProductList.jsx'));
const TransactionListPage = lazy(() => import('@/modules/transactions/TransactionListPage.jsx'));

function NotFound() {
  return <p className="py-16 text-center text-ink-soft">Halaman tidak ditemukan.</p>;
}

export default function AppRoutes() {
  return (
    <Suspense fallback={<LoadingState />}>
      <Routes>
        {/* Konsumen (publik) */}
        <Route element={<ConsumerLayout />}>
          <Route index element={<CatalogPage />} />
          <Route path="products/:id" element={<ProductDetailPage />} />
          <Route path="checkout" element={<CheckoutPage />} />
          <Route path="payment/redirect" element={<PaymentRedirect />} />
          <Route path="payment/success" element={<PaymentSuccessPage />} />
        </Route>

        {/* Admin */}
        <Route path="admin/login" element={<LoginPage />} />
        <Route path="admin" element={<ProtectedRoute />}>
          <Route element={<AdminLayout />}>
            <Route index element={<DashboardPage />} />
            <Route path="products" element={<AdminProductList />} />
            <Route path="transactions" element={<TransactionListPage />} />
          </Route>
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
    </Suspense>
  );
}
