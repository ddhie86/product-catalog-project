// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createProduct, updateProduct, uploadProductImage } from './services/adminProductApi';
import { fetchCategories } from '@/modules/catalog/services/catalogApi';
import { Button } from '@/components/Button.jsx';
import { Modal } from '@/components/Modal.jsx';
import { SelectField, TextAreaField, TextField } from '@/components/FormFields.jsx';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import ProductImage from '@/components/ProductImage.jsx';

const toForm = (p) => ({
  name: p?.name ?? '',
  sku: p?.sku ?? '',
  description: p?.description ?? '',
  price: p?.price != null ? String(p.price) : '',
  stock: p?.stock != null ? String(p.stock) : '0',
  category_id: p?.category_id ?? '',
  image_url: p?.image_url ?? '',
});

function validate(f, isEdit) {
  const e = {};
  if (!f.name.trim()) e.name = 'Nama produk wajib diisi.';
  if (!isEdit && !f.sku.trim()) e.sku = 'SKU wajib diisi.';
  if (!(Number(f.price) > 0)) e.price = 'Harga harus lebih dari 0.';
  const stock = Number(f.stock);
  if (!Number.isInteger(stock) || stock < 0) e.stock = 'Stok harus bilangan bulat, minimal 0.';
  if (f.image_url && !/^https?:\/\//i.test(f.image_url)) e.image_url = 'URL gambar harus diawali http:// atau https://.';
  return e;
}

function ProductForm({ product, onClose }) {
  const isEdit = Boolean(product);
  const { hasPermission } = useAuth();
  const queryClient = useQueryClient();
  const [form, setForm] = useState(() => toForm(product));
  const [errors, setErrors] = useState({});
  const [selectedFile, setSelectedFile] = useState(null);
  const [localPreview, setLocalPreview] = useState('');
  const [fileError, setFileError] = useState('');
  const [uploadProgress, setUploadProgress] = useState(0);

  const categories = useQuery({ queryKey: ['categories'], queryFn: fetchCategories, staleTime: 5 * 60_000 });

  const uploadMutation = useMutation({
    mutationFn: (file) =>
      uploadProductImage(file, (event) => {
        if (event.total) setUploadProgress(Math.round((event.loaded / event.total) * 100));
      }),
    onSuccess: (result) => {
      setForm((current) => ({ ...current, image_url: result.image_url }));
      setSelectedFile(null);
      setFileError('');
    },
  });

  const canUploadImage = hasPermission('products:create') || hasPermission('products:edit');

  useEffect(() => {
    if (!selectedFile) {
      setLocalPreview('');
      return undefined;
    }
    const objectUrl = URL.createObjectURL(selectedFile);
    setLocalPreview(objectUrl);
    return () => URL.revokeObjectURL(objectUrl);
  }, [selectedFile]);

  const mutation = useMutation({
    mutationFn: (payload) => (isEdit ? updateProduct(product.id, payload) : createProduct(payload)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-products'] });
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      onClose();
    },
  });

  const onChange = (e) => setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  function onSelectImage(event) {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
      setSelectedFile(null);
      setFileError('Pilih gambar JPEG, PNG, atau WebP.');
      event.target.value = '';
      return;
    }
    uploadMutation.reset();
    setUploadProgress(0);
    setFileError('');
    setSelectedFile(file);
  }

  function uploadImage() {
    if (!selectedFile || !canUploadImage) return;
    uploadMutation.mutate(selectedFile);
  }

  function removeImage() {
    setForm((current) => ({ ...current, image_url: '' }));
    setSelectedFile(null);
    setFileError('');
    uploadMutation.reset();
  }

  function onSubmit(e) {
    e.preventDefault();
    const found = validate(form, isEdit);
    setErrors(found);
    if (Object.keys(found).length > 0) return;

    const payload = {
      name: form.name.trim(),
      description: form.description,
      price: Number(form.price),
      stock: Number(form.stock),
      category_id: form.category_id || null,
      image_url: form.image_url.trim(),
    };
    if (!isEdit) payload.sku = form.sku.trim();
    mutation.mutate(payload);
  }

  return (
    <form onSubmit={onSubmit} noValidate className="space-y-4">
      <TextField label="Nama produk" name="name" value={form.name} onChange={onChange} error={errors.name} />
      <TextField label="SKU" name="sku" value={form.sku} onChange={onChange} error={errors.sku} disabled={isEdit} hint={isEdit ? 'SKU tidak dapat diubah.' : undefined} />
      <TextAreaField label="Deskripsi" name="description" value={form.description} onChange={onChange} />
      <div className="grid grid-cols-2 gap-4">
        <TextField label="Harga (Rp)" name="price" type="number" min="0" step="any" value={form.price} onChange={onChange} error={errors.price} />
        <TextField label="Stok" name="stock" type="number" min="0" step="1" value={form.stock} onChange={onChange} error={errors.stock} />
      </div>
      <SelectField
        label="Kategori"
        name="category_id"
        value={form.category_id}
        onChange={onChange}
        placeholder="Tanpa kategori"
        options={(categories.data ?? []).map((c) => ({ value: c.id, label: c.name }))}
      />
      <div className="space-y-2">
        <label htmlFor="product-image-file" className="block text-sm font-medium">Gambar produk</label>
        <input
          id="product-image-file"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          onChange={onSelectImage}
          disabled={!canUploadImage || uploadMutation.isPending || mutation.isPending}
          aria-describedby={fileError || uploadMutation.isError ? 'product-image-error' : 'product-image-hint'}
          className="block w-full rounded-md border border-line bg-white px-3 py-2 text-sm file:mr-3 file:rounded file:border-0 file:bg-paper file:px-3 file:py-1.5 file:text-sm file:font-medium"
        />
        <p id="product-image-hint" className="text-xs text-ink-soft">
          JPEG, PNG, atau WebP. File akan diunggah ke Google Drive sebelum produk disimpan.
        </p>
        {canUploadImage && (
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="secondary"
              onClick={uploadImage}
              disabled={!selectedFile || uploadMutation.isPending || mutation.isPending}
              loading={uploadMutation.isPending}
            >
              Unggah gambar
            </Button>
            {(form.image_url || selectedFile) && (
              <Button variant="secondary" onClick={removeImage} disabled={uploadMutation.isPending || mutation.isPending}>
                Hapus gambar
              </Button>
            )}
          </div>
        )}
        {uploadMutation.isPending && (
          <div className="space-y-1">
            <progress aria-label="Progres upload gambar" max="100" value={uploadProgress} className="h-2 w-full accent-brand" />
            <p className="text-xs text-ink-soft">Mengunggah gambar{uploadProgress > 0 ? ` (${uploadProgress}%)` : '…'}</p>
          </div>
        )}
        {(fileError || uploadMutation.isError) && (
          <p id="product-image-error" role="alert" className="text-sm text-danger">
            {fileError || uploadMutation.error.message}
          </p>
        )}
        {(localPreview || form.image_url) && (
          <div className="flex items-start gap-3">
            <ProductImage
              src={localPreview || form.image_url}
              alt="Pratinjau gambar produk"
              className="h-24 w-24 rounded-md border border-line bg-paper object-cover"
            />
            <p className="break-all text-xs text-ink-soft">{selectedFile?.name || form.image_url}</p>
          </div>
        )}
      </div>

      {mutation.isError && (
        <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
          {mutation.error.message}
        </p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button variant="secondary" onClick={onClose}>
          Batal
        </Button>
        <Button type="submit" loading={mutation.isPending} disabled={uploadMutation.isPending}>
          Simpan produk
        </Button>
      </div>
    </form>
  );
}

/** product = null -> mode tambah; product = objek -> mode edit. */
export default function ProductFormModal({ open, product, onClose }) {
  return (
    <Modal open={open} title={product ? 'Ubah produk' : 'Tambah produk'} onClose={onClose}>
      <ProductForm product={product} onClose={onClose} />
    </Modal>
  );
}
