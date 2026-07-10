// Package version holds the application version, set at build time via
// -ldflags "-X .../internal/version.Version=..." from the repo-root VERSION
// file (read by both the Makefile and the Dockerfile), so the Go binary, the
// CLI --version output, and the Docker image tag all agree on one value.
package version

// Version is overridden at build time; "dev" is the fallback for
// `go run`/`go build` invocations that don't pass -ldflags.
var Version = "dev"
