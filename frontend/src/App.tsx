// App.tsx - Top-level React Router tree for the SPA.
//
// Routes nest under ShellLayout (nav chrome). Home loads the live dashboard;
// /performance shows per-account analytics from the Go backend.
//
// CONNECTIONS:
//  - Layout: layout/ShellLayout.tsx
//  - Pages:  pages/HomePage.tsx, pages/PerformancePage.tsx
//  - Mount:  main.tsx (BrowserRouter)

import { Route, Routes } from 'react-router-dom';
import ShellLayout from './layout/ShellLayout';
import HomePage from './pages/HomePage';
import PerformancePage from './pages/PerformancePage';

export default function App() {
  return (
    <Routes>
      <Route element={<ShellLayout />}>
        <Route index element={<HomePage />} />
        <Route path="performance" element={<PerformancePage />} />
      </Route>
    </Routes>
  );
}
