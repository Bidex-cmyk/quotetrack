import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../services/api'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [sent, setSent] = useState(false)
  const [busy, setBusy] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (!email.trim()) {
      setError('Email is required.')
      return
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      setError('Please enter a valid email address.')
      return
    }
    setBusy(true)
    try {
      await api.forgotPassword(email.trim())
      setSent(true)
    } catch (err) {
      // Even on error, show success to avoid leaking info.
      setSent(true)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-brand">
          <span className="brand-mark">Q</span>
          <h1>Forgot your password?</h1>
        </div>
        {sent ? (
          <>
            <p className="auth-tagline">
              If an account exists for that email, you&apos;ll receive password reset instructions.
            </p>
            <p className="auth-alt" style={{ marginTop: '1rem' }}>
              <Link to="/login">Back to login</Link>
            </p>
          </>
        ) : (
          <>
            <p className="auth-tagline">
              Enter your email and we&apos;ll send you a password reset link.
            </p>
            <form onSubmit={handleSubmit} className="auth-form" noValidate>
              <label className="field">
                <span>Email</span>
                <input
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value)
                    if (error) setError('')
                  }}
                  placeholder="you@business.com"
                  aria-invalid={!!error}
                  aria-describedby={error ? 'forgot-email-error' : undefined}
                />
                {error && (
                  <span className="field-error" id="forgot-email-error" role="alert">
                    {error}
                  </span>
                )}
              </label>
              <button className="btn btn-primary btn-block" disabled={busy} type="submit">
                {busy ? 'Sending…' : 'Send reset link'}
              </button>
            </form>
            <p className="auth-alt" style={{ marginTop: '1rem' }}>
              <Link to="/login">Back to login</Link>
            </p>
          </>
        )}
      </div>
    </div>
  )
}
