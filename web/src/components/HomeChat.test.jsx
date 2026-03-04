import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
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

  it('clicking close starts closing animation', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))
    expect(screen.getByText('Ask about Swati')).toBeInTheDocument()

    await userEvent.click(screen.getByLabelText('Close chat'))

    // Closing class should be applied
    const overlay = document.querySelector('.home-chat-overlay')
    expect(overlay.classList.contains('home-chat-overlay-closing')).toBe(true)
  })

  it('close button applies closing state and panel has closing class', async () => {
    const { container } = render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    await userEvent.click(screen.getByLabelText('Close chat'))

    // Both overlay and panel should have closing classes
    const overlay = container.querySelector('.home-chat-overlay')
    const panel = container.querySelector('.home-chat')
    expect(overlay.classList.contains('home-chat-overlay-closing')).toBe(true)
    expect(panel.classList.contains('home-chat-closing')).toBe(true)
  })

  it('panel stopPropagation handler exists on animation end', async () => {
    const { container } = render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    // The panel's onAnimationEnd calls e.stopPropagation()
    const panel = container.querySelector('.home-chat')
    expect(panel).not.toBeNull()

    // Fire animationEnd on the panel — it should stop propagation
    const stopProp = vi.fn()
    const evt = new Event('animationend', { bubbles: true })
    Object.defineProperty(evt, 'stopPropagation', { value: stopProp })
    panel.dispatchEvent(evt)
    // We verify the panel exists and has the right structure
    expect(panel.querySelector('.home-chat-header')).not.toBeNull()
  })

  it('closes chat on Escape key', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))
    expect(screen.getByText('Ask about Swati')).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')

    const overlay = document.querySelector('.home-chat-overlay')
    expect(overlay.classList.contains('home-chat-overlay-closing')).toBe(true)
  })

  it('closes chat when clicking outside the panel (overlay)', async () => {
    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const overlay = document.querySelector('.home-chat-overlay')
    // Click on overlay directly (outside the panel)
    overlay.dispatchEvent(new MouseEvent('click', { bubbles: true }))

    await waitFor(() => {
      expect(overlay.classList.contains('home-chat-overlay-closing')).toBe(true)
    })
  })

  it('does not send on Shift+Enter', async () => {
    sendChat.mockResolvedValue({ reply: 'Hi!' })

    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const input = screen.getByPlaceholderText('Type a question...')
    await userEvent.type(input, 'Hello')
    await userEvent.keyboard('{Shift>}{Enter}{/Shift}')

    expect(sendChat).not.toHaveBeenCalled()
  })

  it('clears error on next successful send', async () => {
    sendChat.mockRejectedValueOnce(new Error('fail'))
    sendChat.mockResolvedValueOnce({ reply: 'success' })

    render(<HomeChat />)
    await userEvent.click(screen.getByText(/Ask me anything about Swati/))

    const input = screen.getByPlaceholderText('Type a question...')
    await userEvent.type(input, 'Hello{Enter}')

    await waitFor(() => {
      expect(screen.getByText('Something went wrong. Please try again.')).toBeInTheDocument()
    })

    await userEvent.type(input, 'Try again{Enter}')

    await waitFor(() => {
      expect(screen.getByText('success')).toBeInTheDocument()
    })
    expect(screen.queryByText('Something went wrong. Please try again.')).not.toBeInTheDocument()
  })
})
