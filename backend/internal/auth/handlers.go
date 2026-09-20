package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"quotetrack/backend/internal/httpapi"
	"quotetrack/backend/internal/models"
)

// EmailSender abstracts sending emails so the auth package doesn't depend on a
// specific delivery provider. Implementations live in the mailer package.
type EmailSender interface {
	SendForgotPassword(toEmail, resetURL string) error
}

// Handlers bundles auth HTTP handlers and their dependencies.
type Handlers struct {
	store       Store
	tokens      *TokenManager
	email       EmailSender
	frontendURL string
}

func NewHandlers(store Store, tokens *TokenManager, email EmailSender, frontendURL string) *Handlers {
	return &Handlers{store: store, tokens: tokens, email: email, frontendURL: frontendURL}
}

type credentialsReq struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	BusinessName string `json:"business_name"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Signup creates a new account and returns a token.
func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	var req credentialsReq
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.BusinessName = strings.TrimSpace(req.BusinessName)

	if !validEmail(req.Email) {
		httpapi.Error(w, http.StatusBadRequest, "A valid email is required")
		return
	}
	if len(req.Password) < 8 {
		httpapi.Error(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}
	if req.BusinessName == "" {
		httpapi.Error(w, http.StatusBadRequest, "Business name is required")
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not create account")
		return
	}

	user, err := h.store.CreateUser(r.Context(), req.Email, hash, req.BusinessName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			httpapi.Error(w, http.StatusConflict, "An account with this email already exists")
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not create account")
		return
	}

	token, err := h.tokens.Create(user.ID)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not create session")
		return
	}
	httpapi.JSON(w, http.StatusCreated, authResponse{Token: token, User: user})
}

// Login authenticates a user and returns a token.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req credentialsReq
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, hash, err := h.store.UserByEmail(r.Context(), req.Email)
	if err != nil {
		httpapi.Error(w, http.StatusUnauthorized, "Email or password is incorrect.")
		return
	}
	if !CheckPassword(hash, req.Password) {
		httpapi.Error(w, http.StatusUnauthorized, "Email or password is incorrect.")
		return
	}

	token, err := h.tokens.Create(user.ID)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not create session")
		return
	}
	httpapi.JSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

// Me returns the currently authenticated user.
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		httpapi.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

// UpdateMe updates the authenticated user's profile (settings page).
func (h *Handlers) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		httpapi.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		Email        *string `json:"email"`
		BusinessName *string `json:"business_name"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if !validEmail(email) {
			httpapi.Error(w, http.StatusBadRequest, "A valid email is required")
			return
		}
		user.Email = email
	}
	if req.BusinessName != nil {
		name := strings.TrimSpace(*req.BusinessName)
		if name == "" {
			httpapi.Error(w, http.StatusBadRequest, "Business name cannot be empty")
			return
		}
		user.BusinessName = name
	}

	updated, err := h.store.UpdateUser(r.Context(), user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			httpapi.Error(w, http.StatusConflict, "Another account already uses this email")
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not update profile")
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{"user": updated})
}

// ---------- Password Reset ----------

// resetReq is the request body for POST /api/auth/forgot-password.
type forgotReq struct {
	Email string `json:"email"`
}

// resetReq is the request body for POST /api/auth/reset-password.
type resetReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ForgotPassword issues a password-reset token for the given email.
// It always returns a generic message to avoid leaking account existence.
func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotReq
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	generic := "If an account exists for that email, you'll receive password reset instructions."

	if !validEmail(req.Email) {
		// Return the same generic message so we don't reveal anything.
		httpapi.JSON(w, http.StatusOK, map[string]string{"message": generic})
		return
	}

	user, _, err := h.store.UserByEmail(r.Context(), req.Email)
	if err != nil {
		// User not found — still return generic message.
		httpapi.JSON(w, http.StatusOK, map[string]string{"message": generic})
		return
	}

	// Generate a cryptographically secure random token.
	rawToken, err := generateResetToken()
	if err != nil {
		httpapi.JSON(w, http.StatusOK, map[string]string{"message": generic})
		return
	}

	// Store only the SHA-256 hash of the token.
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := h.store.CreateResetToken(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
		httpapi.JSON(w, http.StatusOK, map[string]string{"message": generic})
		return
	}

	// Build the reset URL.
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.frontendURL, rawToken)

	// Send the email if a sender is configured.
	if h.email != nil {
		_ = h.email.SendForgotPassword(user.Email, resetURL)
	}

	// Always return the same generic message.
	httpapi.JSON(w, http.StatusOK, map[string]string{"message": generic})
}

// ResetPassword validates a reset token and updates the user's password.
func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetReq
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	req.Password = strings.TrimSpace(req.Password)

	if req.Token == "" {
		httpapi.Error(w, http.StatusBadRequest, "Reset token is required")
		return
	}
	if len(req.Password) < 8 {
		httpapi.Error(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	tokenHash := hashToken(req.Token)

	rt, err := h.store.FindResetToken(r.Context(), tokenHash)
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "This password reset link is invalid or has expired.")
		return
	}

	// Check expiry.
	if time.Now().After(rt.ExpiresAt) {
		httpapi.Error(w, http.StatusBadRequest, "This password reset link is invalid or has expired.")
		return
	}

	// Check if already used.
	if rt.UsedAt != nil {
		httpapi.Error(w, http.StatusBadRequest, "This password reset link is invalid or has expired.")
		return
	}

	// Hash the new password.
	hash, err := HashPassword(req.Password)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not reset password")
		return
	}

	// Update password.
	if err := h.store.UpdatePassword(r.Context(), rt.UserID, hash); err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not reset password")
		return
	}

	// Mark the token as used.
	_ = h.store.MarkResetTokenUsed(r.Context(), rt.ID)

	httpapi.JSON(w, http.StatusOK, map[string]string{
		"message": "Your password has been reset successfully.",
	})
}

// ---------- helpers ----------

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func validEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	return dot > 1
}

// generateResetToken returns a 32-byte hex-encoded random token.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hex digest of a token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
