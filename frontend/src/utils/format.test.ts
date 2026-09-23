import { describe, it, expect } from 'vitest'
import { formatMoney } from './format'

describe('formatMoney', () => {
  it('formats dollars as USD', () => {
    expect(formatMoney(50000)).toBe('$50,000.00')
    expect(formatMoney(1234.56)).toBe('$1,234.56')
    expect(formatMoney(0)).toBe('$0.00')
  })
})
