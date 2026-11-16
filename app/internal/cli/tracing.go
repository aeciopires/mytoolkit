package cli

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/aeciopires/mytoolkit/internal/telemetry"
	"github.com/aeciopires/mytoolkit/internal/version"
)

// tracingShutdownTimeout bounds the final span flush on exit, so an
// unreachable collector can't hang shutdown.
const tracingShutdownTimeout = 5 * time.Second

// setupTracing installs OpenTelemetry tracing for a long-running command
// (see internal/telemetry for the OTEL_* environment variables). The
// returned func flushes pending spans and is always safe to defer.
func setupTracing(ctx context.Context, serviceName string) (func(), error) {
	shutdown, err := telemetry.Setup(ctx, serviceName, version.Version)
	if err != nil {
		return func() {}, err
	}
	if telemetry.Enabled() {
		log.Info().Msg("OpenTelemetry tracing enabled (OTLP/HTTP export)")
	}
	return func() {
		// The command's ctx is usually already canceled here (SIGTERM).
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), tracingShutdownTimeout)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Warn().Err(err).Msg("flushing traces on shutdown failed")
		}
	}, nil
}
