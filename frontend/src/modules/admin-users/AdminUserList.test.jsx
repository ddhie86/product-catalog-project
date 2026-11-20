import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, expect, it, vi } from 'vitest';
import AdminUserList from './AdminUserList.jsx';

const api = vi.hoisted(() => ({
  getAdminUsers: vi.fn(),
  createAdminUser: vi.fn(),
  updateAdminUser: vi.fn(),
  deleteAdminUser: vi.fn(),
}));

const roleApi = vi.hoisted(() => ({
  getAdminRoles: vi.fn(),
}));

vi.mock('./services/adminUserApi', () => api);
vi.mock('@/modules/admin-roles/services/adminRoleApi', () => roleApi);
vi.mock('@/modules/auth/AuthContext.jsx', () => ({
  useAuth: () => ({ hasPermission: () => true }),
}));

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <AdminUserList />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

it('membuat akun admin melalui form dengan email yang sudah dipangkas', async () => {
  api.getAdminUsers.mockResolvedValue([]);
  api.createAdminUser.mockResolvedValue({ id: 'admin-1', email: 'admin@example.com' });
  roleApi.getAdminRoles.mockResolvedValue({
    roles: [{ id: 'role-1', name: 'Super Admin', is_system: true }],
    permissions: [],
  });
  renderPage();

  const addButtons = await screen.findAllByRole('button', { name: 'Tambah admin' });
  fireEvent.click(addButtons[0]);
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: ' admin@example.com ' } });
  fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'strong-password' } });
  fireEvent.click(await screen.findByLabelText(/Super Admin/));
  fireEvent.click(screen.getByRole('button', { name: 'Simpan admin' }));

  await waitFor(() =>
    expect(api.createAdminUser).toHaveBeenCalledWith({
      email: 'admin@example.com',
      password: 'strong-password',
      role_ids: ['role-1'],
    }),
  );
});
