import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import type { User } from '../types'
import { api, clearSession, getStoredUser, getToken, setSession } from '../services/api'

interface AuthContextValue {
  user: User | null
  initializing: boolean
  login: (email: string, password: string) => Promise<void>
  signup: (email: string, password: string, businessName: string) => Promise<void>
  logout: () => void
  setUser: (user: User) => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => getStoredUser())
  const [initializing, setInitializing] = useState(true)

  useEffect(() => {
    let active = true
    async function verify() {
      if (!getToken()) {
        setInitializing(false)
        return
      }
      try {
        const { user } = await api.me()
        if (active) {
          setUser(user)
          setSession(getToken() ?? '', user)
        }
      } catch {
        if (active) {
          clearSession()
          setUser(null)
        }
      } finally {
        if (active) setInitializing(false)
      }
    }
    verify()
    return () => {
      active = false
    }
  }, [])

  const login = async (email: string, password: string) => {
    const res = await api.login(email, password)
    setSession(res.token, res.user)
    setUser(res.user)
  }

  const signup = async (email: string, password: string, businessName: string) => {
    const res = await api.signup(email, password, businessName)
    setSession(res.token, res.user)
    setUser(res.user)
  }

  const logout = () => {
    clearSession()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, initializing, login, signup, logout, setUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}