package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"quotetrack/backend/internal/httpapi"
	"quotetrack/backend/internal/models"
)

// Handlers bundles auth HTTP handlers and their dependencies.
type Handlers struct {
	store  Store
	tokens *TokenManager
}

func NewHandlers(store Store, tokens *TokenManager) *Handlers {
	return &Handlers{store: store, tokens: tokens}
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
		httpapi.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if !CheckPassword(hash, req.Password) {
		httpapi.Error(w, http.StatusUnauthorized, "Invalid email or password")
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