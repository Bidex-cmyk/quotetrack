import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../services/api'
import type { PublicQuote } from '../types'
import { Spinner } from '../components/Spinner'
import StatusBadge from '../components/StatusBadge'
import { formatDate, formatMoney } from '../utils/format'

export default function PublicQuotePage() {
  const { publicId } = useParams<{ publicId: string }>()
  const [quote, setQuote] = useState<PublicQuote | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!publicId) return
    api
      .publicQuote(publicId)
      .then(setQuote)
      .catch((err) => setError(err instanceof Error ? err.message : 'Quote not found'))
      .finally(() => setLoading(false))
  }, [publicId])

  if (loading) return <Spinner size={40} />
  if (error || !quote) return <div className="center-screen"><p className="muted">Quote not found or expired.</p></div>

  return (
    <div className="public-quote-page">
      <div className="public-quote">
        <div className="public-quote-header">
          <div className="brand-mark large">Q</div>
          <div>
            <h2>{quote.business_name}</h2>
            <p className="muted">{quote.customer_name}</p>
          </div>
          <StatusBadge status={quote.status} />
        </div>

        <div className="public-quote-body">
          <h3>{quote.title}</h3>
          <p className="muted">Quote {quote.quote_number} &middot; {formatDate(quote.quote_date)}</p>
        </div>

        {quote.notes && <p className="notes-text">{quote.notes}</p>}

        <table className="table">
          <thead>
            <tr>
              <th>Description</th>
              <th className="num">Qty</th>
              <th className="num">Unit price</th>
              <th className="num">Total</th>
            </tr>
          </thead>
          <tbody>
            {quote.items.map((it) => (
              <tr key={it.id}>
                <td>{it.description}</td>
                <td className="num">{it.quantity}</td>
                <td className="num">{formatMoney(it.unit_price)}</td>
                <td className="num">{formatMoney(it.total)}</td>
              </tr>
            ))}
            <tr className="table-subtotal-row">
              <td colSpan={3}>Total</td>
              <td className="num"><strong>{formatMoney(quote.total)}</strong></td>
            </tr>
          </tbody>
        </table>

        {quote.expiry_date && (
          <p className="muted" style={{ marginTop: '1.5rem' }}>
            This quote is valid until {formatDate(quote.expiry_date)}.
          </p>
        )}
      </div>
    </div>
  )
}