import { useState, useRef, useEffect, useCallback } from 'react'
import { sendChat } from '../api'
import { config } from '../config'
import './HomeChat.css'

export default function HomeChat() {
  const [open, setOpen] = useState(false)
  const [closing, setClosing] = useState(false)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const messagesEndRef = useRef(null)
  const panelRef = useRef(null)
  const overlayRef = useRef(null)

  // Track whether overlay has ever been opened (to avoid rendering it on first load)
  const [hasOpened, setHasOpened] = useState(false)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, loading])

  const handleOpen = () => {
    setHasOpened(true)
    setOpen(true)
  }

  const handleClose = useCallback(() => {
    if (closing) return
    setClosing(true)
  }, [closing])

  // When the overlay's own close animation ends, hide it
  const handleAnimationEnd = (e) => {
    if (closing && e.target === overlayRef.current) {
      setOpen(false)
      setClosing(false)
    }
  }

  // Close on Escape
  useEffect(() => {
    if (!open) return
    const onKey = (e) => {
      if (e.key === 'Escape') handleClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, handleClose])

  // Close on click outside the panel
  const handleOverlayClick = (e) => {
    if (panelRef.current && !panelRef.current.contains(e.target)) {
      handleClose()
    }
  }

  const handleSend = async () => {
    const text = input.trim()
    if (!text || loading) return

    const userMsg = { role: 'user', content: text }
    const updatedMessages = [...messages, userMsg]
    setMessages(updatedMessages)
    setInput('')
    setError(null)
    setLoading(true)

    try {
      const toSend = updatedMessages.slice(-10)
      const data = await sendChat(toSend)
      setMessages((prev) => [...prev, { role: 'assistant', content: data.reply }])
    } catch {
      setError('Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const showOverlay = open || closing

  return (
    <>
      {/* Trigger — always mounted, hidden when overlay is visible */}
      <div
        className="home-chat-trigger"
        style={{ display: showOverlay ? 'none' : undefined }}
        onClick={handleOpen}
      >
        <span className="home-chat-trigger-icon">&#128172;</span>
        <span className="home-chat-trigger-text">Ask me anything about {config.SITE_FIRST_NAME}...</span>
      </div>

      {/* Overlay — always mounted after first open, toggled via CSS */}
      {hasOpened && (
        <div
          className={`home-chat-overlay ${showOverlay ? '' : 'home-chat-overlay-hidden'} ${closing ? 'home-chat-overlay-closing' : ''}`}
          onClick={handleOverlayClick}
          onAnimationEnd={handleAnimationEnd}
          ref={overlayRef}
        >
          <div
            className={`home-chat ${closing ? 'home-chat-closing' : ''}`}
            ref={panelRef}
            onAnimationEnd={(e) => e.stopPropagation()}
          >
            <div className="home-chat-header">
              <span className="home-chat-icon">&#128172;</span>
              <span className="home-chat-title">Ask about {config.SITE_FIRST_NAME}</span>
              <button className="home-chat-close" onClick={handleClose} aria-label="Close chat">&times;</button>
            </div>
            <div className="home-chat-messages">
              <div className="home-chat-welcome">
                Ask me anything about {config.SITE_FIRST_NAME}'s experience &mdash; I'll only answer
                based on information available on this website.
              </div>
              {messages.map((msg, i) => (
                <div key={i} className={`home-chat-msg home-chat-msg-${msg.role}`}>
                  {msg.content}
                </div>
              ))}
              {loading && (
                <div className="home-chat-msg home-chat-msg-assistant home-chat-typing">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              )}
              {error && <div className="home-chat-error">{error}</div>}
              <div ref={messagesEndRef} />
            </div>
            <div className="home-chat-input-area">
              <input
                className="home-chat-input"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type a question..."
                disabled={loading}
                autoFocus={open}
              />
              <button
                className="home-chat-send"
                onClick={handleSend}
                disabled={loading || !input.trim()}
              >
                Send
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
