// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, expect, it, vi } from 'vitest';
import ProductFormModal from './ProductFormModal.jsx';

const api = vi.hoisted(() => ({
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
  uploadProductImage: vi.fn(),
}));
const auth = vi.hoisted(() => ({ permissions: ['products:create', 'products:edit'] }));

vi.mock('./services/adminProductApi', () => api);
vi.mock('@/modules/catalog/services/catalogApi', () => ({
  fetchCategories: vi.fn().mockResolvedValue([]),
}));
vi.mock('@/modules/auth/AuthContext.jsx', () => ({
  useAuth: () => ({ hasPermission: (permission) => auth.permissions.includes(permission) }),
}));

function renderForm(product = null) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const onClose = vi.fn();
  render(
    <QueryClientProvider client={queryClient}>
      <ProductFormModal open product={product} onClose={onClose} />
    </QueryClientProvider>,
  );
  return onClose;
}

beforeEach(() => {
  vi.clearAllMocks();
  auth.permissions = ['products:create', 'products:edit'];
  vi.stubGlobal('URL', {
    ...URL,
    createObjectURL: vi.fn(() => 'blob:product-preview'),
    revokeObjectURL: vi.fn(),
  });
  api.createProduct.mockResolvedValue({ id: 'product-1' });
  api.updateProduct.mockResolvedValue({ id: 'product-1' });
});

it('previews and uploads a selected image, then saves its URL with a new product', async () => {
  const imageURL = 'https://drive.google.com/uc?export=view&id=image-1';
  api.uploadProductImage.mockImplementation(async (_file, onUploadProgress) => {
    onUploadProgress({ loaded: 50, total: 100 });
    return { image_url: imageURL, file_id: 'image-1', name: 'image.webp' };
  });
  renderForm();

  const file = new File(['webp-data'], 'product.webp', { type: 'image/webp' });
  fireEvent.change(screen.getByLabelText('Gambar produk'), { target: { files: [file] } });
  expect(screen.getByRole('img', { name: 'Pratinjau gambar produk' })).toHaveAttribute('src', 'blob:product-preview');

  fireEvent.click(screen.getByRole('button', { name: 'Unggah gambar' }));
  await waitFor(() => expect(api.uploadProductImage).toHaveBeenCalledWith(file, expect.any(Function)));
  await waitFor(() => expect(screen.getByRole('img', { name: 'Pratinjau gambar produk' })).toHaveAttribute('src', imageURL));

  fireEvent.change(screen.getByLabelText('Nama produk'), { target: { value: 'Produk baru' } });
  fireEvent.change(screen.getByLabelText('SKU'), { target: { value: 'SKU-1' } });
  fireEvent.change(screen.getByLabelText('Harga (Rp)'), { target: { value: '12000' } });
  fireEvent.click(screen.getByRole('button', { name: 'Simpan produk' }));

  await waitFor(() => expect(api.createProduct).toHaveBeenCalledWith({
    name: 'Produk baru',
    sku: 'SKU-1',
    description: '',
    price: 12000,
    stock: 0,
    category_id: null,
    image_url: imageURL,
  }));
});

it('shows an upload error without losing the selected image preview', async () => {
  api.uploadProductImage.mockRejectedValue(new Error('Upload gagal.'));
  renderForm();

  const file = new File(['png-data'], 'product.png', { type: 'image/png' });
  fireEvent.change(screen.getByLabelText('Gambar produk'), { target: { files: [file] } });
  fireEvent.click(screen.getByRole('button', { name: 'Unggah gambar' }));

  expect(await screen.findByRole('alert')).toHaveTextContent('Upload gagal.');
  expect(screen.getByRole('img', { name: 'Pratinjau gambar produk' })).toHaveAttribute('src', 'blob:product-preview');
});

it('preserves the current image URL when editing product details', async () => {
  renderForm({
    id: 'product-1',
    name: 'Produk lama',
    sku: 'SKU-1',
    description: 'Deskripsi',
    price: 12000,
    stock: 2,
    image_url: 'https://images.example.com/product.png',
  });

  fireEvent.click(screen.getByRole('button', { name: 'Simpan produk' }));
  await waitFor(() => expect(api.updateProduct).toHaveBeenCalledWith('product-1', {
    name: 'Produk lama',
    description: 'Deskripsi',
    price: 12000,
    stock: 2,
    category_id: null,
    image_url: 'https://images.example.com/product.png',
  }));
});
