import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { suggestExperienceBullets } from '../api'
import './ExperienceForm.css'

const emptyForm = {
  company: '',
  role: '',
  location: '',
  startDate: '',
  endDate: '',
  bullets: '',
}

export default function ExperienceForm({ experience, onSave, onCancel }) {
  const { isAdmin, getToken } = useAuth()
  const [form, setForm] = useState(emptyForm)
  const [suggesting, setSuggesting] = useState(false)
  const [suggestions, setSuggestions] = useState(null)
  const [suggestError, setSuggestError] = useState(null)
  const [roleDescription, setRoleDescription] = useState('')
  const [showRoleInput, setShowRoleInput] = useState(false)

  useEffect(() => {
    if (experience) {
      setForm({
        company: experience.company || '',
        role: experience.role || '',
        location: experience.location || '',
        startDate: experience.startDate || '',
        endDate: experience.endDate || '',
        bullets: (experience.bullets || []).join('\n'),
      })
    } else {
      setForm(emptyForm)
    }
  }, [experience])

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value })
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!form.company.trim() || !form.role.trim()) return

    onSave({
      id: experience?.id || crypto.randomUUID(),
      company: form.company.trim(),
      role: form.role.trim(),
      location: form.location.trim(),
      startDate: form.startDate.trim(),
      endDate: form.endDate.trim(),
      bullets: form.bullets
        .split('\n')
        .map(b => b.trim())
        .filter(Boolean),
    })
  }

  const bulletsArray = form.bullets
    .split('\n')
    .map(b => b.trim())
    .filter(Boolean)

  const canSuggest = isAdmin && experience && bulletsArray.length > 0

  const handleSuggest = async () => {
    setSuggesting(true)
    setSuggestError(null)
    setSuggestions(null)
    try {
      const token = getToken()
      const payload = {
        company: form.company.trim(),
        role: form.role.trim(),
        location: form.location.trim(),
        start_date: form.startDate.trim(),
        end_date: form.endDate.trim(),
        bullets: bulletsArray,
      }
      if (roleDescription.trim()) {
        payload.role_description = roleDescription.trim()
      }
      const result = await suggestExperienceBullets(payload, token)
      setSuggestions(result.bullets)
    } catch {
      setSuggestError('Failed to get suggestions. Please try again.')
    } finally {
      setSuggesting(false)
    }
  }

  const handleAcceptSuggestions = () => {
    setForm({ ...form, bullets: suggestions.join('\n') })
    setSuggestions(null)
  }

  const handleDismissSuggestions = () => {
    setSuggestions(null)
  }

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2 className="modal-title">
          {experience ? 'Edit Experience' : 'Add Experience'}
        </h2>
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <label className="form-label">
              Company *
              <input
                className="form-input"
                name="company"
                value={form.company}
                onChange={handleChange}
                placeholder="e.g. Google"
                required
              />
            </label>
            <label className="form-label">
              Role *
              <input
                className="form-input"
                name="role"
                value={form.role}
                onChange={handleChange}
                placeholder="e.g. Senior Product Manager"
                required
              />
            </label>
          </div>
          <div className="form-row">
            <label className="form-label">
              Location
              <input
                className="form-input"
                name="location"
                value={form.location}
                onChange={handleChange}
                placeholder="e.g. London, UK"
              />
            </label>
          </div>
          <div className="form-row">
            <label className="form-label">
              Start Date
              <input
                className="form-input"
                name="startDate"
                value={form.startDate}
                onChange={handleChange}
                placeholder="e.g. Jan 2023"
              />
            </label>
            <label className="form-label">
              End Date
              <input
                className="form-input"
                name="endDate"
                value={form.endDate}
                onChange={handleChange}
                placeholder="e.g. Present"
              />
            </label>
          </div>
          <label className="form-label">
            Key Achievements (one per line)
            <textarea
              className="form-textarea"
              name="bullets"
              value={form.bullets}
              onChange={handleChange}
              rows={6}
              placeholder={"Led cross-functional team of 10...\nDelivered 30% improvement in..."}
            />
          </label>

          {canSuggest && (
            <div className="suggest-section">
              {suggestError && (
                <div className="suggest-error" role="alert">
                  <span>{suggestError}</span>
                  <button
                    type="button"
                    className="suggest-error-close"
                    onClick={() => setSuggestError(null)}
                  >
                    &times;
                  </button>
                </div>
              )}

              {!showRoleInput ? (
                <button
                  type="button"
                  className="suggest-role-toggle"
                  onClick={() => setShowRoleInput(true)}
                >
                  + Add role description
                </button>
              ) : (
                <label className="form-label">
                  Role Description (optional)
                  <textarea
                    className="form-textarea"
                    value={roleDescription}
                    onChange={e => setRoleDescription(e.target.value)}
                    rows={3}
                    placeholder="Paste a job description to tailor suggestions..."
                  />
                </label>
              )}

              <button
                type="button"
                className="btn btn-suggest"
                onClick={handleSuggest}
                disabled={suggesting}
              >
                {suggesting ? 'Getting suggestions...' : 'Suggest Improvements'}
              </button>

              {suggestions && (
                <div className="suggest-panel">
                  <h4 className="suggest-panel-title">Suggested Improvements</h4>
                  <ul className="suggest-list">
                    {suggestions.map((bullet, i) => (
                      <li key={i} className="suggest-item">{bullet}</li>
                    ))}
                  </ul>
                  <div className="suggest-actions">
                    <button
                      type="button"
                      className="btn btn-primary btn-sm"
                      onClick={handleAcceptSuggestions}
                    >
                      Accept
                    </button>
                    <button
                      type="button"
                      className="btn btn-secondary btn-sm"
                      onClick={handleDismissSuggestions}
                    >
                      Dismiss
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          <div className="form-actions">
            <button type="button" className="btn btn-secondary" onClick={onCancel}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary">
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
