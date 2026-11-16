package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	stdlog "log"
	"log/slog"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

func restoreGlobals(t *testing.T) {
	t.Helper()
	prevLog, prevSlog := log.Logger, slog.Default()
	prevFmt, prevLevel := zerolog.TimeFieldFormat, zerolog.GlobalLevel()
	prevOut, prevFlags := stdlog.Writer(), stdlog.Flags()
	prevOtel := otel.GetErrorHandler()
	t.Cleanup(func() {
		log.Logger = prevLog
		slog.SetDefault(prevSlog)
		zerolog.TimeFieldFormat = prevFmt
		zerolog.SetGlobalLevel(prevLevel)
		stdlog.SetOutput(prevOut)
		stdlog.SetFlags(prevFlags)
		otel.SetErrorHandler(prevOtel)
	})
}

// lines decodes every output line as JSON — the whole point of the
// package is that nothing it captures is plain text.
func lines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for line := range strings.Lines(buf.String()) {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("non-JSON log line %q: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

func TestSetupEmitsUniformJSONFromEverySource(t *testing.T) {
	restoreGlobals(t)
	var buf bytes.Buffer
	Setup(&buf, "debug", "mytoolkit", "9.9.9")

	log.Info().Str("tool", "base64").Msg("zerolog line")
	stdlog.Println("http: superfluous response.WriteHeader call")
	slog.Warn("slog line", "k", "v")
	otel.Handle(errors.New("export failed"))
	StdLogger(log.Logger, "http").Print("tls handshake error")

	got := lines(t, &buf)
	if len(got) != 5 {
		t.Fatalf("got %d lines, want 5:\n%s", len(got), buf.String())
	}
	wantLevels := []string{"info", "error", "warn", "warn", "error"}
	for i, m := range got {
		if m["service"] != "mytoolkit" || m["version"] != "9.9.9" {
			t.Errorf("line %d lacks service/version: %v", i, m)
		}
		if m["level"] != wantLevels[i] {
			t.Errorf("line %d level = %v, want %s", i, m["level"], wantLevels[i])
		}
		if _, ok := m["time"].(string); !ok {
			t.Errorf("line %d lacks time: %v", i, m)
		}
		if msg, _ := m["message"].(string); msg == "" || strings.HasSuffix(msg, "\n") {
			t.Errorf("line %d message = %q", i, msg)
		}
	}
	if got[1]["logger"] != "stdlib" || got[2]["logger"] != "slog" || got[3]["logger"] != "otel" || got[4]["logger"] != "http" {
		t.Errorf("logger fields = %v %v %v %v", got[1]["logger"], got[2]["logger"], got[3]["logger"], got[4]["logger"])
	}
	if got[2]["k"] != "v" {
		t.Errorf("slog attribute lost: %v", got[2])
	}
}

func TestSetupHonorsLevel(t *testing.T) {
	restoreGlobals(t)
	var buf bytes.Buffer
	Setup(&buf, "warn", "svc", "v")
	log.Info().Msg("hidden")
	log.Warn().Msg("shown")
	if got := lines(t, &buf); len(got) != 1 || got[0]["message"] != "shown" {
		t.Errorf("got %v", got)
	}
}

func TestParseLevel(t *testing.T) {
	for in, want := range map[string]zerolog.Level{
		"debug": zerolog.DebugLevel, "warn": zerolog.WarnLevel, "": zerolog.InfoLevel, "bogus": zerolog.InfoLevel,
	} {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}
