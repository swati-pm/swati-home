import { useState, useEffect } from 'react'
import './BlogForm.css'

const emptyForm = {
  title: '',
  summary: '',
  url: '',
  date: '',
}

export default function BlogForm({ blog, onSave, onCancel, onDelete }) {
  const [form, setForm] = useState(emptyForm)

  useEffect(() => {
    if (blog) {
      setForm({
        title: blog.title || '',
        summary: blog.summary || '',
        url: blog.url || '',
        date: blog.date || '',
      })
    } else {
      setForm(emptyForm)
    }
  }, [blog])

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value })
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!form.title.trim()) return

    onSave({
      id: blog?.id || crypto.randomUUID(),
      title: form.title.trim(),
      summary: form.summary.trim(),
      url: form.url.trim(),
      date: form.date,
    })
  }

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2 className="modal-title">
          {blog ? 'Edit Blog Post' : 'Add Blog Post'}
        </h2>
        <form onSubmit={handleSubmit}>
          <label className="form-label">
            Title *
            <input
              className="form-input"
              name="title"
              value={form.title}
              onChange={handleChange}
              placeholder="e.g. Building AI Governance into Product Workflows"
              required
            />
          </label>
          <label className="form-label">
            Summary
            <textarea
              className="form-textarea"
              name="summary"
              value={form.summary}
              onChange={handleChange}
              rows={3}
              placeholder="A brief description of the blog post"
            />
          </label>
          <div className="form-row">
            <label className="form-label">
              URL (link to full post)
              <input
                className="form-input"
                name="url"
                value={form.url}
                onChange={handleChange}
                placeholder="https://medium.com/..."
              />
            </label>
            <label className="form-label">
              Date
              <input
                className="form-input"
                name="date"
                type="date"
                value={form.date}
                onChange={handleChange}
              />
            </label>
          </div>
          <div className="form-actions">
            {blog && onDelete && (
              <button
                type="button"
                className="btn btn-danger"
                onClick={() => onDelete(blog.id)}
              >
                Delete
              </button>
            )}
            <div className="form-actions-right">
              <button type="button" className="btn btn-secondary" onClick={onCancel}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary">
                Save
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  )
}
