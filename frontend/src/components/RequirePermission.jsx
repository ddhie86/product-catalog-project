import { Navigate } from 'react-router-dom';
import { useAuth } from '@/modules/auth/AuthContext.jsx';

export default function RequirePermission({ permission, children }) {
  const { hasPermission } = useAuth();
  return hasPermission(permission) ? children : <Navigate to="/admin/no-access" replace />;
}
