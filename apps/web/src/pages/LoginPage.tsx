import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import { useAuthStore } from '../store/auth'

export function LoginPage() {
  const [email, setEmail] = useState('admin@logs.local')
  const [password, setPassword] = useState('admin123')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const setToken = useAuthStore((s) => s.setToken)
  const nav = useNavigate()

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const { data } = await api.post('/auth/login', { email, password })
      setToken(data.token)
      nav('/traces')
    } catch {
      setError('Credenciais inválidas')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen grid place-items-center p-4 bg-gradient-to-br from-slate-100 to-slate-200">
      <form onSubmit={onSubmit} className="w-full max-w-md bg-white rounded-2xl shadow-xl border border-slate-200 p-8 space-y-5">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Logs</h1>
          <p className="text-slate-500 mt-1">Visual Log Tracing Platform</p>
        </div>
        <div className="space-y-3">
          <input className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-indigo-200" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email" />
          <input className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-indigo-200" type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="senha" />
        </div>
        {error && <p className="text-red-600 text-sm">{error}</p>}
        <button disabled={loading} className="w-full bg-indigo-600 text-white rounded-lg py-2 font-medium hover:bg-indigo-500 transition disabled:opacity-70">{loading ? 'Entrando...' : 'Entrar'}</button>
      </form>
    </div>
  )
}
