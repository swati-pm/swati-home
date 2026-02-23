import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ExperienceCard from './ExperienceCard'

const mockExperience = {
  id: '1',
  company: 'bunny.net',
  role: 'Product Department Lead',
  location: 'London, UK',
  startDate: 'Jan 2025',
  endDate: 'Present',
  bullets: [
    'Led product strategy across multiple teams',
    'Shipped AI-powered features',
  ],
}

describe('ExperienceCard', () => {
  it('renders company name and role', () => {
    render(<ExperienceCard experience={mockExperience} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('bunny.net')).toBeInTheDocument()
    expect(screen.getByText('Product Department Lead')).toBeInTheDocument()
  })

  it('renders date range', () => {
    render(<ExperienceCard experience={mockExperience} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('Jan 2025 \u2013 Present')).toBeInTheDocument()
  })

  it('renders location', () => {
    render(<ExperienceCard experience={mockExperience} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('London, UK')).toBeInTheDocument()
  })

  it('renders bullet points', () => {
    render(<ExperienceCard experience={mockExperience} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('Led product strategy across multiple teams')).toBeInTheDocument()
    expect(screen.getByText('Shipped AI-powered features')).toBeInTheDocument()
  })

  it('does not show Edit button for non-admin', () => {
    render(<ExperienceCard experience={mockExperience} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.queryByText('Edit')).not.toBeInTheDocument()
  })

  it('shows Edit button for admin', () => {
    render(<ExperienceCard experience={mockExperience} onEdit={vi.fn()} isAdmin={true} />)
    expect(screen.getByText('Edit')).toBeInTheDocument()
  })

  it('calls onEdit with experience when Edit is clicked', async () => {
    const onEdit = vi.fn()
    render(<ExperienceCard experience={mockExperience} onEdit={onEdit} isAdmin={true} />)
    await userEvent.click(screen.getByText('Edit'))
    expect(onEdit).toHaveBeenCalledWith(mockExperience)
  })

  it('handles missing dates gracefully', () => {
    const noDate = { ...mockExperience, startDate: '', endDate: '' }
    render(<ExperienceCard experience={noDate} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.getByText('bunny.net')).toBeInTheDocument()
  })

  it('handles empty bullets', () => {
    const noBullets = { ...mockExperience, bullets: [] }
    render(<ExperienceCard experience={noBullets} onEdit={vi.fn()} isAdmin={false} />)
    expect(screen.queryByRole('list')).not.toBeInTheDocument()
  })
})
