import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <div className="center-screen">
      <div className="not-found">
        <h2>404</h2>
        <p className="muted">Page not found.</p>
        <Link to="/app" className="btn btn-ghost">
          Back to dashboard
        </Link>
      </div>
    </div>
  )
}