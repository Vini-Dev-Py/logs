import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import { useAuthStore } from '../store/auth'

export function LoginPage() {
  const [email,setEmail]=useState('admin@logs.local'); const [password,setPassword]=useState('admin123'); const [error,setError]=useState('');
  const setToken = useAuthStore((s)=>s.setToken); const nav=useNavigate()
  async function onSubmit(e: React.FormEvent){e.preventDefault(); try{const {data}=await api.post('/auth/login',{email,password}); setToken(data.token); nav('/traces')}catch{setError('Credenciais inválidas')}}
  return <div className='center'><form className='card' onSubmit={onSubmit}><h2>Logs</h2><input value={email} onChange={e=>setEmail(e.target.value)} placeholder='email'/><input type='password' value={password} onChange={e=>setPassword(e.target.value)} placeholder='senha'/><button>Entrar</button>{error&&<p>{error}</p>}</form></div>
}
