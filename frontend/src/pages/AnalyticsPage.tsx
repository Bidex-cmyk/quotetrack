import { useEffect, useState } from 'react'
import { api } from '../services/api'
import type { Analytics } from '../types'
import { Spinner } from '../components/Spinner'
import EmptyState from '../components/EmptyState'
import { formatMoney, percent } from '../utils/format'

export default function AnalyticsPage() {
  const [data, setData] = useState<Analytics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .analytics()
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load analytics'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <Spinner size={40} />
  if (error || !data)
    return <EmptyState title="Could not load analytics" message={error || 'No data'} />

  const rows: Array<{ label: string; count: number; value: string }> = [
    { label: 'Total quotes', count: data.total_quotes, value: formatMoney(data.total_value) },
    { label: 'Won', count: data.won, value: formatMoney(data.won_value) },
    { label: 'Lost', count: data.lost, value: formatMoney(data.lost_value) },
    { label: 'Waiting', count: data.waiting, value: formatMoney(data.waiting_value) },
  ]

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>Analytics</h2>
          <p className="muted">Your win/loss picture at a glance.</p>
        </div>
      </div>

      <div className="stat-grid">
        <div className="stat-card stat-highlight">
          <span className="stat-value">{percent(data.win_rate)}</span>
          <span className="stat-label">Win rate</span>
        </div>
        {rows.map((r) => (
          <div className="stat-card" key={r.label}>
            <span className="stat-value">{r.count}</span>
            <span className="stat-label">{r.label}</span>
            <span className="stat-sub">{r.value}</span>
          </div>
        ))}
      </div>

      <div className="card">
        <h3>Pipeline value</h3>
        <div className="bar-stack">
          <div
            className="bar-seg bar-won"
            style={{ flexGrow: data.total_value > 0 ? data.won_value / data.total_value : 0 }}
            title={`Won: ${formatMoney(data.won_value)}`}
          />
          <div
            className="bar-seg bar-waiting"
            style={{ flexGrow: data.total_value > 0 ? data.waiting_value / data.total_value : 0 }}
            title={`Waiting: ${formatMoney(data.waiting_value)}`}
          />
          <div
            className="bar-seg bar-lost"
            style={{ flexGrow: data.total_value > 0 ? data.lost_value / data.total_value : 0 }}
            title={`Lost: ${formatMoney(data.lost_value)}`}
          />
        </div>
        <div className="bar-legend">
          <span><i className="dot dot-won" /> Won {formatMoney(data.won_value)}</span>
          <span><i className="dot dot-waiting" /> Waiting {formatMoney(data.waiting_value)}</span>
          <span><i className="dot dot-lost" /> Lost {formatMoney(data.lost_value)}</span>
        </div>
      </div>
    </div>
  )
}