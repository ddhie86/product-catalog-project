// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { lazy, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';

import ConsumerLayout from '@/components/ConsumerLayout.jsx';
import AdminLayout from '@/components/AdminLayout.jsx';
import ProtectedRoute from '@/components/ProtectedRoute.jsx';
import RequirePermission from '@/components/RequirePermission.jsx';
import { LoadingState } from '@/components/States.jsx';
import { useAuth } from '@/modules/auth/AuthContext.jsx';

import CatalogPage from '@/modules/catalog/CatalogPage.jsx';
import ProductDetailPage from '@/modules/catalog/ProductDetailPage.jsx';
import CheckoutPage from '@/modules/checkout/CheckoutPage.jsx';
import PaymentRedirect from '@/modules/checkout/PaymentRedirect.jsx';
import PaymentSuccessPage from '@/modules/checkout/PaymentSuccessPage.jsx';
import LoginPage from '@/modules/auth/pages/LoginPage.jsx';
import ContactUsPage from '@/modules/contact/ContactUsPage.jsx';

// Halaman admin dipisah ke chunk sendiri; konsumen tidak ikut mengunduhnya.
const DashboardPage = lazy(() => import('@/modules/admin-dashboard/DashboardPage.jsx'));
const AdminProductList = lazy(() => import('@/modules/admin-products/AdminProductList.jsx'));
const AdminCategoryList = lazy(() => import('@/modules/admin-categories/AdminCategoryList.jsx'));
const AdminUserList = lazy(() => import('@/modules/admin-users/AdminUserList.jsx'));
const AdminRoleList = lazy(() => import('@/modules/admin-roles/AdminRoleList.jsx'));
const TransactionListPage = lazy(() => import('@/modules/transactions/TransactionListPage.jsx'));
const ContactMessageListPage = lazy(() => import('@/modules/admin-contact-messages/ContactMessageListPage.jsx'));
const ActivityLogPage = lazy(() => import('@/modules/admin-activity-logs/ActivityLogPage.jsx'));

function AdminHome() {
  const { hasPermission } = useAuth();
  if (hasPermission('dashboard:view')) return <DashboardPage />;
  const firstAllowed = [
    ['products:view', '/admin/products'],
    ['categories:view', '/admin/categories'],
    ['transactions:view', '/admin/transactions'],
    ['users:view', '/admin/users'],
    ['roles:view', '/admin/roles'],
    ['contact_messages:view', '/admin/contact-messages'],
    ['activity_logs:view', '/admin/activity-logs'],
  ].find(([permission]) => hasPermission(permission));
  return <Navigate to={firstAllowed?.[1] ?? '/admin/no-access'} replace />;
}

function NoAccessPage() {
  return (
    <div className="rounded-lg border border-line bg-white p-8 text-center">
      <h1 className="text-xl font-semibold">Akses tidak tersedia</h1>
      <p className="mt-2 text-sm text-ink-soft">Role akun Anda belum memiliki izin untuk membuka menu ini.</p>
    </div>
  );
}

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
          <Route path="contact-us" element={<ContactUsPage />} />
        </Route>

        {/* Admin */}
        <Route path="admin/login" element={<LoginPage />} />
        <Route path="admin" element={<ProtectedRoute />}>
          <Route element={<AdminLayout />}>
            <Route index element={<AdminHome />} />
            <Route path="no-access" element={<NoAccessPage />} />
            <Route path="products" element={<RequirePermission permission="products:view"><AdminProductList /></RequirePermission>} />
            <Route path="categories" element={<RequirePermission permission="categories:view"><AdminCategoryList /></RequirePermission>} />
            <Route path="users" element={<RequirePermission permission="users:view"><AdminUserList /></RequirePermission>} />
            <Route path="roles" element={<RequirePermission permission="roles:view"><AdminRoleList /></RequirePermission>} />
            <Route path="transactions" element={<RequirePermission permission="transactions:view"><TransactionListPage /></RequirePermission>} />
            <Route path="contact-messages" element={<RequirePermission permission="contact_messages:view"><ContactMessageListPage /></RequirePermission>} />
            <Route path="activity-logs" element={<RequirePermission permission="activity_logs:view"><ActivityLogPage /></RequirePermission>} />
          </Route>
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
    </Suspense>
  );
}
