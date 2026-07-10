// Package jwttool implements the JWT Encode/Decode tool's pure logic.
// Named jwttool (not jwt) to avoid colliding with the golang-jwt/jwt import.
package jwttool

import (
	"encoding/json"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

// SupportedAlgorithms lists every signing algorithm this tool accepts, in
// the order shown in the web UI's combobox. DefaultAlgorithm is what the
// tool has always defaulted to (HS256) — kept as the default when adding
// the rest of this list, per the feature request that the default not
// change.
var SupportedAlgorithms = []string{
	"HS256", "HS384", "HS512",
	"RS256", "RS384", "RS512",
	"PS256", "PS384", "PS512",
	"ES256", "ES384", "ES512",
	"EdDSA",
}

const DefaultAlgorithm = "HS256"

type DecodeResult struct {
	Header    map[string]any `json:"header"`
	Claims    map[string]any `json:"claims"`
	Signature string         `json:"signature"`
	Valid     *bool          `json:"valid,omitempty"`
}

// Decode parses a JWT without requiring a key (for inspection). If secret
// or key is non-empty, it additionally attempts verification and sets
// Valid: secret is used for HMAC-family tokens (HS256/384/512), key (a
// PEM-encoded public key) for every other supported algorithm. Which one
// applies is determined by the token's own "alg" header, not by the
// caller — pasting a value into the wrong field simply fails verification
// (Valid=false), it does not silently succeed.
func Decode(token string, secret string, key string) (DecodeResult, error) {
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

	if secret != "" || key != "" {
		_, verifyErr := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			return verificationKey(t, secret, key)
		})
		valid := verifyErr == nil
		result.Valid = &valid
	}

	return result, nil
}

func verificationKey(t *jwt.Token, secret, key string) (any, error) {
	switch t.Method.(type) {
	case *jwt.SigningMethodHMAC:
		return []byte(secret), nil
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		return jwt.ParseRSAPublicKeyFromPEM([]byte(key))
	case *jwt.SigningMethodECDSA:
		return jwt.ParseECPublicKeyFromPEM([]byte(key))
	case *jwt.SigningMethodEd25519:
		return jwt.ParseEdPublicKeyFromPEM([]byte(key))
	default:
		return nil, apperr.New(400, "UNSUPPORTED_ALGORITHM", "unsupported signing method: "+t.Method.Alg())
	}
}

// Encode signs claims into a JWT. HMAC algorithms (HS256/HS384/HS512) sign
// with secret as a raw shared-secret string; every other supported
// algorithm signs with key, a PEM-encoded private key (RSA for RS*/PS*,
// EC for ES*, Ed25519 for EdDSA).
func Encode(claims map[string]any, secret string, key string, algorithm string) (string, error) {
	if len(claims) == 0 {
		return "", apperr.New(400, "EMPTY_CLAIMS", "claims must not be empty")
	}
	method, signingKey, err := signingMethodAndKey(algorithm, secret, key)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(method, jwt.MapClaims(claims))
	signed, err := token.SignedString(signingKey)
	if err != nil {
		return "", apperr.Newf(500, "SIGNING_FAILED", "%s", err.Error())
	}
	return signed, nil
}

func signingMethodAndKey(algorithm, secret, key string) (jwt.SigningMethod, any, error) {
	switch algorithm {
	case "", DefaultAlgorithm:
		return jwt.SigningMethodHS256, []byte(secret), nil
	case "HS384":
		return jwt.SigningMethodHS384, []byte(secret), nil
	case "HS512":
		return jwt.SigningMethodHS512, []byte(secret), nil
	case "RS256":
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid RSA private key: %s", err.Error())
		}
		return jwt.SigningMethodRS256, k, nil
	case "RS384":
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid RSA private key: %s", err.Error())
		}
		return jwt.SigningMethodRS384, k, nil
	case "RS512":
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid RSA private key: %s", err.Error())
		}
		return jwt.SigningMethodRS512, k, nil
	case "PS256":
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid RSA private key: %s", err.Error())
		}
		return jwt.SigningMethodPS256, k, nil
	case "PS384":
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid RSA private key: %s", err.Error())
		}
		return jwt.SigningMethodPS384, k, nil
	case "PS512":
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid RSA private key: %s", err.Error())
		}
		return jwt.SigningMethodPS512, k, nil
	case "ES256":
		k, err := jwt.ParseECPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid ECDSA private key: %s", err.Error())
		}
		return jwt.SigningMethodES256, k, nil
	case "ES384":
		k, err := jwt.ParseECPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid ECDSA private key: %s", err.Error())
		}
		return jwt.SigningMethodES384, k, nil
	case "ES512":
		k, err := jwt.ParseECPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid ECDSA private key: %s", err.Error())
		}
		return jwt.SigningMethodES512, k, nil
	case "EdDSA":
		k, err := jwt.ParseEdPrivateKeyFromPEM([]byte(key))
		if err != nil {
			return nil, nil, apperr.Newf(400, "INVALID_KEY", "invalid Ed25519 private key: %s", err.Error())
		}
		return jwt.SigningMethodEdDSA, k, nil
	default:
		return nil, nil, apperr.New(400, "UNSUPPORTED_ALGORITHM", "algorithm must be one of: "+strings.Join(SupportedAlgorithms, ", "))
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
