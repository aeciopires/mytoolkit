// Package jwttool implements the JWT tool's pure logic: decode/verify,
// sign (encode) with registered and custom claims, offline key generation,
// and JWK/JWKS conversion, all on top of github.com/golang-jwt/jwt/v5.
// Named jwttool (not jwt) to avoid colliding with the golang-jwt/jwt import.
package jwttool

import (
	"cmp"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

// SupportedAlgorithms lists every signing algorithm this tool accepts, in
// the order shown in the web UI's combobox. "none" is deliberately absent:
// an unsigned token is never produced, and never reported as valid.
var SupportedAlgorithms = []string{
	"HS256", "HS384", "HS512",
	"RS256", "RS384", "RS512",
	"PS256", "PS384", "PS512",
	"ES256", "ES384", "ES512",
	"EdDSA",
}

// DefaultAlgorithm is used when no algorithm is given. It must stay HS256
// (see .skills/jwt/SKILL.md).
const DefaultAlgorithm = "HS256"

// DecodeResult is Decode's output. Valid is nil when no secret/key was
// supplied (inspection only); otherwise it reports whether the signature
// and the time-based claims (exp/nbf/iat) verified, and VerifyError says
// why when they did not.
type DecodeResult struct {
	Header      map[string]any `json:"header"`
	Claims      map[string]any `json:"claims"`
	Signature   string         `json:"signature"`
	Valid       *bool          `json:"valid,omitempty"`
	VerifyError string         `json:"verify_error,omitempty"`
}

// Decode parses a JWT without requiring a key (for inspection). If secret
// or key is non-empty, it additionally verifies the token and sets Valid:
// secret is used for HMAC-family tokens (HS256/384/512); key is used for
// every other algorithm and may be a PEM public key, private key or
// certificate (several concatenated PEM blocks are tried in turn), a JWK,
// or a JWKS — in which case the key whose "kid" matches the token header
// is selected. Which of secret/key applies is determined by the token's
// own "alg" header, not by the caller — pasting a value into the wrong
// field simply fails verification (Valid=false), it never silently
// succeeds.
func Decode(token, secret, key string) (DecodeResult, error) {
	token = strings.TrimSpace(token)
	if strings.Count(token, ".") != 2 {
		return DecodeResult{}, apperr.New(400, "INVALID_TOKEN", "token contains an invalid number of segments")
	}

	claims := jwt.MapClaims{}
	parsed, parts, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return DecodeResult{}, apperr.Newf(400, "INVALID_TOKEN", "%s", err)
	}

	result := DecodeResult{
		Header:    parsed.Header,
		Claims:    claims,
		Signature: parts[2],
	}
	if secret == "" && key == "" {
		return result, nil
	}

	_, verifyErr := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return verificationKey(t, secret, key)
	}, jwt.WithValidMethods(SupportedAlgorithms), jwt.WithIssuedAt())
	valid := verifyErr == nil
	result.Valid = &valid
	if verifyErr != nil {
		result.VerifyError = verifyErr.Error()
	}
	return result, nil
}

// verificationKey is Decode's jwt.Keyfunc. It returns an error (→
// Valid=false) instead of an empty key when the material for the token's
// algorithm family is missing: an empty HMAC key would otherwise verify a
// token forged with an empty secret.
func verificationKey(t *jwt.Token, secret, key string) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); ok {
		if secret == "" {
			return nil, fmt.Errorf("a secret is required to verify a %s token", t.Method.Alg())
		}
		return []byte(secret), nil
	}
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("a public key, JWK or JWKS is required to verify a %s token", t.Method.Alg())
	}

	keys, err := parseKeys(key)
	if err != nil {
		return nil, err
	}
	kid, _ := t.Header["kid"].(string)
	var set jwt.VerificationKeySet
	for _, k := range keys {
		if kid != "" && k.kid != "" && k.kid != kid {
			continue
		}
		if k.alg != "" && k.alg != t.Method.Alg() {
			continue
		}
		pub, err := publicKeyOf(k.key)
		if err != nil {
			continue
		}
		if checkKeyFitsAlgorithm(pub, t.Method.Alg()) != nil {
			continue
		}
		set.Keys = append(set.Keys, pub)
	}
	switch len(set.Keys) {
	case 0:
		if kid != "" {
			return nil, fmt.Errorf("no %s key with kid %q found", t.Method.Alg(), kid)
		}
		return nil, fmt.Errorf("no key suitable for %s found", t.Method.Alg())
	case 1:
		return set.Keys[0], nil
	default:
		return set, nil
	}
}

// EncodeOptions configures Encode.
type EncodeOptions struct {
	// Algorithm is one of SupportedAlgorithms; empty means DefaultAlgorithm.
	Algorithm string
	// Secret is the HMAC shared secret for HS*; required (non-empty) there.
	Secret string
	// Key is the private key for every other algorithm: PEM (PKCS#1,
	// PKCS#8 or SEC 1) or a private JWK, e.g. GenerateKey's output.
	Key string
	// KeyID sets the "kid" header. Empty falls back to the private JWK's
	// own "kid" when Key is a JWK; a PEM key carries none.
	KeyID string
}

