// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery } from '@tanstack/react-query';
import { processCheckout } from './services/checkoutApi';
import { fetchProductById } from '@/modules/catalog/services/catalogApi';
import { Button } from '@/components/Button.jsx';
import { TextAreaField, TextField } from '@/components/FormFields.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { formatRupiah } from '@/lib/format';
import ProductImage from '@/components/ProductImage.jsx';

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function validate(f, maxQty) {
  const e = {};
  if (!f.customer_name.trim()) e.customer_name = 'Nama wajib diisi.';
  if (!EMAIL_RE.test(f.customer_email)) e.customer_email = 'Masukkan email yang valid.';
  if (!/^[0-9+\-\s]{8,20}$/.test(f.customer_phone)) e.customer_phone = 'Masukkan nomor HP yang valid.';
  if (!f.shipping_address.trim()) e.shipping_address = 'Alamat wajib diisi.';
  const qty = Number(f.quantity);
  if (!Number.isInteger(qty) || qty < 1) e.quantity = 'Jumlah minimal 1.';
  else if (qty > maxQty) e.quantity = `Jumlah maksimal ${maxQty} sesuai stok.`;
  return e;
}

export default function CheckoutPage() {
  const [params] = useSearchParams();
  const productId = params.get('product');
  const navigate = useNavigate();

  const product = useQuery({
    queryKey: ['product', productId],
    queryFn: () => fetchProductById(productId),
    enabled: Boolean(productId),
  });

  const [form, setForm] = useState({
    customer_name: '',
    customer_email: '',
    customer_phone: '',
    shipping_address: '',
    quantity: params.get('qty') ?? '1',
  });
  const [errors, setErrors] = useState({});

  const mutation = useMutation({
    mutationFn: processCheckout,
    onSuccess: (res) =>
      navigate('/payment/redirect', {
        replace: true,
        state: { invoiceNumber: res.invoice_number, totalAmount: res.total_amount, paymentUrl: res.payment_url },
      }),
  });

  if (!productId) {
    return (
      <EmptyState
        title="Belum ada produk yang dipilih"
        description="Pilih produk dari katalog, lalu tekan Beli sekarang."
        action={
          <Link to="/" className="text-sm font-medium text-brand">
            Ke katalog
          </Link>
        }
      />
    );
  }
  if (product.isPending) return <LoadingState label="Memuat pesanan" />;
  if (product.isError) return <ErrorState error={product.error} onRetry={product.refetch} />;

  const p = product.data;
  const qty = Number(form.quantity) || 0;
  const onChange = (e) => setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  function onSubmit(e) {
    e.preventDefault();
    const found = validate(form, p.stock);
    setErrors(found);
    if (Object.keys(found).length > 0) return;
    mutation.mutate({
      customer_name: form.customer_name.trim(),
      customer_email: form.customer_email.trim(),
      customer_phone: form.customer_phone.trim(),
      shipping_address: form.shipping_address.trim(),
      items: [{ product_id: p.id, quantity: Number(form.quantity) }],
    });
  }

  return (
    <div className="grid gap-8 md:grid-cols-[1fr_20rem]">
      <form onSubmit={onSubmit} noValidate className="space-y-4">
        <h1 className="text-2xl font-semibold">Data pembeli</h1>
        <TextField label="Nama lengkap" name="customer_name" autoComplete="name" value={form.customer_name} onChange={onChange} error={errors.customer_name} />
        <TextField label="Email" name="customer_email" type="email" autoComplete="email" value={form.customer_email} onChange={onChange} error={errors.customer_email} hint="Konfirmasi pesanan dan tanda terima dikirim ke email ini." />
        <TextField label="Nomor HP" name="customer_phone" type="tel" autoComplete="tel" value={form.customer_phone} onChange={onChange} error={errors.customer_phone} />
        <TextAreaField label="Alamat pengiriman" name="shipping_address" autoComplete="street-address" value={form.shipping_address} onChange={onChange} error={errors.shipping_address} />
        <TextField label="Jumlah" name="quantity" type="number" min={1} max={p.stock} value={form.quantity} onChange={onChange} error={errors.quantity} className="max-w-[8rem]" />

        {mutation.isError && (
          <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
            {mutation.error.message}
          </p>
        )}
        <Button type="submit" loading={mutation.isPending}>
          Buat pesanan dan bayar
        </Button>
      </form>

      <aside className="h-fit rounded-lg border border-line bg-white p-5">
        <h2 className="font-semibold">Ringkasan pesanan</h2>
        <div className="mt-3 flex gap-3 text-sm">
          <div className="h-16 w-16 shrink-0 overflow-hidden rounded bg-paper">
            <ProductImage src={p.image_url} alt={p.name} className="h-full w-full object-cover" />
          </div>
          <div>
            <p className="font-medium">{p.name}</p>
            <p className="text-ink-soft">
              {qty} x {formatRupiah(p.price)}
            </p>
          </div>
        </div>
        <div className="mt-4 flex justify-between border-t border-line pt-4 font-semibold">
          <span>Total</span>
          <span>{formatRupiah(p.price * qty)}</span>
        </div>
      </aside>
    </div>
  );
}
