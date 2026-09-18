package quotes

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"quotetrack/backend/internal/auth"
	"quotetrack/backend/internal/httpapi"
	"quotetrack/backend/internal/models"
)

// Handlers exposes HTTP handlers for quote management.
type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

func userID(r *http.Request) string {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		return ""
	}
	return u.ID
}

// List handles GET /api/quotes with optional filters.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lq := ListQuery{
		Status:     strings.TrimSpace(q.Get("status")),
		CustomerID: strings.TrimSpace(q.Get("customer_id")),
		Search:     strings.TrimSpace(q.Get("q")),
	}
	if id := lq.CustomerID; id != "" {
		if _, err := uuid.Parse(id); err != nil {
			httpapi.Error(w, http.StatusBadRequest, "Invalid customer_id")
			return
		}
	}

	list, err := h.store.List(r.Context(), userID(r), lq)
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not list quotes")
		return
	}
	if list == nil {
		list = []QuoteRow{}
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{"quotes": list})
}

// Get handles GET /api/quotes/{id}.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}
	q, err := h.store.Get(r.Context(), userID(r), id)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not get quote")
		return
	}
	httpapi.JSON(w, http.StatusOK, q)
}

// quotePayload is the JSON shape accepted by Create and Update.
type quotePayload struct {
	CustomerID   string  `json:"customer_id"`
	Title        string  `json:"title"`
	Notes        string  `json:"notes"`
	Status       string  `json:"status"`
	QuoteDate    string  `json:"quote_date"`
	ExpiryDate   *string `json:"expiry_date"`
	FollowUpDate *string `json:"follow_up_date"`
	Items        []struct {
		Description string  `json:"description"`
		Quantity    float64 `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
	} `json:"items"`
}

// Create handles POST /api/quotes.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var p quotePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	in, errMsg := buildInput(userID(r), &p)
	if errMsg != "" {
		httpapi.Error(w, http.StatusBadRequest, errMsg)
		return
	}
	publicID, err := auth.RandomPublicID()
	if err != nil {
		httpapi.Error(w, http.StatusInternalServerError, "Could not create quote")
		return
	}
	in.PublicID = publicID

	q, err := h.store.Create(r.Context(), in)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.Error(w, http.StatusBadRequest, "Customer not found")
			return
		}
		log.Printf("quotes create: %v", err)
		httpapi.Error(w, http.StatusInternalServerError, "Could not create quote")
		return
	}
	httpapi.JSON(w, http.StatusCreated, q)
}

// Update handles PUT /api/quotes/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}
	var p quotePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	in, errMsg := buildInput(userID(r), &p)
	if errMsg != "" {
		httpapi.Error(w, http.StatusBadRequest, errMsg)
		return
	}

	q, err := h.store.Update(r.Context(), userID(r), id, in)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not update quote")
		return
	}
	httpapi.JSON(w, http.StatusOK, q)
}

// SetStatus handles PATCH /api/quotes/{id}/status.
func (h *Handlers) SetStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	status := models.QuoteStatus(strings.TrimSpace(body.Status))
	if !status.Valid() {
		httpapi.Error(w, http.StatusBadRequest, "Invalid status")
		return
	}
	q, err := h.store.SetStatus(r.Context(), userID(r), id, status)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		log.Printf("quotes set-status: %v", err)
		httpapi.Error(w, http.StatusInternalServerError, "Could not update status")
		return
	}
	httpapi.JSON(w, http.StatusOK, q)
}

// SetFollowUp handles PATCH /api/quotes/{id}/follow-up.
func (h *Handlers) SetFollowUp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}
	var body struct {
		FollowUpDate *string `json:"follow_up_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var date *string
	if body.FollowUpDate != nil && strings.TrimSpace(*body.FollowUpDate) != "" {
		v := strings.TrimSpace(*body.FollowUpDate)
		if !validDate(v) {
			httpapi.Error(w, http.StatusBadRequest, "Follow-up date must be YYYY-MM-DD")
			return
		}
		date = &v
	}
	q, err := h.store.SetFollowUp(r.Context(), userID(r), id, date)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not update follow-up date")
		return
	}
	httpapi.JSON(w, http.StatusOK, q)
}

// Delete handles DELETE /api/quotes/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}
	err := h.store.Delete(r.Context(), userID(r), id)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		httpapi.Error(w, http.StatusInternalServerError, "Could not delete quote")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func buildInput(uid string, p *quotePayload) (*CreateQuoteInput, string) {
	if _, err := uuid.Parse(p.CustomerID); err != nil {
		return nil, "A valid customer_id is required"
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, "Title is required"
	}
	if len(p.Items) == 0 {
		return nil, "Quote must contain at least one line item"
	}
	status := models.QuoteStatus(strings.TrimSpace(p.Status))
	if status == "" {
		status = models.StatusDraft
	}
	if !status.Valid() {
		return nil, "Invalid status"
	}
	quoteDate := strings.TrimSpace(p.QuoteDate)
	if quoteDate == "" {
		quoteDate = time.Now().Format("2006-01-02")
	}
	if !validDate(quoteDate) {
		return nil, "Quote date must be YYYY-MM-DD"
	}
	for _, d := range []*string{p.ExpiryDate, p.FollowUpDate} {
		if d != nil && strings.TrimSpace(*d) != "" && !validDate(strings.TrimSpace(*d)) {
			return nil, "Dates must be YYYY-MM-DD"
		}
	}
	expiry := strPtr(strings.TrimSpace(deref(p.ExpiryDate)))
	followUp := strPtr(strings.TrimSpace(deref(p.FollowUpDate)))

	items := make([]models.QuoteItem, 0, len(p.Items))
	for _, it := range p.Items {
		desc := strings.TrimSpace(it.Description)
		if desc == "" {
			return nil, "Line item descriptions are required"
		}
		if it.Quantity <= 0 {
			return nil, "Line item quantities must be greater than zero"
		}
		if it.UnitPrice < 0 {
			return nil, "Line item unit prices cannot be negative"
		}
		var up models.Cents
		up.FromDollars(it.UnitPrice)
		items = append(items, models.QuoteItem{
			Description: desc,
			Quantity:    it.Quantity,
			UnitPrice:   up,
		})
	}

	return &CreateQuoteInput{
		UserID:       uid,
		CustomerID:   p.CustomerID,
		Title:        title,
		Notes:        strings.TrimSpace(p.Notes),
		Status:       status,
		QuoteDate:    quoteDate,
		ExpiryDate:   expiry,
		FollowUpDate: followUp,
		Items:        items,
	}, ""
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}