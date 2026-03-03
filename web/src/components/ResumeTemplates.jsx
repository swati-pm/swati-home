import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { listTemplates, setActiveTemplate, getResumeDownloadURL } from '../api'
import './ResumeTemplates.css'

export default function ResumeTemplates() {
  const { isAdmin, getToken } = useAuth()
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(true)
  const [updating, setUpdating] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [error, setError] = useState(null)

  const showError = (msg) => {
    setError(msg)
    setTimeout(() => setError(null), 5000)
  }

  const fetchTemplates = async () => {
    try {
      const data = await listTemplates(getToken())
      setTemplates(data)
    } catch (err) {
      showError('Failed to load templates. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchTemplates() }, [])

  const handleSetActive = async (name) => {
    setUpdating(true)
    setError(null)
    try {
      await setActiveTemplate(name, getToken())
      await fetchTemplates()
    } catch (err) {
      showError('Failed to update template. Please try again.')
    } finally {
      setUpdating(false)
    }
  }

  const handleDownload = async () => {
    setDownloading(true)
    setError(null)
    const token = getToken()
    const url = getResumeDownloadURL(token)
    try {
      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error(res.statusText)
      const blob = await res.blob()
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = 'resume.pdf'
      a.click()
      URL.revokeObjectURL(a.href)
    } catch (err) {
      showError('Failed to download resume. Please try again.')
    } finally {
      setDownloading(false)
    }
  }

  if (!isAdmin) {
    return (
      <section className="rt-section">
        <p className="rt-denied">Access denied. Admin login required.</p>
      </section>
    )
  }

  if (loading) {
    return (
      <section className="rt-section">
        <p>Loading...</p>
      </section>
    )
  }

  return (
    <section className="rt-section">
      {error && (
        <div className="rt-error" role="alert">
          <span>{error}</span>
          <button className="rt-error-close" onClick={() => setError(null)}>&times;</button>
        </div>
      )}
      <div className="rt-header">
        <h2 className="rt-title">Resume Templates</h2>
        <button className="rt-download" onClick={handleDownload} disabled={downloading}>
          {downloading ? 'Downloading...' : 'Download PDF'}
        </button>
      </div>
      <div className="rt-grid">
        {templates.map((t) => (
          <div key={t.name} className={`rt-card ${t.active ? 'rt-card--active' : ''}`}>
            <div className="rt-card-top">
              <h3 className="rt-card-name">{t.name}</h3>
              {t.active && <span className="rt-badge">Active</span>}
            </div>
            <p className="rt-card-desc">{t.description}</p>
            {!t.active && (
              <button
                className="rt-card-btn"
                onClick={() => handleSetActive(t.name)}
                disabled={updating}
              >
                Set Active
              </button>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}
