import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Header from './Header'
import { renderWithAuth } from '../test/helpers'

describe('Header', () => {
  it('renders brand name', () => {
    renderWithAuth(<Header currentPage="home" />)
    expect(screen.getByText('Swati Aggarwal')).toBeInTheDocument()
  })

  it('renders navigation links', () => {
    renderWithAuth(<Header currentPage="home" />)
    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Experience')).toBeInTheDocument()
    expect(screen.getByText('Blog')).toBeInTheDocument()
    expect(screen.getByText('Contact')).toBeInTheDocument()
  })

  it('highlights the active page link', () => {
    renderWithAuth(<Header currentPage="experience" />)
    const link = screen.getByText('Experience')
    expect(link).toHaveClass('active')
  })

  it('does not highlight inactive page links', () => {
    renderWithAuth(<Header currentPage="home" />)
    const link = screen.getByText('Experience')
    expect(link).not.toHaveClass('active')
  })

  it('does not show admin-only elements for non-admin users', () => {
    renderWithAuth(<Header currentPage="home" />)
    expect(screen.queryByText('Contact Requests')).not.toBeInTheDocument()
    expect(screen.queryByText('Sign Out')).not.toBeInTheDocument()
  })

  it('shows Contact Requests link and Sign Out for admin', () => {
    renderWithAuth(<Header currentPage="home" />, {
      authValue: { isAdmin: true, signOut: vi.fn() },
    })
    expect(screen.getByText('Contact Requests')).toBeInTheDocument()
    expect(screen.getByText('Sign Out')).toBeInTheDocument()
  })

  it('calls signOut when Sign Out button is clicked', async () => {
    const signOut = vi.fn()
    renderWithAuth(<Header currentPage="home" />, {
      authValue: { isAdmin: true, signOut },
    })

    await userEvent.click(screen.getByText('Sign Out'))
    expect(signOut).toHaveBeenCalledOnce()
  })
})
