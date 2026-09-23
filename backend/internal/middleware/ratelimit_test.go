package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestLimiter returns a limiter and a 200-echo handler wrapped by it.
// Tests send far fewer requests than the refill rate allows within the test
// duration, so no clock control is needed.
func newTestLimiter(perMinute, burst int) (*RateLimiter, http.Handler) {
	l := NewAuthRateLimiter(perMinute, burst)
	return l, l.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
}

const rateLimitErrBody = `{"error":"Too many requests. Please try again later."}`

func doRequest(h http.Handler, path, forwardedFor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = "203.0.113.10:4321"
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRateLimitUnderLimitAllows(t *testing.T) {
	l, h := newTestLimiter(5, 5)
	defer l.Close()

	for i := 1; i <= 5; i++ {
		rec := doRequest(h, "/api/auth/login", "198.51.100.7")
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d under limit: got %d, want 200", i, rec.Code)
		}
	}
}

func TestRateLimitOverLimitReturns429WithExactBody(t *testing.T) {
	l, h := newTestLimiter(5, 5)
	defer l.Close()

	for i := 1; i <= 5; i++ {
		if rec := doRequest(h, "/api/auth/login", "198.51.100.7"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, rec.Code)
		}
	}

	rec := doRequest(h, "/api/auth/login", "198.51.100.7")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request 6 over limit: got %d, want 429", rec.Code)
	}
	if got := rec.Body.String(); got != rateLimitErrBody+"\n" {
		t.Errorf("429 body = %q, want %q", got, rateLimitErrBody+"\n")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
	}
}

func TestRateLimitDoesNotApplyToNonAuthPaths(t *testing.T) {
	l, h := newTestLimiter(5, 5)
	defer l.Close()

	for _, path := range []string{"/health", "/api/public/quotes/abc", "/api/quotes"} {
		for i := 1; i <= 20; i++ {
			rec := doRequest(h, path, "198.51.100.7")
			if rec.Code != http.StatusOK {
				t.Fatalf("%s request %d: got %d, want 200 (must not be rate limited)", path, i, rec.Code)
			}
		}
	}
}

func TestRateLimitBucketsArePerIPAndPerEndpoint(t *testing.T) {
	l, h := newTestLimiter(5, 5)
	defer l.Close()

	// Exhaust the budget for one IP on one endpoint.
	for i := 1; i <= 5; i++ {
		if rec := doRequest(h, "/api/auth/login", "198.51.100.7"); rec.Code != http.StatusOK {
			t.Fatalf("setup request %d: got %d, want 200", i, rec.Code)
		}
	}
	if rec := doRequest(h, "/api/auth/login", "198.51.100.7"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("exhausted IP should get 429, got %d", rec.Code)
	}

	// A different client IP still has its own budget...
	if rec := doRequest(h, "/api/auth/login", "198.51.100.8"); rec.Code != http.StatusOK {
		t.Errorf("different IP: got %d, want 200", rec.Code)
	}
	// ...and the same IP on a different endpoint does too.
	if rec := doRequest(h, "/api/auth/signup", "198.51.100.7"); rec.Code != http.StatusOK {
		t.Errorf("same IP, different endpoint: got %d, want 200", rec.Code)
	}
}

func TestRateLimitInvalidForwardedForFallsBackToRemoteAddr(t *testing.T) {
	l, h := newTestLimiter(5, 5)
	defer l.Close()

	// Garbage XFF must not panic and must fall back to RemoteAddr, which is
	// shared with the default test RemoteAddr — so these five exhaust that
	// bucket together.
	for i := 1; i <= 5; i++ {
		if rec := doRequest(h, "/api/auth/login", "not-an-ip"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, rec.Code)
		}
	}
	if rec := doRequest(h, "/api/auth/login", "not-an-ip"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request 6: got %d, want 429", rec.Code)
	}
	// And a valid XFF is a separate bucket from the fallback RemoteAddr bucket.
	if rec := doRequest(h, "/api/auth/login", "198.51.100.9"); rec.Code != http.StatusOK {
		t.Errorf("valid XFF bucket: got %d, want 200", rec.Code)
	}
}

func TestRateLimitExactJSONBodyMatchesSpec(t *testing.T) {
	// Guard against accidental wording drift: the spec pins this string.
	if !strings.Contains(rateLimitErrBody, "Too many requests. Please try again later.") {
		t.Fatalf("body drifted from spec: %s", rateLimitErrBody)
	}
}
