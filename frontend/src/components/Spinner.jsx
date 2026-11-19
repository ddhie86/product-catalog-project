export default function Spinner({ size = 'md', label = 'Memuat' }) {
  const dim = size === 'sm' ? 'h-4 w-4' : 'h-6 w-6';
  return (
    <span role="status" className="inline-flex items-center">
      <span className={`${dim} animate-spin rounded-full border-2 border-current border-t-transparent`} />
      <span className="sr-only">{label}</span>
    </span>
  );
}
