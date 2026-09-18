package dashboard

import (
	"net/http"
	"time"

	"quotetrack/backend/internal/auth"
	"quotetrack/backend/internal/httpapi"
	"quotetrack/backend/internal/quotes"
)

// Handlers exposes the dashboard HTTP endpoint.
type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// Get handles GET /api/dashboard.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		httpapi.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx := r.Context()
	stats, err := h.store.GetStats(ctx, user.ID)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}

	followUps, err := h.store.FollowUpsDue(ctx, user.ID, time.Now())
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not load follow-ups")
		return
	}
	if followUps == nil {
		followUps = []quotes.QuoteRow{}
	}

	httpapi.JSON(w, http.StatusOK, map[string]interface{}{
		"stats":      stats,
		"follow_ups": followUps,
	})
}

// Analytics handles GET /api/analytics.
func (h *Handlers) Analytics(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		httpapi.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	a, err := h.store.GetAnalytics(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not load analytics")
		return
	}
	httpapi.JSON(w, http.StatusOK, a)
}