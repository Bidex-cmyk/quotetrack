package models

import (
	"encoding/json"
	"math"
	"strconv"
)

// Cents is an integer money amount (minor currency units), used to avoid
// floating-point drift. JSON (un)marshalling uses whole currency units
// (e.g. dollars) so the API stays human friendly.

func (c *Cents) FromDollars(d float64) {
	*c = Cents(math.Round(d * 100))
}

func (c Cents) Dollars() float64 {
	return math.Round(float64(c)) / 100
}

// MarshalJSON emits money as currency units (e.g. 250.5).
func (c Cents) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Dollars())
}

// UnmarshalJSON reads currency units from a JSON number or string.
func (c *Cents) UnmarshalJSON(b []byte) error {
	var n float64
	if err := json.Unmarshal(b, &n); err == nil {
		c.FromDollars(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		c.FromDollars(f)
		return nil
	}
	return nil
}
