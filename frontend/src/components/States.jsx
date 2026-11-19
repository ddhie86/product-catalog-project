import Spinner from './Spinner.jsx';
import { Button } from './Button.jsx';

export function LoadingState({ label = 'Memuat data' }) {
  return (
    <div className="flex items-center justify-center gap-2 py-16 text-ink-soft">
      <Spinner label={label} />
      <span aria-hidden="true">{label}</span>
    </div>
  );
}

export function ErrorState({ error, onRetry }) {
  return (
    <div role="alert" className="rounded-lg border border-danger/30 bg-danger-tint p-6 text-sm">
      <p className="font-semibold text-danger">Data tidak bisa dimuat</p>
      <p className="mt-1 text-ink">{error?.message || 'Terjadi kesalahan tak terduga.'}</p>
      {onRetry && (
        <Button variant="secondary" className="mt-4" onClick={onRetry}>
          Coba lagi
        </Button>
      )}
    </div>
  );
}

export function EmptyState({ title, description, action }) {
  return (
    <div className="rounded-lg border border-dashed border-line bg-white p-10 text-center">
      <p className="font-semibold">{title}</p>
      {description && <p className="mx-auto mt-1 max-w-sm text-sm text-ink-soft">{description}</p>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
