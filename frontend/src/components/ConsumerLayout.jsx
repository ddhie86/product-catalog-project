// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { Link, Outlet, useLocation } from 'react-router-dom';

const navLink = 'text-sm font-medium text-ink-soft transition-colors hover:text-brand';

export default function ConsumerLayout() {
  const location = useLocation();
  const isCatalog = location.pathname === '/';

  return (
    <div className="flex min-h-screen flex-col bg-white">
      <div className="bg-ink py-2 text-center text-[11px] font-medium tracking-wide text-white sm:text-xs">
        TEMUKAN PRODUK FAVORIT ANDA DENGAN MUDAH
      </div>

      <header className="sticky top-0 z-30 border-b border-line/80 bg-white/95 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
          <Link to="/" aria-label="Product Catalog - Beranda" className="shrink-0 text-xl font-bold tracking-tight text-ink sm:text-2xl">
            Product<span className="text-brand">Catalog</span><span className="text-brand">.</span>
          </Link>

          <nav aria-label="Navigasi utama" className="hidden items-center gap-7 md:flex">
            <Link to="/" className={isCatalog ? 'text-sm font-semibold text-brand' : navLink}>Beranda</Link>
            <Link to="/#categories" className={navLink}>Kategori</Link>
            <Link to="/#catalog" className={navLink}>Semua Produk</Link>
          </nav>

          <div className="flex items-center gap-3">
            <Link to="/#catalog-search" aria-label="Cari produk" className="flex h-10 w-10 items-center justify-center rounded-full border border-line text-ink transition hover:border-brand hover:text-brand">
              <svg viewBox="0 0 24 24" aria-hidden="true" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
                <circle cx="10.8" cy="10.8" r="6.8" /><path d="m16 16 4.2 4.2" />
              </svg>
            </Link>
            <Link to="/admin" className="inline-flex rounded-full bg-brand px-3 py-2.5 text-xs font-semibold text-white transition hover:bg-brand-dark sm:px-4">
              Admin
            </Link>
          </div>
        </div>
        <nav aria-label="Navigasi mobile" className="border-t border-line/70 px-4 py-2.5 md:hidden">
          <div className="mx-auto flex max-w-7xl items-center justify-center gap-8 text-xs font-semibold">
            <Link to="/" className={isCatalog ? 'text-brand' : navLink}>Beranda</Link>
            <Link to="/#categories" className={navLink}>Kategori</Link>
            <Link to="/#catalog" className={navLink}>Semua Produk</Link>
          </div>
        </nav>
      </header>

      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-7 sm:px-6 sm:py-10 lg:px-8">
        <Outlet />
      </main>

      <footer className="mt-12 bg-ink text-white">
        <div className="mx-auto grid max-w-7xl gap-8 px-4 py-10 sm:grid-cols-2 sm:px-6 lg:grid-cols-3 lg:px-8">
          <div>
            <Link to="/" className="text-xl font-bold tracking-tight">Product<span className="text-blue-300">Catalog</span>.</Link>
            <p className="mt-3 max-w-xs text-sm leading-6 text-slate-300">
              Temukan pilihan produk berkualitas untuk kebutuhan sehari-hari Anda.
            </p>
          </div>
          <div>
            <h2 className="text-sm font-semibold">Jelajahi</h2>
            <div className="mt-3 flex flex-col items-start gap-2 text-sm text-slate-300">
              <Link to="/" className="hover:text-white">Beranda</Link>
              <Link to="/#categories" className="hover:text-white">Kategori produk</Link>
              <Link to="/#catalog" className="hover:text-white">Semua produk</Link>
            </div>
          </div>
          <div>
            <h2 className="text-sm font-semibold">Informasi</h2>
            <p className="mt-3 text-sm leading-6 text-slate-300">Butuh bantuan? Hubungi tim kami melalui kanal kontak toko.</p>
            <Link to="/admin" className="mt-3 inline-flex text-sm text-blue-300 hover:text-white">Masuk sebagai admin</Link>
          </div>
        </div>
        <div className="border-t border-white/10">
          <p className="mx-auto max-w-7xl px-4 py-4 text-xs text-slate-400 sm:px-6 lg:px-8">
            © {new Date().getFullYear()} Product Catalog. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
