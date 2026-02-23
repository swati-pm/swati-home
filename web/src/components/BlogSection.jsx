import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { listBlogs, createBlog, updateBlog, deleteBlog } from '../api'
import BlogCard from './BlogCard'
import BlogForm from './BlogForm'
import './BlogSection.css'

export default function BlogSection() {
  const { isAdmin, getToken } = useAuth()
  const [blogs, setBlogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState(null)
  const [showForm, setShowForm] = useState(false)

  const fetchData = async () => {
    try {
      const data = await listBlogs()
      setBlogs(data)
    } catch (err) {
      console.error('Failed to load blogs:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [])

  const handleAdd = () => {
    setEditing(undefined)
    setShowForm(true)
  }

  const handleEdit = (blog) => {
    setEditing(blog)
    setShowForm(true)
  }

  const handleSave = async (data) => {
    try {
      const token = getToken()
      if (editing) {
        await updateBlog(data, token)
      } else {
        await createBlog(data, token)
      }
      await fetchData()
    } catch (err) {
      console.error('Failed to save blog:', err)
    }
    setShowForm(false)
    setEditing(null)
  }

  const handleDelete = async (id) => {
    try {
      await deleteBlog(id, getToken())
      await fetchData()
    } catch (err) {
      console.error('Failed to delete blog:', err)
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
      <section className="blog-section">
        <p>Loading...</p>
      </section>
    )
  }

  return (
    <section className="blog-section">
      <div className="blog-header">
        <h2 className="blog-title">Blog</h2>
        {isAdmin && (
          <button className="add-btn" onClick={handleAdd}>
            + Add Post
          </button>
        )}
      </div>
      {blogs.length === 0 ? (
        <p className="blog-empty">No posts yet.</p>
      ) : (
        <div className="blog-list">
          {blogs.map(blog => (
            <BlogCard key={blog.id} blog={blog} onEdit={handleEdit} isAdmin={isAdmin} />
          ))}
        </div>
      )}
      {showForm && (
        <BlogForm
          blog={editing}
          onSave={handleSave}
          onDelete={handleDelete}
          onCancel={handleCancel}
        />
      )}
    </section>
  )
}
