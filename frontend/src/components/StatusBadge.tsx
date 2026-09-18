import type { QuoteStatus } from '../types'

const labels: Record<QuoteStatus, string> = {
  draft: 'Draft',
  sent: 'Sent',
  waiting: 'Waiting',
  won: 'Won',
  lost: 'Lost',
  expired: 'Expired',
}

export default function StatusBadge({ status }: { status: QuoteStatus }) {
  return <span className={`status-badge status-${status}`}>{labels[status] ?? status}</span>
}

export function statusLabel(status: QuoteStatus): string {
  return labels[status] ?? status
}