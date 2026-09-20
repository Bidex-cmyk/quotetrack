package models

import (
	"encoding/json"
	"math"
	"testing"
)

func TestCentsFromDollars(t *testing.T) {
	cases := []struct {
		dollars  float64
		expected Cents
	}{
		{0, 0},
		{1, 100},
		{1.5, 150},
		{0.01, 1},
		{19.99, 1999},
		{123.456, 12346}, // rounds
	}
	for _, tc := range cases {
		var c Cents
		c.FromDollars(tc.dollars)
		if c != tc.expected {
			t.Errorf("FromDollars(%v) = %d, want %d", tc.dollars, c, tc.expected)
		}
	}
}

func TestCentsMarshalJSON(t *testing.T) {
	var c Cents = 12345
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	want := "123.45"
	if string(b) != want {
		t.Errorf("marshal = %s, want %s", b, want)
	}
}

func TestCentsUnmarshalJSON(t *testing.T) {
	var c Cents
	if err := json.Unmarshal([]byte("19.99"), &c); err != nil {
		t.Fatal(err)
	}
	if c != 1999 {
		t.Errorf("unmarshal = %d, want 1999", c)
	}
}

func TestDollarsRoundTrip(t *testing.T) {
	values := []float64{0.01, 1.99, 123.45, 999999.99, 0.10}
	for _, v := range values {
		var c Cents
		c.FromDollars(v)
		got := c.Dollars()
		if math.Abs(got-v) > 0.0001 {
			t.Errorf("round trip %v -> %v", v, got)
		}
	}
}

func TestQuoteStatusValid(t *testing.T) {
	valid := []QuoteStatus{StatusDraft, StatusSent, StatusWaiting, StatusWon, StatusLost, StatusExpired}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if (QuoteStatus("bogus")).Valid() {
		t.Error("bogus status should be invalid")
	}
	if (QuoteStatus("")).Valid() {
		t.Error("empty status should be invalid")
	}
}
