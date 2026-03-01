import { config } from './config'

const API = config.API_BASE

function authHeaders(token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  return headers
}

// Experience API

export async function listExperiences() {
  const res = await fetch(`${API}/experiences`)
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function createExperience(experience, token) {
  const res = await fetch(`${API}/experiences`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(experience),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function updateExperience(experience, token) {
  const res = await fetch(`${API}/experiences/${experience.id}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(experience),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function deleteExperience(id, token) {
  const res = await fetch(`${API}/experiences/${id}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
  if (!res.ok) throw new Error(res.statusText)
}

// Blog API

export async function listBlogs() {
  const res = await fetch(`${API}/blogs`)
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function createBlog(blog, token) {
  const res = await fetch(`${API}/blogs`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(blog),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function updateBlog(blog, token) {
  const res = await fetch(`${API}/blogs/${blog.id}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(blog),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function deleteBlog(id, token) {
  const res = await fetch(`${API}/blogs/${id}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
  if (!res.ok) throw new Error(res.statusText)
}

// Contact API

export async function submitContact(contact) {
  const res = await fetch(`${API}/contacts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(contact),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function listContacts(token) {
  const res = await fetch(`${API}/contacts`, {
    headers: authHeaders(token),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function deleteContact(id, token) {
  const res = await fetch(`${API}/contacts/${id}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
  if (!res.ok) throw new Error(res.statusText)
}

// Resume API

export async function listTemplates(token) {
  const res = await fetch(`${API}/resume/templates`, {
    headers: authHeaders(token),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export async function setActiveTemplate(template, token) {
  const res = await fetch(`${API}/resume/template`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify({ template }),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}

export function getResumeDownloadURL(token) {
  return `${API}/resume/download`
}

// Chat API

export async function sendChat(messages) {
  const res = await fetch(`${API}/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ messages }),
  })
  if (!res.ok) throw new Error(res.statusText)
  return res.json()
}
