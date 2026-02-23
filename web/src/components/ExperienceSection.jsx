import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { listExperiences, createExperience, updateExperience } from '../api'
import ExperienceCard from './ExperienceCard'
import ExperienceForm from './ExperienceForm'
import './ExperienceSection.css'

export default function ExperienceSection() {
  const { isAdmin, getToken } = useAuth()
  const [experiences, setExperiences] = useState([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState(null)
  const [showForm, setShowForm] = useState(false)

  const fetchData = async () => {
    try {
      const data = await listExperiences()
      setExperiences(data)
    } catch (err) {
      console.error('Failed to load experiences:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [])

  const handleAdd = () => {
    setEditing(undefined)
    setShowForm(true)
  }

  const handleEdit = (exp) => {
    setEditing(exp)
    setShowForm(true)
  }

  const handleSave = async (data) => {
    try {
      const token = getToken()
      if (editing) {
        await updateExperience(toApiFormat(data), token)
      } else {
        await createExperience(toApiFormat(data), token)
      }
      await fetchData()
    } catch (err) {
      console.error('Failed to save experience:', err)
    }
    setShowForm(false)
    setEditing(null)
  }

  const handleCancel = () => {
    setShowForm(false)
    setEditing(null)
  }

  if (loading) {
    return (
      <section className="experience-section">
        <p>Loading...</p>
      </section>
    )
  }

  return (
    <section className="experience-section">
      <div className="experience-header">
        <h2 className="experience-title">Professional Experience</h2>
        {isAdmin && (
          <button className="add-btn" onClick={handleAdd}>
            + Add Experience
          </button>
        )}
      </div>
      <div className="experience-list">
        {experiences.map(exp => (
          <ExperienceCard key={exp.id} experience={fromApiFormat(exp)} onEdit={handleEdit} isAdmin={isAdmin} />
        ))}
      </div>
      {showForm && (
        <ExperienceForm
          experience={editing}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      )}
    </section>
  )
}

// Convert between frontend camelCase and API snake_case
function toApiFormat(exp) {
  return {
    id: exp.id,
    company: exp.company,
    role: exp.role,
    location: exp.location,
    start_date: exp.startDate,
    end_date: exp.endDate,
    bullets: exp.bullets,
  }
}

function fromApiFormat(exp) {
  return {
    id: exp.id,
    company: exp.company,
    role: exp.role,
    location: exp.location,
    startDate: exp.startDate || exp.start_date || '',
    endDate: exp.endDate || exp.end_date || '',
    bullets: exp.bullets || [],
  }
}
