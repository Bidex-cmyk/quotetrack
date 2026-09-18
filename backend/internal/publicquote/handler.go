package publicquote

import (
	"log"
	"net/http"
	"strings"

	"quotetrack/backend/internal/httpapi"
	"quotetrack/backend/internal/models"
	"quotetrack/backend/internal/quotes"
)

// Handlers serves the unauthenticated public quote view.
type Handlers struct {
	store *quotes.Store
}

func NewHandlers(store *quotes.Store) *Handlers {
	return &Handlers{store: store}
}

// Get handles GET /api/public/quotes/{publicId}.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	publicID := strings.TrimSpace(r.PathValue("publicId"))
	if publicID == "" {
		httpapi.NotFound(w)
		return
	}
	q, businessName, err := h.store.GetByPublicID(r.Context(), publicID)
	if err != nil {
		if httpapi.IsRecordNotFound(err) {
			httpapi.NotFound(w)
			return
		}
		log.Printf("public quote: %v", err)
		httpapi.Error(w, http.StatusInternalServerError, "Could not load quote")
		return
	}

	resp := publicQuoteResponse{
		BusinessName: businessName,
		CustomerName: quoteCustomerName(q),
		QuoteNumber:  q.QuoteNumber,
		Title:        q.Title,
		Notes:        q.Notes,
		Status:       string(q.Status),
		QuoteDate:    q.QuoteDate,
		ExpiryDate:   q.ExpiryDate,
		Total:        q.Total,
		Items:        q.Items,
	}
	httpapi.JSON(w, http.StatusOK, resp)
}

type publicQuoteResponse struct {
	BusinessName string             `json:"business_name"`
	CustomerName string             `json:"customer_name"`
	QuoteNumber  string             `json:"quote_number"`
	Title        string             `json:"title"`
	Notes        string             `json:"notes"`
	Status       string             `json:"status"`
	QuoteDate    string             `json:"quote_date"`
	ExpiryDate   *string            `json:"expiry_date"`
	Total        models.Cents       `json:"total"`
	Items        []models.QuoteItem `json:"items"`
}

func quoteCustomerName(q *models.Quote) string {
	if q.Customer != nil {
		return q.Customer.Name
	}
	return ""
}