import { Link, Outlet } from 'react-router-dom';

export default function ConsumerLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-line bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-4">
          <Link to="/" className="text-lg font-semibold text-brand">
            Product Catalog
          </Link>
          <Link to="/admin" className="text-sm text-ink-soft hover:text-ink">
            Masuk sebagai admin
          </Link>
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
        <Outlet />
      </main>
      <footer className="border-t border-line py-6 text-center text-sm text-ink-soft">Product Catalog</footer>
    </div>
  );
}
