import './ExperienceCard.css'

export default function ExperienceCard({ experience, onEdit, isAdmin }) {
  const { company, role, location, startDate, endDate, bullets } = experience
  const dateRange = startDate && endDate
    ? `${startDate} \u2013 ${endDate}`
    : startDate || ''

  return (
    <div className="exp-card">
      <div className="exp-card-header">
        <div className="exp-card-info">
          <h3 className="exp-card-company">{company}</h3>
          <p className="exp-card-role">{role}</p>
          <div className="exp-card-meta">
            {dateRange && <span className="exp-card-date">{dateRange}</span>}
            {location && <span className="exp-card-location">{location}</span>}
          </div>
        </div>
        {isAdmin && (
          <button className="exp-card-edit" onClick={() => onEdit(experience)}>
            Edit
          </button>
        )}
      </div>
      {bullets.length > 0 && (
        <ul className="exp-card-bullets">
          {bullets.map((bullet, i) => (
            <li key={i}>{bullet}</li>
          ))}
        </ul>
      )}
    </div>
  )
}
