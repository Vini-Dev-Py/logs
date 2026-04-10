import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'
import { useAuthStore } from '../store/auth'

export function LoginPage() {
  const { t } = useTranslation()
  const [email, setEmail] = useState('admin@logs.local')
  const [password, setPassword] = useState('admin123')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const setAuth = useAuthStore((s) => s.setAuth)
  const nav = useNavigate()

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const { data } = await api.post('/auth/login', { email, password })
      setAuth(data.token, data.user)
      nav('/traces')
    } catch {
      setError(t('login.error'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen grid place-items-center p-4 bg-slate-50 dark:bg-slate-950">
      <motion.form
        initial={{ y: 24, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ duration: 0.35, ease: 'easeOut' }}
        onSubmit={onSubmit}
        className="w-full max-w-sm bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-lg p-8 space-y-6"
      >
        <div>
          <div className="flex items-center gap-2 mb-1">
            <span className="text-indigo-500 text-xl">◈</span>
            <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
              {t('login.title')}
            </h1>
          </div>
          <p className="text-sm text-slate-500 dark:text-slate-400">{t('login.subtitle')}</p>
        </div>

        <div className="space-y-3">
          <input
            id="login-email"
            className="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={t('login.email')}
            autoComplete="email"
          />
          <input
            id="login-password"
            className="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white px-3 py-2.5 text-sm outline-none focus:ring-2 focus:ring-indigo-500 transition"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder={t('login.password')}
            autoComplete="current-password"
          />
        </div>

        {error && (
          <motion.p
            initial={{ opacity: 0, y: -4 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-red-500 text-sm"
          >
            {error}
          </motion.p>
        )}

        <motion.button
          id="login-submit"
          type="submit"
          disabled={loading}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          className="w-full py-2.5 rounded-lg font-semibold text-white text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-70 transition-colors"
        >
          {loading ? t('login.submitting') : t('login.submit')}
        </motion.button>
      </motion.form>
    </div>
  )
}
