package quotes

import (
	"testing"

	"quotetrack/backend/internal/models"
)

func TestTotals(t *testing.T) {
	items := []models.QuoteItem{
		{Description: "a", Quantity: 2, UnitPrice: 1000},   // 2 * 10.00 = 20.00
		{Description: "b", Quantity: 1.5, UnitPrice: 999},  // 1.5 * 9.99 = 14.985 -> 14.99
	}
	var subtotal, total models.Cents
	subtotal, total = Totals(items)
	if subtotal != 3499 {
		t.Errorf("subtotal = %d, want 3499", subtotal)
	}
	if total != subtotal {
		t.Errorf("total = %d, want subtotal %d", total, subtotal)
	}
	if items[0].Total != 2000 {
		t.Errorf("item[0].Total = %d, want 2000", items[0].Total)
	}
	if items[1].Total != 1499 {
		t.Errorf("item[1].Total = %d, want 1499", items[1].Total)
	}
}

func TestTotalOf(t *testing.T) {
	sum := TotalOf([]models.QuoteItem{
		{Total: 100},
		{Total: 200},
		{Total: 50},
	})
	if sum != 350 {
		t.Errorf("TotalOf = %d, want 350", sum)
	}
}

func TestQuoteNumber(t *testing.T) {
	if got := QuoteNumber(1); got != "QT-0001" {
		t.Errorf("got %s", got)
	}
	if got := QuoteNumber(123); got != "QT-0123" {
		t.Errorf("got %s", got)
	}
}