import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../services/api'
import type { QuoteRow } from '../types'
import { Spinner } from '../components/Spinner'
import EmptyState from '../components/EmptyState'
import StatusBadge from '../components/StatusBadge'
import { formatDate, formatMoney } from '../utils/format'

const statusFilters: Array<{ value: string; label: string }> = [
  { value: '', label: 'All' },
  { value: 'draft', label: 'Draft' },
  { value: 'sent', label: 'Sent' },
  { value: 'waiting', label: 'Waiting' },
  { value: 'won', label: 'Won' },
  { value: 'lost', label: 'Lost' },
  { value: 'expired', label: 'Expired' },
]

export default function QuotesPage() {
  const [quotes, setQuotes] = useState<QuoteRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [status, setStatus] = useState('')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')

  useEffect(() => {
    const t = setTimeout(() => setDebounced(search.trim().toLowerCase()), 250)
    return () => clearTimeout(t)
  }, [search])

  const load = useCallback(() => {
    setLoading(true)
    const params: Record<string, string> = {}
    if (status) params.status = status
    if (debounced) params.search = debounced
    api
      .listQuotes(params)
      .then((res) => setQuotes(res.quotes))
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load quotes'))
      .finally(() => setLoading(false))
  }, [status, debounced])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>Quotes</h2>
          <p className="muted">Every quote you've sent, tracked.</p>
        </div>
        <Link to="/app/quotes/new" className="btn btn-primary">
          New quote
        </Link>
      </div>

      <div className="toolbar">
        <input
          type="search"
          className="search-input"
          placeholder="Search quotes…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="segmented">
          {statusFilters.map((f) => (
            <button
              key={f.value}
              className={`seg-btn${status === f.value ? ' active' : ''}`}
              onClick={() => setStatus(f.value)}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <Spinner size={32} />
      ) : error ? (
        <EmptyState title="Could not load quotes" message={error} />
      ) : quotes.length === 0 ? (
        <EmptyState
          title={status || debounced ? 'No quotes match your filters' : 'No quotes yet'}
          message={
            status || debounced
              ? 'Try clearing the filters.'
              : 'Create a quote to send to your first customer.'
          }
          action={
            status || debounced ? undefined : (
              <Link to="/app/quotes/new" className="btn btn-primary">
                New quote
              </Link>
            )
          }
        />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Number</th>
                <th>Title</th>
                <th>Customer</th>
                <th>Status</th>
                <th>Date</th>
                <th>Follow-up</th>
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
                  <td>{q.customer?.name ?? '—'}</td>
                  <td><StatusBadge status={q.status} /></td>
                  <td>{formatDate(q.quote_date)}</td>
                  <td>{formatDate(q.follow_up_date)}</td>
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
    </div>
  )
}