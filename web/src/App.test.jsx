import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { screen } from '@testing-library/react'
import App from './App'
import { renderWithAuth } from './test/helpers'

// Mock all page components to isolate routing logic
vi.mock('./components/Header', () => ({
  default: ({ currentPage }) => <header data-testid="header" data-page={currentPage}>Header</header>,
}))
vi.mock('./components/Home', () => ({
  default: () => <div data-testid="home-page">Home</div>,
}))
vi.mock('./components/ExperienceSection', () => ({
  default: () => <div data-testid="experience-page">Experience</div>,
}))
vi.mock('./components/BlogSection', () => ({
  default: () => <div data-testid="blog-page">Blog</div>,
}))
vi.mock('./components/ContactForm', () => ({
  default: () => <div data-testid="contact-page">Contact</div>,
}))
vi.mock('./components/ContactRequests', () => ({
  default: () => <div data-testid="contact-requests-page">ContactRequests</div>,
}))
vi.mock('./components/AdminLogin', () => ({
  default: () => <div data-testid="admin-page">Admin</div>,
}))

describe('App', () => {
  beforeEach(() => {
    window.location.hash = '#/'
  })

  afterEach(() => {
    window.location.hash = ''
  })

  it('renders the header', () => {
    renderWithAuth(<App />)
    expect(screen.getByTestId('header')).toBeInTheDocument()
  })

  it('renders the footer', () => {
    renderWithAuth(<App />)
    expect(screen.getByText(/Swati Aggarwal/)).toBeInTheDocument()
  })

  it('renders home page by default', () => {
    renderWithAuth(<App />)
    expect(screen.getByTestId('home-page')).toBeInTheDocument()
  })

  it('renders experience page on #/experience hash', () => {
    window.location.hash = '#/experience'
    renderWithAuth(<App />)
    expect(screen.getByTestId('experience-page')).toBeInTheDocument()
  })

  it('renders blog page on #/blog hash', () => {
    window.location.hash = '#/blog'
    renderWithAuth(<App />)
    expect(screen.getByTestId('blog-page')).toBeInTheDocument()
  })

  it('renders contact page on #/contact hash', () => {
    window.location.hash = '#/contact'
    renderWithAuth(<App />)
    expect(screen.getByTestId('contact-page')).toBeInTheDocument()
  })

  it('renders admin page on #/admin hash', () => {
    window.location.hash = '#/admin'
    renderWithAuth(<App />)
    expect(screen.getByTestId('admin-page')).toBeInTheDocument()
  })

  it('renders contact-requests page on #/contact-requests hash', () => {
    window.location.hash = '#/contact-requests'
    renderWithAuth(<App />)
    expect(screen.getByTestId('contact-requests-page')).toBeInTheDocument()
  })
})
