// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Link } from 'react-router-dom';
import { formatRupiah } from '@/lib/format';
import { StockLabel } from './StockLabel.jsx';
import ProductImage from '@/components/ProductImage.jsx';

export default function ProductCard({ product }) {
  return (
    <article className="group flex h-full flex-col">
      <Link
        to={`/products/${product.id}`}
        aria-label={`Lihat detail ${product.name}`}
        className="relative block overflow-hidden rounded-2xl bg-paper"
      >
        <div className="aspect-[4/4.5] overflow-hidden">
          <ProductImage
            src={product.image_url}
            alt={product.name}
            loading="lazy"
            className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
          />
        </div>
        <span className="absolute bottom-3 right-3 rounded-full bg-white/95 px-3 py-1.5 text-[11px] font-semibold text-ink opacity-100 shadow-sm transition sm:translate-y-2 sm:opacity-0 sm:group-hover:translate-y-0 sm:group-hover:opacity-100">
          Lihat detail →
        </span>
      </Link>

      <div className="flex flex-1 flex-col pt-3">
        <h3 className="line-clamp-2 min-h-10 text-sm font-semibold leading-5 text-ink transition group-hover:text-brand sm:text-base">
          <Link to={`/products/${product.id}`}>{product.name}</Link>
        </h3>
        <p className="mt-1.5 text-base font-bold tracking-tight text-ink sm:text-lg">{formatRupiah(product.price)}</p>
        <div className="mt-2">
          <StockLabel stock={product.stock} />
        </div>
      </div>
    </article>
  );
}
