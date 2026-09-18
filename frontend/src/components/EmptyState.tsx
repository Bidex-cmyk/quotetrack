import type { ReactNode } from 'react'

export default function EmptyState({
  title,
  message,
  action,
  compact = false,
}: {
  title: string
  message?: string
  action?: ReactNode
  compact?: boolean
}) {
  return (
    <div className={`empty-state${compact ? ' empty-state-compact' : ''}`}>
      <h3>{title}</h3>
      {message && <p>{message}</p>}
      {action && <div className="empty-state-action">{action}</div>}
    </div>
  )
}