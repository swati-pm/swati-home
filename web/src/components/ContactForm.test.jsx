import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ContactForm from './ContactForm'

vi.mock('../api', () => ({
  submitContact: vi.fn(),
}))

import { submitContact } from '../api'

describe('ContactForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the contact form', () => {
    render(<ContactForm />)
    expect(screen.getByText('Contact Me')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Your name')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('you@example.com')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('What is this about?')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Your message...')).toBeInTheDocument()
  })

  it('submits form and shows success message', async () => {
    submitContact.mockResolvedValue({ id: '1' })

    render(<ContactForm />)

    await userEvent.type(screen.getByPlaceholderText('Your name'), 'Jane Doe')
    await userEvent.type(screen.getByPlaceholderText('you@example.com'), 'jane@example.com')
    await userEvent.type(screen.getByPlaceholderText('What is this about?'), 'Hello')
    await userEvent.type(screen.getByPlaceholderText('Your message...'), 'Test message')
    await userEvent.click(screen.getByText('Send Message'))

    await waitFor(() => {
      expect(screen.getByText('Message Sent!')).toBeInTheDocument()
    })

    expect(submitContact).toHaveBeenCalledWith({
      name: 'Jane Doe',
      email: 'jane@example.com',
      phone: '',
      subject: 'Hello',
      message: 'Test message',
    })
  })

  it('shows error message when submission fails', async () => {
    submitContact.mockRejectedValue(new Error('Network error'))

    render(<ContactForm />)

    await userEvent.type(screen.getByPlaceholderText('Your name'), 'Jane')
    await userEvent.type(screen.getByPlaceholderText('you@example.com'), 'j@e.com')
    await userEvent.type(screen.getByPlaceholderText('What is this about?'), 'Hi')
    await userEvent.type(screen.getByPlaceholderText('Your message...'), 'msg')
    await userEvent.click(screen.getByText('Send Message'))

    await waitFor(() => {
      expect(screen.getByText('Failed to send message. Please try again.')).toBeInTheDocument()
    })
  })

  it('shows Back to Home link after successful submission', async () => {
    submitContact.mockResolvedValue({ id: '1' })

    render(<ContactForm />)

    await userEvent.type(screen.getByPlaceholderText('Your name'), 'Jane')
    await userEvent.type(screen.getByPlaceholderText('you@example.com'), 'j@e.com')
    await userEvent.type(screen.getByPlaceholderText('What is this about?'), 'Hi')
    await userEvent.type(screen.getByPlaceholderText('Your message...'), 'msg')
    await userEvent.click(screen.getByText('Send Message'))

    await waitFor(() => {
      expect(screen.getByText('Back to Home')).toBeInTheDocument()
    })
  })
})
