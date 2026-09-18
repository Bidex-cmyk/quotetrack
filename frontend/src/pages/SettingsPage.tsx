import { useState, type FormEvent } from 'react'
import { api } from '../services/api'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import { Spinner } from '../components/Spinner'

export default function SettingsPage() {
  const { user, setUser } = useAuth()
  const { toast } = useToast()
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
          <p className="muted">Your business profile.</p>
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
    </div>
  )
}