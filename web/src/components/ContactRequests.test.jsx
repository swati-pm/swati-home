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
})
