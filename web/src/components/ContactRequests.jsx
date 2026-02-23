import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { listContacts, deleteContact } from '../api'
import './ContactRequests.css'

export default function ContactRequests() {
  const { isAdmin, getToken } = useAuth()
  const [contacts, setContacts] = useState([])
  const [loading, setLoading] = useState(true)

  const fetchData = async () => {
    try {
      const data = await listContacts(getToken())
      setContacts(data)
    } catch (err) {
      console.error('Failed to load contacts:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [])

  const handleDelete = async (id) => {
    if (!confirm('Delete this contact request?')) return
    try {
      await deleteContact(id, getToken())
      await fetchData()
    } catch (err) {
      console.error('Failed to delete contact:', err)
    }
  }

  if (!isAdmin) {
    return (
      <section className="cr-section">
        <p className="cr-denied">Access denied. Admin login required.</p>
      </section>
    )
  }

  if (loading) {
    return (
      <section className="cr-section">
        <p>Loading...</p>
      </section>
    )
  }

  return (
    <section className="cr-section">
      <div className="cr-header">
        <h2 className="cr-title">Contact Requests</h2>
        <span className="cr-count">{contacts.length} request{contacts.length !== 1 ? 's' : ''}</span>
      </div>
      {contacts.length === 0 ? (
        <p className="cr-empty">No contact requests yet.</p>
      ) : (
        <div className="cr-list">
          {contacts.map(c => (
            <div key={c.id} className="cr-card">
              <div className="cr-card-top">
                <div className="cr-card-info">
                  <h3 className="cr-card-subject">{c.subject}</h3>
                  <div className="cr-card-meta">
                    <span className="cr-card-name">{c.name}</span>
                    <a href={`mailto:${c.email}`} className="cr-card-email">{c.email}</a>
                    {c.phone && <span className="cr-card-phone">{c.phone}</span>}
                  </div>
                </div>
                <div className="cr-card-actions">
                  {c.created_at && (
                    <span className="cr-card-date">
                      {new Date(c.created_at).toLocaleDateString('en-GB', {
                        day: 'numeric', month: 'short', year: 'numeric',
                      })}
                    </span>
                  )}
                  <button className="cr-card-delete" onClick={() => handleDelete(c.id)}>
                    Delete
                  </button>
                </div>
              </div>
              <p className="cr-card-message">{c.message}</p>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
