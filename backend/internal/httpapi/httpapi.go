package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes a JSON error envelope: {"error": "..."}.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// NotFound returns a 404 JSON response.
func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "Resource not found")
}

// IsRecordNotFound reports whether err is a pgx "no rows" error.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// NotFoundStatus maps an error to a 404 if it is a not-found error.
// The caller may pass any error; non-not-found errors return false and the
// caller decides how to handle them.
func NotFoundStatus(err error) (int, bool) {
	if IsRecordNotFound(err) {
		return http.StatusNotFound, true
	}
	return 0, false
}