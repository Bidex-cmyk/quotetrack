import { useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api } from '../services/api'
import PasswordInput from '../components/PasswordInput'

function passwordStrength(pw: string): { level: 'weak' | 'fair' | 'strong'; label: string; color: string } {
  let score = 0
  if (pw.length >= 8) score++
  if (pw.length >= 12) score++
  if (/[A-Z]/.test(pw)) score++
  if (/[a-z]/.test(pw)) score++
  if (/[0-9]/.test(pw)) score++
  if (/[^A-Za-z0-9]/.test(pw)) score++
  if (score <= 2) return { level: 'weak', label: 'Weak', color: '#dc2626' }
  if (score <= 4) return { level: 'fair', label: 'Fair', color: '#d97706' }
  return { level: 'strong', label: 'Strong', color: '#16a34a' }
}

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') || ''

  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<{ password?: string; confirmPassword?: string }>({})
  const [busy, setBusy] = useState(false)
  const [success, setSuccess] = useState(false)

  const hasLetter = /[a-z]/i.test(password)
  const hasNumber = /[0-9]/.test(password)
  const hasMinLength = password.length >= 8
  const strength = passwordStrength(password)

  if (!token) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <div className="auth-brand">
            <span className="brand-mark">Q</span>
            <h1>Invalid link</h1>
          </div>
          <p className="auth-tagline">
            This password reset link is invalid or has expired.
          </p>
          <p className="auth-alt" style={{ marginTop: '1rem' }}>
            <Link to="/forgot-password">Request a new link</Link>
          </p>
          <p className="auth-alt">
            <Link to="/login">Back to login</Link>
          </p>
        </div>
      </div>
    )
  }

  const validate = (): boolean => {
    const errs: typeof fieldErrors = {}
    if (!password) {
      errs.password = 'Password is required.'
    } else if (password.length < 8) {
      errs.password = 'Password must be at least 8 characters.'
    } else if (!hasLetter) {
      errs.password = 'Password must contain at least one letter.'
    } else if (!hasNumber) {
      errs.password = 'Password must contain at least one number.'
    }
    if (!confirmPassword) {
      errs.confirmPassword = 'Please confirm your password.'
    } else if (confirmPassword !== password) {
      errs.confirmPassword = 'Passwords do not match.'
    }
    setFieldErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (!validate()) return
    setBusy(true)
    try {
      await api.resetPassword(token, password)
      setSuccess(true)
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message || 'Could not reset password. Please try again.')
      } else {
        setError('Something went wrong. Please try again.')
      }
    } finally {
      setBusy(false)
    }
  }

  if (success) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <div className="auth-brand">
            <span className="brand-mark">Q</span>
            <h1>Password reset successful</h1>
          </div>
          <p className="auth-tagline">
            Your password has been reset successfully.
          </p>
          <p className="auth-alt" style={{ marginTop: '1rem' }}>
            <Link to="/login" className="btn btn-primary btn-block">Continue to login</Link>
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-brand">
          <span className="brand-mark">Q</span>
          <h1>Reset your password</h1>
        </div>
        <p className="auth-tagline">Enter your new password below.</p>
        <form onSubmit={handleSubmit} className="auth-form" noValidate>
          <PasswordInput
            label="New password"
            id="reset-password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => {
              setPassword(e.target.value)
              if (fieldErrors.password) setFieldErrors((p) => ({ ...p, password: undefined }))
            }}
            placeholder="At least 8 characters"
            error={fieldErrors.password}
          />
          {password.length > 0 && (
            <div className="password-strength" aria-live="polite">
              <div className="password-strength-bar">
                <div
                  className={`password-strength-fill strength-${strength.level}`}
                  style={{ width: strength.level === 'weak' ? '33%' : strength.level === 'fair' ? '66%' : '100%', backgroundColor: strength.color }}
                />
              </div>
              <span className="password-strength-label" style={{ color: strength.color }}>
                {strength.label}
              </span>
            </div>
          )}
          {password.length > 0 && (
            <ul className="password-requirements" aria-label="Password requirements">
              <li className={hasMinLength ? 'req-met' : 'req-unmet'}>
                {hasMinLength ? '✓' : '○'} At least 8 characters
              </li>
              <li className={hasLetter ? 'req-met' : 'req-unmet'}>
                {hasLetter ? '✓' : '○'} At least one letter
              </li>
              <li className={hasNumber ? 'req-met' : 'req-unmet'}>
                {hasNumber ? '✓' : '○'} At least one number
              </li>
            </ul>
          )}
          <PasswordInput
            label="Confirm new password"
            id="reset-confirm-password"
            autoComplete="new-password"
            value={confirmPassword}
            onChange={(e) => {
              setConfirmPassword(e.target.value)
              if (fieldErrors.confirmPassword) setFieldErrors((p) => ({ ...p, confirmPassword: undefined }))
            }}
            placeholder="Re-enter your password"
            error={fieldErrors.confirmPassword}
          />
          {error && <p className="form-error" role="alert">{error}</p>}
          <button className="btn btn-primary btn-block" disabled={busy} type="submit">
            {busy ? 'Resetting…' : 'Reset password'}
          </button>
        </form>
        <p className="auth-alt" style={{ marginTop: '1rem' }}>
          <Link to="/login">Back to login</Link>
        </p>
      </div>
    </div>
  )
}
