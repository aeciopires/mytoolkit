// Package hashgen implements the Hash Generator tool's pure logic.
package hashgen

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Algorithm string `json:"algorithm"`
}

func Generate(input []byte, opts Options) (string, error) {
	algo := opts.Algorithm
	if algo == "" {
		algo = "sha256"
	}
	h, err := newHash(algo)
	if err != nil {
		return "", err
	}
	h.Write(input)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func newHash(algo string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, apperr.New(400, "UNSUPPORTED_ALGORITHM", "algorithm must be one of: md5, sha1, sha256, sha512")
	}
}
