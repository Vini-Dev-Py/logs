import { create } from 'zustand'

type AuthState = { token: string; setToken: (t: string) => void }
export const useAuthStore = create<AuthState>((set) => ({ token: localStorage.getItem('token') || '', setToken: (token) => { localStorage.setItem('token', token); set({ token }) } }))
