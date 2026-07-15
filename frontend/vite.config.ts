import path from 'path';
import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  // Local `make up` defaults to :8080. Point at the Ubuntu tunnel for remote-backed local UI:
  //   VITE_DEV_BACKEND_URL=https://api-staging-eth-amm-sim.magro.dev npm run dev
  const backend = (env.VITE_DEV_BACKEND_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');

  return {
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
  };
});
