import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAuthStore } from '../store/auth'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, val: string) => { store[key] = val },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

const mockUser = {
  id: 'user-1',
  name: 'Admin User',
  email: 'admin@logs.local',
  role: 'admin',
  permissions: ['traces:read', 'users:manage'],
}

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorageMock.clear()
    // Reset store state between tests
    useAuthStore.setState({ token: '', user: null })
  })

  it('starts with empty token and null user', () => {
    const { result } = renderHook(() => useAuthStore())
    expect(result.current.token).toBe('')
    expect(result.current.user).toBeNull()
  })

  it('setAuth stores token and user in state', () => {
    const { result } = renderHook(() => useAuthStore())
    act(() => {
      result.current.setAuth('my-token-123', mockUser)
    })
    expect(result.current.token).toBe('my-token-123')
    expect(result.current.user).toEqual(mockUser)
  })

  it('setAuth persists to localStorage', () => {
    const { result } = renderHook(() => useAuthStore())
    act(() => {
      result.current.setAuth('my-token-123', mockUser)
    })
    expect(localStorageMock.getItem('token')).toBe('my-token-123')
    expect(JSON.parse(localStorageMock.getItem('user')!)).toEqual(mockUser)
  })

  it('logout clears token, user and localStorage', () => {
    const { result } = renderHook(() => useAuthStore())
    act(() => {
      result.current.setAuth('my-token-123', mockUser)
    })
    act(() => {
      result.current.logout()
    })
    expect(result.current.token).toBe('')
    expect(result.current.user).toBeNull()
    expect(localStorageMock.getItem('token')).toBeNull()
    expect(localStorageMock.getItem('user')).toBeNull()
  })
})
