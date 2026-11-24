// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
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
        description="Produk ini mungkin sudah tidak tersedia."
        action={<Link to="/" className="text-sm font-semibold text-brand">Kembali ke katalog</Link>}
      />
    );
  }
  if (isError) return <ErrorState error={error} onRetry={refetch} />;

  const outOfStock = product.stock <= 0;
  const safeQty = Math.min(Math.max(1, qty), Math.max(product.stock, 1));

  return (
    <article className="space-y-7">
      <nav aria-label="Breadcrumb" className="flex items-center gap-2 text-xs text-ink-soft">
        <Link to="/" className="transition hover:text-brand">Beranda</Link>
        <span>/</span>
        <span className="font-medium text-ink">Detail produk</span>
      </nav>

      <div className="grid gap-8 lg:grid-cols-2 lg:gap-14">
        <div className="overflow-hidden rounded-3xl bg-paper">
          <div className="aspect-square">
            {product.image_url ? (
              <img src={product.image_url} alt={product.name} className="h-full w-full object-cover" />
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-ink-soft">Foto produk belum tersedia</div>
            )}
          </div>
        </div>

        <div className="flex flex-col py-1 sm:py-3">
          <Link to="/" className="mb-6 inline-flex items-center gap-2 text-sm font-medium text-ink-soft transition hover:text-brand">
            <span aria-hidden="true">←</span> Kembali ke katalog
          </Link>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-brand">Pilihan untuk Anda</p>
          <h1 className="mt-3 text-3xl font-bold leading-tight tracking-tight text-ink sm:text-4xl">{product.name}</h1>
          <p className="mt-5 text-2xl font-bold tracking-tight text-brand sm:text-3xl">{formatRupiah(product.price)}</p>
          <div className="mt-3"><StockLabel stock={product.stock} /></div>

          <div className="my-7 border-y border-line py-6">
            <h2 className="text-sm font-semibold text-ink">Deskripsi produk</h2>
            <p className="mt-3 whitespace-pre-line text-sm leading-7 text-ink-soft">
              {product.description || 'Informasi produk akan segera tersedia.'}
            </p>
          </div>

          <div className="mt-auto space-y-4">
            <div className="flex items-center gap-4">
              <label htmlFor="qty" className="text-sm font-semibold text-ink">Jumlah</label>
              <div className="flex items-center rounded-lg border border-line bg-white">
                <button type="button" aria-label="Kurangi jumlah" disabled={outOfStock || safeQty <= 1} onClick={() => setQty((value) => Math.max(1, value - 1))} className="h-10 w-10 text-lg text-ink-soft transition hover:text-brand disabled:opacity-30">−</button>
                <input
                  id="qty"
                  type="number"
                  min={1}
                  max={Math.max(product.stock, 1)}
                  value={safeQty}
                  disabled={outOfStock}
                  onChange={(event) => setQty(Number(event.target.value) || 1)}
                  className="h-10 w-14 border-x border-line text-center text-sm font-semibold text-ink outline-none disabled:bg-paper"
                />
                <button type="button" aria-label="Tambah jumlah" disabled={outOfStock || safeQty >= product.stock} onClick={() => setQty((value) => Math.min(product.stock, value + 1))} className="h-10 w-10 text-lg text-ink-soft transition hover:text-brand disabled:opacity-30">+</button>
              </div>
            </div>
            <Button disabled={outOfStock} className="w-full rounded-xl py-3.5" onClick={() => navigate(`/checkout?product=${product.id}&qty=${safeQty}`)}>
              {outOfStock ? 'Stok sedang habis' : 'Beli sekarang'}
            </Button>
            <p className="text-center text-xs text-ink-soft">Informasi stok dan harga mengikuti data terbaru dari toko.</p>
          </div>
        </div>
      </div>
    </article>
  );
}
