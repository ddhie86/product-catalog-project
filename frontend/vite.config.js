// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, 'src') },
  },
  server: {
    port: 5173,
    allowedHosts: ['.ngrok-free.dev'],
    // Dev: /api diteruskan ke backend Go sehingga tidak perlu setup CORS.
    proxy: { '/api': 'http://localhost:8080' },
    hmr: {
      clientPort: 443,
    },
  },
  test: {
    environment: 'jsdom',
    environmentOptions: { jsdom: { url: 'http://localhost/' } },
    globals: true,
    setupFiles: './src/test/setup.js',
  },
});
