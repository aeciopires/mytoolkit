package cli

import (
	"errors"

	"github.com/rs/zerolog/log"
)

// loggedError marks an error the long-running commands (`serve`, `mcp`)
// already wrote to the JSON log, so main doesn't print it a second time as
// a plain-text "Error: ..." line — those commands' stderr must stay
// one-JSON-object-per-line for Loki. One-shot CLI tool commands keep the
// human-readable "Error: ..." output.
type loggedError struct{ error }

func (e loggedError) Unwrap() error { return e.error }

// logFatal logs err as a JSON error line and returns it marked as logged.
func logFatal(err error, msg string) error {
	log.Error().Err(err).Msg(msg)
	return loggedError{err}
}

// AlreadyLogged reports whether err was already written to the JSON log.
func AlreadyLogged(err error) bool {
	_, ok := errors.AsType[loggedError](err)
	return ok
}
