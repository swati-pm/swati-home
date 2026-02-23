export const config = {
  GOOGLE_CLIENT_ID: import.meta.env.VITE_GOOGLE_CLIENT_ID || '',
  ADMIN_EMAIL: import.meta.env.VITE_ADMIN_EMAIL || '',
  API_BASE: import.meta.env.VITE_API_BASE || '/api',
  SITE_NAME: import.meta.env.VITE_SITE_NAME || 'Swati Aggarwal',
  SITE_FIRST_NAME: import.meta.env.VITE_SITE_FIRST_NAME || 'Swati',
  SESSION_KEY: import.meta.env.VITE_SESSION_KEY || 'swati-home-auth',
}
