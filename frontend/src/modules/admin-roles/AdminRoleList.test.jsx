import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, expect, it, vi } from 'vitest';
import AdminRoleList from './AdminRoleList.jsx';

const api = vi.hoisted(() => ({
  getAdminRoles: vi.fn(),
  createRole: vi.fn(),
  updateRole: vi.fn(),
  deleteRole: vi.fn(),
}));

vi.mock('./services/adminRoleApi', () => api);
vi.mock('@/modules/auth/AuthContext.jsx', () => ({
  useAuth: () => ({ hasPermission: () => true, refreshPermissions: vi.fn() }),
}));

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <AdminRoleList />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

it('creates a role and automatically includes the menu view permission', async () => {
  api.getAdminRoles.mockResolvedValue({
    roles: [],
    permissions: [
      { key: 'products:view', menu: 'products', menu_label: 'Produk', action: 'view', label: 'Lihat' },
      { key: 'products:edit', menu: 'products', menu_label: 'Produk', action: 'edit', label: 'Ubah' },
    ],
  });
  api.createRole.mockResolvedValue({ id: 'role-1', name: 'Editor' });
  renderPage();

  const addButtons = await screen.findAllByRole('button', { name: 'Tambah role' });
  fireEvent.click(addButtons[0]);
  fireEvent.change(screen.getByLabelText('Nama role'), { target: { value: 'Editor' } });
  fireEvent.click(screen.getByLabelText('Ubah'));
  fireEvent.click(screen.getByRole('button', { name: 'Simpan role' }));

  await waitFor(() => {
    expect(api.createRole).toHaveBeenCalledWith({
      name: 'Editor',
      permissions: ['products:edit', 'products:view'],
    });
  });
});
