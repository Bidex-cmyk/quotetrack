package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"quotetrack/backend/internal/httpapi"
)

// authPathPrefix is the only path prefix subject to rate limiting.
const authPathPrefix = "/api/auth/"

// cleanupInterval is how often the janitor sweeps idle entries.
const cleanupInterval = 5 * time.Minute

// idleTTL is how long an untouched bucket must be idle before the janitor
// drops it. It is comfortably longer than the refill period (1 minute at the
// default 5/min), so evicting an idle key cannot hand out extra tokens to a
// client that is actively retrying.
const idleTTL = 10 * time.Minute

// RateLimiter limits requests to auth endpoints per client IP, in memory.
// It is intentionally in-process: QuoteTrack runs as a single Render
// instance with no Redis, so a shared store would buy nothing.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	limit    rate.Limit
	burst    int
	stop     chan struct{}
	stopOnce sync.Once
	now      func() time.Time // injectable for tests
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewAuthRateLimiter returns a limiter that allows perMinute requests per
// minute (sustained) with the given burst, keyed by client IP + endpoint
// path, enforced only under /api/auth/.
func NewAuthRateLimiter(perMinute, burst int) *RateLimiter {
	if perMinute < 1 {
		perMinute = 1
	}
	if burst < 1 {
		burst = 1
	}
	l := &RateLimiter{
		buckets: make(map[string]*bucket),
		limit:   rate.Limit(perMinute) / 60.0, // rate.Limit is events per second
		burst:   burst,
		stop:    make(chan struct{}),
		now:     time.Now,
	}
	go l.janitor()
	return l
}

// Wrap applies the rate limit to /api/auth/* paths and passes everything
// else (health checks, public quote links, authenticated CRUD) through
// untouched.
func (l *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, authPathPrefix) {
			key := clientIP(r) + "|" + r.URL.Path // per IP, per endpoint
			if !l.allow(key) {
				httpapi.Error(w, http.StatusTooManyRequests,
					"Too many requests. Please try again later.")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// allow consumes one token and reports whether the request may proceed.
func (l *RateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.buckets[key] = b
	}
	b.lastSeen = l.now()
	return b.limiter.Allow()
}

// janitor periodically drops idle buckets so the map cannot grow without
// bound under a flood of distinct client IPs.
func (l *RateLimiter) janitor() {
	t := time.NewTicker(cleanupInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			l.cleanup()
		case <-l.stop:
			return
		}
	}
}

func (l *RateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := l.now().Add(-idleTTL)
	for k, b := range l.buckets {
		if b.lastSeen.Before(cutoff) {
			delete(l.buckets, k)
		}
	}
}

// Close stops the janitor goroutine. The server keeps one limiter for the
// life of the process, so this is mainly for tests.
func (l *RateLimiter) Close() {
	l.stopOnce.Do(func() { close(l.stop) })
}

// clientIP extracts the client IP for rate-limit keying.
//
// Render's proxy sits in front of the app, so RemoteAddr alone would be the
// proxy's address and every user would share one bucket. We therefore read
// X-Forwarded-For and trust only its LEFTMOST value, and only if it parses
// as a valid IP; otherwise we fall back to RemoteAddr. Assumption: Render's
// router sets X-Forwarded-For to the real client address (overriding or
// prepending it ahead of anything a client sent), which makes the leftmost
// entry trustworthy here. If the app is ever exposed directly to the
// internet, clients could spoof that header — bind rate limiting to
// RemoteAddr in that case.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if ip := net.ParseIP(first); ip != nil {
			return ip.String()
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
