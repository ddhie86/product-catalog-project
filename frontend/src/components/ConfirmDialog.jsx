// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Button } from './Button.jsx';
import { Modal } from './Modal.jsx';

export function ConfirmDialog({ open, title, message, error, confirmLabel = 'Hapus', loading, onConfirm, onClose }) {
  return (
    <Modal open={open} title={title} onClose={onClose}>
      <p className="text-sm text-ink-soft">{message}</p>
      {error && (
        <p role="alert" className="mt-4 rounded-md bg-danger-tint p-3 text-sm text-danger">
          {error}
        </p>
      )}
      <div className="mt-6 flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose}>
          Batal
        </Button>
        <Button variant="danger" loading={loading} onClick={onConfirm}>
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  );
}
