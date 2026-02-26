import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'

export function TracesPage() {
  const [status, setStatus] = useState('')
  const [service, setService] = useState('')
  const timeRange = useMemo(() => {
    const now = new Date()
    return {
      from: new Date(now.getTime() - 24 * 3600 * 1000).toISOString(),
      to: now.toISOString(),
    }
  }, [])

  const q = useQuery({
    queryKey: ['traces', status, service],
    queryFn: async () => {
      const params = new URLSearchParams({ from: timeRange.from, to: timeRange.to, status, service })
      return (await api.get(`/traces?${params.toString()}`)).data.items || []
    },
  })

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">Traces</h2>
      </header>
      <div className="bg-white border border-slate-200 rounded-xl p-4 flex gap-3">
        <select className="border rounded-lg px-3 py-2" value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">Todos status</option>
          <option value="OK">OK</option>
          <option value="ERROR">ERROR</option>
        </select>
        <input className="border rounded-lg px-3 py-2" placeholder="Serviço" value={service} onChange={(e) => setService(e.target.value)} />
      </div>
      <div className="bg-white border border-slate-200 rounded-xl overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-slate-600">
            <tr>
              <th className="text-left p-3">startedAt</th><th className="text-left p-3">method/path</th><th className="text-left p-3">status</th><th className="text-left p-3">duration</th><th className="text-left p-3">service</th>
            </tr>
          </thead>
          <tbody>
            {q.data?.map((t: any) => (
              <tr key={t.traceId} className="border-t hover:bg-slate-50">
                <td className="p-3">{new Date(t.startedAt).toLocaleString()}</td>
                <td className="p-3"><Link className="text-indigo-600 hover:underline" to={`/traces/${t.traceId}`}>{t.httpMethod} {t.httpPath}</Link></td>
                <td className="p-3"><span className={`px-2 py-1 rounded-full text-xs ${t.status === 'ERROR' ? 'bg-red-100 text-red-700' : 'bg-emerald-100 text-emerald-700'}`}>{t.status}</span></td>
                <td className="p-3">{t.durationMs}ms</td>
                <td className="p-3">{t.serviceName}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
