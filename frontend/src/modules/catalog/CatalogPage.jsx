// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { fetchCatalogProducts, fetchCategories } from './services/catalogApi';
import ProductCard from './ProductCard.jsx';
import { Button } from '@/components/Button.jsx';
import { SelectField } from '@/components/FormFields.jsx';
import { Pagination } from '@/components/Pagination.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { SORT_OPTIONS } from '@/lib/constants';

const PAGE_SIZE = 12;

export default function CatalogPage() {
  // Filter disimpan di URL agar hasil pencarian bisa dibagikan & tombol back berfungsi.
  const [searchParams, setSearchParams] = useSearchParams();
  const q = searchParams.get('q') ?? '';
  const category = searchParams.get('category') ?? '';
  const sort = searchParams.get('sort') ?? 'newest';
  const page = Number(searchParams.get('page') ?? 1) || 1;

  const [searchInput, setSearchInput] = useState(q);

  const update = (patch) => {
    const next = new URLSearchParams(searchParams);
    Object.entries({ page: 1, ...patch }).forEach(([k, v]) => (v ? next.set(k, v) : next.delete(k)));
    if (next.get('page') === '1') next.delete('page');
    setSearchParams(next);
  };

  const products = useQuery({
    queryKey: ['catalog', { q, category, sort, page }],
    queryFn: () => fetchCatalogProducts({ q, category, sort, page, limit: PAGE_SIZE }),
    placeholderData: keepPreviousData,
  });
  const categories = useQuery({ queryKey: ['categories'], queryFn: fetchCategories, staleTime: 5 * 60_000 });

  const items = products.data?.data ?? [];

  return (
    <div className="grid gap-8 md:grid-cols-[14rem_1fr]">
      <aside className="space-y-6">
        <form
          role="search"
          onSubmit={(e) => {
            e.preventDefault();
            update({ q: searchInput.trim() });
          }}
          className="space-y-2"
        >
          <label htmlFor="catalog-search" className="block text-sm font-medium">
            Cari produk
          </label>
          <input
            id="catalog-search"
            type="search"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Nama atau kata kunci"
            className="block w-full rounded-md border border-line bg-white px-3 py-2 text-sm"
          />
          <Button type="submit" className="w-full">
            Cari
          </Button>
        </form>

        <fieldset>
          <legend className="mb-2 text-sm font-medium">Kategori</legend>
          <div className="space-y-1 text-sm">
            <label className="flex items-center gap-2">
              <input type="radio" name="category" checked={category === ''} onChange={() => update({ category: '' })} />
              Semua kategori
            </label>
            {(categories.data ?? []).map((c) => (
              <label key={c.id} className="flex items-center gap-2">
                <input type="radio" name="category" checked={category === c.id} onChange={() => update({ category: c.id })} />
                {c.name}
              </label>
            ))}
          </div>
        </fieldset>
      </aside>

      <section aria-live="polite">
        <div className="mb-4 flex flex-wrap items-end justify-between gap-4">
          <h1 className="text-2xl font-semibold">Katalog produk</h1>
          <div className="w-48">
            <SelectField label="Urutkan" value={sort} onChange={(e) => update({ sort: e.target.value })} options={SORT_OPTIONS} />
          </div>
        </div>

        {products.isPending && <LoadingState label="Memuat produk" />}
        {products.isError && <ErrorState error={products.error} onRetry={products.refetch} />}
        {products.isSuccess && items.length === 0 && (
          <EmptyState
            title="Produk tidak ditemukan"
            description="Coba kata kunci lain atau hapus filter kategori."
            action={
              <Button
                variant="secondary"
                onClick={() => {
                  setSearchInput('');
                  setSearchParams({});
                }}
              >
                Hapus semua filter
              </Button>
            }
          />
        )}
        {items.length > 0 && (
          <ul className="grid grid-cols-2 gap-4 lg:grid-cols-3">
            {items.map((p) => (
              <li key={p.id}>
                <ProductCard product={p} />
              </li>
            ))}
          </ul>
        )}
        <Pagination meta={products.data?.meta} onPageChange={(p) => update({ page: p })} />
      </section>
    </div>
  );
}
