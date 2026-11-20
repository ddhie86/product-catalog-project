import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { AUTH_EXPIRED_EVENT, tokenStorage } from '@/services/apiClient';
import * as authApi from './services/authApi';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => tokenStorage.get());
  const [permissions, setPermissions] = useState([]);
  const [permissionsLoading, setPermissionsLoading] = useState(Boolean(token));
  const [permissionsError, setPermissionsError] = useState(null);
  const permissionRequest = useRef(0);

  // Interceptor memancarkan event ini saat backend menjawab 401 untuk endpoint admin.
  useEffect(() => {
    const onExpired = () => setToken(null);
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired);
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired);
  }, []);

  const refreshPermissions = useCallback(async ({ silent = false } = {}) => {
    const requestID = ++permissionRequest.current;
    if (!token) {
      if (requestID === permissionRequest.current) {
        setPermissions([]);
        setPermissionsError(null);
        setPermissionsLoading(false);
      }
      return;
    }
    if (!silent) setPermissionsLoading(true);
    try {
      const currentPermissions = await authApi.getCurrentPermissions();
      if (requestID === permissionRequest.current) {
        setPermissions(currentPermissions);
        setPermissionsError(null);
      }
    } catch (error) {
      if (requestID === permissionRequest.current) {
        setPermissions([]);
        setPermissionsError(error);
      }
    } finally {
      if (requestID === permissionRequest.current && !silent) setPermissionsLoading(false);
    }
  }, [token]);

  useEffect(() => {
    refreshPermissions();
  }, [refreshPermissions]);

  const login = useCallback(async (credentials) => {
    const { access_token: accessToken } = await authApi.login(credentials);
    tokenStorage.set(accessToken);
    setPermissionsLoading(true);
    setToken(accessToken);
  }, []);

  const logout = useCallback(() => {
    tokenStorage.clear();
    permissionRequest.current += 1;
    setToken(null);
    setPermissions([]);
    setPermissionsError(null);
  }, []);

  const hasPermission = useCallback((permission) => permissions.includes(permission), [permissions]);
  const value = useMemo(() => ({
    isAuthenticated: Boolean(token),
    permissions,
    permissionsLoading,
    permissionsError,
    hasPermission,
    refreshPermissions,
    login,
    logout,
  }), [token, permissions, permissionsLoading, permissionsError, hasPermission, refreshPermissions, login, logout]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
