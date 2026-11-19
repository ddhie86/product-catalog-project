import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchProductById } from './services/catalogApi';
import { StockLabel } from './StockLabel.jsx';
import { Button } from '@/components/Button.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { formatRupiah } from '@/lib/format';

export default function ProductDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [qty, setQty] = useState(1);

  const { data: product, isPending, isError, error, refetch } = useQuery({
    queryKey: ['product', id],
    queryFn: () => fetchProductById(id),
  });

  if (isPending) return <LoadingState label="Memuat produk" />;
  if (isError && error.status === 404) {
    return (
      <EmptyState
        title="Produk tidak ditemukan"
        description="Produk ini mungkin sudah dihapus."
        action={
          <Link to="/" className="text-sm font-medium text-brand">
            Kembali ke katalog
          </Link>
        }
      />
    );
  }
  if (isError) return <ErrorState error={error} onRetry={refetch} />;

  const outOfStock = product.stock <= 0;
  const safeQty = Math.min(Math.max(1, qty), Math.max(product.stock, 1));

  return (
    <article className="grid gap-8 md:grid-cols-2">
      <div className="aspect-square overflow-hidden rounded-lg border border-line bg-white">
        {product.image_url ? (
          <img src={product.image_url} alt={product.name} className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full items-center justify-center text-ink-soft">Tanpa foto</div>
        )}
      </div>

      <div>
        <Link to="/" className="text-sm text-ink-soft hover:text-ink">
          Kembali ke katalog
        </Link>
        <h1 className="mt-2 text-3xl font-semibold">{product.name}</h1>
        <p className="mt-2 text-2xl font-semibold text-brand">{formatRupiah(product.price)}</p>
        <div className="mt-2">
          <StockLabel stock={product.stock} />
        </div>

        <p className="mt-6 max-w-prose whitespace-pre-line text-sm leading-relaxed">{product.description || 'Belum ada deskripsi.'}</p>

        <div className="mt-8 flex items-end gap-3">
          <div>
            <label htmlFor="qty" className="mb-1 block text-sm font-medium">
              Jumlah
            </label>
            <input
              id="qty"
              type="number"
              min={1}
              max={Math.max(product.stock, 1)}
              value={safeQty}
              disabled={outOfStock}
              onChange={(e) => setQty(Number(e.target.value) || 1)}
              className="w-24 rounded-md border border-line bg-white px-3 py-2 text-sm disabled:bg-paper"
            />
          </div>
          <Button disabled={outOfStock} onClick={() => navigate(`/checkout?product=${product.id}&qty=${safeQty}`)}>
            {outOfStock ? 'Stok habis' : 'Beli sekarang'}
          </Button>
        </div>
      </div>
    </article>
  );
}
