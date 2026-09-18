package quotes

import (
	"math"

	"quotetrack/backend/internal/models"
)

// Totals computes per-item and aggregate money amounts for a quote and
// mutates items with their line totals.
func Totals(items []models.QuoteItem) (subtotal, total models.Cents) {
	for i := range items {
		t := models.Cents(math.Round(items[i].Quantity * float64(items[i].UnitPrice)))
		items[i].Total = t
		subtotal += t
	}
	return subtotal, subtotal
}

// TotalOf returns the cumulative total for a slice of items without
// mutating them (used when items already carry totals from the DB).
func TotalOf(items []models.QuoteItem) models.Cents {
	var s models.Cents
	for _, it := range items {
		s += it.Total
	}
	return s
}