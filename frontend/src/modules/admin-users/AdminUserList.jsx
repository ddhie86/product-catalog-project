// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { deleteAdminUser, getAdminUsers } from './services/adminUserApi';
import AdminUserFormModal from './AdminUserFormModal.jsx';
import { Button } from '@/components/Button.jsx';
import { ConfirmDialog } from '@/components/ConfirmDialog.jsx';
import { PageHeader } from '@/components/PageHeader.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';

const dateFormatter = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' });

export default function AdminUserList() {
  const { hasPermission } = useAuth();
  const queryClient = useQueryClient();
  const [formState, setFormState] = useState({ open: false, user: null });
  const [toDelete, setToDelete] = useState(null);
  const list = useQuery({ queryKey: ['admin-users'], queryFn: getAdminUsers });

  const remove = useMutation({
    mutationFn: deleteAdminUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setToDelete(null);
    },
  });

  const users = list.data ?? [];
  const openForm = (user = null) => setFormState({ open: true, user });
  const closeForm = () => setFormState({ open: false, user: null });

  return (
    <>
      <PageHeader
        title="Admin"
        description="Kelola akun yang dapat mengakses panel administrasi."
        actions={hasPermission('users:create') && <Button onClick={() => openForm()}>Tambah admin</Button>}
      />

      {list.isPending && <LoadingState label="Memuat akun admin" />}
      {list.isError && <ErrorState error={list.error} onRetry={list.refetch} />}
      {list.isSuccess && users.length === 0 && (
        <EmptyState
          title="Belum ada akun admin"
          description="Tambahkan akun admin untuk memberikan akses ke panel."
          action={hasPermission('users:create') && <Button onClick={() => openForm()}>Tambah admin</Button>}
        />
      )}

      {users.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-line bg-white">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">Daftar akun admin</caption>
            <thead className="border-b border-line bg-paper text-ink-soft">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Email</th>
                <th scope="col" className="px-4 py-3 font-medium">Dibuat</th>
                <th scope="col" className="px-4 py-3 font-medium"><span className="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {users.map((user) => (
                <tr key={user.id}>
                  <td className="px-4 py-3">
                    <p className="font-medium">{user.email}</p>
                    <p className="mt-1 text-xs text-ink-soft">{(user.roles ?? []).map((role) => role.name).join(', ') || 'Tanpa role'}</p>
                  </td>
                  <td className="px-4 py-3 text-ink-soft">
                    {user.created_at ? dateFormatter.format(new Date(user.created_at)) : '—'}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex justify-end gap-2">
                      {hasPermission('users:edit') && (
                        <Button variant="secondary" onClick={() => openForm(user)} aria-label={`Ubah ${user.email}`}>
                          Ubah
                        </Button>
                      )}
                      {hasPermission('users:delete') && (
                        <Button variant="danger" onClick={() => setToDelete(user)} aria-label={`Hapus ${user.email}`}>
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

      <AdminUserFormModal open={formState.open} user={formState.user} onClose={closeForm} />
      <ConfirmDialog
        open={Boolean(toDelete)}
        title="Hapus akun admin?"
        message={toDelete ? `Akun "${toDelete.email}" akan dihapus. Akun yang sedang digunakan atau admin terakhir tidak dapat dihapus.` : ''}
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
