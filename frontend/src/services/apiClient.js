import axios from 'axios';

const TOKEN_KEY = 'pc_admin_token';

// Catatan keamanan: localStorage rentan terhadap XSS. Untuk hardening production,
// pertimbangkan httpOnly cookie yang diterbitkan backend.
export const tokenStorage = {
  get: () => localStorage.getItem(TOKEN_KEY),
  set: (t) => localStorage.setItem(TOKEN_KEY, t),
  clear: () => localStorage.removeItem(TOKEN_KEY),
};

export const AUTH_EXPIRED_EVENT = 'auth:expired';

export class ApiError extends Error {
  constructor(message, { status, code } = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15_000,
});

const isAdminUrl = (url = '') => url.startsWith('/admin') && url !== '/admin/login';

apiClient.interceptors.request.use((config) => {
  const token = tokenStorage.get();
  if (token && isAdminUrl(config.url)) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (res) => res,
  (error) => {
    const status = error.response?.status;
    const body = error.response?.data?.error;

    if (status === 401 && isAdminUrl(error.config?.url)) {
      tokenStorage.clear();
      window.dispatchEvent(new Event(AUTH_EXPIRED_EVENT));
    }
    const message = body?.message || (error.request ? 'Tidak dapat terhubung ke server.' : error.message);
    return Promise.reject(new ApiError(message, { status, code: body?.code }));
  },
);

/**
 * Membuka envelope backend { success, data, meta } dan mengembalikan { data, meta }.
 */
export async function request(config) {
  const res = await apiClient.request(config);
  return { data: res.data?.data, meta: res.data?.meta };
}
