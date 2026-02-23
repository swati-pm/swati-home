import { render } from '@testing-library/react'
import { AuthContext } from '../contexts/AuthContext'

// Re-export AuthContext so tests can import from here
export { AuthContext }

/**
 * Render a component wrapped with AuthContext.Provider
 */
export function renderWithAuth(ui, { authValue = {}, ...options } = {}) {
  const defaultAuth = {
    user: null,
    isAdmin: false,
    isLoading: false,
    signIn: vi.fn(),
    signOut: vi.fn(),
    getToken: vi.fn(() => 'test-token'),
    ...authValue,
  }

  return render(
    <AuthContext.Provider value={defaultAuth}>
      {ui}
    </AuthContext.Provider>,
    options,
  )
}
