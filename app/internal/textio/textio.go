// Package textio provides shared file-or-stdio read/write helpers used by
// every CLI subcommand's --in/--out flags.
package textio

import (
	"io"
	"os"
)

// Read returns the contents of path, or reads from stdin if path is "" or "-".
func Read(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

// Write writes data to path, or to stdout if path is "" or "-".
func Write(path string, data []byte) error {
	if path == "" || path == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WritePrivate is Write for secret material (private keys, HMAC secrets):
// a file is created with mode 0600 so it isn't world-readable.
func WritePrivate(path string, data []byte) error {
	if path == "" || path == "-" {
		return Write(path, data)
	}
	return os.WriteFile(path, data, 0o600)
}
