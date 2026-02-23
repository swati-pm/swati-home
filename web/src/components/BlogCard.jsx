import './BlogCard.css'

export default function BlogCard({ blog, onEdit, isAdmin }) {
  const { title, summary, url, date } = blog

  const formattedDate = date
    ? new Date(date + 'T00:00:00').toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
      })
    : ''

  const content = (
    <>
      <div className="blog-card-top">
        <div className="blog-card-info">
          <h3 className="blog-card-title">{title}</h3>
          {formattedDate && <span className="blog-card-date">{formattedDate}</span>}
        </div>
        {isAdmin && (
          <button className="blog-card-edit" onClick={(e) => { e.preventDefault(); onEdit(blog) }}>
            Edit
          </button>
        )}
      </div>
      {summary && <p className="blog-card-summary">{summary}</p>}
    </>
  )

  if (url) {
    return (
      <a href={url} target="_blank" rel="noopener noreferrer" className="blog-card blog-card-linked">
        {content}
      </a>
    )
  }

  return <div className="blog-card">{content}</div>
}
