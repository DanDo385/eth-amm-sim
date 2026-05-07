import path from 'path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

const backend = (process.env.VITE_DEV_BACKEND_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  build: {
    chunkSizeWarningLimit: 1200,
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        target: backend,
        changeOrigin: true,
        rewrite: (p) => {
          const stripped = p.replace(/^\/api/, '');
          return stripped.length ? stripped : '/';
        },
      },
    },
  },
  preview: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        target: backend,
        changeOrigin: true,
        rewrite: (p) => {
          const stripped = p.replace(/^\/api/, '');
          return stripped.length ? stripped : '/';
        },
      },
    },
  },
});
