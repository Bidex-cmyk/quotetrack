import { useState, type FormEvent } from 'react'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import { useTheme, type Theme } from '../contexts/ThemeContext'
import { Spinner } from '../components/Spinner'

const themeOptions: Array<{ value: Theme; label: string; icon: string; description: string }> = [
  { value: 'light', label: 'Light', icon: '☀️', description: 'Always use light theme' },
  { value: 'dark', label: 'Dark', icon: '🌙', description: 'Always use dark theme' },
  { value: 'system', label: 'System', icon: '💻', description: 'Match your device settings' },
]

export default function SettingsPage() {
  const { user, setUser } = useAuth()
  const { toast } = useToast()
  const { theme, setTheme } = useTheme()
  const [businessName, setBusinessName] = useState(user?.business_name ?? '')
  const [email, setEmail] = useState(user?.email ?? '')
  const [busy, setBusy] = useState(false)

  if (!user) return <Spinner size={40} />

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    try {
      const { user: updated } = await api.updateMe({
        business_name: businessName.trim(),
        email: email.trim(),
      })
      setUser(updated)
      toast('Settings saved')
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Could not save settings', 'error')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>Settings</h2>
          <p className="muted">Your business profile and preferences.</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-card form-card-narrow">
        <label className="field">
          <span>Business name</span>
          <input
            type="text"
            value={businessName}
            onChange={(e) => setBusinessName(e.target.value)}
          />
        </label>
        <label className="field">
          <span>Email</span>
          <input
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </label>
        <div className="form-actions">
          <button className="btn btn-primary" disabled={busy} type="submit">
            {busy ? 'Saving…' : 'Save settings'}
          </button>
        </div>
      </form>

      <div className="section" style={{ marginTop: '1.5rem' }}>
        <div className="section-head">
          <h3>Appearance</h3>
        </div>
        <div className="form-card form-card-narrow">
          <div className="field">
            <span>Theme</span>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: '.5rem',
                marginTop: '.25rem',
              }}
            >
              {themeOptions.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  className={`btn ${theme === opt.value ? 'btn-primary' : 'btn-ghost'}`}
                  onClick={() => setTheme(opt.value)}
                  aria-pressed={theme === opt.value}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: '.2rem',
                    padding: '.65rem .5rem',
                    minHeight: '60px',
                  }}
                >
                  <span style={{ fontSize: '1.1rem' }}>{opt.icon}</span>
                  <span style={{ fontSize: '.85rem', fontWeight: 600 }}>{opt.label}</span>
                  <span
                    style={{
                      fontSize: '.72rem',
                      opacity: 0.8,
                      textAlign: 'center',
                    }}
                  >
                    {opt.description}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
