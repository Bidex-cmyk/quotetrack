package auth

import (
	"net/http"

	"quotetrack/backend/internal/httpapi"
)

// Middleware authenticates requests via a Bearer JWT and injects the
// authenticated user into the request context.
func (tm *TokenManager) Middleware(st Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				httpapi.Error(w, http.StatusUnauthorized, "Authentication required")
				return
			}
			userID, err := tm.Verify(token)
			if err != nil {
				httpapi.Error(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			user, err := st.UserByID(r.Context(), userID)
			if err != nil {
				httpapi.Error(w, http.StatusUnauthorized, "Account not found")
				return
			}
			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
		})
	}
}
