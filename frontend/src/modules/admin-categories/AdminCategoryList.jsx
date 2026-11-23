// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { deleteCategory, getAdminCategories } from './services/adminCategoryApi';
import CategoryFormModal from './CategoryFormModal.jsx';
import { Button } from '@/components/Button.jsx';
import { ConfirmDialog } from '@/components/ConfirmDialog.jsx';
import { PageHeader } from '@/components/PageHeader.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';

export default function AdminCategoryList() {
  const { hasPermission } = useAuth();
  const queryClient = useQueryClient();
  const [formState, setFormState] = useState({ open: false, category: null });
  const [toDelete, setToDelete] = useState(null);
  const list = useQuery({ queryKey: ['categories'], queryFn: getAdminCategories });

  const remove = useMutation({
    mutationFn: deleteCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      setToDelete(null);
    },
  });

  const categories = list.data ?? [];
  const openForm = (category = null) => setFormState({ open: true, category });
  const closeForm = () => setFormState({ open: false, category: null });

  return (
    <>
      <PageHeader
        title="Kategori"
        description="Kelola kategori produk yang tersedia di katalog."
        actions={hasPermission('categories:create') && <Button onClick={() => openForm()}>Tambah kategori</Button>}
      />

      {list.isPending && <LoadingState label="Memuat kategori" />}
      {list.isError && <ErrorState error={list.error} onRetry={list.refetch} />}
      {list.isSuccess && categories.length === 0 && (
        <EmptyState
          title="Belum ada kategori"
          description="Tambahkan kategori agar produk dapat dikelompokkan di katalog."
          action={hasPermission('categories:create') && <Button onClick={() => openForm()}>Tambah kategori</Button>}
        />
      )}

      {categories.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-line bg-white">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">Daftar kategori produk</caption>
            <thead className="border-b border-line bg-paper text-ink-soft">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Nama kategori</th>
                <th scope="col" className="px-4 py-3 font-medium"><span className="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {categories.map((category) => (
                <tr key={category.id}>
                  <td className="px-4 py-3 font-medium">{category.name}</td>
                  <td className="px-4 py-3">
                    <div className="flex justify-end gap-2">
                      {hasPermission('categories:edit') && (
                        <Button variant="secondary" onClick={() => openForm(category)} aria-label={`Ubah ${category.name}`}>
                          Ubah
                        </Button>
                      )}
                      {hasPermission('categories:delete') && (
                        <Button variant="danger" onClick={() => setToDelete(category)} aria-label={`Hapus ${category.name}`}>
                          Hapus
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <CategoryFormModal open={formState.open} category={formState.category} onClose={closeForm} />
      <ConfirmDialog
        open={Boolean(toDelete)}
        title="Hapus kategori?"
        message={toDelete ? `Kategori "${toDelete.name}" akan dihapus. Kategori yang masih digunakan produk tidak dapat dihapus.` : ''}
        error={remove.error?.message}
        loading={remove.isPending}
        onConfirm={() => remove.mutate(toDelete.id)}
        onClose={() => {
          setToDelete(null);
          remove.reset();
        }}
      />
    </>
  );
}
