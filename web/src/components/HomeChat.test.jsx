import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import HomeChat from './HomeChat'

// Mock scrollIntoView which jsdom doesn't implement
Element.prototype.scrollIntoView = vi.fn()

vi.mock('../api', () => ({
  sendChat: vi.fn(),
}))

import { sendChat } from '../api'

describe('HomeChat', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the trigger with placeholder text', () => {
    render(<HomeChat />)
    expect(screen.getByText(/Ask me anything about Swati/)).toBeInTheDocument()
  })

  it('opens the overlay when trigger is clicked', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    expect(screen.getByText('Ask about Swati')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Type a question...')).toBeInTheDocument()
  })

  it('shows welcome message in the chat panel', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    expect(screen.getByText(/I'll only answer/)).toBeInTheDocument()
  })

  it('sends a message and displays the reply', async () => {
    sendChat.mockResolvedValue({ reply: 'She is a product leader.' })

    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const input = screen.getByPlaceholderText('Type a question...')
    await userEvent.type(input, 'Who is Swati?')
    await userEvent.click(screen.getByText('Send'))

    expect(screen.getByText('Who is Swati?')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('She is a product leader.')).toBeInTheDocument()
    })
  })

  it('shows error message when chat fails', async () => {
    sendChat.mockRejectedValue(new Error('fail'))

    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const input = screen.getByPlaceholderText('Type a question...')
    await userEvent.type(input, 'Hello')
    await userEvent.click(screen.getByText('Send'))

    await waitFor(() => {
      expect(screen.getByText('Something went wrong. Please try again.')).toBeInTheDocument()
    })
  })

  it('does not send empty messages', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const sendBtn = screen.getByText('Send')
    expect(sendBtn).toBeDisabled()
  })

  it('sends message on Enter key', async () => {
    sendChat.mockResolvedValue({ reply: 'Hi!' })

    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const input = screen.getByPlaceholderText('Type a question...')
    await userEvent.type(input, 'Hello{Enter}')

    expect(sendChat).toHaveBeenCalledWith([{ role: 'user', content: 'Hello' }])
  })

  it('renders close button in the chat panel', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    expect(screen.getByLabelText('Close chat')).toBeInTheDocument()
  })
})
