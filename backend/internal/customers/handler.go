package customers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"quotetrack/backend/internal/auth"
	"quotetrack/backend/internal/httpapi"
	"quotetrack/backend/internal/models"
)

func http_api_log_error(err error) {
	log.Printf("customers: %v", err)
}

// Handlers exposes HTTP handlers for customer CRUD.
type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

func withUserID(r *http.Request) string {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		return ""
	}
	return u.ID
}

// List handles GET /api/customers?q=.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	userID := withUserID(r)
	search := strings.TrimSpace(r.URL.Query().Get("q"))

	customers, err := h.store.List(r.Context(), userID, search)
	if err != nil {
		http_api_log_error(err)
		httpapi.Error(w, http.StatusInternalServerError, "Could not list customers")
		return
	}
	if customers == nil {
		customers = []models.CustomerWithStats{}
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{"customers": customers})
}

// Get handles GET /api/customers/{id}.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	userID := withUserID(r)
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}
	c, err := h.store.Get(r.Context(), userID, id)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		log.Printf("customers get: %v", err)
		httpapi.Error(w, http.StatusInternalServerError, "Could not get customer")
		return
	}
	httpapi.JSON(w, http.StatusOK, c)
}

// Create handles POST /api/customers.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	userID := withUserID(r)
	var c models.Customer
	if err := decodeJSON(r, &c); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		httpapi.Error(w, http.StatusBadRequest, "Name is required")
		return
	}
	c.Phone = strings.TrimSpace(c.Phone)
	c.Email = strings.TrimSpace(c.Email)
	c.Company = strings.TrimSpace(c.Company)
	c.Notes = strings.TrimSpace(c.Notes)

	created, err := h.store.Create(r.Context(), userID, &c)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not create customer")
		return
	}
	httpapi.JSON(w, http.StatusCreated, created)
}

// Update handles PUT /api/customers/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	userID := withUserID(r)
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var c models.Customer
	if err := decodeJSON(r, &c); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	c.ID = id
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		httpapi.Error(w, http.StatusBadRequest, "Name is required")
		return
	}
	c.Phone = strings.TrimSpace(c.Phone)
	c.Email = strings.TrimSpace(c.Email)
	c.Company = strings.TrimSpace(c.Company)
	c.Notes = strings.TrimSpace(c.Notes)

	updated, err := h.store.Update(r.Context(), userID, &c)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not update customer")
		return
	}
	httpapi.JSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/customers/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID := withUserID(r)
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}
	err := h.store.Delete(r.Context(), userID, id)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not delete customer")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
