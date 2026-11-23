// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createProduct, updateProduct } from './services/adminProductApi';
import { fetchCategories } from '@/modules/catalog/services/catalogApi';
import { Button } from '@/components/Button.jsx';
import { Modal } from '@/components/Modal.jsx';
import { SelectField, TextAreaField, TextField } from '@/components/FormFields.jsx';

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
  const queryClient = useQueryClient();
  const [form, setForm] = useState(() => toForm(product));
  const [errors, setErrors] = useState({});

  const categories = useQuery({ queryKey: ['categories'], queryFn: fetchCategories, staleTime: 5 * 60_000 });

  const mutation = useMutation({
    mutationFn: (payload) => (isEdit ? updateProduct(product.id, payload) : createProduct(payload)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-products'] });
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      onClose();
    },
  });

  const onChange = (e) => setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

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
      <TextField label="URL gambar" name="image_url" type="url" value={form.image_url} onChange={onChange} error={errors.image_url} />

      {mutation.isError && (
        <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
          {mutation.error.message}
        </p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button variant="secondary" onClick={onClose}>
          Batal
        </Button>
        <Button type="submit" loading={mutation.isPending}>
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
