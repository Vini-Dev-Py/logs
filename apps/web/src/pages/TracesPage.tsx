import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'

export function TracesPage(){
  const now = new Date(); const from = new Date(now.getTime()-24*3600*1000).toISOString(); const to = now.toISOString()
  const q = useQuery({queryKey:['traces'], queryFn: async()=> (await api.get(`/traces?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`)).data.items || []})
  return <div><h2>Traces</h2><table><thead><tr><th>Início</th><th>Operação</th><th>Status</th><th>Duração</th><th>Serviço</th></tr></thead><tbody>{q.data?.map((t:any)=><tr key={t.traceId}><td>{new Date(t.startedAt).toLocaleString()}</td><td><Link to={`/traces/${t.traceId}`}>{t.httpMethod} {t.httpPath}</Link></td><td>{t.status}</td><td>{t.durationMs}ms</td><td>{t.serviceName}</td></tr>)}</tbody></table></div>
}
