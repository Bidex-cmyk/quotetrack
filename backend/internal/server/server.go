package server

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/auth"
	"quotetrack/backend/internal/config"
	"quotetrack/backend/internal/customers"
	"quotetrack/backend/internal/dashboard"
	"quotetrack/backend/internal/middleware"
	"quotetrack/backend/internal/publicquote"
	"quotetrack/backend/internal/quotes"
)

// New builds the HTTP handler with all routes wired.
func New(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	ttl, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		ttl = 7 * 24 * time.Hour
	}
	tokens := auth.NewTokenManager(cfg.JWTSecret, ttl)

	authStore := auth.NewStore(pool)
	authHandlers := auth.NewHandlers(authStore, tokens)

	customerStore := customers.NewStore(pool)
	customerHandlers := customers.NewHandlers(customerStore)

	quoteStore := quotes.NewStore(pool)
	quoteHandlers := quotes.NewHandlers(quoteStore)

	dashStore := dashboard.NewStore(pool, quoteStore)
	dashHandlers := dashboard.NewHandlers(dashStore)

	publicHandlers := publicquote.NewHandlers(quoteStore)

	// Public routes.
	mux.HandleFunc("POST /api/auth/signup", authHandlers.Signup)
	mux.HandleFunc("POST /api/auth/login", authHandlers.Login)
	mux.HandleFunc("GET /api/public/quotes/{publicId}", publicHandlers.Get)

	// Authenticated routes.
	api := http.NewServeMux()
	api.HandleFunc("GET /api/me", authHandlers.Me)
	api.HandleFunc("PATCH /api/me", authHandlers.UpdateMe)

	api.HandleFunc("GET /api/customers", customerHandlers.List)
	api.HandleFunc("POST /api/customers", customerHandlers.Create)
	api.HandleFunc("GET /api/customers/{id}", customerHandlers.Get)
	api.HandleFunc("PUT /api/customers/{id}", customerHandlers.Update)
	api.HandleFunc("DELETE /api/customers/{id}", customerHandlers.Delete)

	api.HandleFunc("GET /api/quotes", quoteHandlers.List)
	api.HandleFunc("POST /api/quotes", quoteHandlers.Create)
	api.HandleFunc("GET /api/quotes/{id}", quoteHandlers.Get)
	api.HandleFunc("PUT /api/quotes/{id}", quoteHandlers.Update)
	api.HandleFunc("DELETE /api/quotes/{id}", quoteHandlers.Delete)
	api.HandleFunc("PATCH /api/quotes/{id}/status", quoteHandlers.SetStatus)
	api.HandleFunc("PATCH /api/quotes/{id}/follow-up", quoteHandlers.SetFollowUp)

	api.HandleFunc("GET /api/dashboard", dashHandlers.Get)
	api.HandleFunc("GET /api/analytics", dashHandlers.Analytics)

	// Fallback for unknown /api routes.
	api.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	mux.Handle("/api/", tokens.Middleware(authStore)(api))

	return middleware.Recover(middleware.Logger(middleware.CORS(cfg.FrontendOrigin)(mux)))
}