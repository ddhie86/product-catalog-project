import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createRole, updateRole } from './services/adminRoleApi';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { Button } from '@/components/Button.jsx';
import { Modal } from '@/components/Modal.jsx';
import { TextField } from '@/components/FormFields.jsx';

function RoleForm({ role, permissions, onClose }) {
  const isEdit = Boolean(role);
  const queryClient = useQueryClient();
  const { refreshPermissions } = useAuth();
  const [name, setName] = useState(role?.name ?? '');
  const [selected, setSelected] = useState(() => new Set(role?.permissions ?? []));
  const [validationErrors, setValidationErrors] = useState({});

  const mutation = useMutation({
    mutationFn: (payload) => (isEdit ? updateRole(role.id, payload) : createRole(payload)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['admin-roles'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      await refreshPermissions({ silent: true });
      onClose();
    },
  });

  function togglePermission(permission, checked) {
    setSelected((current) => {
      const next = new Set(current);
      if (checked) {
        next.add(permission.key);
        if (permission.action !== 'view') {
          const viewKey = `${permission.menu}:view`;
          if (permissions.some((item) => item.key === viewKey)) next.add(viewKey);
        }
      } else {
        next.delete(permission.key);
        if (permission.action === 'view') {
          for (const item of permissions) {
            if (item.menu === permission.menu) next.delete(item.key);
          }
        }
      }
      return next;
    });
  }

  function onSubmit(e) {
    e.preventDefault();
    const trimmedName = name.trim();
    const errors = {};
    if (!trimmedName) {
      errors.name = 'Nama role wajib diisi.';
    }
    if (selected.size === 0) {
      errors.permissions = 'Pilih minimal satu permission.';
    }
    setValidationErrors(errors);
    if (Object.keys(errors).length > 0) {
      return;
    }
    mutation.mutate({ name: trimmedName, permissions: [...selected] });
  }

  const groups = permissions.reduce((result, permission) => {
    const group = result.find((item) => item.menu === permission.menu);
    if (group) group.permissions.push(permission);
    else result.push({ menu: permission.menu, label: permission.menu_label, permissions: [permission] });
    return result;
  }, []);

  return (
    <form onSubmit={onSubmit} noValidate className="space-y-5">
      <TextField
        label="Nama role"
        name="name"
        value={name}
        maxLength={64}
        onChange={(e) => setName(e.target.value)}
        error={validationErrors.name}
        autoFocus
      />
      {groups.map((group) => (
        <fieldset key={group.menu} className="space-y-2 rounded-md border border-line p-3">
          <legend className="px-1 text-sm font-semibold">{group.label}</legend>
          <div className="grid gap-2 sm:grid-cols-2">
            {group.permissions.map((permission) => (
              <label key={permission.key} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={selected.has(permission.key)}
                  onChange={(e) => togglePermission(permission, e.target.checked)}
                />
                <span>{permission.label}</span>
              </label>
            ))}
          </div>
        </fieldset>
      ))}
      {validationErrors.permissions && <p className="text-sm text-danger">{validationErrors.permissions}</p>}
      {mutation.isError && (
        <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
          {mutation.error.message}
        </p>
      )}
      <div className="flex justify-end gap-2 pt-2">
        <Button variant="secondary" onClick={onClose}>Batal</Button>
        <Button type="submit" loading={mutation.isPending}>Simpan role</Button>
      </div>
    </form>
  );
}

export default function RoleFormModal({ open, role, permissions, onClose }) {
  return (
    <Modal open={open} title={role ? 'Ubah role' : 'Tambah role'} onClose={onClose}>
      <RoleForm key={role?.id ?? 'new'} role={role} permissions={permissions} onClose={onClose} />
    </Modal>
  );
}
