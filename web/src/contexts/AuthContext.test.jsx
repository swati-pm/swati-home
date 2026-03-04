import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider, useAuth } from './AuthContext'

// Test component that exposes auth context
function AuthConsumer() {
  const { user, isAdmin, isLoading, signIn, signOut, getToken } = useAuth()
  return (
    <div>
      <span data-testid="loading">{String(isLoading)}</span>
      <span data-testid="admin">{String(isAdmin)}</span>
      <span data-testid="user">{user ? user.name : 'null'}</span>
      <span data-testid="token">{getToken() || 'null'}</span>
      <button onClick={signIn}>Sign In</button>
      <button onClick={signOut}>Sign Out</button>
    </div>
  )
}

describe('AuthContext', () => {
  beforeEach(() => {
    sessionStorage.clear()
    delete window.google
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('useAuth throws when used outside AuthProvider', () => {
    // Suppress console.error for expected error
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    expect(() => render(<AuthConsumer />)).toThrow('useAuth must be used within AuthProvider')
    consoleSpy.mockRestore()
  })

  it('provides default values when no session exists', () => {
    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    expect(screen.getByTestId('user')).toHaveTextContent('null')
    expect(screen.getByTestId('admin')).toHaveTextContent('false')
    expect(screen.getByTestId('token')).toHaveTextContent('null')
  })

  it('restores session from sessionStorage', () => {
    sessionStorage.setItem(
      'swati-home-auth',
      JSON.stringify({
        email: 'test@example.com',
        name: 'Test User',
        picture: '',
        token: 'stored-token',
      })
    )

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    expect(screen.getByTestId('user')).toHaveTextContent('Test User')
    expect(screen.getByTestId('token')).toHaveTextContent('stored-token')
  })

  it('handles corrupted sessionStorage gracefully', () => {
    sessionStorage.setItem('swati-home-auth', 'not-valid-json')

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    expect(screen.getByTestId('user')).toHaveTextContent('null')
  })

  it('signOut clears user and sessionStorage', async () => {
    sessionStorage.setItem(
      'swati-home-auth',
      JSON.stringify({
        email: 'admin@test.com',
        name: 'Admin',
        picture: '',
        token: 'tok',
      })
    )

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    expect(screen.getByTestId('user')).toHaveTextContent('Admin')

    await userEvent.click(screen.getByText('Sign Out'))

    expect(screen.getByTestId('user')).toHaveTextContent('null')
    expect(sessionStorage.getItem('swati-home-auth')).toBeNull()
  })

  it('signOut calls google disableAutoSelect if available', async () => {
    const disableAutoSelect = vi.fn()
    window.google = { accounts: { id: { disableAutoSelect, initialize: vi.fn() } } }

    sessionStorage.setItem(
      'swati-home-auth',
      JSON.stringify({ email: 'a@t.com', name: 'A', picture: '', token: 't' })
    )

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    await userEvent.click(screen.getByText('Sign Out'))

    expect(disableAutoSelect).toHaveBeenCalled()
  })

  it('signIn calls google prompt if available', async () => {
    const prompt = vi.fn()
    window.google = { accounts: { id: { prompt, initialize: vi.fn() } } }

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    await userEvent.click(screen.getByText('Sign In'))

    expect(prompt).toHaveBeenCalled()
  })

  it('signIn does nothing when google is not available', async () => {
    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    // Should not throw
    await userEvent.click(screen.getByText('Sign In'))
    expect(screen.getByTestId('user')).toHaveTextContent('null')
  })

  it('initializes GSI when google is already available', () => {
    const initialize = vi.fn()
    window.google = { accounts: { id: { initialize, prompt: vi.fn() } } }

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    expect(initialize).toHaveBeenCalledWith(
      expect.objectContaining({
        callback: expect.any(Function),
        auto_select: false,
      })
    )
  })

  it('waits for GSI script to load via interval', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })

    const initialize = vi.fn()

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    // GSI not loaded yet
    expect(initialize).not.toHaveBeenCalled()

    // Simulate GSI loading after a delay
    window.google = { accounts: { id: { initialize, prompt: vi.fn() } } }

    await act(async () => {
      vi.advanceTimersByTime(200)
    })

    expect(initialize).toHaveBeenCalled()

    vi.useRealTimers()
  })

  it('isLoading becomes false after mount', async () => {
    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })
  })
})
