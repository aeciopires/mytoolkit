// Package qrcode implements the QR Code Generator tool's pure logic.
package qrcode

import (
	qr "github.com/skip2/go-qrcode"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

const maxTextBytes = 2900

type Options struct {
	Size int `json:"size"`
}

// Generate returns PNG-encoded bytes for a QR code of text.
func Generate(text string, opts Options) ([]byte, error) {
	if len(text) == 0 {
		return nil, apperr.New(400, "EMPTY_INPUT", "text must not be empty")
	}
	if len(text) > maxTextBytes {
		return nil, apperr.Newf(400, "INPUT_TOO_LARGE", "text must not exceed %d bytes", maxTextBytes)
	}

	size := opts.Size
	if size <= 0 {
		size = 256
	}
	if size > 2048 {
		size = 2048
	}

	png, err := qr.Encode(text, qr.Medium, size)
	if err != nil {
		return nil, apperr.Newf(400, "QR_ENCODE_FAILED", "%s", err.Error())
	}
	return png, nil
}
