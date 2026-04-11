import { AnimatePresence, motion } from "framer-motion";
import {
    BarChart2,
    ChevronLeft,
    ChevronRight,
    LogOut,
    Moon,
    Settings,
    Sun,
    Workflow,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuthStore } from "../store/auth";
import { useThemeStore } from "../store/theme";

const SIDEBAR_COLLAPSED_KEY = "sidebar_collapsed";

function getInitialCollapsedState(): boolean {
  try {
    const saved = localStorage.getItem(SIDEBAR_COLLAPSED_KEY);
    return saved ? JSON.parse(saved) : false;
  } catch {
    return false;
  }
}

interface NavItemConfig {
  to: string;
  icon: React.ElementType;
  label: string;
}

const baseNavLinks: NavItemConfig[] = [
  { to: "/dashboard", icon: BarChart2, label: "nav.dashboard" },
  { to: "/traces", icon: Workflow, label: "nav.traces" },
];

export function Layout() {
  const { t } = useTranslation();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const { theme, toggleTheme } = useThemeStore();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(getInitialCollapsedState);

  // Persist collapsed state
  useEffect(() => {
    try {
      localStorage.setItem(SIDEBAR_COLLAPSED_KEY, JSON.stringify(collapsed));
    } catch {}
  }, [collapsed]);

  const toggleSidebar = () => setCollapsed((prev) => !prev);

  const allLinks: NavItemConfig[] = [
    ...baseNavLinks,
    ...(user?.permissions?.includes("users:manage")
      ? [{ to: "/settings/users", icon: Settings, label: "nav.settings" }]
      : []),
  ];

  const sidebarWidth = collapsed ? "w-16" : "w-64";

  return (
    <div className="h-screen flex bg-slate-50 dark:bg-slate-950 overflow-hidden">
      {/* ── Sidebar ── */}
      <motion.aside
        initial={false}
        animate={{ width: collapsed ? 64 : 256 }}
        transition={{ type: "spring", stiffness: 400, damping: 40 }}
        className={`${sidebarWidth} shrink-0 bg-gradient-to-b from-slate-900 via-slate-900 to-slate-950 dark:from-slate-950 dark:via-slate-950 dark:to-black border-r border-slate-800/50 text-white flex flex-col relative overflow-hidden`}
      >
        {/* Background decoration */}
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_left,_var(--tw-gradient-stops))] from-indigo-500/5 via-transparent to-transparent pointer-events-none" />

        {/* ── Header: Logo + Toggle ── */}
        <div className="relative z-10 flex items-center justify-between px-4 py-5 border-b border-slate-800/50">
          <motion.div
            animate={{ opacity: collapsed ? 0 : 1, x: collapsed ? -10 : 0 }}
            transition={{ duration: 0.2 }}
            className="flex items-center gap-2.5"
          >
            <div className="w-8 h-8 rounded-lg bg-indigo-500/20 border border-indigo-500/30 flex items-center justify-center">
              <span className="text-indigo-400 text-sm font-bold">◈</span>
            </div>
            <span className="text-lg font-bold tracking-tight bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
              Logs
            </span>
          </motion.div>

          {/* Toggle button - always visible */}
          <button
            onClick={toggleSidebar}
            className="absolute -right-3 top-6 w-6 h-6 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-700 transition-colors z-20 shadow-lg"
            title={collapsed ? "Expandir sidebar" : "Recolher sidebar"}
          >
            {collapsed ? (
              <ChevronRight size={12} strokeWidth={2.5} />
            ) : (
              <ChevronLeft size={12} strokeWidth={2.5} />
            )}
          </button>
        </div>

        {/* ── Navigation ── */}
        <nav className="relative z-10 flex-1 px-2.5 py-4 space-y-1 overflow-y-auto scrollbar-thin">
          {allLinks.map(({ to, icon: Icon, label }) => {
            const isActive = location.pathname.startsWith(to);
            return (
              <NavLink
                key={to}
                to={to}
                className="group relative flex items-center"
                title={collapsed ? t(label) : undefined}
              >
                <div
                  className={`relative w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${
                    isActive
                      ? "bg-indigo-500/15 text-white shadow-lg shadow-indigo-500/10"
                      : "text-slate-400 hover:text-white hover:bg-slate-800/60"
                  }`}
                >
                  {/* Active indicator */}
                  {isActive && (
                    <motion.div
                      layoutId="active-indicator"
                      className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 bg-indigo-500 rounded-r-full"
                      transition={{
                        type: "spring",
                        stiffness: 400,
                        damping: 30,
                      }}
                    />
                  )}

                  <Icon
                    size={18}
                    strokeWidth={isActive ? 2.5 : 1.8}
                    className={`shrink-0 transition-colors ${
                      isActive
                        ? "text-indigo-400"
                        : "text-slate-500 group-hover:text-slate-300"
                    }`}
                  />

                  <AnimatePresence initial={false}>
                    {!collapsed && (
                      <motion.span
                        initial={{ opacity: 0, width: 0 }}
                        animate={{ opacity: 1, width: "auto" }}
                        exit={{ opacity: 0, width: 0 }}
                        transition={{ duration: 0.15 }}
                        className="overflow-hidden whitespace-nowrap"
                      >
                        {t(label)}
                      </motion.span>
                    )}
                  </AnimatePresence>
                </div>
              </NavLink>
            );
          })}
        </nav>

        {/* ── Footer ── */}
        <div className="relative z-10 px-2.5 py-3 border-t border-slate-800/50 space-y-1.5">
          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            title={
              collapsed
                ? theme === "light"
                  ? "Dark mode"
                  : "Light mode"
                : undefined
            }
            className="group relative w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-slate-400 hover:text-white hover:bg-slate-800/60 transition-all duration-200"
          >
            <motion.div
              key={theme}
              initial={{ rotate: -90, opacity: 0, scale: 0.7 }}
              animate={{ rotate: 0, opacity: 1, scale: 1 }}
              transition={{ duration: 0.2 }}
            >
              {theme === "light" ? (
                <Moon size={18} strokeWidth={1.8} />
              ) : (
                <Sun size={18} strokeWidth={1.8} />
              )}
            </motion.div>

            <AnimatePresence initial={false}>
              {!collapsed && (
                <motion.span
                  initial={{ opacity: 0, width: 0 }}
                  animate={{ opacity: 1, width: "auto" }}
                  exit={{ opacity: 0, width: 0 }}
                  transition={{ duration: 0.15 }}
                  className="overflow-hidden whitespace-nowrap"
                >
                  {theme === "light" ? "Dark mode" : "Light mode"}
                </motion.span>
              )}
            </AnimatePresence>
          </button>

          {/* User + logout */}
          {user && (
            <div
              className={`flex items-center rounded-xl border border-slate-800/50 bg-slate-800/30 overflow-hidden transition-all duration-200 ${
                collapsed ? "px-2 py-2" : "px-3 py-2.5"
              }`}
            >
              {/* Avatar */}
              <div className="shrink-0 w-7 h-7 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-xs font-bold">
                {user.name.charAt(0).toUpperCase()}
              </div>

              <AnimatePresence initial={false}>
                {!collapsed && (
                  <motion.div
                    initial={{ opacity: 0, width: 0 }}
                    animate={{ opacity: 1, width: "auto" }}
                    exit={{ opacity: 0, width: 0 }}
                    transition={{ duration: 0.15 }}
                    className="overflow-hidden"
                  >
                    <p className="text-xs font-medium text-white truncate ml-2">
                      {user.name}
                    </p>
                    <p className="text-[10px] text-slate-500 truncate ml-2">
                      {user.email}
                    </p>
                  </motion.div>
                )}
              </AnimatePresence>

              <button
                onClick={logout}
                title={collapsed ? "Sair" : undefined}
                className="shrink-0 text-slate-500 hover:text-red-400 transition-colors ml-auto"
              >
                <LogOut size={collapsed ? 14 : 13} strokeWidth={1.8} />
              </button>
            </div>
          )}
        </div>
      </motion.aside>

      {/* ── Main content ── */}
      <main className="flex-1 overflow-auto">
        <div className="p-6 min-h-full">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
