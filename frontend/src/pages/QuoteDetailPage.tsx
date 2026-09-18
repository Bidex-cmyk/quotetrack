import { useCallback, useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useParams } from 'react-router-dom'
import { api } from '../services/api'
import type { Quote } from '../types'
import { Spinner } from '../components/Spinner'
import EmptyState from '../components/EmptyState'
import StatusBadge, { statusLabel } from '../components/StatusBadge'
import Modal from '../components/Modal'
import { useToast } from '../contexts/ToastContext'
import { addDaysISO, formatDate, formatMoney, todayISO } from '../utils/format'

export default function QuoteDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const [quote, setQuote] = useState<Quote | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [followUpModal, setFollowUpModal] = useState(false)
  const [followUpDate, setFollowUpDate] = useState(addDaysISO(7))
  const [busy, setBusy] = useState(false)

  const load = useCallback(() => {
    if (!id) return
    setLoading(true)
    api
      .getQuote(id)
      .then(setQuote)
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load quote'))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  if (!id) return <Navigate to="/app/quotes" replace />
  if (loading) return <Spinner size={40} />
  if (error || !quote) return <EmptyState title="Could not load quote" message={error || 'Not found'} />

  const changeStatus = async (status: Quote['status']) => {
    try {
      const updated = await api.setStatus(quote.id, status)
      setQuote(updated)
      toast(`Quote marked as ${statusLabel(status)}`)
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Could not update status', 'error')
    }
  }

  const saveFollowUp = async () => {
    setBusy(true)
    try {
      const updated = await api.setFollowUp(quote.id, followUpDate || null)
      setQuote(updated)
      setFollowUpModal(false)
      toast(followUpDate ? 'Follow-up date set' : 'Follow-up cleared')
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Could not update follow-up', 'error')
    } finally {
      setBusy(false)
    }
  }

  const handleDelete = async () => {
    setBusy(true)
    try {
      await api.deleteQuote(quote.id)
      toast('Quote deleted')
      navigate('/app/quotes')
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Could not delete', 'error')
      setBusy(false)
    }
  }

  const shareUrl = `${window.location.origin}/quote/${quote.public_id}`
  const isOpen = !['won', 'lost'].includes(quote.status)

  return (
    <div>
      <div className="page-header">
        <div>
          <div className="quote-title-row">
            <h2>{quote.quote_number}</h2>
            <StatusBadge status={quote.status} />
          </div>
          <p className="muted">{quote.title}</p>
        </div>
        <div className="header-actions">
          <Link to={`/app/quotes/${quote.id}/edit`} className="btn btn-ghost">
            Edit
          </Link>
          <button className="btn btn-danger" onClick={() => setConfirmDelete(true)}>
            Delete
          </button>
        </div>
      </div>

      <div className="quote-toolbar">
        {isOpen && (
          <>
            <button className="btn btn-success" onClick={() => changeStatus('won')}>
              ✓ Mark won
            </button>
            <button className="btn btn-ghost btn-danger-text" onClick={() => changeStatus('lost')}>
              Mark lost
            </button>
          </>
        )}
        <button className="btn btn-ghost" onClick={() => setFollowUpModal(true)}>
          {quote.follow_up_date ? 'Change follow-up' : 'Set follow-up'}
        </button>
        <a className="btn btn-primary" href={shareUrl} target="_blank" rel="noreferrer">
          Share ↗
        </a>
      </div>

      <div className="detail-grid">
        <div className="card">
          <h3>Customer</h3>
          <dl className="detail-list">
            <dt>Name</dt>
            <dd>{quote.customer?.name ?? '—'}</dd>
            <dt>Phone</dt>
            <dd>{quote.customer?.phone || '—'}</dd>
            <dt>Email</dt>
            <dd>{quote.customer?.email || '—'}</dd>
            <dt>Company</dt>
            <dd>{quote.customer?.company || '—'}</dd>
          </dl>
          {quote.customer && (
            <Link to={`/app/customers/${quote.customer.id}`} className="row-link">
              View customer →
            </Link>
          )}
        </div>
        <div className="card">
          <h3>Details</h3>
          <dl className="detail-list">
            <dt>Quote date</dt>
            <dd>{formatDate(quote.quote_date)}</dd>
            <dt>Expiry date</dt>
            <dd>{formatDate(quote.expiry_date)}</dd>
            <dt>Follow-up</dt>
            <dd>{formatDate(quote.follow_up_date)}</dd>
            <dt>Status</dt>
            <dd><StatusBadge status={quote.status} /></dd>
          </dl>
        </div>
      </div>

      {quote.notes && (
        <div className="card">
          <h3>Notes</h3>
          <p className="notes-text">{quote.notes}</p>
        </div>
      )}

      <div className="card">
        <h3>Line items</h3>
        {quote.items.length === 0 ? (
          <p className="muted">No line items.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>#</th>
                <th>Description</th>
                <th className="num">Qty</th>
                <th className="num">Unit price</th>
                <th className="num">Total</th>
              </tr>
            </thead>
            <tbody>
              {quote.items.map((it) => (
                <tr key={it.id}>
                  <td>{it.position + 1}</td>
                  <td>{it.description}</td>
                  <td className="num">{it.quantity}</td>
                  <td className="num">{formatMoney(it.unit_price)}</td>
                  <td className="num">{formatMoney(it.total)}</td>
                </tr>
              ))}
              <tr className="table-subtotal-row">
                <td colSpan={4}>Total</td>
                <td className="num"><strong>{formatMoney(quote.total)}</strong></td>
              </tr>
            </tbody>
          </table>
        )}
      </div>

      <Modal open={followUpModal} onClose={() => setFollowUpModal(false)} title="Set a follow-up">
        <p className="muted">
          QuoteTrack will remind you to reach out to this customer on this date.
        </p>
        <label className="field">
          <span>Follow-up date</span>
          <input
            type="date"
            value={followUpDate}
            min={todayISO()}
            onChange={(e) => setFollowUpDate(e.target.value)}
          />
        </label>
        <div className="quick-snooze">
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setFollowUpDate(addDaysISO(7))}>
            +7 days
          </button>
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setFollowUpDate(addDaysISO(14))}>
            +14 days
          </button>
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setFollowUpDate(addDaysISO(30))}>
            +30 days
          </button>
        </div>
        <div className="form-actions">
          <button className="btn btn-ghost" onClick={() => setFollowUpModal(false)}>
            Cancel
          </button>
          <button className="btn btn-primary" onClick={saveFollowUp} disabled={busy}>
            {busy ? 'Saving…' : 'Save'}
          </button>
        </div>
      </Modal>

      <Modal open={confirmDelete} onClose={() => setConfirmDelete(false)} title="Delete quote">
        <p>
          Delete <strong>{quote.quote_number} — {quote.title}</strong>? This cannot be undone.
        </p>
        <div className="form-actions">
          <button className="btn btn-ghost" onClick={() => setConfirmDelete(false)}>
            Cancel
          </button>
          <button className="btn btn-danger" onClick={handleDelete} disabled={busy}>
            {busy ? 'Deleting…' : 'Delete quote'}
          </button>
        </div>
      </Modal>
    </div>
  )
}