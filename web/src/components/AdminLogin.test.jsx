import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdminLogin from './AdminLogin'
import { renderWithAuth } from '../test/helpers'

describe('AdminLogin', () => {
  it('shows sign in prompt when not logged in', () => {
    renderWithAuth(<AdminLogin />)
    expect(screen.getByText('Admin Sign In')).toBeInTheDocument()
    expect(screen.getByText('Sign in with Google')).toBeInTheDocument()
  })

  it('calls signIn when button is clicked', async () => {
    const signIn = vi.fn()
    renderWithAuth(<AdminLogin />, { authValue: { signIn } })

    await userEvent.click(screen.getByText('Sign in with Google'))
    expect(signIn).toHaveBeenCalledOnce()
  })

  it('shows access denied for logged-in non-admin user', () => {
    renderWithAuth(<AdminLogin />, {
      authValue: {
        user: { email: 'other@test.com', name: 'Other' },
        isAdmin: false,
      },
    })
    expect(screen.getByText('Access Denied')).toBeInTheDocument()
  })

  it('shows sign out button for non-admin user', async () => {
    const signOut = vi.fn()
    renderWithAuth(<AdminLogin />, {
      authValue: {
        user: { email: 'other@test.com', name: 'Other' },
        isAdmin: false,
        signOut,
      },
    })

    await userEvent.click(screen.getByText('Sign Out'))
    expect(signOut).toHaveBeenCalledOnce()
  })

  it('returns null while loading', () => {
    const { container } = renderWithAuth(<AdminLogin />, {
      authValue: { isLoading: true },
    })
    expect(container.innerHTML).toBe('')
  })
})
