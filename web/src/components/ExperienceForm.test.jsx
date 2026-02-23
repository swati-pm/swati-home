import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ExperienceForm from './ExperienceForm'

describe('ExperienceForm', () => {
  it('shows "Add Experience" title when no experience is provided', () => {
    render(<ExperienceForm onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Add Experience')).toBeInTheDocument()
  })

  it('shows "Edit Experience" title when experience is provided', () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', location: '', startDate: '', endDate: '', bullets: [] }
    render(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Edit Experience')).toBeInTheDocument()
  })

  it('populates form fields when editing', () => {
    const exp = {
      id: '1',
      company: 'ACME Corp',
      role: 'Senior PM',
      location: 'NYC',
      startDate: 'Jan 2020',
      endDate: 'Dec 2023',
      bullets: ['Led team', 'Shipped products'],
    }
    render(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByDisplayValue('ACME Corp')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Senior PM')).toBeInTheDocument()
    expect(screen.getByDisplayValue('NYC')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Jan 2020')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Dec 2023')).toBeInTheDocument()
    const textarea = screen.getByPlaceholderText(/Led cross-functional team/)
    expect(textarea.value).toBe('Led team\nShipped products')
  })

  it('calls onCancel when Cancel button is clicked', async () => {
    const onCancel = vi.fn()
    render(<ExperienceForm onSave={vi.fn()} onCancel={onCancel} />)
    await userEvent.click(screen.getByText('Cancel'))
    expect(onCancel).toHaveBeenCalledOnce()
  })

  it('calls onSave with form data on submit', async () => {
    const onSave = vi.fn()
    render(<ExperienceForm onSave={onSave} onCancel={vi.fn()} />)

    await userEvent.type(screen.getByPlaceholderText('e.g. Google'), 'TestCo')
    await userEvent.type(screen.getByPlaceholderText('e.g. Senior Product Manager'), 'Engineer')
    await userEvent.type(screen.getByPlaceholderText('e.g. London, UK'), 'Berlin')
    await userEvent.click(screen.getByText('Save'))

    expect(onSave).toHaveBeenCalledOnce()
    const saved = onSave.mock.calls[0][0]
    expect(saved.company).toBe('TestCo')
    expect(saved.role).toBe('Engineer')
    expect(saved.location).toBe('Berlin')
    expect(saved.id).toBeDefined()
  })

  it('does not submit without required fields', async () => {
    const onSave = vi.fn()
    render(<ExperienceForm onSave={onSave} onCancel={vi.fn()} />)
    await userEvent.click(screen.getByText('Save'))
    expect(onSave).not.toHaveBeenCalled()
  })

  it('calls onCancel when clicking overlay background', async () => {
    const onCancel = vi.fn()
    const { container } = render(<ExperienceForm onSave={vi.fn()} onCancel={onCancel} />)
    await userEvent.click(container.querySelector('.modal-overlay'))
    expect(onCancel).toHaveBeenCalled()
  })
})
