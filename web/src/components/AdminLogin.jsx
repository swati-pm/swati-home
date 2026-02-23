import { useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import './AdminLogin.css'

export default function AdminLogin() {
  const { user, isAdmin, isLoading, signIn, signOut } = useAuth()

  // Redirect admin to home after successful login
  useEffect(() => {
    if (!isLoading && isAdmin) {
      window.location.hash = '#/'
    }
  }, [isAdmin, isLoading])

  if (isLoading) return null

  // Logged in but not admin
  if (user && !isAdmin) {
    return (
      <section className="admin-login">
        <div className="admin-card">
          <h2 className="admin-title">Access Denied</h2>
          <p className="admin-text">
            This account does not have admin privileges.
          </p>
          <button className="admin-btn admin-btn-secondary" onClick={signOut}>
            Sign Out
          </button>
        </div>
      </section>
    )
  }

  // Not logged in — show sign in prompt
  return (
    <section className="admin-login">
      <div className="admin-card">
        <h2 className="admin-title">Admin Sign In</h2>
        <p className="admin-text">
          Sign in with Google to manage site content.
        </p>
        <button className="admin-btn" onClick={signIn}>
          Sign in with Google
        </button>
      </div>
    </section>
  )
}
