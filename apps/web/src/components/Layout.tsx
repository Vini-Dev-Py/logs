import { motion } from 'framer-motion'
import { Settings, Workflow } from 'lucide-react'
import { Link, NavLink, Outlet } from 'react-router-dom'

export function Layout() {
  return (
    <div className="min-h-screen bg-slate-100 flex">
      <motion.aside
        initial={{ x: -16, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        className="w-72 bg-slate-900 text-white p-5 flex flex-col gap-8"
      >
        <Link to="/traces" className="text-2xl font-semibold tracking-tight">Logs</Link>
        <nav className="space-y-2">
          <NavLink to="/traces" className={({isActive}) => `flex items-center gap-3 px-3 py-2 rounded-lg transition ${isActive ? 'bg-slate-700' : 'hover:bg-slate-800'}`}>
            <Workflow size={18} /> Traces
          </NavLink>
          <button className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-slate-300 hover:bg-slate-800 transition">
            <Settings size={18} /> Settings
          </button>
        </nav>
      </motion.aside>
      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  )
}
