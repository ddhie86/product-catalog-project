import { request } from '@/services/apiClient';

/**
 * payload: { customer_name, customer_email, customer_phone, shipping_address,
 *            items: [{ product_id, quantity }] }
 * -> { transaction_id, invoice_number, total_amount, status, payment_url }
 */
export async function processCheckout(payload) {
  const { data } = await request({ method: 'post', url: '/checkout', data: payload });
  return data;
}
