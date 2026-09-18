package quotes

import "fmt"

// QuoteNumber formats a sequential counter into a professional quote number.
func QuoteNumber(seq int) string {
	return fmt.Sprintf("QT-%04d", seq)
}