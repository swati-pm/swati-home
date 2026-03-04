import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ResumeTemplates from './ResumeTemplates'
import { renderWithAuth } from '../test/helpers'

vi.mock('../api', () => ({
  listTemplates: vi.fn(),
  setActiveTemplate: vi.fn(),
  getResumeDownloadURL: vi.fn(() => '/api/resume/download'),
}))

import { listTemplates, setActiveTemplate } from '../api'

describe('ResumeTemplates', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows access denied for non-admin users', () => {
    renderWithAuth(<ResumeTemplates />)
    expect(screen.getByText(/Access denied/)).toBeInTheDocument()
  })

  it('shows loading state initially for admin', () => {
    listTemplates.mockReturnValue(new Promise(() => {})) // never resolves
    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('displays template cards for admin', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional layout', active: true },
      { name: 'modern', description: 'Contemporary design', active: false },
      { name: 'compact', description: 'Dense layout', active: false },
    ])

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('classic')).toBeInTheDocument()
    })
    expect(screen.getByText('modern')).toBeInTheDocument()
    expect(screen.getByText('compact')).toBeInTheDocument()
    expect(screen.getByText('Active')).toBeInTheDocument()
    expect(screen.getByText('Traditional layout')).toBeInTheDocument()
  })

  it('shows Set Active buttons for inactive templates', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional', active: true },
      { name: 'modern', description: 'Modern', active: false },
    ])

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('classic')).toBeInTheDocument()
    })

    const buttons = screen.getAllByText('Set Active')
    expect(buttons).toHaveLength(1)
  })

  it('calls setActiveTemplate when Set Active is clicked', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional', active: true },
      { name: 'modern', description: 'Modern', active: false },
    ])
    setActiveTemplate.mockResolvedValue({ active_template: 'modern' })

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Set Active')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Set Active'))
    expect(setActiveTemplate).toHaveBeenCalledWith('modern', 'test-token')
  })

  it('shows Download PDF button for admin', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional', active: true },
    ])

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Download PDF')).toBeInTheDocument()
    })
  })

  it('shows error when loading templates fails', async () => {
    listTemplates.mockRejectedValue(new Error('Network error'))

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
    expect(screen.getByText('Failed to load templates. Please try again.')).toBeInTheDocument()
  })

  it('shows error when setting active template fails', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional', active: true },
      { name: 'modern', description: 'Modern', active: false },
    ])
    setActiveTemplate.mockRejectedValue(new Error('Server error'))

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Set Active')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Set Active'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
    expect(screen.getByText('Failed to update template. Please try again.')).toBeInTheDocument()
  })

  it('shows error when download fails', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional', active: true },
    ])

    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      statusText: 'Bad Gateway',
    })

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Download PDF')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Download PDF'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
    expect(screen.getByText('Failed to download resume. Please try again.')).toBeInTheDocument()
  })

  it('dismisses error when close button is clicked', async () => {
    listTemplates.mockRejectedValue(new Error('Network error'))

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('\u00d7'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('downloads PDF successfully', async () => {
    listTemplates.mockResolvedValue([
      { name: 'classic', description: 'Traditional', active: true },
    ])

    const fakeBlob = new Blob(['%PDF-1.4'], { type: 'application/pdf' })
    const fakeURL = 'blob:http://localhost/fake-pdf'
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      blob: () => Promise.resolve(fakeBlob),
    })
    const createObjURL = vi.spyOn(URL, 'createObjectURL').mockReturnValue(fakeURL)
    const revokeObjURL = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})

    const mockAnchor = { href: '', download: '', click: vi.fn() }
    const createEl = vi.spyOn(document, 'createElement').mockImplementation((tag) => {
      if (tag === 'a') return mockAnchor
      return document.createElement.wrappedMethod
        ? document.createElement.wrappedMethod.call(document, tag)
        : Object.getPrototypeOf(document).createElement.call(document, tag)
    })

    renderWithAuth(<ResumeTemplates />, { authValue: { isAdmin: true } })

    await waitFor(() => {
      expect(screen.getByText('Download PDF')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Download PDF'))

    await waitFor(() => {
      expect(mockAnchor.click).toHaveBeenCalled()
    })
    expect(createObjURL).toHaveBeenCalledWith(fakeBlob)
    expect(mockAnchor.download).toBe('resume.pdf')
    expect(revokeObjURL).toHaveBeenCalled()

    fetchSpy.mockRestore()
    createObjURL.mockRestore()
    revokeObjURL.mockRestore()
    createEl.mockRestore()
  })
})
