import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import BlogCard from './BlogCard'

const mockBlog = {
  id: '1',
  title: 'AI Governance in Product',
  summary: 'How to build responsible AI workflows.',
  url: 'https://example.com/blog',
  date: '2025-01-15',
}

describe('BlogCard', () => {
  it('renders blog title', () => {
    render(<BlogCard blog={mockBlog} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('AI Governance in Product')).toBeInTheDocument()
  })

  it('renders blog summary', () => {
    render(<BlogCard blog={mockBlog} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('How to build responsible AI workflows.')).toBeInTheDocument()
  })

  it('renders formatted date', () => {
    render(<BlogCard blog={mockBlog} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('15 Jan 2025')).toBeInTheDocument()
  })

  it('renders as a link when url is provided', () => {
    render(<BlogCard blog={mockBlog} onEdit={vi.fn()} isAdmin={false} />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', 'https://example.com/blog')
    expect(link).toHaveAttribute('target', '_blank')
  })

  it('renders as div when no url is provided', () => {
    const noUrl = { ...mockBlog, url: '' }
    render(<BlogCard blog={noUrl} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })

  it('does not show Edit button for non-admin', () => {
    render(<BlogCard blog={mockBlog} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.queryByText('Edit')).not.toBeInTheDocument()
  })

  it('shows Edit button for admin', () => {
    render(<BlogCard blog={mockBlog} onEdit={vi.fn()} isAdmin={true} />)
    expect(screen.getByText('Edit')).toBeInTheDocument()
  })

  it('calls onEdit when Edit is clicked', async () => {
    const onEdit = vi.fn()
    render(<BlogCard blog={mockBlog} onEdit={onEdit} isAdmin={true} />)
    await userEvent.click(screen.getByText('Edit'))
    expect(onEdit).toHaveBeenCalledWith(mockBlog)
  })
})
