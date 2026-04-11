import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { LoginPage } from '../pages/LoginPage'
import { useAuthStore } from '../store/auth'

// Mock i18n — returns key as value for simplicity in tests
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (k: string) => k }),
}))

// Mock the API client
vi.mock('../api/client', () => ({
  api: {
    post: vi.fn(),
  },
}))

import { api } from '../api/client'
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const mockApi = api as unknown as { post: ReturnType<typeof vi.fn> }

const mockUser = {
  id: 'u1',
  name: 'Admin',
  email: 'admin@logs.local',
  role: 'admin',
  permissions: ['users:manage'],
}

function renderLogin() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: '', user: null })
  })

  it('renders email, password fields and submit button', () => {
    renderLogin()
    expect(screen.getByPlaceholderText('login.email')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('login.password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'login.submit' })).toBeInTheDocument()
  })

  it('shows error message when login fails', async () => {
    mockApi.post.mockRejectedValueOnce(new Error('unauthorized'))
    renderLogin()

    fireEvent.click(screen.getByRole('button', { name: 'login.submit' }))

    await waitFor(() => {
      expect(screen.getByText('login.error')).toBeInTheDocument()
    })
  })

  it('calls setAuth with token and user on success', async () => {
    mockApi.post.mockResolvedValueOnce({ data: { token: 'tok-abc', user: mockUser } })
    renderLogin()

    fireEvent.click(screen.getByRole('button', { name: 'login.submit' }))

    await waitFor(() => {
      const { token, user } = useAuthStore.getState()
      expect(token).toBe('tok-abc')
      expect(user).toEqual(mockUser)
    })
  })
})
