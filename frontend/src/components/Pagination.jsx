import { Button } from './Button.jsx';

/** meta: { page, total_pages, total_items } dari backend. */
export function Pagination({ meta, onPageChange }) {
  if (!meta || meta.total_pages <= 1) return null;
  return (
    <nav aria-label="Halaman" className="mt-6 flex items-center justify-between gap-4 text-sm">
      <span className="text-ink-soft">
        Halaman {meta.page} dari {meta.total_pages} ({meta.total_items} data)
      </span>
      <div className="flex gap-2">
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
