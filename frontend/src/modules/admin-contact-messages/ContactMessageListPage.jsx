// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useState } from 'react';
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { Button } from '@/components/Button.jsx';
import { Modal } from '@/components/Modal.jsx';
import { PageHeader } from '@/components/PageHeader.jsx';
import { Pagination } from '@/components/Pagination.jsx';
import { EmptyState, ErrorState, LoadingState } from '@/components/States.jsx';
import { TextAreaField } from '@/components/FormFields.jsx';
import { useAuth } from '@/modules/auth/AuthContext.jsx';
import { formatDateTime } from '@/lib/format';
import { fetchContactMessages, replyToContactMessage } from './services/contactMessageApi.js';

export default function ContactMessageListPage() {
  const { hasPermission } = useAuth();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState(null);
  const [reply, setReply] = useState('');
  const messages = useQuery({
    queryKey: ['admin-contact-messages', page],
    queryFn: () => fetchContactMessages({ page, limit: 10 }),
    placeholderData: keepPreviousData,
  });
  const sendReply = useMutation({
    mutationFn: ({ id, message }) => replyToContactMessage(id, message),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-contact-messages'] });
      setSelected(null);
      setReply('');
    },
  });

  const items = messages.data?.data ?? [];
  function closeReply() {
    setSelected(null);
    setReply('');
    sendReply.reset();
  }
  function submitReply(event) {
    event.preventDefault();
    if (!selected || !reply.trim()) return;
    sendReply.mutate({ id: selected.id, message: reply.trim() });
  }

  return (
    <>
      <PageHeader title="Pesan kontak" description="Pantau pesan dari pengunjung dan balas langsung ke email mereka." />

      {messages.isPending && <LoadingState label="Memuat pesan kontak" />}
      {messages.isError && <ErrorState error={messages.error} onRetry={messages.refetch} />}
      {messages.isSuccess && items.length === 0 && (
        <EmptyState title="Belum ada pesan" description="Pesan dari formulir Hubungi kami akan muncul di sini." />
      )}

      {items.length > 0 && (
        <div className="overflow-x-auto rounded-lg border border-line bg-white">
          <table className="w-full min-w-[760px] text-left text-sm">
            <caption className="sr-only">Daftar pesan kontak</caption>
            <thead className="border-b border-line bg-paper text-ink-soft">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Pengirim</th>
                <th scope="col" className="px-4 py-3 font-medium">Tipe / pesan</th>
                <th scope="col" className="px-4 py-3 font-medium">Diterima</th>
                <th scope="col" className="px-4 py-3 font-medium">Status</th>
                {hasPermission('contact_messages:reply') && <th scope="col" className="px-4 py-3"><span className="sr-only">Aksi</span></th>}
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {items.map((item) => (
                <tr key={item.id} className="align-top">
                  <td className="px-4 py-3">
                    <p className="font-medium">{item.name}</p>
                    <p className="mt-1 text-xs text-ink-soft">{item.email}</p>
                  </td>
                  <td className="max-w-lg px-4 py-3">
                    <span className="rounded-full bg-paper px-2.5 py-1 text-xs font-medium">
                      {item.type === 'QUESTION' ? 'Pertanyaan' : 'Keluhan'}
                    </span>
                    <p className="mt-2 max-h-28 overflow-y-auto whitespace-pre-wrap break-words text-ink-soft">{item.message}</p>
                    {item.reply_message && (
                      <p className="mt-2 border-l-2 border-brand pl-3 text-xs text-ink-soft">
                        Balasan terakhir: {item.reply_message}
                      </p>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-ink-soft">{formatDateTime(item.created_at)}</td>
                  <td className="px-4 py-3">
                    <span className={`whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-medium ${item.replied_at ? 'bg-emerald-50 text-emerald-800' : 'bg-amber-50 text-amber-800'}`}>
                      {item.replied_at ? 'Sudah dibalas' : 'Belum dibalas'}
                    </span>
                  </td>
                  {hasPermission('contact_messages:reply') && (
                    <td className="px-4 py-3 text-right">
                      <Button
                        variant="secondary"
                        onClick={() => {
                          sendReply.reset();
                          setSelected(item);
                          setReply('');
                        }}
                      >
                        {item.replied_at ? 'Balas lagi' : 'Balas email'}
                      </Button>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <Pagination meta={messages.data?.meta} onPageChange={setPage} />

      <Modal open={Boolean(selected)} title="Balas pesan" onClose={closeReply}>
        {selected && (
          <form onSubmit={submitReply} className="space-y-4">
            <div className="rounded-md bg-paper p-3 text-sm">
              <p className="font-medium">{selected.name} <span className="font-normal text-ink-soft">({selected.email})</span></p>
              <p className="mt-2 whitespace-pre-wrap text-ink-soft">{selected.message}</p>
            </div>
            <TextAreaField
              label="Isi balasan"
              name="reply"
              rows={5}
              maxLength={5000}
              required
              value={reply}
              onChange={(event) => setReply(event.target.value)}
              error={sendReply.isError ? sendReply.error.message : undefined}
            />
            <p className="text-xs text-ink-soft">Email akan dikirim ke alamat pengirim menggunakan SMTP yang dikonfigurasi.</p>
            <div className="flex justify-end gap-2">
              <Button variant="secondary" onClick={closeReply}>Batal</Button>
              <Button type="submit" loading={sendReply.isPending}>Kirim balasan</Button>
            </div>
          </form>
        )}
      </Modal>
    </>
  );
}
