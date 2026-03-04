import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ExperienceForm from './ExperienceForm'
import { renderWithAuth } from '../test/helpers'

vi.mock('../api', () => ({
  suggestExperienceBullets: vi.fn(),
}))

import { suggestExperienceBullets } from '../api'

describe('ExperienceForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows "Add Experience" title when no experience is provided', () => {
    renderWithAuth(<ExperienceForm onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Add Experience')).toBeInTheDocument()
  })

  it('shows "Edit Experience" title when experience is provided', () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', location: '', startDate: '', endDate: '', bullets: [] }
    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />)
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
    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />)
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
    renderWithAuth(<ExperienceForm onSave={vi.fn()} onCancel={onCancel} />)
    await userEvent.click(screen.getByText('Cancel'))
    expect(onCancel).toHaveBeenCalledOnce()
  })

  it('calls onSave with form data on submit', async () => {
    const onSave = vi.fn()
    renderWithAuth(<ExperienceForm onSave={onSave} onCancel={vi.fn()} />)

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
    renderWithAuth(<ExperienceForm onSave={onSave} onCancel={vi.fn()} />)
    await userEvent.click(screen.getByText('Save'))
    expect(onSave).not.toHaveBeenCalled()
  })

  it('calls onCancel when clicking overlay background', async () => {
    const onCancel = vi.fn()
    const { container } = renderWithAuth(<ExperienceForm onSave={vi.fn()} onCancel={onCancel} />)
    await userEvent.click(container.querySelector('.modal-overlay'))
    expect(onCancel).toHaveBeenCalled()
  })

  // ---------- Suggest feature tests ----------

  it('does not show suggest button when adding new experience', () => {
    renderWithAuth(<ExperienceForm onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })
    expect(screen.queryByText('Suggest Improvements')).not.toBeInTheDocument()
  })

  it('does not show suggest button for non-admin editing', () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.queryByText('Suggest Improvements')).not.toBeInTheDocument()
  })

  it('shows suggest button for admin editing experience with bullets', () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })
    expect(screen.getByText('Suggest Improvements')).toBeInTheDocument()
  })

  it('shows loading state while suggesting', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    suggestExperienceBullets.mockReturnValue(new Promise(() => {})) // never resolves

    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('Suggest Improvements'))
    expect(screen.getByText('Getting suggestions...')).toBeInTheDocument()
  })

  it('displays suggestions and accepts them', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    suggestExperienceBullets.mockResolvedValue({
      bullets: ['Spearheaded team leadership initiatives'],
    })

    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('Suggest Improvements'))

    await waitFor(() => {
      expect(screen.getByText('Suggested Improvements')).toBeInTheDocument()
    })
    expect(screen.getByText('Spearheaded team leadership initiatives')).toBeInTheDocument()

    await userEvent.click(screen.getByText('Accept'))
    const textarea = screen.getByPlaceholderText(/Led cross-functional team/)
    expect(textarea.value).toBe('Spearheaded team leadership initiatives')
    expect(screen.queryByText('Suggested Improvements')).not.toBeInTheDocument()
  })

  it('dismisses suggestions', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    suggestExperienceBullets.mockResolvedValue({
      bullets: ['Improved bullet'],
    })

    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('Suggest Improvements'))
    await waitFor(() => {
      expect(screen.getByText('Suggested Improvements')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Dismiss'))
    expect(screen.queryByText('Suggested Improvements')).not.toBeInTheDocument()
    const textarea = screen.getByPlaceholderText(/Led cross-functional team/)
    expect(textarea.value).toBe('Led team')
  })

  it('shows error when suggestion fails', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    suggestExperienceBullets.mockRejectedValue(new Error('Service Unavailable'))

    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('Suggest Improvements'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
    expect(screen.getByText('Failed to get suggestions. Please try again.')).toBeInTheDocument()
  })

  it('dismisses suggest error', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    suggestExperienceBullets.mockRejectedValue(new Error('fail'))

    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('Suggest Improvements'))
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('\u00d7'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('shows role description input when toggled', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('+ Add role description'))
    expect(screen.getByPlaceholderText(/Paste a job description/)).toBeInTheDocument()
  })

  it('sends role description with suggest request', async () => {
    const exp = { id: '1', company: 'ACME', role: 'PM', bullets: ['Led team'] }
    suggestExperienceBullets.mockResolvedValue({ bullets: ['Better bullet'] })

    renderWithAuth(<ExperienceForm experience={exp} onSave={vi.fn()} onCancel={vi.fn()} />, {
      authValue: { isAdmin: true },
    })

    await userEvent.click(screen.getByText('+ Add role description'))
    await userEvent.type(
      screen.getByPlaceholderText(/Paste a job description/),
      'Senior PM with data skills'
    )
    await userEvent.click(screen.getByText('Suggest Improvements'))

    await waitFor(() => {
      expect(suggestExperienceBullets).toHaveBeenCalledWith(
        expect.objectContaining({ role_description: 'Senior PM with data skills' }),
        'test-token'
      )
    })
  })
})
