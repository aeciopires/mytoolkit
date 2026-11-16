package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestLogFatalLogsJSONAndMarksError(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = prev })

	base := errors.New("listen tcp :8080: bind: address already in use")
	err := logFatal(base, "web server failed")

	if !AlreadyLogged(err) || AlreadyLogged(base) {
		t.Error("AlreadyLogged mismatch")
	}
	if !errors.Is(err, base) {
		t.Error("logFatal must keep the original error in the chain")
	}
	want := `{"level":"error","error":"listen tcp :8080: bind: address already in use","message":"web server failed"}`
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Errorf("log line = %s\nwant      %s", got, want)
	}
}
