import { useEffect, useState, type FormEvent } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import { api } from '../services/api'
import type { CustomerWithStats, NewQuoteItem, Quote, QuoteStatus } from '../types'
import { Spinner } from '../components/Spinner'
import LineItemsEditor from '../components/LineItemsEditor'
import { useToast } from '../contexts/ToastContext'
import { addDaysISO, todayISO } from '../utils/format'

export default function QuoteFormPage() {
  const { id } = useParams<{ id: string }>()
  const isEdit = Boolean(id)
  const location = useLocation()
  const navigate = useNavigate()
  const { toast } = useToast()

  const preCustomerId = (location.state as { customerId?: string } | null)?.customerId

  const [customers, setCustomers] = useState<CustomerWithStats[]>([])
  const [customerId, setCustomerId] = useState(preCustomerId ?? '')
  const [title, setTitle] = useState('')
  const [notes, setNotes] = useState('')
  const [status, setStatus] = useState<QuoteStatus>('draft')
  const [quoteDate, setQuoteDate] = useState(todayISO())
  const [expiryDate, setExpiryDate] = useState(addDaysISO(30))
  const [followUpDate, setFollowUpDate] = useState(addDaysISO(7))
  const [items, setItems] = useState<NewQuoteItem[]>([])
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [loadedQuote, setLoadedQuote] = useState<Quote | null>(null)

  useEffect(() => {
    Promise.all([api.listCustomers()])
      .then(([res]) => setCustomers(res.customers))
      .catch(() => undefined)
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (!id) return
    api
      .getQuote(id)
      .then((q) => {
        setLoadedQuote(q)
        setCustomerId(q.customer_id)
        setTitle(q.title)
        setNotes(q.notes)
        setStatus(q.status)
        setQuoteDate(q.quote_date || todayISO())
        setExpiryDate(q.expiry_date ?? '')
        setFollowUpDate(q.follow_up_date ?? '')
        setItems(
          q.items.map((it) => ({
            description: it.description,
            quantity: it.quantity,
            unit_price: it.unit_price / 100,
          })),
        )
      })
      .catch(() => toast('Could not load quote', 'error'))
  }, [id, toast])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (!customerId) {
      setError('Choose a customer.')
      return
    }
    if (!title.trim()) {
      setError('A quote title is required.')
      return
    }
    if (items.length === 0) {
      setError('Add at least one line item.')
      return
    }
    if (!quoteDate) {
      setError('A quote date is required.')
      return
    }

    const payload = {
      customer_id: customerId,
      title: title.trim(),
      notes: notes.trim(),
      status,
      quote_date: quoteDate,
      expiry_date: expiryDate || null,
      follow_up_date: followUpDate || null,
      items: items.map((it) => ({
        description: it.description.trim(),
        quantity: it.quantity,
        unit_price: it.unit_price,
      })),
    }

    setBusy(true)
    try {
      const saved = isEdit && id ? await api.updateQuote(id, payload) : await api.createQuote(payload)
      toast(isEdit ? 'Quote updated' : 'Quote created')
      navigate(`/app/quotes/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save quote')
    } finally {
      setBusy(false)
    }
  }

  if (loading) return <Spinner size={40} />
  if (isEdit && !loadedQuote) return null

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>{isEdit ? `Edit ${loadedQuote?.quote_number}` : 'New quote'}</h2>
          <p className="muted">
            {isEdit
              ? `Editing ${loadedQuote?.title}`
              : 'Create a quote to send and start tracking follow-ups.'}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-card">
        <div className="form-grid">
          <label className="field">
            <span>Customer *</span>
            <select value={customerId} onChange={(e) => setCustomerId(e.target.value)}>
              <option value="">Choose a customer…</option>
              {customers.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                  {c.company ? ` — ${c.company}` : ''}
                </option>
              ))}
            </select>
            {customers.length === 0 && (
              <span className="field-hint">
                No customers yet. <Link to="/app/customers/new">Create one first</Link>.
              </span>
            )}
          </label>
          <label className="field">
            <span>Title *</span>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Kitchen renovation"
            />
          </label>
          <label className="field">
            <span>Status</span>
            <select value={status} onChange={(e) => setStatus(e.target.value as QuoteStatus)}>
              <option value="draft">Draft</option>
              <option value="sent">Sent</option>
              <option value="waiting">Waiting</option>
            </select>
          </label>
          <label className="field">
            <span>Quote date</span>
            <input type="date" value={quoteDate} onChange={(e) => setQuoteDate(e.target.value)} />
          </label>
          <label className="field">
            <span>Expiry date</span>
            <input type="date" value={expiryDate} onChange={(e) => setExpiryDate(e.target.value)} />
          </label>
          <label className="field">
            <span>Follow-up date</span>
            <input
              type="date"
              value={followUpDate}
              onChange={(e) => setFollowUpDate(e.target.value)}
            />
          </label>
        </div>

        <div className="field">
          <span>Line items</span>
          <LineItemsEditor items={items} onChange={setItems} />
        </div>

        <label className="field">
          <span>Notes</span>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Terms, payment details, anything the customer should know…"
            rows={3}
          />
        </label>

        {error && <p className="form-error">{error}</p>}

        <div className="form-actions">
          <Link to={isEdit && id ? `/app/quotes/${id}` : '/app/quotes'} className="btn btn-ghost">
            Cancel
          </Link>
          <button className="btn btn-primary" disabled={busy} type="submit">
            {busy ? 'Saving…' : isEdit ? 'Save changes' : 'Create quote'}
          </button>
        </div>
      </form>
    </div>
  )
}