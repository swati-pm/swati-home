import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import ContactRequests from './ContactRequests'
import { renderWithAuth } from '../test/helpers'

vi.mock('../api', () => ({
  listContacts: vi.fn(),
  deleteContact: vi.fn(),
}))

import { listContacts, deleteContact } from '../api'

describe('ContactRequests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows access denied for non-admin users', () => {
    renderWithAuth(<ContactRequests />)
    expect(screen.getByText(/Access denied/)).toBeInTheDocument()
  })

  it('shows loading state initially for admin', () => {
    listContacts.mockReturnValue(new Promise(() => {})) // never resolves
    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('displays contact requests for admin', async () => {
    listContacts.mockResolvedValue([
      {
        id: 'c1',
        name: 'John',
        email: 'john@test.com',
        phone: '123',
        subject: 'Inquiry',
        message: 'I want to connect',
        created_at: '2025-01-15T10:00:00Z',
      },
    ])

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Inquiry')).toBeInTheDocument()
    })
    expect(screen.getByText('John')).toBeInTheDocument()
    expect(screen.getByText('john@test.com')).toBeInTheDocument()
    expect(screen.getByText('I want to connect')).toBeInTheDocument()
    expect(screen.getByText('1 request')).toBeInTheDocument()
  })

  it('shows empty message when no contacts exist', async () => {
    listContacts.mockResolvedValue([])

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('No contact requests yet.')).toBeInTheDocument()
    })
  })

  it('shows pluralized count for multiple requests', async () => {
    listContacts.mockResolvedValue([
      { id: 'c1', name: 'A', email: 'a@t.com', subject: 'S1', message: 'M1' },
      { id: 'c2', name: 'B', email: 'b@t.com', subject: 'S2', message: 'M2' },
    ])

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('2 requests')).toBeInTheDocument()
    })
  })

  it('deletes a contact when user confirms', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    listContacts
      .mockResolvedValueOnce([
        { id: 'c1', name: 'John', email: 'j@t.com', subject: 'Hi', message: 'Msg' },
      ])
      .mockResolvedValueOnce([])
    deleteContact.mockResolvedValue({})

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('John')).toBeInTheDocument()
    })

    const user = (await import('@testing-library/user-event')).default
    await user.click(screen.getByText('Delete'))

    expect(deleteContact).toHaveBeenCalledWith('c1', 'test-token')
    await waitFor(() => {
      expect(screen.getByText('No contact requests yet.')).toBeInTheDocument()
    })
    window.confirm.mockRestore()
  })

  it('does not delete when user cancels confirm', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    listContacts.mockResolvedValue([
      { id: 'c1', name: 'John', email: 'j@t.com', subject: 'Hi', message: 'Msg' },
    ])

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('John')).toBeInTheDocument()
    })

    const user = (await import('@testing-library/user-event')).default
    await user.click(screen.getByText('Delete'))

    expect(deleteContact).not.toHaveBeenCalled()
    window.confirm.mockRestore()
  })

  it('handles fetch error gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    listContacts.mockRejectedValue(new Error('Network fail'))

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith('Failed to load contacts:', expect.any(Error))
    })
    consoleSpy.mockRestore()
  })

  it('handles delete error gracefully', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    listContacts.mockResolvedValue([
      { id: 'c1', name: 'John', email: 'j@t.com', subject: 'Hi', message: 'Msg' },
    ])
    deleteContact.mockRejectedValue(new Error('Delete failed'))

    renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('John')).toBeInTheDocument()
    })

    const user = (await import('@testing-library/user-event')).default
    await user.click(screen.getByText('Delete'))

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith('Failed to delete contact:', expect.any(Error))
    })
    consoleSpy.mockRestore()
    window.confirm.mockRestore()
  })

  it('shows contact without phone when phone is empty', async () => {
    listContacts.mockResolvedValue([
      { id: 'c1', name: 'Jane', email: 'j@t.com', subject: 'Test', message: 'Msg' },
    ])

    const { container } = renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Jane')).toBeInTheDocument()
    })
    expect(container.querySelector('.cr-card-phone')).toBeNull()
  })

  it('shows contact without created_at date', async () => {
    listContacts.mockResolvedValue([
      { id: 'c1', name: 'Jane', email: 'j@t.com', subject: 'Test', message: 'Msg' },
    ])

    const { container } = renderWithAuth(<ContactRequests />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Jane')).toBeInTheDocument()
    })
    expect(container.querySelector('.cr-card-date')).toBeNull()
  })
})
