import { useState, useEffect } from 'react'
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
  const [form, setForm] = useState(emptyForm)

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
