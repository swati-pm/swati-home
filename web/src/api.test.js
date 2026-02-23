import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  listExperiences,
  createExperience,
  updateExperience,
  deleteExperience,
  listBlogs,
  createBlog,
  updateBlog,
  deleteBlog,
  submitContact,
  listContacts,
  deleteContact,
  sendChat,
} from './api'

beforeEach(() => {
  vi.restoreAllMocks()
})

describe('api module', () => {
  // ---------- Experience API ----------

  describe('listExperiences', () => {
    it('fetches experiences from /api/experiences', async () => {
      const data = [{ id: '1', company: 'ACME' }]
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(data),
      })

      const result = await listExperiences()
      expect(fetch).toHaveBeenCalledWith('/api/experiences')
      expect(result).toEqual(data)
    })

    it('throws on non-ok response', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: false,
        statusText: 'Internal Server Error',
      })

      await expect(listExperiences()).rejects.toThrow('Internal Server Error')
    })
  })

  describe('createExperience', () => {
    it('sends POST with auth header', async () => {
      const exp = { company: 'ACME', role: 'PM' }
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: '1', ...exp }),
      })

      await createExperience(exp, 'tok123')
      expect(fetch).toHaveBeenCalledWith('/api/experiences', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: 'Bearer tok123',
        },
        body: JSON.stringify(exp),
      })
    })
  })

  describe('updateExperience', () => {
    it('sends PUT to /api/experiences/:id', async () => {
      const exp = { id: 'abc', company: 'ACME' }
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(exp),
      })

      await updateExperience(exp, 'tok')
      expect(fetch).toHaveBeenCalledWith('/api/experiences/abc', expect.objectContaining({ method: 'PUT' }))
    })
  })

  describe('deleteExperience', () => {
    it('sends DELETE to /api/experiences/:id', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true })
      await deleteExperience('abc', 'tok')
      expect(fetch).toHaveBeenCalledWith('/api/experiences/abc', expect.objectContaining({ method: 'DELETE' }))
    })
  })

  // ---------- Blog API ----------

  describe('listBlogs', () => {
    it('fetches blogs from /api/blogs', async () => {
      const data = [{ id: '1', title: 'Test' }]
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(data),
      })

      const result = await listBlogs()
      expect(fetch).toHaveBeenCalledWith('/api/blogs')
      expect(result).toEqual(data)
    })
  })

  describe('createBlog', () => {
    it('sends POST with auth header', async () => {
      const blog = { title: 'Post' }
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: '1', ...blog }),
      })

      await createBlog(blog, 'tok')
      expect(fetch).toHaveBeenCalledWith('/api/blogs', expect.objectContaining({ method: 'POST' }))
    })
  })

  describe('updateBlog', () => {
    it('sends PUT to /api/blogs/:id', async () => {
      const blog = { id: 'b1', title: 'Updated' }
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(blog),
      })

      await updateBlog(blog, 'tok')
      expect(fetch).toHaveBeenCalledWith('/api/blogs/b1', expect.objectContaining({ method: 'PUT' }))
    })
  })

  describe('deleteBlog', () => {
    it('sends DELETE to /api/blogs/:id', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true })
      await deleteBlog('b1', 'tok')
      expect(fetch).toHaveBeenCalledWith('/api/blogs/b1', expect.objectContaining({ method: 'DELETE' }))
    })
  })

  // ---------- Contact API ----------

  describe('submitContact', () => {
    it('sends POST to /api/contacts without auth', async () => {
      const contact = { name: 'Jane', email: 'j@e.com', subject: 'Hi', message: 'Hello' }
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: '1', ...contact }),
      })

      await submitContact(contact)
      expect(fetch).toHaveBeenCalledWith('/api/contacts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(contact),
      })
    })
  })

  describe('listContacts', () => {
    it('sends GET with auth header', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve([]),
      })

      await listContacts('tok')
      expect(fetch).toHaveBeenCalledWith('/api/contacts', {
        headers: {
          'Content-Type': 'application/json',
          Authorization: 'Bearer tok',
        },
      })
    })
  })

  describe('deleteContact', () => {
    it('sends DELETE with auth header', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true })
      await deleteContact('c1', 'tok')
      expect(fetch).toHaveBeenCalledWith('/api/contacts/c1', expect.objectContaining({ method: 'DELETE' }))
    })
  })

  // ---------- Chat API ----------

  describe('sendChat', () => {
    it('sends POST with messages payload', async () => {
      const msgs = [{ role: 'user', content: 'hi' }]
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ reply: 'hello' }),
      })

      const result = await sendChat(msgs)
      expect(fetch).toHaveBeenCalledWith('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: msgs }),
      })
      expect(result).toEqual({ reply: 'hello' })
    })

    it('throws on failure', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue({
        ok: false,
        statusText: 'Too Many Requests',
      })

      await expect(sendChat([])).rejects.toThrow('Too Many Requests')
    })
  })
})
