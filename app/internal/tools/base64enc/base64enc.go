// Package base64enc implements the Base64 Encode/Decode tool's pure logic.
package base64enc

import (
	"encoding/base64"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Decode  bool   `json:"decode"`
	Variant string `json:"variant"` // "standard" | "url", default "standard"
	Padding *bool  `json:"padding"` // default true
}

func Process(input []byte, opts Options) (string, error) {
	if err := apperr.OneOf("variant", variantOrDefault(opts.Variant), "standard", "url"); err != nil {
		return "", err
	}
	padding := true
	if opts.Padding != nil {
		padding = *opts.Padding
	}
	enc := encoding(variantOrDefault(opts.Variant), padding)

	if opts.Decode {
		if len(input) == 0 {
			return "", nil
		}
		out, err := enc.DecodeString(string(input))
		if err != nil {
			return "", apperr.Newf(400, "INVALID_BASE64", "%s", err.Error())
		}
		return string(out), nil
	}

	return enc.EncodeToString(input), nil
}

func variantOrDefault(v string) string {
	if v == "" {
		return "standard"
	}
	return v
}

func encoding(variant string, padding bool) *base64.Encoding {
	if variant == "url" {
		if padding {
			return base64.URLEncoding
		}
		return base64.RawURLEncoding
	}
	if padding {
		return base64.StdEncoding
	}
	return base64.RawStdEncoding
}
