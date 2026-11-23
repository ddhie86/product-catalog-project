// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { deleteRole, getAdminRoles } from './services/adminRoleApi';
import RoleFormModal from './RoleFormModal.jsx';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { Button } from '@/components/Button.jsx';
import { ConfirmDialog } from '@/components/ConfirmDialog.jsx';
import { PageHeader } from '@/components/PageHeader.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';

export default function AdminRoleList() {
  const { hasPermission, refreshPermissions } = useAuth();
  const queryClient = useQueryClient();
  const [formState, setFormState] = useState({ open: false, role: null });
  const [toDelete, setToDelete] = useState(null);
  const list = useQuery({ queryKey: ['admin-roles'], queryFn: getAdminRoles });

  const remove = useMutation({
    mutationFn: deleteRole,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['admin-roles'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      await refreshPermissions({ silent: true });
      setToDelete(null);
    },
  });

  const roles = list.data?.roles ?? [];
  const permissions = list.data?.permissions ?? [];
  const openForm = (role = null) => setFormState({ open: true, role });
  const closeForm = () => setFormState({ open: false, role: null });

  return (
    <>
      <PageHeader
        title="Role dan permission"
        description="Atur akses menu dan aksi untuk setiap role admin."
        actions={hasPermission('roles:create') && <Button onClick={() => openForm()}>Tambah role</Button>}
      />

      {list.isPending && <LoadingState label="Memuat role" />}
      {list.isError && <ErrorState error={list.error} onRetry={list.refetch} />}
      {list.isSuccess && roles.length === 0 && (
        <EmptyState
          title="Belum ada role"
          description="Buat role untuk menentukan hak akses admin."
          action={hasPermission('roles:create') && <Button onClick={() => openForm()}>Tambah role</Button>}
        />
      )}

      {roles.length > 0 && (
        <div className="space-y-4">
          {roles.map((role) => (
            <section key={role.id} className="rounded-lg border border-line bg-white p-4 sm:p-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="font-semibold">{role.name}</h2>
                  {role.is_system && <p className="mt-1 text-xs text-ink-soft">Role sistem — tidak dapat diubah atau dihapus.</p>}
                  {role.is_default && <p className="mt-1 text-xs text-ink-soft">Role bawaan — tidak dapat dihapus.</p>}
                </div>
                <div className="flex gap-2">
                  {hasPermission('roles:edit') && !role.is_system && (
                    <Button variant="secondary" onClick={() => openForm(role)} aria-label={`Ubah ${role.name}`}>
                      Ubah
                    </Button>
                  )}
                  {hasPermission('roles:delete') && !role.is_system && !role.is_default && (
                    <Button variant="danger" onClick={() => setToDelete(role)} aria-label={`Hapus ${role.name}`}>
                      Hapus
                    </Button>
                  )}
                </div>
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                {role.permissions.map((permission) => (
                  <span key={permission} className="rounded-full bg-paper px-2.5 py-1 text-xs text-ink-soft">
                    {permission}
                  </span>
                ))}
                {role.permissions.length === 0 && <span className="text-sm text-ink-soft">Tidak memiliki permission.</span>}
              </div>
            </section>
          ))}
        </div>
      )}

      <RoleFormModal
        open={formState.open}
        role={formState.role}
        permissions={permissions}
        onClose={closeForm}
      />
      <ConfirmDialog
        open={Boolean(toDelete)}
        title="Hapus role?"
        message={toDelete ? `Role "${toDelete.name}" akan dihapus.` : ''}
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
