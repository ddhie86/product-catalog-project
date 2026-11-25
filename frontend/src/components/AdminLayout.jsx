// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { Button } from './Button.jsx';

const nav = [
  { to: '/admin', label: 'Dasbor', end: true, permission: 'dashboard:view' },
  { to: '/admin/products', label: 'Produk', permission: 'products:view' },
  { to: '/admin/categories', label: 'Kategori', permission: 'categories:view' },
  { to: '/admin/transactions', label: 'Transaksi', permission: 'transactions:view' },
  { to: '/admin/users', label: 'Pengguna admin', permission: 'users:view' },
  { to: '/admin/roles', label: 'Role', permission: 'roles:view' },
  { to: '/admin/contact-messages', label: 'Pesan kontak', permission: 'contact_messages:view' },
  { to: '/admin/activity-logs', label: 'Log aktivitas', permission: 'activity_logs:view' },
];

export default function AdminLayout() {
  const { logout, hasPermission } = useAuth();
  const navigate = useNavigate();

  const onLogout = () => {
    logout();
    navigate('/admin/login', { replace: true });
  };

  return (
    <div className="min-h-screen md:flex">
      <aside className="border-b border-line bg-white md:min-h-screen md:w-56 md:border-b-0 md:border-r">
        <div className="px-4 py-4 text-lg font-semibold text-brand">Panel admin</div>
        <nav aria-label="Menu admin" className="flex flex-wrap gap-1 px-2 pb-2 md:flex-col md:flex-nowrap md:pb-0">
          {nav.filter((item) => hasPermission(item.permission)).map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `rounded-md px-3 py-2 text-sm font-medium ${isActive ? 'bg-brand-tint text-brand' : 'text-ink-soft hover:bg-paper'}`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="hidden px-2 pt-6 md:block">
          <Button variant="secondary" className="w-full" onClick={onLogout}>
            Keluar
          </Button>
        </div>
      </aside>
      <main className="flex-1 px-4 py-6 md:px-8">
        <div className="mb-4 flex justify-end md:hidden">
          <Button variant="secondary" onClick={onLogout}>
            Keluar
          </Button>
        </div>
        <Outlet />
      </main>
    </div>
  );
}
