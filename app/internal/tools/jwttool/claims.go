package jwttool

import (
	"bytes"
	"crypto/rand"
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

// timeClaims are registered claims whose value must be a NumericDate
// (RFC 7519 §2). They are set through ClaimsSpec's duration fields, never
// as a free-form custom claim, since a string "exp" would make golang-jwt
// (and every other verifier) reject the token as malformed.
var timeClaims = []string{"exp", "nbf", "iat"}

// ClaimsSpec describes the registered claims (RFC 7519 §4.1) and custom
// claims to add on top of a base claims object before signing. Every
// field is optional; zero values add nothing.
type ClaimsSpec struct {
	Issuer   string   // "iss"
	Subject  string   // "sub"
	Audience []string // "aud": one value is emitted as a string, several as a list

	// ExpiresIn sets "exp" to now+ExpiresIn when positive.
	ExpiresIn time.Duration
	// NotBefore sets "nbf" to now+NotBefore when non-zero (negative values
	// back-date it, e.g. to tolerate clock skew).
	NotBefore time.Duration
	// IssuedAt sets "iat" to now.
	IssuedAt bool
	// ID sets "jti"; GenerateID sets it to a random value instead.
	ID         string
	GenerateID bool

	// Custom holds "name=value" pairs, see ParseClaimPairs.
	Custom []string
}

// BuildClaims returns a copy of base with spec applied. Precedence, lowest
// to highest: base (e.g. a JSON claims object), spec.Custom pairs, then
// the registered-claim fields — so an explicit --sub/--aud always wins.
// now is a parameter (rather than time.Now()) so callers and tests get
// deterministic output.
func BuildClaims(base map[string]any, spec ClaimsSpec, now time.Time) (map[string]any, error) {
	claims := maps.Clone(base)
	if claims == nil {
		claims = map[string]any{}
	}

	custom, err := ParseClaimPairs(spec.Custom)
	if err != nil {
		return nil, err
	}
	maps.Copy(claims, custom)

	setIf := func(name, value string) {
		if value != "" {
			claims[name] = value
		}
	}
	setIf("iss", spec.Issuer)
	setIf("sub", spec.Subject)
	switch len(spec.Audience) {
	case 0:
	case 1:
		claims["aud"] = spec.Audience[0]
	default:
		claims["aud"] = spec.Audience
	}
	if spec.ExpiresIn < 0 {
		return nil, apperr.New(400, "INVALID_CLAIMS", "expires_in must be positive")
	}
	if spec.ExpiresIn > 0 {
		claims["exp"] = now.Add(spec.ExpiresIn).Unix()
	}
	if spec.NotBefore != 0 {
		claims["nbf"] = now.Add(spec.NotBefore).Unix()
	}
	if spec.IssuedAt {
		claims["iat"] = now.Unix()
	}
	switch {
	case spec.GenerateID:
		claims["jti"] = rand.Text()
	case spec.ID != "":
		claims["jti"] = spec.ID
	}
	return claims, nil
}

// ParseClaimPairs turns "name=value" pairs into custom claims whose values
// are a string or a list of strings:
//
//	role=admin              → "role": "admin"
//	groups=dev groups=ops   → "groups": ["dev", "ops"]   (repeated name)
//	scopes[]=read           → "scopes": ["read"]          (forced list)
//
// The value is everything after the first "=", taken verbatim (it may
// itself contain "=" or ","). exp/nbf/iat are rejected — they must be
// numeric dates, set via ClaimsSpec's duration fields instead.
func ParseClaimPairs(pairs []string) (map[string]any, error) {
	out := make(map[string]any, len(pairs))
	for _, pair := range pairs {
		name, value, ok := strings.Cut(pair, "=")
		name = strings.TrimSpace(name)
		name, forceList := strings.CutSuffix(name, "[]")
		if !ok || name == "" {
			return nil, apperr.Newf(400, "INVALID_CLAIMS", "custom claim %q must have the form name=value or name[]=value", pair)
		}
		if slices.Contains(timeClaims, name) {
			return nil, apperr.Newf(400, "INVALID_CLAIMS", "%q must be a numeric date: use the expires_in/not_before/issued_at options (CLI: --exp/--nbf/--iat) instead of a custom claim", name)
		}
		switch prev := out[name].(type) {
		case nil:
			if forceList {
				out[name] = []string{value}
			} else {
				out[name] = value
			}
		case string:
			out[name] = []string{prev, value}
		case []string:
			out[name] = append(prev, value)
		}
	}
	return out, nil
}

// ParseClaimsJSON decodes a claims JSON object for CLI/REST/MCP adapters
// that receive claims as raw text. Blank input (or JSON null) yields an
// empty map rather than an error, so claims can come entirely from a
// ClaimsSpec; Encode still rejects a token whose final claims are empty.
// Numbers are kept exact (json.Number), so large integer claims like
// "exp" are not rounded through float64.
func ParseClaimsJSON(raw []byte) (map[string]any, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	// encoding/json/v2 rejects duplicate member names and invalid UTF-8 —
	// both matter for a JWT: verifiers disagree on which of two "sub"
	// members wins, a classic token-confusion vector. It also rejects
	// trailing data after the object.
	var v any
	if err := json.Unmarshal(raw, &v, exactNumbers); err != nil {
		return nil, apperr.Newf(400, "INVALID_CLAIMS", "invalid claims JSON: %s", err)
	}
	switch claims := v.(type) {
	case nil:
		return map[string]any{}, nil
	case map[string]any:
		return claims, nil
	default:
		return nil, apperr.Newf(400, "INVALID_CLAIMS", "invalid claims JSON: %s", errNotObject)
	}
}

// exactNumbers decodes JSON numbers held in an `any` as json.Number (the
// v1 type golang-jwt understands for NumericDate claims) instead of
// float64, so integers above 2^53 — and exp/iat values — round-trip
// byte-for-byte.
var exactNumbers = json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *any) error {
	if dec.PeekKind() != '0' {
		return errors.ErrUnsupported
	}
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	*v = jsonv1.Number(tok.String())
	return nil
}))

// ParseDuration is time.ParseDuration plus a "d" (24h days) suffix, e.g.
// "30d", and treats "" as zero. Used for expires_in/not_before options.
func ParseDuration(field, s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	if days, ok := strings.CutSuffix(s, "d"); ok {
		n, err := strconv.Atoi(days)
		if err == nil {
			return time.Duration(n) * 24 * time.Hour, nil
		}
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, apperr.Newf(400, "INVALID_OPTION", "%s must be a duration like 15m, 1h or 7d, got %q", field, s)
	}
	return d, nil
}

// SplitLines splits s into trimmed, non-empty lines, skipping lines that
// start with "#". Used for the web UI's one-claim-per-line textarea.
func SplitLines(s string) []string {
	var out []string
	for line := range strings.Lines(s) {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out
}

// SplitList splits s on commas and newlines into trimmed, non-empty
// items. Used for audience lists typed into a single text field.
func SplitList(s string) []string {
	var out []string
	for item := range strings.FieldsFuncSeq(s, func(r rune) bool { return r == ',' || r == '\n' }) {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
