package cli

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/aeciopires/mytoolkit/internal/config"
	"github.com/aeciopires/mytoolkit/internal/health"
	"github.com/aeciopires/mytoolkit/internal/logging"
)

// shutdownTimeout bounds how long in-flight requests get to finish after
// SIGINT/SIGTERM — kept below Kubernetes' default 30s
// terminationGracePeriodSeconds so the pod exits cleanly before SIGKILL.
const shutdownTimeout = 20 * time.Second

// listenAndServe runs handler on addr with production-safe timeouts
// (a bare http.ListenAndServe has none, so a client trickling headers can
// hold a connection open forever — "Slowloris") and shuts down gracefully
// on SIGINT/SIGTERM or when ctx is canceled.
func listenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// No WriteTimeout: the MCP streamable-HTTP transport holds
		// long-lived SSE responses open by design.
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
		// net/http's own errors (TLS handshake, panics in handlers it
		// recovers, superfluous WriteHeader) as JSON, not stdlib text.
		ErrorLog: logging.StdLogger(log.Logger, "http.server"),
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	// Fail /readyz first, then (optionally) keep serving for
	// MYTOOLKIT_SHUTDOWN_DELAY so Kubernetes/load balancers notice and stop
	// sending new requests before srv.Shutdown closes the listener and
	// drains the in-flight ones.
	health.MarkShuttingDown()
	delay := config.ShutdownDelay()
	log.Info().Str("delay", delay.String()).Msg("shutting down")
	if delay > 0 {
		select {
		case <-time.After(delay):
		case err := <-errc:
			return err
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
