// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import ProductCard from './ProductCard.jsx';
import { Button } from '@/components/Button.jsx';
import { SelectField } from '@/components/FormFields.jsx';
import { Pagination } from '@/components/Pagination.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { SORT_OPTIONS } from '@/lib/constants';
import { fetchCatalogProducts, fetchCategories } from './services/catalogApi';

const PAGE_SIZE = 12;

export default function CatalogPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const q = searchParams.get('q') ?? '';
  const category = searchParams.get('category') ?? '';
  const sort = searchParams.get('sort') ?? 'newest';
  const page = Number(searchParams.get('page') ?? 1) || 1;
  const [searchInput, setSearchInput] = useState(q);

  useEffect(() => setSearchInput(q), [q]);

  const update = (patch) => {
    const next = new URLSearchParams(searchParams);
    Object.entries({ page: 1, ...patch }).forEach(([key, value]) => (value ? next.set(key, value) : next.delete(key)));
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
  const categoryItems = categories.data ?? [];

  return (
    <div className="space-y-12 sm:space-y-16">
      <section className="relative isolate overflow-hidden rounded-3xl bg-paper px-6 py-10 sm:px-10 sm:py-14 lg:px-14">
        <div aria-hidden="true" className="absolute -right-20 -top-28 -z-10 h-80 w-80 rounded-full bg-brand/10 blur-3xl" />
        <div aria-hidden="true" className="absolute bottom-[-7rem] right-1/4 -z-10 h-64 w-64 rounded-full bg-slate-400/15 blur-3xl" />
        <div className="max-w-2xl">
          <p className="text-xs font-bold uppercase tracking-[0.22em] text-brand">Pilihan terbaik untuk Anda</p>
          <h1 className="mt-4 text-3xl font-bold leading-tight tracking-tight text-ink sm:text-4xl lg:text-5xl">
            Temukan produk yang tepat untuk <span className="text-brand">setiap kebutuhan.</span>
          </h1>
          <p className="mt-4 max-w-xl text-sm leading-7 text-ink-soft sm:text-base">
            Jelajahi koleksi pilihan kami, temukan produk favorit, dan nikmati pengalaman belanja yang praktis.
          </p>
          <form
            role="search"
            onSubmit={(event) => {
              event.preventDefault();
              update({ q: searchInput.trim() });
            }}
            className="mt-7 flex max-w-xl items-center gap-2 rounded-xl border border-line bg-white p-2 shadow-sm focus-within:border-brand"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true" className="ml-2 h-5 w-5 shrink-0 text-ink-soft" fill="none" stroke="currentColor" strokeWidth="1.8">
              <circle cx="10.8" cy="10.8" r="6.8" /><path d="m16 16 4.2 4.2" />
            </svg>
            <input
              id="catalog-search"
              type="search"
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder="Cari produk yang Anda inginkan..."
              aria-label="Cari produk"
              className="min-w-0 flex-1 border-0 bg-transparent px-1 py-2 text-sm text-ink outline-none placeholder:text-ink-soft/70 focus:ring-0"
            />
            <Button type="submit" className="rounded-lg px-5">Cari</Button>
          </form>
          <div className="mt-5 flex flex-wrap gap-x-5 gap-y-2 text-xs font-medium text-ink-soft">
            <span className="inline-flex items-center gap-2"><span className="h-1.5 w-1.5 rounded-full bg-ok" /> Pilihan produk berkualitas</span>
            <span className="inline-flex items-center gap-2"><span className="h-1.5 w-1.5 rounded-full bg-brand" /> Belanja mudah dan praktis</span>
          </div>
        </div>
      </section>

      {categoryItems.length > 0 && (
        <section id="categories" className="scroll-mt-28">
          <div className="mb-5 flex items-end justify-between gap-4">
            <div>
              <p className="text-xs font-bold uppercase tracking-[0.18em] text-brand">Jelajahi</p>
              <h2 className="mt-1 text-2xl font-bold tracking-tight text-ink">Belanja berdasarkan kategori</h2>
            </div>
            <button type="button" onClick={() => update({ category: '' })} className="text-sm font-semibold text-brand hover:text-brand-dark">Semua kategori</button>
          </div>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            {categoryItems.slice(0, 8).map((item, index) => (
              <button
                key={item.id}
                type="button"
                onClick={() => update({ category: item.id })}
                aria-pressed={category === item.id}
                className={`group flex min-h-24 items-center justify-between rounded-2xl border px-4 py-4 text-left transition hover:-translate-y-0.5 hover:border-brand hover:bg-brand-tint/40 ${category === item.id ? 'border-brand bg-brand-tint' : 'border-line bg-white'}`}
              >
                <span>
                  <span className="block text-sm font-semibold text-ink">{item.name}</span>
                  <span className="mt-1 block text-xs text-ink-soft">Lihat koleksi</span>
                </span>
                <span className={`flex h-10 w-10 items-center justify-center rounded-xl text-sm font-bold ${index % 2 === 0 ? 'bg-paper text-brand' : 'bg-brand-tint text-brand-dark'}`}>
                  {String(item.name || 'P').slice(0, 1).toUpperCase()}
                </span>
              </button>
            ))}
          </div>
        </section>
      )}

      <section id="catalog" className="scroll-mt-28">
        <div className="mb-5 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.18em] text-brand">Koleksi kami</p>
            <h2 className="mt-1 text-2xl font-bold tracking-tight text-ink">Jelajahi semua produk</h2>
            <p className="mt-1 text-sm text-ink-soft">Temukan pilihan yang sesuai dengan kebutuhan Anda.</p>
          </div>
          <div className="w-full sm:w-52">
            <SelectField label="Urutkan produk" value={sort} onChange={(event) => update({ sort: event.target.value })} options={SORT_OPTIONS} />
          </div>
        </div>

        <div className="mb-5 flex flex-wrap items-center gap-2">
          <button type="button" onClick={() => update({ category: '' })} className={`rounded-full px-4 py-2 text-xs font-semibold transition ${category === '' ? 'bg-ink text-white' : 'border border-line bg-white text-ink-soft hover:border-brand hover:text-brand'}`}>Semua produk</button>
          {categoryItems.slice(0, 5).map((item) => (
            <button key={item.id} type="button" onClick={() => update({ category: item.id })} aria-pressed={category === item.id} className={`rounded-full px-4 py-2 text-xs font-semibold transition ${category === item.id ? 'bg-ink text-white' : 'border border-line bg-white text-ink-soft hover:border-brand hover:text-brand'}`}>
              {item.name}
            </button>
          ))}
        </div>

        {products.isPending && <LoadingState label="Memuat produk" />}
        {products.isError && <ErrorState error={products.error} onRetry={products.refetch} />}
        {products.isSuccess && items.length === 0 && (
          <EmptyState
            title="Produk belum ditemukan"
            description="Coba kata kunci lain atau pilih kategori yang berbeda."
            action={<Button variant="secondary" onClick={() => { setSearchInput(''); setSearchParams({}); }}>Hapus semua filter</Button>}
          />
        )}
        {items.length > 0 && (
          <>
            <div aria-live="polite" className="mb-4 flex items-center justify-between text-xs text-ink-soft">
              <span>{products.data?.meta?.total_items ?? items.length} produk ditemukan</span>
              {products.isFetching && <span>Memperbarui...</span>}
            </div>
            <ul className="flex flex-wrap justify-center gap-x-3 gap-y-8 sm:gap-x-5 sm:gap-y-9">
              {items.map((product) => (
                <li key={product.id} className="w-[calc(50%-6px)] max-w-[280px] xl:w-[calc(25%-15px)]">
                  <ProductCard product={product} />
                </li>
              ))}
            </ul>
          </>
        )}
        <Pagination meta={products.data?.meta} onPageChange={(nextPage) => update({ page: nextPage })} />
      </section>

      <section className="overflow-hidden rounded-3xl bg-ink px-6 py-8 text-white sm:flex sm:items-center sm:justify-between sm:px-10">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-blue-300">Temukan favorit Anda</p>
          <h2 className="mt-2 text-2xl font-bold tracking-tight">Pilihan tepat, belanja lebih mudah.</h2>
          <p className="mt-2 max-w-xl text-sm leading-6 text-slate-300">Jelajahi katalog kami dan temukan produk yang paling sesuai untuk Anda.</p>
        </div>
        <a href="#catalog" className="mt-5 inline-flex shrink-0 items-center justify-center rounded-lg bg-white px-5 py-3 text-sm font-semibold text-ink transition hover:bg-blue-50 sm:mt-0">Jelajahi produk <span className="ml-2">→</span></a>
      </section>
    </div>
  );
}
