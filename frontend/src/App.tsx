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