// Encode signs claims into a JWT. HMAC algorithms (HS256/HS384/HS512) sign
// with opts.Secret as a raw shared-secret string; every other supported
// algorithm signs with opts.Key, a private key (RSA for RS*/PS*, EC on the
// matching curve for ES*, Ed25519 for EdDSA). Signing is entirely offline.
func Encode(claims map[string]any, opts EncodeOptions) (string, error) {
	if len(claims) == 0 {
		return "", apperr.New(400, "EMPTY_CLAIMS", "claims must not be empty")
	}
	method, signingKey, jwkKid, err := signingMethodAndKey(opts)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(method, jwt.MapClaims(claims))
	if kid := cmp.Or(opts.KeyID, jwkKid); kid != "" {
		token.Header["kid"] = kid
	}
	signed, err := token.SignedString(signingKey)
	if err != nil {
		return "", apperr.Newf(500, "SIGNING_FAILED", "%s", err)
	}
	return signed, nil
}

// lookupMethod resolves an algorithm name to its golang-jwt signing
// method, restricted to SupportedAlgorithms (so "none" is rejected).
func lookupMethod(alg string) (jwt.SigningMethod, error) {
	if !slices.Contains(SupportedAlgorithms, alg) {
		return nil, apperr.New(400, "UNSUPPORTED_ALGORITHM", "algorithm must be one of: "+strings.Join(SupportedAlgorithms, ", "))
	}
	return jwt.GetSigningMethod(alg), nil
}

func signingMethodAndKey(opts EncodeOptions) (method jwt.SigningMethod, key any, kid string, err error) {
	alg := cmp.Or(opts.Algorithm, DefaultAlgorithm)
	method, err = lookupMethod(alg)
	if err != nil {
		return nil, nil, "", err
	}

	switch method.(type) {
	case *jwt.SigningMethodHMAC:
		if opts.Secret == "" {
			return nil, nil, "", apperr.Newf(400, "MISSING_SECRET", "a non-empty secret is required for %s", alg)
		}
		return method, []byte(opts.Secret), "", nil
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		key, kid, err = parsePrivateKey[*rsa.PrivateKey](opts.Key, "RSA", jwt.ParseRSAPrivateKeyFromPEM)
	case *jwt.SigningMethodECDSA:
		key, kid, err = parsePrivateKey[*ecdsa.PrivateKey](opts.Key, "ECDSA", jwt.ParseECPrivateKeyFromPEM)
	case *jwt.SigningMethodEd25519:
		key, kid, err = parsePrivateKey(opts.Key, "Ed25519", func(b []byte) (ed25519.PrivateKey, error) {
			k, err := jwt.ParseEdPrivateKeyFromPEM(b)
			if err != nil {
				return nil, err
			}
			ed, ok := k.(ed25519.PrivateKey)
			if !ok {
				return nil, jwt.ErrNotEdPrivateKey
			}
			return ed, nil
		})
	}
	if err != nil {
		return nil, nil, "", err
	}
	if err := checkKeyFitsAlgorithm(key, alg); err != nil {
		return nil, nil, "", apperr.Newf(400, "INVALID_KEY", "%s", err)
	}
	return method, key, kid, nil
}

// parsePrivateKey parses a private key of type K from either a private JWK
// (input starting with "{") or PEM text via fromPEM, wrapping failures as
// INVALID_KEY with the key family in the message. For a JWK it also
// returns the JWK's "kid".
func parsePrivateKey[K any](input, family string, fromPEM func([]byte) (K, error)) (K, string, error) {
	var zero K
	if strings.HasPrefix(strings.TrimSpace(input), "{") {
		keys, err := parseJWKInput(strings.TrimSpace(input))
		if err != nil {
			return zero, "", apperr.Newf(400, "INVALID_KEY", "invalid %s private key: %s", family, err)
		}
		if len(keys) != 1 {
			return zero, "", apperr.Newf(400, "INVALID_KEY", "invalid %s private key: expected a single private JWK, got %d keys", family, len(keys))
		}
		k, ok := keys[0].key.(K)
		if !ok {
			return zero, "", apperr.Newf(400, "INVALID_KEY", "invalid %s private key: JWK is not a %s private key", family, family)
		}
		return k, keys[0].kid, nil
	}
	k, err := fromPEM([]byte(input))
	if err != nil {
		return zero, "", apperr.Newf(400, "INVALID_KEY", "invalid %s private key: %s", family, err)
	}
	return k, "", nil
}

// errNotObject reports claims JSON that parsed but isn't an object.
var errNotObject = errors.New("claims must be a JSON object")
