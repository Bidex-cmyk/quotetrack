import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../services/api'
import type { CustomerWithStats } from '../types'
import { Spinner } from '../components/Spinner'
import EmptyState from '../components/EmptyState'
import { formatMoney } from '../utils/format'

export default function CustomersPage() {
  const [customers, setCustomers] = useState<CustomerWithStats[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')

  useEffect(() => {
    const t = setTimeout(() => setDebounced(search.trim().toLowerCase()), 250)
    return () => clearTimeout(t)
  }, [search])

  const load = useCallback(() => {
    setLoading(true)
    api
      .listCustomers(debounced)
      .then((res) => setCustomers(res.customers))
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load customers'))
      .finally(() => setLoading(false))
  }, [debounced])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>Customers</h2>
          <p className="muted">The people you quote for.</p>
        </div>
        <Link to="/app/customers/new" className="btn btn-primary">
          New customer
        </Link>
      </div>

      <div className="toolbar">
        <input
          type="search"
          className="search-input"
          placeholder="Search customers…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {loading ? (
        <Spinner size={32} />
      ) : error ? (
        <EmptyState title="Could not load customers" message={error} />
      ) : customers.length === 0 ? (
        <EmptyState
          title={debounced ? 'No customers match' : 'No customers yet'}
          message={
            debounced
              ? 'Try a different search.'
              : 'Create your first customer to start quoting.'
          }
          action={
            debounced ? undefined : (
              <Link to="/app/customers/new" className="btn btn-primary">
                New customer
              </Link>
            )
          }
        />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Phone</th>
                <th>Company</th>
                <th className="num">Quotes</th>
                <th className="num">Total value</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {customers.map((c) => (
                <tr key={c.id}>
                  <td>
                    <Link to={`/app/customers/${c.id}`} className="row-link">
                      {c.name}
                    </Link>
                  </td>
                  <td>{c.phone || '—'}</td>
                  <td>{c.company || '—'}</td>
                  <td className="num">{c.quote_count}</td>
                  <td className="num">{formatMoney(c.total_value)}</td>
                  <td className="actions-cell">
                    <Link to={`/app/customers/${c.id}`} className="btn btn-ghost btn-sm">
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