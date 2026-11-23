// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { LOW_STOCK_THRESHOLD } from '@/lib/constants';

export function StockLabel({ stock }) {
  if (stock <= 0) return <span className="text-sm font-medium text-danger">Stok habis</span>;
  if (stock < LOW_STOCK_THRESHOLD) return <span className="text-sm font-medium text-signal">Sisa {stock}</span>;
  return <span className="text-sm text-ink-soft">Stok tersedia</span>;
}
