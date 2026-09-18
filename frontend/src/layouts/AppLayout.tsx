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

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="app-header-inner">
          <NavLink to="/app" className="brand">
            <span className="brand-mark">Q</span>
            <span className="brand-name">QuoteTrack</span>
          </NavLink>
          <nav className="app-nav">
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
        </div>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}