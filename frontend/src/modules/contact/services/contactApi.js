// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { request } from '@/services/apiClient.js';

export async function sendContactMessage(payload) {
  const { data } = await request({ method: 'post', url: '/contact-us', data: payload });
  return data;
}
