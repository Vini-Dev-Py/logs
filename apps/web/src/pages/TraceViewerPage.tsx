import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import ReactFlow from 'react-flow-renderer'
import { api } from '../api/client'

export function TraceViewerPage(){
  const {traceId=''} = useParams()
  const q = useQuery({queryKey:['trace',traceId], queryFn: async()=> (await api.get(`/traces/${traceId}`)).data})
  const nodes = (q.data?.nodes||[]).map((n:any, i:number)=>({id:n.id,data:{label:`${n.data.label} (${n.data.status})`}, position:{x:(i%4)*220,y:Math.floor(i/4)*120}, style:{border:n.data.status==='ERROR'?'2px solid red':(n.data.durationMs>1000?'2px solid orange':'1px solid #ccc')}}))
  const ann = (q.data?.annotations||[]).map((a:any)=>({id:`ann-${a.id}`,data:{label:`📝 ${a.text}`},position:{x:a.x,y:a.y},style:{background:'#fffae6'}}))
  return <div style={{height:'80vh'}}><h3>Trace {traceId}</h3><ReactFlow nodes={[...nodes,...ann]} edges={q.data?.edges||[]} fitView /></div>
}
