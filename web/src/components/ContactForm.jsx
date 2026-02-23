import { useState } from 'react'
import { submitContact } from '../api'
import './ContactForm.css'

export default function ContactForm() {
  const [form, setForm] = useState({
    name: '',
    email: '',
    phone: '',
    subject: '',
    message: '',
  })
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState(null)

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await submitContact({
        name: form.name.trim(),
        email: form.email.trim(),
        phone: form.phone.trim(),
        subject: form.subject.trim(),
        message: form.message.trim(),
      })
      setSubmitted(true)
    } catch (err) {
      setError('Failed to send message. Please try again.')
      console.error('Contact submit error:', err)
    } finally {
      setSubmitting(false)
    }
  }

  if (submitted) {
    return (
      <section className="contact-page">
        <div className="contact-success">
          <div className="contact-success-icon">&#10003;</div>
          <h2 className="contact-success-title">Message Sent!</h2>
          <p className="contact-success-text">
            Thank you for reaching out. I'll get back to you soon.
          </p>
          <a href="#/" className="btn btn-primary">Back to Home</a>
        </div>
      </section>
    )
  }

  return (
    <section className="contact-page">
      <h2 className="contact-page-title">Contact Me</h2>
      <p className="contact-page-subtitle">
        Have a question or want to work together? Send me a message.
      </p>
      <form className="contact-form" onSubmit={handleSubmit}>
        <label className="form-label">
          Name *
          <input
            className="form-input"
            name="name"
            value={form.name}
            onChange={handleChange}
            placeholder="Your name"
            required
          />
        </label>
        <div className="form-row">
          <label className="form-label">
            Email *
            <input
              className="form-input"
              name="email"
              type="email"
              value={form.email}
              onChange={handleChange}
              placeholder="you@example.com"
              required
            />
          </label>
          <label className="form-label">
            Phone
            <input
              className="form-input"
              name="phone"
              type="tel"
              value={form.phone}
              onChange={handleChange}
              placeholder="+44 7000 000000"
            />
          </label>
        </div>
        <label className="form-label">
          Subject *
          <input
            className="form-input"
            name="subject"
            value={form.subject}
            onChange={handleChange}
            placeholder="What is this about?"
            required
          />
        </label>
        <label className="form-label">
          Message *
          <textarea
            className="form-textarea"
            name="message"
            value={form.message}
            onChange={handleChange}
            placeholder="Your message..."
            rows={6}
            required
          />
        </label>
        {error && <p className="contact-error">{error}</p>}
        <div className="contact-form-actions">
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? 'Sending...' : 'Send Message'}
          </button>
        </div>
      </form>
    </section>
  )
}
