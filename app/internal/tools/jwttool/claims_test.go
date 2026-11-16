package jwttool

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseClaimPairs(t *testing.T) {
	got, err := ParseClaimPairs([]string{"role=admin", "groups=dev", "groups=ops", "scopes[]=read", "url=https://x/?a=b,c"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"role":   "admin",
		"groups": []string{"dev", "ops"},
		"scopes": []string{"read"},
		"url":    "https://x/?a=b,c",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	for _, bad := range []string{"noequals", "=value", "exp=123", "iat[]=1"} {
		if _, err := ParseClaimPairs([]string{bad}); err == nil {
			t.Errorf("%q: expected an error", bad)
		} else {
			assertCode(t, err, "INVALID_CLAIMS")
		}
	}
}

func TestBuildClaimsPrecedenceAndTimes(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	got, err := BuildClaims(
		map[string]any{"sub": "from-json", "keep": true},
		ClaimsSpec{
			Subject: "from-flag", Issuer: "me", Audience: []string{"a", "b"},
			ExpiresIn: time.Hour, NotBefore: -30 * time.Second, IssuedAt: true, ID: "id-1",
			Custom: []string{"sub=from-custom", "tier=gold"},
		}, now)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"sub": "from-flag", "keep": true, "iss": "me", "aud": []string{"a", "b"},
		"exp": int64(1_700_003_600), "nbf": int64(1_699_999_970), "iat": int64(1_700_000_000),
		"jti": "id-1", "tier": "gold",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}

	single, _ := BuildClaims(nil, ClaimsSpec{Audience: []string{"only"}, GenerateID: true}, now)
	if single["aud"] != "only" {
		t.Errorf("single audience should be a string, got %#v", single["aud"])
	}
	if jti, _ := single["jti"].(string); len(jti) < 16 {
		t.Errorf("generated jti too short: %q", jti)
	}
}

// A token built with exp/nbf in the past/future must decode but fail
// verification with a useful reason.
func TestDecodeChecksTimeClaims(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour)
	claims, _ := BuildClaims(nil, ClaimsSpec{Subject: "x", ExpiresIn: time.Hour}, past)
	token, _ := Encode(claims, EncodeOptions{Secret: "s"})
	r, err := Decode(token, "s", "")
	if err != nil {
		t.Fatal(err)
	}
	mustInvalid(t, r)

	claims, _ = BuildClaims(nil, ClaimsSpec{Subject: "x", ExpiresIn: time.Hour, IssuedAt: true}, time.Now())
	token, _ = Encode(claims, EncodeOptions{Secret: "s"})
	r, _ = Decode(token, "s", "")
	mustValid(t, r)
}

func TestParseClaimsJSON(t *testing.T) {
	for _, in := range []string{"", "  \n", "null"} {
		got, err := ParseClaimsJSON([]byte(in))
		if err != nil || len(got) != 0 {
			t.Errorf("%q: got %v, %v; want empty map", in, got, err)
		}
	}
	for _, in := range []string{"[1]", "{bad", `{"a":1} {"b":2}`} {
		_, err := ParseClaimsJSON([]byte(in))
		assertCode(t, err, "INVALID_CLAIMS")
	}
}

func TestParseDuration(t *testing.T) {
	cases := map[string]time.Duration{"": 0, "90s": 90 * time.Second, "7d": 7 * 24 * time.Hour, "1h30m": 90 * time.Minute}
	for in, want := range cases {
		if got, err := ParseDuration("f", in); err != nil || got != want {
			t.Errorf("%q: got %v, %v; want %v", in, got, err, want)
		}
	}
	_, err := ParseDuration("f", "soon")
	assertCode(t, err, "INVALID_OPTION")
}

func TestSplitHelpers(t *testing.T) {
	if got := SplitLines("a=1\n\n# comment\n  b=2  \n"); !reflect.DeepEqual(got, []string{"a=1", "b=2"}) {
		t.Errorf("SplitLines: %#v", got)
	}
	if got := SplitList(" a, b\nc,,"); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("SplitList: %#v", got)
	}
}

// Regression: HS* signing with an empty secret used to succeed, producing
// a token anyone can forge.
func TestEncodeHMACRequiresSecret(t *testing.T) {
	_, err := Encode(map[string]any{"a": 1}, EncodeOptions{Algorithm: "HS256"})
	assertCode(t, err, "MISSING_SECRET")
}

// Regression: verifying an HS* token while supplying only a key (no
// secret) used to verify against an empty HMAC key, so a token forged
// with an empty secret reported valid=true.
func TestDecodeHMACTokenWithOnlyKeyIsInvalid(t *testing.T) {
	// HS256 with an empty key — what an attacker can mint without knowing
	// any secret.
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "forged"}).SignedString([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	kp, _ := GenerateKey(KeyOptions{Algorithm: "ES256"})
	r, err := Decode(forged, "", kp.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	mustInvalid(t, r)
}

func TestDecodeRejectsAlgNone(t *testing.T) {
	const unsigned = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ4In0."
	r, err := Decode(unsigned, "s", "")
	if err != nil {
		t.Fatal(err)
	}
	mustInvalid(t, r)
}

func TestParseClaimsJSONStrict(t *testing.T) {
	// Duplicate member names are rejected (token-confusion vector).
	_, err := ParseClaimsJSON([]byte(`{"sub":"alice","sub":"admin"}`))
	assertCode(t, err, "INVALID_CLAIMS")

	// Integers above 2^53 are kept exact in the signed payload.
	claims, err := ParseClaimsJSON([]byte(`{"n":9007199254740993,"nested":{"x":1.5}}`))
	if err != nil {
		t.Fatal(err)
	}
	token, _ := Encode(claims, EncodeOptions{Secret: "s"})
	payload := strings.Split(token, ".")[1]
	raw, _ := base64.RawURLEncoding.DecodeString(payload)
	if !strings.Contains(string(raw), `"n":9007199254740993`) {
		t.Fatalf("large integer was not preserved exactly: %s", raw)
	}
}
