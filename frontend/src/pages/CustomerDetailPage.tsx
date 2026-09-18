import { useCallback, useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useParams } from 'react-router-dom'
import { api } from '../services/api'
import type { CustomerWithStats, QuoteRow } from '../types'
import { Spinner } from '../components/Spinner'
import EmptyState from '../components/EmptyState'
import StatusBadge from '../components/StatusBadge'
import Modal from '../components/Modal'
import { useToast } from '../contexts/ToastContext'
import { formatDate, formatMoney } from '../utils/format'

export default function CustomerDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const [customer, setCustomer] = useState<CustomerWithStats | null>(null)
  const [quotes, setQuotes] = useState<QuoteRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editForm, setEditForm] = useState({
    name: '',
    phone: '',
    email: '',
    company: '',
    notes: '',
  })
  const [busy, setBusy] = useState(false)

  const load = useCallback(() => {
    if (!id) return
    setLoading(true)
    Promise.all([api.getCustomer(id), api.listQuotes({ customer_id: id })])
      .then(([c, qres]) => {
        setCustomer(c)
        setQuotes(qres.quotes)
        setEditForm({
          name: c.name,
          phone: c.phone,
          email: c.email,
          company: c.company,
          notes: c.notes,
        })
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load customer'))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  if (!id) return <Navigate to="/app/customers" replace />
  if (loading) return <Spinner size={40} />
  if (error || !customer)
    return <EmptyState title="Could not load customer" message={error || 'Not found'} />

  const saveEdit = async () => {
    if (!editForm.name.trim()) {
      toast('Name is required', 'error')
      return
    }
    setBusy(true)
    try {
      await api.updateCustomer(customer.id, editForm)
      toast('Customer updated')
      setEditing(false)
      load()
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Could not update', 'error')
    } finally {
      setBusy(false)
    }
  }

  const handleDelete = async () => {
    setBusy(true)
    try {
      await api.deleteCustomer(customer.id)
      toast('Customer deleted')
      navigate('/app/customers')
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Could not delete', 'error')
      setBusy(false)
    }
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>{customer.name}</h2>
          <p className="muted">
            {customer.company || customer.phone || customer.email || 'No contact details'}
          </p>
        </div>
        <div className="header-actions">
          <button className="btn btn-ghost" onClick={() => setEditing(true)}>
            Edit
          </button>
          <button
            className="btn btn-danger"
            onClick={() => setConfirmDelete(true)}
            disabled={quotes.length > 0}
            title={quotes.length > 0 ? 'Delete quotes first' : 'Delete customer'}
          >
            Delete
          </button>
        </div>
      </div>

      <div className="detail-grid">
        <div className="card">
          <h3>Details</h3>
          <dl className="detail-list">
            <dt>Phone</dt>
            <dd>{customer.phone || '—'}</dd>
            <dt>Email</dt>
            <dd>{customer.email || '—'}</dd>
            <dt>Company</dt>
            <dd>{customer.company || '—'}</dd>
            <dt>Notes</dt>
            <dd>{customer.notes || '—'}</dd>
          </dl>
        </div>
        <div className="card">
          <h3>Quoting</h3>
          <dl className="detail-list">
            <dt>Quotes</dt>
            <dd>{customer.quote_count}</dd>
            <dt>Total value</dt>
            <dd>{formatMoney(customer.total_value)}</dd>
            <dt>Customer since</dt>
            <dd>{formatDate(customer.created_at.slice(0, 10))}</dd>
          </dl>
        </div>
      </div>

      <section className="section">
        <div className="section-head">
          <h3>Quotes</h3>
          <Link to="/app/quotes/new" state={{ customerId: customer.id }} className="btn btn-sm btn-primary">
            New quote
          </Link>
        </div>
        {quotes.length === 0 ? (
          <EmptyState compact title="No quotes yet" message="Quote this customer to get started." />
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Number</th>
                  <th>Title</th>
                  <th>Status</th>
                  <th>Date</th>
                  <th className="num">Total</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {quotes.map((q) => (
                  <tr key={q.id}>
                    <td>{q.quote_number}</td>
                    <td>
                      <Link to={`/app/quotes/${q.id}`} className="row-link">
                        {q.title}
                      </Link>
                    </td>
                    <td><StatusBadge status={q.status} /></td>
                    <td>{formatDate(q.quote_date)}</td>
                    <td className="num">{formatMoney(q.total)}</td>
                    <td className="actions-cell">
                      <Link to={`/app/quotes/${q.id}`} className="btn btn-ghost btn-sm">
                        View
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <Modal open={editing} onClose={() => setEditing(false)} title="Edit customer">
        <label className="field">
          <span>Name *</span>
          <input
            type="text"
            value={editForm.name}
            onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))}
          />
        </label>
        <label className="field">
          <span>Phone</span>
          <input
            type="tel"
            value={editForm.phone}
            onChange={(e) => setEditForm((f) => ({ ...f, phone: e.target.value }))}
          />
        </label>
        <label className="field">
          <span>Email</span>
          <input
            type="email"
            value={editForm.email}
            onChange={(e) => setEditForm((f) => ({ ...f, email: e.target.value }))}
          />
        </label>
        <label className="field">
          <span>Company</span>
          <input
            type="text"
            value={editForm.company}
            onChange={(e) => setEditForm((f) => ({ ...f, company: e.target.value }))}
          />
        </label>
        <label className="field">
          <span>Notes</span>
          <textarea
            rows={3}
            value={editForm.notes}
            onChange={(e) => setEditForm((f) => ({ ...f, notes: e.target.value }))}
          />
        </label>
        <div className="form-actions">
          <button className="btn btn-ghost" onClick={() => setEditing(false)}>
            Cancel
          </button>
          <button className="btn btn-primary" onClick={saveEdit} disabled={busy}>
            {busy ? 'Saving…' : 'Save'}
          </button>
        </div>
      </Modal>

      <Modal open={confirmDelete} onClose={() => setConfirmDelete(false)} title="Delete customer">
        <p>
          Are you sure you want to delete <strong>{customer.name}</strong>? This cannot be undone.
        </p>
        <div className="form-actions">
          <button className="btn btn-ghost" onClick={() => setConfirmDelete(false)}>
            Cancel
          </button>
          <button className="btn btn-danger" onClick={handleDelete} disabled={busy}>
            {busy ? 'Deleting…' : 'Delete customer'}
          </button>
        </div>
      </Modal>
    </div>
  )
}