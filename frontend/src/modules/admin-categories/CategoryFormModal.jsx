// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createCategory, updateCategory } from './services/adminCategoryApi';
import { Button } from '@/components/Button.jsx';
import { Modal } from '@/components/Modal.jsx';
import { TextField } from '@/components/FormFields.jsx';

function CategoryForm({ category, onClose }) {
  const isEdit = Boolean(category);
  const queryClient = useQueryClient();
  const [name, setName] = useState(category?.name ?? '');
  const [validationError, setValidationError] = useState('');

  const mutation = useMutation({
    mutationFn: (payload) => (isEdit ? updateCategory(category.id, payload) : createCategory(payload)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      queryClient.invalidateQueries({ queryKey: ['admin-products'] });
      onClose();
    },
  });

  function onSubmit(e) {
    e.preventDefault();
    const trimmedName = name.trim();
    if (!trimmedName) {
      setValidationError('Nama kategori wajib diisi.');
      return;
    }
    if (trimmedName.length > 255) {
      setValidationError('Nama kategori maksimal 255 karakter.');
      return;
    }
    setValidationError('');
    mutation.mutate({ name: trimmedName });
  }

  return (
    <form onSubmit={onSubmit} noValidate className="space-y-4">
      <TextField
        label="Nama kategori"
        name="name"
        value={name}
        maxLength={255}
        onChange={(e) => setName(e.target.value)}
        error={validationError}
        autoFocus
      />
      {mutation.isError && (
        <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
          {mutation.error.message}
        </p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button variant="secondary" onClick={onClose}>Batal</Button>
        <Button type="submit" loading={mutation.isPending}>Simpan kategori</Button>
      </div>
    </form>
  );
}

export default function CategoryFormModal({ open, category, onClose }) {
  return (
    <Modal open={open} title={category ? 'Ubah kategori' : 'Tambah kategori'} onClose={onClose}>
      <CategoryForm key={category?.id ?? 'new'} category={category} onClose={onClose} />
    </Modal>
  );
}
