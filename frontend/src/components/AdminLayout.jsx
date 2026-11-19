import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { Button } from './Button.jsx';

const nav = [
  { to: '/admin', label: 'Dasbor', end: true },
  { to: '/admin/products', label: 'Produk' },
  { to: '/admin/transactions', label: 'Transaksi' },
];

export default function AdminLayout() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  const onLogout = () => {
    logout();
    navigate('/admin/login', { replace: true });
  };

  return (
    <div className="min-h-screen md:flex">
      <aside className="border-b border-line bg-white md:min-h-screen md:w-56 md:border-b-0 md:border-r">
        <div className="px-4 py-4 text-lg font-semibold text-brand">Panel admin</div>
        <nav aria-label="Menu admin" className="flex gap-1 px-2 pb-2 md:flex-col md:pb-0">
          {nav.map((item) => (
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
