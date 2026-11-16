// Package logging configures the process-wide structured JSON logger used
// by the long-running commands (`serve`, `mcp`).
//
// Every log line — MyToolkit's own (zerolog), the standard library's `log`
// package (net/http server errors), `log/slog` (third-party libraries) and
// OpenTelemetry SDK errors — is written to stderr as one JSON object per
// line with a common shape, so Grafana Alloy/Loki can parse all of it
// with a single `json` stage:
//
//	{"level":"info","service":"mytoolkit","version":"1.3.0",
//	 "time":"2026-11-16T09:39:29.123456789Z","message":"...", ...fields}
//
// Logs never go to stdout: CLI tool output and the MCP stdio JSON-RPC
// stream own stdout.
package logging

import (
	"io"
	stdlog "log"
	"log/slog"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

// Setup installs the global JSON logger writing to w at the given level
// (debug, info, warn, error; anything else means info) and returns it.
func Setup(w io.Writer, level, service, version string) zerolog.Logger {
	// Nanosecond timestamps keep Loki's per-stream ordering exact when
	// several lines are written within the same second.
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(ParseLevel(level))

	logger := zerolog.New(w).With().
		Timestamp().
		Str("service", service).
		Str("version", version).
		Logger()
	log.Logger = logger

	// log/slog users (libraries) -> the same JSON shape.
	slog.SetDefault(slog.New(zerolog.NewSlogHandler(logger.With().Str("logger", "slog").Logger())))

	// Standard library `log` (e.g. http.Server's "http: TLS handshake
	// error"/"superfluous WriteHeader") -> JSON at error level instead of
	// plain text. Must come after slog.SetDefault, which otherwise
	// redirects the stdlib logger into the slog handler at info level.
	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
	stdlog.SetOutput(levelWriter{logger.With().Str("logger", "stdlib").Logger(), zerolog.ErrorLevel})

	// OpenTelemetry SDK errors (e.g. a failed span export) would otherwise
	// go through the stdlib logger as text.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Warn().Err(err).Str("logger", "otel").Msg("opentelemetry error")
	}))
	return logger
}

// ParseLevel maps a level name to a zerolog level, defaulting to info.
func ParseLevel(level string) zerolog.Level {
	l, err := zerolog.ParseLevel(level)
	if err != nil || l == zerolog.NoLevel {
		return zerolog.InfoLevel
	}
	return l
}

// StdLogger returns a standard library *log.Logger that writes through
// logger at error level, for APIs that take one (http.Server.ErrorLog).
func StdLogger(logger zerolog.Logger, component string) *stdlog.Logger {
	return stdlog.New(levelWriter{logger.With().Str("logger", component).Logger(), zerolog.ErrorLevel}, "", 0)
}

// levelWriter turns each Write (one stdlib log line) into a JSON event at
// a fixed level. zerolog.Logger's own Write omits the "level" field,
// which would leave those lines unlabeled in Loki.
type levelWriter struct {
	logger zerolog.Logger
	level  zerolog.Level
}

func (w levelWriter) Write(p []byte) (int, error) {
	w.logger.WithLevel(w.level).Msg(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}
