// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';

import { Button } from '@/components/Button.jsx';
import { SelectField, TextAreaField, TextField } from '@/components/FormFields.jsx';
import { sendContactMessage } from './services/contactApi.js';

const MESSAGE_TYPES = [
  { value: 'QUESTION', label: 'Pertanyaan' },
  { value: 'COMPLAINT', label: 'Keluhan' },
];

const initialForm = { name: '', email: '', type: '', message: '' };

export default function ContactUsPage() {
  const [form, setForm] = useState(initialForm);
  const mutation = useMutation({ mutationFn: sendContactMessage });
  const onChange = (event) => setForm((current) => ({ ...current, [event.target.name]: event.target.value }));

  function onSubmit(event) {
    event.preventDefault();
    mutation.mutate(
      { ...form, name: form.name.trim(), email: form.email.trim(), message: form.message.trim() },
      { onSuccess: () => setForm(initialForm) },
    );
  }

  return (
    <div className="mx-auto max-w-2xl">
      <div className="mb-8">
        <p className="text-xs font-bold uppercase tracking-[0.22em] text-brand">Kami siap membantu</p>
        <h1 className="mt-2 text-3xl font-bold tracking-tight">Hubungi kami</h1>
        <p className="mt-3 text-sm leading-6 text-ink-soft">
          Kirim pertanyaan atau keluhan Anda. Tim support kami akan menindaklanjutinya melalui email.
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-5 rounded-xl border border-line bg-white p-5 sm:p-8">
        <TextField label="Nama" name="name" autoComplete="name" maxLength={255} required value={form.name} onChange={onChange} />
        <TextField label="Email" name="email" type="email" autoComplete="email" maxLength={255} required value={form.email} onChange={onChange} />
        <SelectField
          label="Tipe message"
          name="type"
          options={MESSAGE_TYPES}
          placeholder="Pilih tipe message"
          required
          value={form.type}
          onChange={onChange}
        />
        <TextAreaField label="Message" name="message" rows={6} maxLength={5000} required value={form.message} onChange={onChange} />

        {mutation.isError && (
          <p role="alert" className="rounded-md bg-danger-tint p-3 text-sm text-danger">
            Pesan belum berhasil dikirim. {mutation.error.message}
          </p>
        )}
        {mutation.isSuccess && (
          <p role="status" className="rounded-md bg-emerald-50 p-3 text-sm text-emerald-800">
            Pesan berhasil dikirim. Terima kasih telah menghubungi kami.
          </p>
        )}
        <Button type="submit" loading={mutation.isPending}>Kirim pesan</Button>
      </form>
    </div>
  );
}
