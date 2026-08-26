// main.tsx - Vite SPA entry: mount App inside BrowserRouter + global styles.
//
// CONNECTIONS:
//  - App routes: App.tsx
//  - Theme/CSS:  globals.css (Tailwind + demo trade-row styles)
//  - Hosted UI:  https://eth-amm-sim.vercel.app

import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './globals.css';

createRoot(document.getElementById('root')!).render(
  <BrowserRouter
    future={{
      v7_startTransition: true,
      v7_relativeSplatPath: true,
    }}
  >
    <App />
  </BrowserRouter>,
);
