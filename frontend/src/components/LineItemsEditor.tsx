import type { NewQuoteItem } from '../types'
import { formatMoney } from '../utils/format'

export function lineTotal(item: NewQuoteItem): number {
  return Math.round(item.quantity * item.unit_price * 100) / 100
}

export default function LineItemsEditor({
  items,
  onChange,
}: {
  items: NewQuoteItem[]
  onChange: (items: NewQuoteItem[]) => void
}) {
  const update = (index: number, patch: Partial<NewQuoteItem>) => {
    const next = items.map((it, i) => (i === index ? { ...it, ...patch } : it))
    onChange(next)
  }

  const add = () => {
    onChange([...items, { description: '', quantity: 1, unit_price: 0 }])
  }

  const remove = (index: number) => {
    onChange(items.filter((_, i) => i !== index))
  }

  const total = items.reduce((sum, it) => sum + lineTotal(it), 0)

  return (
    <div className="line-items-editor">
      {items.map((item, i) => (
        <div className="line-item" key={i}>
          <input
            type="text"
            className="line-item-desc"
            placeholder="Description"
            value={item.description}
            onChange={(e) => update(i, { description: e.target.value })}
          />
          <input
            type="number"
            className="line-item-qty"
            min="0"
            step="0.01"
            placeholder="Qty"
            value={item.quantity}
            onChange={(e) => update(i, { quantity: parseFloat(e.target.value) || 0 })}
          />
          <span className="line-item-x">&times;</span>
          <input
            type="number"
            className="line-item-price"
            min="0"
            step="0.01"
            placeholder="Price"
            value={item.unit_price}
            onChange={(e) => update(i, { unit_price: parseFloat(e.target.value) || 0 })}
          />
          <span className="line-item-total">{formatMoney(lineTotal(item))}</span>
          <button
            className="btn btn-ghost btn-icon"
            onClick={() => remove(i)}
            aria-label="Remove item"
          >
            &times;
          </button>
        </div>
      ))}
      <div className="line-items-footer">
        <button className="btn btn-ghost" onClick={add}>
          + Add line
        </button>
        <span className="line-total">
          Total: <strong>{formatMoney(total)}</strong>
        </span>
      </div>
    </div>
  )
}
