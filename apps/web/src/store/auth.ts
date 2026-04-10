import { create } from 'zustand'

type User = { id: string; name: string; email: string; role: string; permissions: string[] };
type AuthState = { 
  token: string; 
  user: User | null;
  setAuth: (t: string, u: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({ 
  token: localStorage.getItem('token') || '', 
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  setAuth: (token, user) => { 
    localStorage.setItem('token', token); 
    localStorage.setItem('user', JSON.stringify(user));
    set({ token, user });
  },
  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    set({ token: '', user: null });
  }
}))
