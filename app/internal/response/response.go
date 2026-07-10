// Package response implements the shared JSON success/error envelope used
// by REST handlers and middleware. It is a leaf package (depends only on
// apperr) so both internal/httpapi and internal/httpapi/middleware can
// import it without an import cycle.
package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type successResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
	Meta    meta `json:"meta"`
}

type meta struct {
	Tool       string  `json:"tool"`
	DurationMs float64 `json:"duration_ms"`
}

type errorResponse struct {
	Success bool      `json:"success"`
	Error   errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteSuccess writes the shared success envelope.
func WriteSuccess(w http.ResponseWriter, tool string, data any, duration time.Duration) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(successResponse{
		Success: true,
		Data:    data,
		Meta:    meta{Tool: tool, DurationMs: float64(duration.Microseconds()) / 1000.0},
	})
}

// WriteError writes the shared error envelope, mapping *apperr.Error to its
// status/code and falling back to 500/INTERNAL for any other error.
func WriteError(w http.ResponseWriter, err error) {
	var ae *apperr.Error
	status := http.StatusInternalServerError
	code := "INTERNAL"
	msg := "internal server error"
	if errors.As(err, &ae) {
		status = ae.Status
		code = ae.Code
		msg = ae.Message
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Success: false,
		Error:   errorBody{Code: code, Message: msg},
	})
}

// StatusOf returns the HTTP status an error maps to, for logging/metrics.
func StatusOf(err error) int {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Status
	}
	return http.StatusInternalServerError
}
