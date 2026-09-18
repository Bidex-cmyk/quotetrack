import type { QuoteStatus } from '../types'
import { formatMoney, todayISO } from '../utils/format'

export default function FollowUpCard({
  title,
  quoteNumber,
  customerName,
  status,
  followUpDate,
  total,
  onView,
  onSnooze,
  onComplete,
}: {
  title: string
  quoteNumber: string
  customerName: string
  status: QuoteStatus
  followUpDate: string | null
  total: number
  onView: () => void
  onSnooze: () => void
  onComplete: () => void
}) {
  const overdue = followUpDate !== null && followUpDate < todayISO()
  return (
    <div className={`follow-up-card${overdue ? ' overdue' : ''}`}>
      <div className="follow-up-card-main">
        <div className="follow-up-card-top">
          <span className="follow-up-quote-number">{quoteNumber}</span>
          <span className={`status-badge status-small status-${status}`}>{status}</span>
        </div>
        <h4>{title}</h4>
        <p className="muted">{customerName}</p>
        {followUpDate && (
          <p className="follow-up-due">
            {overdue ? 'Overdue' : 'Due'}: {followUpDate}
          </p>
        )}
        <p className="follow-up-total">{formatMoney(total)}</p>
      </div>
      <div className="follow-up-card-actions">
        <button className="btn btn-primary btn-sm" onClick={onView}>
          View
        </button>
        <button className="btn btn-ghost btn-sm" onClick={onSnooze}>
          Snooze 7 days
        </button>
        <button className="btn btn-success btn-sm" onClick={onComplete}>
          Mark won
        </button>
      </div>
    </div>
  )
}