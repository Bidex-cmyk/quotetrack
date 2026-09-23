import { Link, Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { Spinner } from '../components/Spinner'

export default function LandingPage() {
  const { user, initializing } = useAuth()

  if (initializing) {
    return (
      <div className="center-screen">
        <Spinner size={40} />
      </div>
    )
  }

  if (user) {
    return <Navigate to="/app" replace />
  }

  return (
    <div className="landing-page">
      {/* Section 1 — Top nav */}
      <header className="landing-nav">
        <div className="landing-nav-inner">
          <Link to="/" className="brand" aria-label="QuoteTrack home">
            <span className="brand-mark">Q</span>
            <span className="brand-name">QuoteTrack</span>
          </Link>
          <nav className="landing-nav-actions" aria-label="Primary">
            <Link to="/login" className="landing-link">
              Log in
            </Link>
            <Link to="/signup" className="btn btn-primary">
              Get started
            </Link>
          </nav>
        </div>
      </header>

      {/* Section 2 — Hero */}
      <section className="landing-hero">
        <h1 className="landing-hero-title">Send quotes that get answers.</h1>
        <p className="landing-hero-sub">
          Create a professional quote in under a minute, share it as a link, and know the moment your
          client opens it.
        </p>
        <div className="landing-ctas">
          <Link to="/signup" className="btn btn-primary">
            Get started free
          </Link>
          <a href="#how" className="btn btn-ghost">
            See how it works
          </a>
        </div>
        <p className="landing-muted">No signup required to look around.</p>
      </section>

      {/* Section 3 — How it works */}
      <section id="how" className="landing-how">
        <div className="landing-how-inner">
          <h2 className="landing-section-title">How it works</h2>
          <div className="landing-steps">
            {/* Step 1 */}
            <div className="landing-step">
              <div className="landing-step-badge">1</div>
              <div className="landing-step-icon" aria-hidden="true">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                  <circle cx="9" cy="7" r="4" />
                  <path d="M19 8v6M22 11l-3 3-3-3" />
                </svg>
              </div>
              <h3>Add your customer</h3>
              <p className="muted">Name, phone, email — done in ten seconds.</p>
            </div>
            {/* Step 2 */}
            <div className="landing-step">
              <div className="landing-step-badge">2</div>
              <div className="landing-step-icon" aria-hidden="true">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                  <line x1="8" y1="13" x2="16" y2="13" />
                  <line x1="8" y1="17" x2="16" y2="17" />
                  <line x1="12" y1="9" x2="12" y2="9" />
                </svg>
              </div>
              <h3>Build the quote</h3>
              <p className="muted">Add line items, set prices, pick an expiry date.</p>
            </div>
            {/* Step 3 */}
            <div className="landing-step">
              <div className="landing-step-badge">3</div>
              <div className="landing-step-icon" aria-hidden="true">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                  <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
                </svg>
              </div>
              <h3>Send the link</h3>
              <p className="muted">Share a link. Your client opens it on any device — no login needed on their end.</p>
            </div>
          </div>
        </div>
      </section>

      {/* Section 4 — Why it's different */}
      <section className="landing-why">
        <h2 className="landing-section-title">Why it&apos;s different</h2>
        <ul className="landing-why-list">
          <li>No bloat — no accounting, no inventory, no CRM. Just quoting.</li>
          <li>Works on your phone — no app to install.</li>
          <li>Know when to follow up — sent, won, or waiting.</li>
        </ul>
      </section>

      {/* Section 5 — Final CTA */}
      <section className="landing-final">
        <h2 className="landing-section-title">Ready to send your first quote?</h2>
        <Link to="/signup" className="btn btn-primary">
          Create your first quote — it&apos;s free
        </Link>
        <p className="landing-muted">Takes about a minute.</p>
      </section>

      {/* Footer */}
      <footer className="landing-footer">Made with QuoteTrack · © 2026</footer>
    </div>
  )
}
