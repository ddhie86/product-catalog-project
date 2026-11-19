import { useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { deleteProduct, getAdminProducts } from './services/adminProductApi';
import ProductFormModal from './ProductFormModal.jsx';
import { Button } from '@/components/Button.jsx';
import { ConfirmDialog } from '@/components/ConfirmDialog.jsx';
import { PageHeader } from '@/components/PageHeader.jsx';
import { Pagination } from '@/components/Pagination.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { LOW_STOCK_THRESHOLD } from '@/lib/constants';
import { formatRupiah } from '@/lib/format';

export default function AdminProductList() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [q, setQ] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [formState, setFormState] = useState({ open: false, product: null });
  const [toDelete, setToDelete] = useState(null);

  const list = useQuery({
    queryKey: ['admin-products', { q, page }],
    queryFn: () => getAdminProducts({ q, page, limit: 10 }),
    placeholderData: keepPreviousData,
  });

  const remove = useMutation({
    mutationFn: (id) => deleteProduct(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-products'] });
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      setToDelete(null);
    },
  });

  const items = list.data?.data ?? [];
  const openForm = (product = null) => setFormState({ open: true, product });
  const closeForm = () => setFormState({ open: false, product: null });

  return (
    <>
      <PageHeader title="Produk" description="Kelola katalog yang tampil di toko." actions={<Button onClick={() => openForm()}>Tambah produk</Button>} />

      <form
        role="search"
        className="mb-4 flex max-w-md gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1);
          setQ(searchInput.trim());
        }}
      >
        <label htmlFor="admin-product-search" className="sr-only">
          Cari produk
        </label>
        <input
          id="admin-product-search"
          type="search"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          placeholder="Cari nama atau SKU"
          className="block w-full rounded-md border border-line bg-white px-3 py-2 text-sm"
        />
        <Button type="submit" variant="secondary">
          Cari
        </Button>
      </form>

      {list.isPending && <LoadingState label="Memuat produk" />}
      {list.isError && <ErrorState error={list.error} onRetry={list.refetch} />}
      {list.isSuccess && items.length === 0 && (
        <EmptyState title="Belum ada produk" description="Tambahkan produk pertama agar tampil di katalog." action={<Button onClick={() => openForm()}>Tambah produk</Button>} />
      )}

      {items.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-line bg-white">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">Daftar produk</caption>
            <thead className="border-b border-line bg-paper text-ink-soft">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Produk</th>
                <th scope="col" className="px-4 py-3 font-medium">SKU</th>
                <th scope="col" className="px-4 py-3 text-right font-medium">Harga</th>
                <th scope="col" className="px-4 py-3 text-right font-medium">Stok</th>
                <th scope="col" className="px-4 py-3 font-medium"><span className="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {items.map((p) => (
                <tr key={p.id}>
                  <td className="px-4 py-3 font-medium">{p.name}</td>
                  <td className="px-4 py-3 text-ink-soft">{p.sku}</td>
                  <td className="px-4 py-3 text-right">{formatRupiah(p.price)}</td>
                  <td className={`px-4 py-3 text-right ${p.stock < LOW_STOCK_THRESHOLD ? 'font-semibold text-signal' : ''}`}>{p.stock}</td>
                  <td className="px-4 py-3">
                    <div className="flex justify-end gap-2">
                      <Button variant="secondary" onClick={() => openForm(p)} aria-label={`Ubah ${p.name}`}>
                        Ubah
                      </Button>
                      <Button variant="danger" onClick={() => setToDelete(p)} aria-label={`Hapus ${p.name}`}>
                        Hapus
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <Pagination meta={list.data?.meta} onPageChange={setPage} />

      <ProductFormModal open={formState.open} product={formState.product} onClose={closeForm} />
      <ConfirmDialog
        open={Boolean(toDelete)}
        title="Hapus produk?"
        message={toDelete ? `"${toDelete.name}" akan disembunyikan dari katalog. Riwayat transaksinya tetap tersimpan.` : ''}
        loading={remove.isPending}
        onConfirm={() => remove.mutate(toDelete.id)}
        onClose={() => setToDelete(null)}
      />
      {remove.isError && (
        <p role="alert" className="mt-4 text-sm text-danger">
          {remove.error.message}
        </p>
      )}
    </>
  );
}
