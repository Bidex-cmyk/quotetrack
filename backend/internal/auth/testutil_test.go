package auth

import (
	"net/http"
	"net/http/httptest"
)

func newGetRequest(authHeader string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}
	return r
}
