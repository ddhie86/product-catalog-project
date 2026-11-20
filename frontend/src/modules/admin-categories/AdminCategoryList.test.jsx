import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, expect, it, vi } from 'vitest';
import AdminCategoryList from './AdminCategoryList.jsx';

const api = vi.hoisted(() => ({
  getAdminCategories: vi.fn(),
  createCategory: vi.fn(),
  updateCategory: vi.fn(),
  deleteCategory: vi.fn(),
}));
const auth = vi.hoisted(() => ({ permissions: [] }));

vi.mock('./services/adminCategoryApi', () => api);
vi.mock('@/modules/auth/AuthContext.jsx', () => ({
  useAuth: () => ({ hasPermission: (permission) => auth.permissions.includes(permission) }),
}));

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <AdminCategoryList />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  auth.permissions = ['categories:view', 'categories:create', 'categories:edit', 'categories:delete'];
});

it('membuat kategori melalui form admin', async () => {
  api.getAdminCategories.mockResolvedValue([]);
  api.createCategory.mockResolvedValue({ id: 'cat-1', name: 'Minuman' });
  renderPage();

  const addButtons = await screen.findAllByRole('button', { name: 'Tambah kategori' });
  fireEvent.click(addButtons[0]);
  fireEvent.change(screen.getByLabelText('Nama kategori'), { target: { value: '  Minuman  ' } });
  fireEvent.click(screen.getByRole('button', { name: 'Simpan kategori' }));

  await waitFor(() => expect(api.createCategory).toHaveBeenCalledWith({ name: 'Minuman' }));
});

it('menampilkan alasan saat kategori yang masih dipakai gagal dihapus', async () => {
  api.getAdminCategories.mockResolvedValue([{ id: 'cat-1', name: 'Minuman' }]);
  api.deleteCategory.mockRejectedValue(new Error('Kategori masih digunakan produk.'));
  renderPage();

  fireEvent.click(await screen.findByRole('button', { name: 'Hapus Minuman' }));
  fireEvent.click(screen.getByRole('button', { name: 'Hapus' }));

  expect(await screen.findByRole('alert')).toHaveTextContent('Kategori masih digunakan produk.');
});

it('menyembunyikan aksi kategori saat role hanya memiliki izin lihat', async () => {
  auth.permissions = ['categories:view'];
  api.getAdminCategories.mockResolvedValue([{ id: 'cat-1', name: 'Minuman' }]);
  renderPage();

  expect(await screen.findByText('Minuman')).toBeInTheDocument();
  expect(screen.queryByRole('button', { name: 'Tambah kategori' })).not.toBeInTheDocument();
  expect(screen.queryByRole('button', { name: 'Ubah Minuman' })).not.toBeInTheDocument();
  expect(screen.queryByRole('button', { name: 'Hapus Minuman' })).not.toBeInTheDocument();
});
