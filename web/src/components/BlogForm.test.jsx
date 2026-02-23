import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import BlogForm from './BlogForm'

describe('BlogForm', () => {
  it('shows "Add Blog Post" title when no blog is provided', () => {
    render(<BlogForm onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Add Blog Post')).toBeInTheDocument()
  })

  it('shows "Edit Blog Post" title when blog is provided', () => {
    const blog = { id: '1', title: 'Test', summary: '', url: '', date: '' }
    render(<BlogForm blog={blog} onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Edit Blog Post')).toBeInTheDocument()
  })

  it('populates form fields when editing', () => {
    const blog = {
      id: '1',
      title: 'My Post',
      summary: 'Summary here',
      url: 'https://example.com',
      date: '2025-01-15',
    }
    render(<BlogForm blog={blog} onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByDisplayValue('My Post')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Summary here')).toBeInTheDocument()
    expect(screen.getByDisplayValue('https://example.com')).toBeInTheDocument()
    expect(screen.getByDisplayValue('2025-01-15')).toBeInTheDocument()
  })

  it('calls onSave with form data on submit', async () => {
    const onSave = vi.fn()
    render(<BlogForm onSave={onSave} onCancel={vi.fn()} />)

    await userEvent.type(screen.getByPlaceholderText(/Building AI Governance/), 'New Post')
    await userEvent.click(screen.getByText('Save'))

    expect(onSave).toHaveBeenCalledOnce()
    expect(onSave.mock.calls[0][0].title).toBe('New Post')
  })

  it('calls onCancel when Cancel is clicked', async () => {
    const onCancel = vi.fn()
    render(<BlogForm onSave={vi.fn()} onCancel={onCancel} />)
    await userEvent.click(screen.getByText('Cancel'))
    expect(onCancel).toHaveBeenCalledOnce()
  })

  it('shows Delete button when editing with onDelete', () => {
    const blog = { id: '1', title: 'Test', summary: '', url: '', date: '' }
    render(<BlogForm blog={blog} onSave={vi.fn()} onCancel={vi.fn()} onDelete={vi.fn()} />)
    expect(screen.getByText('Delete')).toBeInTheDocument()
  })

  it('does not show Delete button when adding new blog', () => {
    render(<BlogForm onSave={vi.fn()} onCancel={vi.fn()} onDelete={vi.fn()} />)
    expect(screen.queryByText('Delete')).not.toBeInTheDocument()
  })

  it('calls onDelete with blog id when Delete is clicked', async () => {
    const onDelete = vi.fn()
    const blog = { id: 'b1', title: 'Test', summary: '', url: '', date: '' }
    render(<BlogForm blog={blog} onSave={vi.fn()} onCancel={vi.fn()} onDelete={onDelete} />)
    await userEvent.click(screen.getByText('Delete'))
    expect(onDelete).toHaveBeenCalledWith('b1')
  })
})
