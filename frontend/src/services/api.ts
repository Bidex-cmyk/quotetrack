import type {
  Analytics,
  AuthResponse,
  Customer,
  CustomerWithStats,
  DashboardData,
  NewQuoteItem,
  PublicQuote,
  Quote,
  QuoteRow,
  QuoteStatus,
  User,
} from '../types'

const TOKEN_KEY = 'quotetrack_token'
const USER_KEY = 'quotetrack_user'

export const apiBase = '/api'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setSession(token: string, user: User): void {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function getStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export function clearSession(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

export async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> | undefined),
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${apiBase}${path}`, { ...options, headers })
  if (res.status === 204) return undefined as T

  const body = await res.json().catch(() => null)
  if (!res.ok) {
    const message =
      (body && (body as { error?: string }).error) ||
      `Request failed (${res.status})`
    throw new ApiError(message, res.status)
  }
  return body as T
}

export const api = {
  // Auth
  signup: (email: string, password: string, businessName: string) =>
    request<AuthResponse>('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({ email, password, business_name: businessName }),
    }),
  login: (email: string, password: string) =>
    request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  me: () => request<{ user: User }>('/me'),
  updateMe: (patch: { email?: string; business_name?: string }) =>
    request<{ user: User }>('/me', {
      method: 'PATCH',
      body: JSON.stringify(patch),
    }),

  // Password reset
  forgotPassword: (email: string) =>
    request<{ message: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),
  resetPassword: (token: string, password: string) =>
    request<{ message: string }>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, password }),
    }),

  // Customers
  listCustomers: (q = '') =>
    request<{ customers: CustomerWithStats[] }>(
      `/customers${q ? `?q=${encodeURIComponent(q)}` : ''}`,
    ),
  getCustomer: (id: string) => request<CustomerWithStats>(`/customers/${id}`),
  createCustomer: (c: Partial<Customer>) =>
    request<CustomerWithStats>('/customers', {
      method: 'POST',
      body: JSON.stringify({
        name: c.name ?? '',
        phone: c.phone ?? '',
        email: c.email ?? '',
        company: c.company ?? '',
        notes: c.notes ?? '',
      }),
    }),
  updateCustomer: (id: string, c: Partial<Customer>) =>
    request<Customer>(`/customers/${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        name: c.name ?? '',
        phone: c.phone ?? '',
        email: c.email ?? '',
        company: c.company ?? '',
        notes: c.notes ?? '',
      }),
    }),
  deleteCustomer: (id: string) => request<void>(`/customers/${id}`, { method: 'DELETE' }),

  // Quotes
  listQuotes: (params: Record<string, string> = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request<{ quotes: QuoteRow[] }>(`/quotes${qs ? `?${qs}` : ''}`)
  },
  getQuote: (id: string) => request<Quote>(`/quotes/${id}`),
  createQuote: (payload: {
    customer_id: string
    title: string
    notes?: string
    status?: QuoteStatus
    quote_date?: string
    expiry_date?: string | null
    follow_up_date?: string | null
    items: NewQuoteItem[]
  }) =>
    request<Quote>('/quotes', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  updateQuote: (
    id: string,
    payload: {
      customer_id: string
      title: string
      notes?: string
      quote_date?: string
      expiry_date?: string | null
      follow_up_date?: string | null
      items: NewQuoteItem[]
    },
  ) =>
    request<Quote>(`/quotes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),
  setStatus: (id: string, status: QuoteStatus) =>
    request<Quote>(`/quotes/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),
  setFollowUp: (id: string, follow_up_date: string | null) =>
    request<Quote>(`/quotes/${id}/follow-up`, {
      method: 'PATCH',
      body: JSON.stringify({ follow_up_date }),
    }),
  deleteQuote: (id: string) => request<void>(`/quotes/${id}`, { method: 'DELETE' }),

  // Dashboard + analytics
  dashboard: () => request<DashboardData>('/dashboard'),
  analytics: () => request<Analytics>('/analytics'),

  // Public (no auth)
  publicQuote: (publicId: string) =>
    request<PublicQuote>(`/public/quotes/${publicId}`),
}
