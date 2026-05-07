import { NavLink, Outlet } from 'react-router-dom';

export default function ShellLayout() {
  return (
    <div className="min-h-screen">
      <nav className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3">
          <div className="flex items-center space-x-2">
            <span className="text-xl font-bold text-white">ETH-AMM-SIM</span>
            <span className="rounded bg-blue-600 px-2 py-0.5 text-xs text-white">DEMO</span>
          </div>
          <div className="flex items-center space-x-6">
            <NavLink
              to="/"
              end
              className={({ isActive }) =>
                `${isActive ? 'text-white' : 'text-gray-300'} transition hover:text-white`
              }
            >
              Dashboard
            </NavLink>
            <NavLink
              to="/performance"
              className={({ isActive }) =>
                `${isActive ? 'text-white' : 'text-gray-300'} transition hover:text-white`
              }
            >
              Performance
            </NavLink>
          </div>
        </div>
      </nav>
      <main className="mx-auto max-w-[1800px] px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
