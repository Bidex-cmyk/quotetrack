import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import type { DashboardData } from '../types'
import { Spinner } from '../components/Spinner'
import EmptyState from '../components/EmptyState'
import FollowUpCard from '../components/FollowUpCard'
import { useToast } from '../contexts/ToastContext'
import { addDaysISO } from '../utils/format'

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { toast } = useToast()
  const navigate = useNavigate()

  const load = useCallback(() => {
    setLoading(true)
    api
      .dashboard()
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : 'Could not load dashboard'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const snooze = async (id: string) => {
    await api.setFollowUp(id, addDaysISO(7))
    toast('Follow-up snoozed by 7 days')
    load()
  }

  const markWon = async (id: string) => {
    await api.setStatus(id, 'won')
    toast('Quote marked as won 🎉')
    load()
  }

  if (loading) return <Spinner size={40} />
  if (error) return <EmptyState title="Could not load dashboard" message={error} />
  if (!data) return null

  const { stats, follow_ups } = data

  const cards = [
    { label: 'Total quotes', value: String(stats.total_quotes) },
    { label: 'Open (waiting)', value: String(stats.waiting) },
    { label: 'Follow-ups due today', value: String(stats.follow_ups_due_today) },
    { label: 'Won', value: String(stats.won) },
    { label: 'Lost', value: String(stats.lost) },
  ]

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>Dashboard</h2>
          <p className="muted">Your quotes at a glance.</p>
        </div>
        <Link to="/app/quotes/new" className="btn btn-primary">
          New quote
        </Link>
      </div>

      <div className="stat-grid">
        {cards.map((c) => (
          <div className="stat-card" key={c.label}>
            <span className="stat-value">{c.value}</span>
            <span className="stat-label">{c.label}</span>
          </div>
        ))}
      </div>

      <section className="section">
        <div className="section-head">
          <h3>Follow-ups due</h3>
          <span className="muted">Quotes that need a follow-up call</span>
        </div>
        {follow_ups.length === 0 ? (
          <EmptyState
            compact
            title="Nothing to follow up"
            message="You're all caught up. New quotes with a follow-up date will show up here."
          />
        ) : (
          <div className="follow-up-list">
            {follow_ups.map((q) => (
              <FollowUpCard
                key={q.id}
                title={q.title}
                quoteNumber={q.quote_number}
                customerName={q.customer?.name ?? ''}
                status={q.status}
                followUpDate={q.follow_up_date}
                total={q.total}
                onView={() => navigate(`/app/quotes/${q.id}`)}
                onSnooze={() => snooze(q.id)}
                onComplete={() => markWon(q.id)}
              />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}