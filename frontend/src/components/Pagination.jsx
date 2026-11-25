// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Button } from './Button.jsx';

/** meta: { page, total_pages, total_items } dari backend. */
export function Pagination({ meta, onPageChange }) {
  if (!meta || meta.total_pages <= 1) return null;
  return (
    <nav aria-label="Halaman" className="mt-8 flex flex-col gap-3 text-sm sm:flex-row sm:items-center sm:justify-between sm:gap-4">
      <span className="text-ink-soft">
        Halaman {meta.page} dari {meta.total_pages} ({meta.total_items} data)
      </span>
      <div className="grid grid-cols-2 gap-2 sm:flex">
        <Button variant="secondary" disabled={meta.page <= 1} onClick={() => onPageChange(meta.page - 1)}>
          Sebelumnya
        </Button>
        <Button variant="secondary" disabled={meta.page >= meta.total_pages} onClick={() => onPageChange(meta.page + 1)}>
          Berikutnya
        </Button>
      </div>
    </nav>
  );
}
