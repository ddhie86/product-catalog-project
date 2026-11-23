// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Link } from 'react-router-dom';
import { formatRupiah } from '@/lib/format';
import { StockLabel } from './StockLabel.jsx';

export default function ProductCard({ product }) {
  return (
    <Link to={`/products/${product.id}`} className="group flex flex-col overflow-hidden rounded-lg border border-line bg-white hover:border-brand">
      <div className="aspect-square bg-paper">
        {product.image_url ? (
          <img src={product.image_url} alt={product.name} loading="lazy" className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-ink-soft">Tanpa foto</div>
        )}
      </div>
      <div className="flex flex-1 flex-col gap-1 p-4">
        <h3 className="line-clamp-2 font-medium group-hover:text-brand">{product.name}</h3>
        <p className="font-semibold">{formatRupiah(product.price)}</p>
        <div className="mt-auto pt-1">
          <StockLabel stock={product.stock} />
        </div>
      </div>
    </Link>
  );
}
