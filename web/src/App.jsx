import { useState, useEffect } from 'react'
import Header from './components/Header'
import Home from './components/Home'
import ExperienceSection from './components/ExperienceSection'
import BlogSection from './components/BlogSection'
import ContactForm from './components/ContactForm'
import ContactRequests from './components/ContactRequests'
import AdminLogin from './components/AdminLogin'
import { config } from './config'
import './App.css'

function getPage() {
  const hash = window.location.hash || '#/'
  if (hash.startsWith('#/experience')) return 'experience'
  if (hash.startsWith('#/blog')) return 'blog'
  if (hash.startsWith('#/contact-requests')) return 'contact-requests'
  if (hash.startsWith('#/contact')) return 'contact'
  if (hash.startsWith('#/admin')) return 'admin'
  return 'home'
}

export default function App() {
  const [page, setPage] = useState(getPage)

  useEffect(() => {
    const onHashChange = () => setPage(getPage())
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  }, [])

  return (
    <div className="app">
      <Header currentPage={page} />
      <main>
        {page === 'home' && <Home />}
        {page === 'experience' && <ExperienceSection />}
        {page === 'blog' && <BlogSection />}
        {page === 'contact' && <ContactForm />}
        {page === 'contact-requests' && <ContactRequests />}
        {page === 'admin' && <AdminLogin />}
      </main>
      <footer className="footer">
        <p>&copy; {new Date().getFullYear()} {config.SITE_NAME}</p>
      </footer>
    </div>
  )
}
