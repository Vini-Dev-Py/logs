import { motion, AnimatePresence } from 'framer-motion'
import { Settings, Workflow, BarChart2, Sun, Moon, LogOut } from 'lucide-react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../store/auth'
import { useThemeStore } from '../store/theme'

const baseNavLinks = [
  { to: '/dashboard', icon: BarChart2, label: 'nav.dashboard' },
  { to: '/traces', icon: Workflow, label: 'nav.traces' },
]

export function Layout() {
  const { t } = useTranslation()
  const user = useAuthStore(s => s.user)
  const logout = useAuthStore(s => s.logout)
  const { theme, toggleTheme } = useThemeStore()
  const location = useLocation()

  const allLinks = [
    ...baseNavLinks,
    ...(user?.permissions?.includes('users:manage')
      ? [{ to: '/settings/users', icon: Settings, label: 'nav.settings' }]
      : []),
  ]

  return (
    <div className="min-h-screen flex bg-slate-50 dark:bg-slate-950">
      {/* ── Sidebar ── */}
      <motion.aside
        initial={{ x: -20, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ duration: 0.25, ease: 'easeOut' }}
        className="w-60 shrink-0 bg-slate-900 dark:bg-slate-950 border-r border-slate-800 text-white flex flex-col p-4 gap-6"
      >
        {/* Logo */}
        <div className="flex items-center gap-2 px-2 py-1">
          <span className="text-indigo-400 text-lg">◈</span>
          <span className="text-base font-bold tracking-tight">Logs</span>
        </div>

        {/* Nav */}
        <nav className="flex-1 space-y-0.5 relative">
          {allLinks.map(({ to, icon: Icon, label }) => {
            const isActive = location.pathname.startsWith(to)
            return (
              <NavLink
                key={to}
                to={to}
                className="relative flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors duration-150"
              >
                {isActive && (
                  <motion.span
                    layoutId="sidebar-active-pill"
                    className="absolute inset-0 rounded-lg bg-slate-700 dark:bg-slate-800"
                    transition={{ type: 'spring', stiffness: 380, damping: 32 }}
                  />
                )}
                <Icon
                  size={15}
                  className={`relative z-10 ${isActive ? 'text-white' : 'text-slate-400'}`}
                />
                <span className={`relative z-10 ${isActive ? 'text-white' : 'text-slate-400'}`}>
                  {t(label)}
                </span>
              </NavLink>
            )
          })}
        </nav>

        {/* Footer */}
        <div className="space-y-1">
          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            title={theme === 'light' ? 'Dark mode' : 'Light mode'}
            className="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-slate-400 hover:bg-slate-800 transition-colors"
          >
            <AnimatePresence mode="wait" initial={false}>
              <motion.span
                key={theme}
                initial={{ rotate: -90, opacity: 0, scale: 0.8 }}
                animate={{ rotate: 0, opacity: 1, scale: 1 }}
                exit={{ rotate: 90, opacity: 0, scale: 0.8 }}
                transition={{ duration: 0.18 }}
              >
                {theme === 'light' ? <Moon size={15} /> : <Sun size={15} />}
              </motion.span>
            </AnimatePresence>
            <span>{theme === 'light' ? 'Dark mode' : 'Light mode'}</span>
          </button>

          {/* User + logout */}
          {user && (
            <div className="flex items-center gap-2 px-3 py-2 rounded-lg border border-slate-800">
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium text-white truncate">{user.name}</p>
                <p className="text-xs text-slate-500 truncate">{user.email}</p>
              </div>
              <button
                onClick={logout}
                title="Sair"
                className="text-slate-500 hover:text-red-400 transition-colors"
              >
                <LogOut size={14} />
              </button>
            </div>
          )}
        </div>
      </motion.aside>

      {/* ── Main content ── */}
      <main className="flex-1 p-6 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
