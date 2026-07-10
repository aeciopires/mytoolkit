// Package handlers adapts pure internal/tools functions into REST handlers.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
)

type request struct {
	Input   string          `json:"input"`
	Options json.RawMessage `json:"options"`
}

// Wrap adapts a func(input []byte, opts Opts) (string, error) tool function
// into a REST handler using the shared request/response envelope.
func Wrap[Opts any](slug string, fn func(input []byte, opts Opts) (string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var req request
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
				return
			}
		}

		var opts Opts
		if len(req.Options) > 0 {
			if err := json.Unmarshal(req.Options, &opts); err != nil {
				response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_OPTIONS", "invalid options object"))
				return
			}
		}

		out, err := fn([]byte(req.Input), opts)
		if err != nil {
			response.WriteError(w, err)
			return
		}

		response.WriteSuccess(w, slug, map[string]string{"output": out}, time.Since(start))
	}
}
