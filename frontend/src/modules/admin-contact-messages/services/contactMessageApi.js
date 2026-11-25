// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient.js';

export async function fetchContactMessages(params = {}) {
  const { data, meta } = await request({ method: 'get', url: '/admin/contact-messages', params });
  return { data, meta };
}

export async function replyToContactMessage(id, message) {
  const { data } = await request({
    method: 'post',
    url: `/admin/contact-messages/${id}/reply`,
    data: { message },
  });
  return data;
}
