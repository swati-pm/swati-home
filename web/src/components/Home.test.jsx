import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import Home from './Home'

// Mock HomeChat to isolate Home component
vi.mock('./HomeChat', () => ({
  default: () => <div data-testid="home-chat">MockChat</div>,
}))

describe('Home', () => {
  it('renders greeting with name', () => {
    render(<Home />)
    expect(screen.getByText('Swati')).toBeInTheDocument()
    expect(screen.getByText(/Hi, I'm/)).toBeInTheDocument()
  })

  it('renders the bio summary text', () => {
    render(<Home />)
    expect(screen.getByText(/Product leader with 15\+ years/)).toBeInTheDocument()
  })

  it('renders the HomeChat component', () => {
    render(<Home />)
    expect(screen.getByTestId('home-chat')).toBeInTheDocument()
  })
})
