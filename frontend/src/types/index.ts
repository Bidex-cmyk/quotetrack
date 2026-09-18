export interface User {
  id: string
  email: string
  business_name: string
  created_at: string
  updated_at: string
}

export interface Customer {
  id: string
  user_id: string
  name: string
  phone: string
  email: string
  company: string
  notes: string
  created_at: string
  updated_at: string
}

export interface CustomerWithStats extends Customer {
  quote_count: number
  total_value: number
}

export interface QuoteItem {
  id: string
  position: number
  description: string
  quantity: number
  unit_price: number
  total: number
}

export type QuoteStatus =
  | 'draft'
  | 'sent'
  | 'waiting'
  | 'won'
  | 'lost'
  | 'expired'

export interface Quote {
  id: string
  user_id: string
  customer_id: string
  public_id: string
  quote_number: string
  title: string
  notes: string
  status: QuoteStatus
  quote_date: string
  expiry_date: string | null
  follow_up_date: string | null
  subtotal: number
  total: number
  created_at: string
  updated_at: string
  items: QuoteItem[]
  customer?: Customer
  item_count?: number
}

export interface QuoteRow extends Quote {
  item_count: number
}

export interface NewQuoteItem {
  description: string
  quantity: number
  unit_price: number
}

export interface DashboardStats {
  total_quotes: number
  waiting: number
  follow_ups_due_today: number
  won: number
  lost: number
}

export interface DashboardData {
  stats: DashboardStats
  follow_ups: QuoteRow[]
}

export interface Analytics {
  total_quotes: number
  total_value: number
  won: number
  won_value: number
  lost: number
  lost_value: number
  waiting: number
  waiting_value: number
  win_rate: number
}

export interface PublicQuote {
  business_name: string
  customer_name: string
  quote_number: string
  title: string
  notes: string
  status: QuoteStatus
  quote_date: string
  expiry_date: string | null
  total: number
  items: QuoteItem[]
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: string
}