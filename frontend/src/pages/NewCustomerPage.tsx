import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { useToast } from '../contexts/ToastContext'

export default function NewCustomerPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [form, setForm] = useState({ name: '', phone: '', email: '', company: '', notes: '' })
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const set = (key: keyof typeof form) => (e: { target: { value: string } }) =>
    setForm((f) => ({ ...f, [key]: e.target.value }))

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (!form.name.trim()) {
      setError('A customer name is required.')
      return
    }
    setBusy(true)
    try {
      const created = await api.createCustomer(form)
      toast('Customer created')
      navigate(`/app/customers/${created.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create customer')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>New customer</h2>
          <p className="muted">Add someone you might quote.</p>
        </div>
      </div>
      <form onSubmit={handleSubmit} className="form-card">
        <label className="field">
          <span>Name *</span>
          <input type="text" value={form.name} onChange={set('name')} placeholder="Jane Smith" />
        </label>
        <label className="field">
          <span>Phone</span>
          <input type="tel" value={form.phone} onChange={set('phone')} placeholder="+1 555 123 4567" />
        </label>
        <label className="field">
          <span>Email</span>
          <input type="email" value={form.email} onChange={set('email')} placeholder="jane@example.com" />
        </label>
        <label className="field">
          <span>Company</span>
          <input type="text" value={form.company} onChange={set('company')} placeholder="Smith & Co" />
        </label>
        <label className="field">
          <span>Notes</span>
          <textarea
            value={form.notes}
            onChange={set('notes')}
            placeholder="Anything worth remembering…"
            rows={4}
          />
        </label>
        {error && <p className="form-error">{error}</p>}
        <div className="form-actions">
          <button type="button" className="btn btn-ghost" onClick={() => navigate(-1)}>
            Cancel
          </button>
          <button className="btn btn-primary" disabled={busy} type="submit">
            {busy ? 'Saving…' : 'Create customer'}
          </button>
        </div>
      </form>
    </div>
  )
}