// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { ErrorState, LoadingState } from './States.jsx';

/** Membungkus route admin; redirect ke /admin/login bila belum masuk. */
export default function ProtectedRoute() {
  const { isAuthenticated, permissionsLoading, permissionsError, refreshPermissions } = useAuth();
  const location = useLocation();
  if (!isAuthenticated) {
    return <Navigate to="/admin/login" replace state={{ from: location }} />;
  }
  if (permissionsLoading) return <LoadingState label="Memeriksa akses" />;
  if (permissionsError) return <ErrorState error={permissionsError} onRetry={refreshPermissions} />;
  return <Outlet />;
}
