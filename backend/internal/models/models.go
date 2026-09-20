package models

import "time"

// Cents is an integer money amount (minor currency units) used to avoid
// floating point drift. JSON serialisation converts to/from whole currency
// units (e.g. dollars) for a human-friendly API.
type Cents int64

// User is a business account holder.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	BusinessName string    `json:"business_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Customer is a person/business the user sells to.
type Customer struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomerWithStats is a customer plus lightweight aggregate metadata.
type CustomerWithStats struct {
	Customer
	QuoteCount int   `json:"quote_count"`
	TotalValue Cents `json:"total_value"`
}

// QuoteItem is one line on a quote.
type QuoteItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   Cents   `json:"unit_price"`
	Total       Cents   `json:"total"`
	Position    int     `json:"position"`
}

// Quote is a price quotation for a customer.
type Quote struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	CustomerID   string      `json:"customer_id"`
	PublicID     string      `json:"public_id"`
	QuoteNumber  string      `json:"quote_number"`
	Title        string      `json:"title"`
	Notes        string      `json:"notes"`
	Status       QuoteStatus `json:"status"`
	QuoteDate    string      `json:"quote_date"`
	ExpiryDate   *string     `json:"expiry_date"`
	FollowUpDate *string     `json:"follow_up_date"`
	Subtotal     Cents       `json:"subtotal"`
	Total        Cents       `json:"total"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Items        []QuoteItem `json:"items"`
	Customer     *Customer   `json:"customer,omitempty"`
}

// QuoteStatus is the lifecycle state of a quote.
type QuoteStatus string

const (
	StatusDraft   QuoteStatus = "draft"
	StatusSent    QuoteStatus = "sent"
	StatusWaiting QuoteStatus = "waiting"
	StatusWon     QuoteStatus = "won"
	StatusLost    QuoteStatus = "lost"
	StatusExpired QuoteStatus = "expired"
)

var validStatuses = map[QuoteStatus]bool{
	StatusDraft:   true,
	StatusSent:    true,
	StatusWaiting: true,
	StatusWon:     true,
	StatusLost:    true,
	StatusExpired: true,
}

// Valid reports whether s is an allowed quote status.
func (s QuoteStatus) Valid() bool {
	return validStatuses[s]
}

// ResetToken represents a password-reset token row.
type ResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"` // never expose the hash
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
