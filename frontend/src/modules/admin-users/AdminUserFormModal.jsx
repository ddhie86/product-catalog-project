// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createAdminUser, updateAdminUser } from './services/adminUserApi';
import { getAdminRoles } from '@/modules/admin-roles/services/adminRoleApi';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { Button } from '@/components/Button.jsx';
import { Modal } from '@/components/Modal.jsx';
import { TextField } from '@/components/FormFields.jsx';

function AdminUserForm({ user, onClose }) {
  const isEdit = Boolean(user);
  const queryClient = useQueryClient();
  const { hasPermission, refreshPermissions } = useAuth();
  const canAssignRoles = hasPermission('users:assign_roles');
  const [email, setEmail] = useState(user?.email ?? '');
  const [password, setPassword] = useState('');
  const [roleIDs, setRoleIDs] = useState(() => (user?.roles ?? []).map((role) => role.id));
  const [errors, setErrors] = useState({});
  const rolesQuery = useQuery({ queryKey: ['admin-roles'], queryFn: getAdminRoles, enabled: canAssignRoles });

  const mutation = useMutation({
    mutationFn: (payload) => (isEdit ? updateAdminUser(user.id, payload) : createAdminUser(payload)),
    onSuccess: async () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      if (isEdit) await refreshPermissions({ silent: true });
      onClose();
    },
  });

  function onSubmit(e) {
    e.preventDefault();
    const trimmedEmail = email.trim();
    const validationErrors = {};
    if (!trimmedEmail || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      validationErrors.email = 'Masukkan alamat email yang valid.';
    }
    if ((!isEdit || password) && (password.trim() === '' || password.length < 12 || password.length > 72)) {
      validationErrors.password = 'Password harus berisi 12–72 karakter.';
    }
    if (canAssignRoles && roleIDs.length === 0) validationErrors.roles = 'Pilih minimal satu role.';
    setErrors(validationErrors);
    if (Object.keys(validationErrors).length > 0) {
      return;
    }

    const payload = { email: trimmedEmail };
    if (canAssignRoles) payload.role_ids = roleIDs;
    if (password) payload.password = password;
    mutation.mutate(payload);
  }

  return (
    <form onSubmit={onSubmit} noValidate className="space-y-4">
      <TextField
        label="Email"
        name="email"
        type="email"
        autoComplete="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        error={errors.email}
        autoFocus
      />
      <TextField
        label={isEdit ? 'Password baru (opsional)' : 'Password'}
        name="password"
        type="password"
        autoComplete={isEdit ? 'new-password' : 'new-password'}
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        hint={isEdit ? 'Kosongkan jika tidak ingin mengganti password.' : 'Gunakan 12–72 karakter.'}
        error={errors.password}
      />
      {canAssignRoles && rolesQuery.isPending && <p className="text-sm text-ink-soft">Memuat daftar role…</p>}
      {canAssignRoles && rolesQuery.isError && <p role="alert" className="text-sm text-danger">{rolesQuery.error.message}</p>}
      {canAssignRoles && rolesQuery.isSuccess && (
        <fieldset className="space-y-2">
          <legend className="text-sm font-medium">Role</legend>
          {rolesQuery.data.roles.map((role) => (
            <label key={role.id} className="flex items-start gap-2 rounded-md border border-line p-3 text-sm">
              <input
                type="checkbox"
                checked={roleIDs.includes(role.id)}
                onChange={(e) => setRoleIDs((current) =>
                  e.target.checked ? [...current, role.id] : current.filter((id) => id !== role.id))}
              />
              <span>
                <span className="block font-medium">{role.name}</span>
                {role.is_system && <span className="text-xs text-ink-soft">Role sistem</span>}
              </span>
            </label>
          ))}
          {errors.roles && <p className="text-xs text-danger">{errors.roles}</p>}
        </fieldset>
      )}
      {!canAssignRoles && (
        <p className="text-sm text-ink-soft">
          {isEdit
            ? `Role saat ini: ${(user.roles ?? []).map((role) => role.name).join(', ') || 'tanpa role'}.`
            : 'Akun baru akan menggunakan role bawaan.'}
        </p>
      )}
      {mutation.isError && (
        <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
          {mutation.error.message}
        </p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button variant="secondary" onClick={onClose}>Batal</Button>
        <Button type="submit" loading={mutation.isPending}>Simpan admin</Button>
      </div>
    </form>
  );
}

export default function AdminUserFormModal({ open, user, onClose }) {
  return (
    <Modal open={open} title={user ? 'Ubah admin' : 'Tambah admin'} onClose={onClose}>
      <AdminUserForm key={user?.id ?? 'new'} user={user} onClose={onClose} />
    </Modal>
  );
}
