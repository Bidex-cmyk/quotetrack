import { useEffect, useRef, useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

const navItems = [
  { to: '/app', label: 'Dashboard' },
  { to: '/app/quotes', label: 'Quotes' },
  { to: '/app/customers', label: 'Customers' },
  { to: '/app/analytics', label: 'Analytics' },
  { to: '/app/settings', label: 'Settings' },
]

export default function AppLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)
  const btnRef = useRef<HTMLButtonElement>(null)

  const closeMenu = () => {
    setMenuOpen(false)
    // focus returns to hamburger after closing
    requestAnimationFrame(() => btnRef.current?.focus())
  }

  const handleLogout = () => {
    setMenuOpen(false)
    logout()
    navigate('/login')
  }

  useEffect(() => {
    if (!menuOpen) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeMenu()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [menuOpen])

  // prevent background scroll when menu open (subtle, not required but nice)
  useEffect(() => {
    if (menuOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [menuOpen])

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="app-header-inner">
          <NavLink to="/app" className="brand">
            <span className="brand-mark">Q</span>
            <span className="brand-name">QuoteTrack</span>
          </NavLink>
          <nav className="app-nav" aria-label="Primary">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
                end={item.to === '/app'}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
          <div className="app-header-right">
            <span className="business-name" title={user?.business_name}>
              {user?.business_name}
            </span>
            <button className="btn btn-ghost btn-sm" onClick={handleLogout}>
              Log out
            </button>
          </div>
          <button
            ref={btnRef}
            className="mobile-menu-btn"
            aria-label={menuOpen ? 'Close menu' : 'Open menu'}
            aria-expanded={menuOpen}
            onClick={() => setMenuOpen((v) => !v)}
          >
            {menuOpen ? (
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            ) : (
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
                <line x1="3" y1="6" x2="21" y2="6" />
                <line x1="3" y1="12" x2="21" y2="12" />
                <line x1="3" y1="18" x2="21" y2="18" />
              </svg>
            )}
          </button>
        </div>
        {menuOpen && (
          <>
            <div className="mobile-menu-backdrop" onClick={closeMenu} aria-hidden="true" />
            <nav className="mobile-menu" aria-label="Primary">
              {navItems.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.to === '/app'}
                  className={({ isActive }) => `mobile-nav-link${isActive ? ' active' : ''}`}
                  onClick={closeMenu}
                >
                  {item.label}
                </NavLink>
              ))}
              <div className="mobile-menu-divider" />
              {user?.business_name && (
                <div className="mobile-business-name" title={user.business_name}>
                  {user.business_name}
                </div>
              )}
              <button className="mobile-logout-btn" onClick={handleLogout}>
                Log out
              </button>
            </nav>
          </>
        )}
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}