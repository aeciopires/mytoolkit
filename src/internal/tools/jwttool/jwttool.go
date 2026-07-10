// Package jwttool implements the JWT Encode/Decode tool's pure logic.
// Named jwttool (not jwt) to avoid colliding with the golang-jwt/jwt import.
package jwttool

import (
	"encoding/json"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type DecodeResult struct {
	Header    map[string]any `json:"header"`
	Claims    map[string]any `json:"claims"`
	Signature string         `json:"signature"`
	Valid     *bool          `json:"valid,omitempty"`
}

// Decode parses a JWT without requiring the signing secret (for inspection).
// If secret is non-empty, it additionally attempts verification and sets Valid.
func Decode(token string, secret string) (DecodeResult, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return DecodeResult{}, apperr.New(400, "INVALID_TOKEN", "token contains an invalid number of segments")
	}

	claims := jwt.MapClaims{}
	parsed, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return DecodeResult{}, apperr.Newf(400, "INVALID_TOKEN", "%s", err.Error())
	}

	result := DecodeResult{
		Header:    parsed.Header,
		Claims:    map[string]any(claims),
		Signature: parts[2],
	}

	if secret != "" {
		_, verifyErr := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		valid := verifyErr == nil
		result.Valid = &valid
	}

	return result, nil
}

// Encode signs claims into a JWT using an HMAC algorithm (HS256/HS384/HS512).
func Encode(claims map[string]any, secret string, algorithm string) (string, error) {
	if len(claims) == 0 {
		return "", apperr.New(400, "EMPTY_CLAIMS", "claims must not be empty")
	}
	method, err := hmacMethod(algorithm)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(method, jwt.MapClaims(claims))
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", apperr.Newf(500, "SIGNING_FAILED", "%s", err.Error())
	}
	return signed, nil
}

func hmacMethod(algorithm string) (*jwt.SigningMethodHMAC, error) {
	switch algorithm {
	case "", "HS256":
		return jwt.SigningMethodHS256, nil
	case "HS384":
		return jwt.SigningMethodHS384, nil
	case "HS512":
		return jwt.SigningMethodHS512, nil
	default:
		return nil, apperr.New(400, "UNSUPPORTED_ALGORITHM", "algorithm must be one of: HS256, HS384, HS512")
	}
}

// ParseClaimsJSON is a small helper for CLI/REST adapters that receive
// claims as raw JSON text rather than a decoded map.
func ParseClaimsJSON(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, apperr.New(400, "EMPTY_CLAIMS", "claims must not be empty")
	}
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, apperr.Newf(400, "INVALID_TOKEN", "invalid claims JSON: %s", err.Error())
	}
	return claims, nil
}
