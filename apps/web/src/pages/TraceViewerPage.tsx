import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import ReactFlow from 'react-flow-renderer'
import { api } from '../api/client'

export function TraceViewerPage() {
  const { traceId = '' } = useParams()
  const [note, setNote] = useState('')
  const q = useQuery({ queryKey: ['trace', traceId], queryFn: async () => (await api.get(`/traces/${traceId}`)).data })

  const nodes = useMemo(() => {
    const base = (q.data?.nodes || []).map((n: any, i: number) => ({
      id: n.id,
      data: { label: `${n.data.label} • ${n.data.type}` },
      position: { x: (i % 4) * 260, y: Math.floor(i / 4) * 120 },
      style: { border: n.data.status === 'ERROR' ? '2px solid #dc2626' : (n.data.durationMs > 1000 ? '2px solid #f59e0b' : '1px solid #cbd5e1'), borderRadius: 12, padding: 8, background: 'white' },
    }))
    const anns = (q.data?.annotations || []).map((a: any) => ({ id: `ann-${a.id}`, data: { label: `📝 ${a.text}` }, position: { x: a.x, y: a.y }, style: { background: '#fffbeb', border: '1px solid #f59e0b', borderRadius: 10 } }))
    return [...base, ...anns]
  }, [q.data])

  async function addAnnotation() {
    if (!note.trim()) return
    await api.post(`/traces/${traceId}/annotations`, { nodeId: 'annotation', text: note, x: 80, y: 80 })
    setNote('')
    q.refetch()
  }

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between">
        <h3 className="text-xl font-semibold">Trace {traceId}</h3>
        <div className="flex gap-2">
          <input className="border rounded-lg px-3 py-2" value={note} onChange={(e) => setNote(e.target.value)} placeholder="Add note" />
          <button onClick={addAnnotation} className="px-3 py-2 rounded-lg bg-indigo-600 text-white">Salvar nota</button>
        </div>
      </header>
      <div className="h-[78vh] bg-white border border-slate-200 rounded-xl overflow-hidden">
        <ReactFlow nodes={nodes} edges={q.data?.edges || []} fitView />
      </div>
    </section>
  )
}
