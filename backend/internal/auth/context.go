package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"quotetrack/backend/internal/models"
)

// contextKey is an unexported type for context values.
type contextKey string

const userContextKey contextKey = "user"

// Store is the subset of persistence the auth handlers need.
type Store interface {
	CreateUser(ctx context.Context, email, passwordHash, businessName string) (*models.User, error)
	UserByEmail(ctx context.Context, email string) (*models.User, string, error)
	UserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	CreateResetToken(ctx context.Context, userID, tokenHash string, expiresAt interface{}) error
	FindResetToken(ctx context.Context, tokenHash string) (*models.ResetToken, error)
	MarkResetTokenUsed(ctx context.Context, tokenID string) error
}

// UserFromContext returns the authenticated user stored by the middleware,
// or nil if none is present.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(userContextKey).(*models.User)
	return u
}

func withUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// bearerToken extracts the token from an Authorization header.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

var errNoToken = errors.New("missing or invalid authorization token")
