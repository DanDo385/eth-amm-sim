/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Direct REST origin (bypasses /api proxy). Leave unset on Vercel. */
  readonly VITE_API_URL?: string;
  /** WebSocket URL. Production default is the Ubuntu Cloudflare Tunnel wss URL. */
  readonly VITE_WS_URL?: string;
  /** Vite dev/preview proxy target for /api/* (default http://127.0.0.1:8080). */
  readonly VITE_DEV_BACKEND_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
