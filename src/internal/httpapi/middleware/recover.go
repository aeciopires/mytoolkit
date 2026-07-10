package middleware

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
)

// Recover converts a panic in a downstream handler into a structured 500
// JSON error response instead of crashing the server.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().Interface("panic", rec).Str("path", r.URL.Path).Msg("panic recovered")
				response.WriteError(w, apperr.New(http.StatusInternalServerError, "INTERNAL", "internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
