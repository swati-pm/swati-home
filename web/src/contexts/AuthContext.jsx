import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { config } from '../config'

export const AuthContext = createContext(null)

function decodeJwtPayload(token) {
  const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
  return JSON.parse(atob(base64))
}

const SESSION_KEY = config.SESSION_KEY

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [isLoading, setIsLoading] = useState(true)

  const isAdmin = user?.email === config.ADMIN_EMAIL

  const saveSession = (userData) => {
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(userData))
  }

  const clearSession = () => {
    sessionStorage.removeItem(SESSION_KEY)
  }

  const handleCredentialResponse = useCallback((response) => {
    const payload = decodeJwtPayload(response.credential)
    const userData = {
      email: payload.email,
      name: payload.name,
      picture: payload.picture,
      token: response.credential,
    }
    setUser(userData)
    saveSession(userData)
  }, [])

  const signIn = useCallback(() => {
    if (window.google?.accounts?.id) {
      window.google.accounts.id.prompt()
    }
  }, [])

  const signOut = useCallback(() => {
    setUser(null)
    clearSession()
    if (window.google?.accounts?.id) {
      window.google.accounts.id.disableAutoSelect()
    }
  }, [])

  // Restore session on mount
  useEffect(() => {
    try {
      const stored = sessionStorage.getItem(SESSION_KEY)
      if (stored) {
        setUser(JSON.parse(stored))
      }
    } catch {}
    setIsLoading(false)
  }, [])

  // Initialize Google Identity Services
  useEffect(() => {
    const initGsi = () => {
      if (!window.google?.accounts?.id) return

      window.google.accounts.id.initialize({
        client_id: config.GOOGLE_CLIENT_ID,
        callback: handleCredentialResponse,
        auto_select: false,
      })
    }

    if (window.google?.accounts?.id) {
      initGsi()
    } else {
      // GSI script may still be loading — wait for it
      const interval = setInterval(() => {
        if (window.google?.accounts?.id) {
          initGsi()
          clearInterval(interval)
        }
      }, 100)
      return () => clearInterval(interval)
    }
  }, [handleCredentialResponse])

  return (
    <AuthContext.Provider value={{ user, isAdmin, isLoading, signIn, signOut, getToken: () => user?.token }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
