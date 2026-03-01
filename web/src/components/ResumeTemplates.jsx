import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { listTemplates, setActiveTemplate, getResumeDownloadURL } from '../api'
import './ResumeTemplates.css'

export default function ResumeTemplates() {
  const { isAdmin, getToken } = useAuth()
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(true)
  const [updating, setUpdating] = useState(false)

  const fetchTemplates = async () => {
    try {
      const data = await listTemplates(getToken())
      setTemplates(data)
    } catch (err) {
      console.error('Failed to load templates:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchTemplates() }, [])

  const handleSetActive = async (name) => {
    setUpdating(true)
    try {
      await setActiveTemplate(name, getToken())
      await fetchTemplates()
    } catch (err) {
      console.error('Failed to set template:', err)
    } finally {
      setUpdating(false)
    }
  }

  const handleDownload = async () => {
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
      console.error('Failed to download resume:', err)
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
      <div className="rt-header">
        <h2 className="rt-title">Resume Templates</h2>
        <button className="rt-download" onClick={handleDownload}>
          Download PDF
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
