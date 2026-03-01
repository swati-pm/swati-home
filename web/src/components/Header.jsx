import { useAuth } from '../contexts/AuthContext'
import { config } from '../config'
import './Header.css'

export default function Header({ currentPage }) {
  const { isAdmin, signOut } = useAuth()

  return (
    <header className="header">
      <div className="header-content">
        <a href="#/" className="header-brand">{config.SITE_NAME}</a>
        <nav className="header-nav">
          <a href="#/" className={`header-link ${currentPage === 'home' ? 'active' : ''}`}>
            Home
          </a>
          <a href="#/experience" className={`header-link ${currentPage === 'experience' ? 'active' : ''}`}>
            Experience
          </a>
          <a href="#/blog" className={`header-link ${currentPage === 'blog' ? 'active' : ''}`}>
            Blog
          </a>
          <a href="#/contact" className={`header-link ${currentPage === 'contact' ? 'active' : ''}`}>
            Contact
          </a>
          {isAdmin && (
            <a href="#/contact-requests" className={`header-link ${currentPage === 'contact-requests' ? 'active' : ''}`}>
              Contact Requests
            </a>
          )}
          {isAdmin && (
            <a href="#/resumes" className={`header-link ${currentPage === 'resumes' ? 'active' : ''}`}>
              Resumes
            </a>
          )}
          {isAdmin && (
            <button className="header-signout" onClick={signOut}>
              Sign Out
            </button>
          )}
        </nav>
      </div>
    </header>
  )
}
