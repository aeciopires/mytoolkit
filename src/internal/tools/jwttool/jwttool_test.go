package jwttool

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	claims := map[string]any{"sub": "123"}
	token, err := Encode(claims, "mysecret", "HS256")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Decode(token, "mysecret")
	if err != nil {
		t.Fatal(err)
	}
	if result.Claims["sub"] != "123" {
		t.Errorf("claims not decoded correctly: %+v", result.Claims)
	}
	if result.Valid == nil || !*result.Valid {
		t.Error("expected Valid=true with correct secret")
	}
}

func TestDecodeWrongSecret(t *testing.T) {
	token, err := Encode(map[string]any{"sub": "123"}, "mysecret", "HS256")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(token, "wrongsecret")
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || *result.Valid {
		t.Error("expected Valid=false with wrong secret")
	}
}

func TestDecodeWithoutSecret(t *testing.T) {
	token, _ := Encode(map[string]any{"sub": "123"}, "mysecret", "HS256")
	result, err := Decode(token, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid != nil {
		t.Error("expected Valid to be nil when no secret is supplied")
	}
	if result.Header["alg"] != "HS256" {
		t.Errorf("unexpected header: %+v", result.Header)
	}
}

func TestDecodeMalformedToken(t *testing.T) {
	if _, err := Decode("not-a-jwt", ""); err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestEncodeEmptyClaims(t *testing.T) {
	if _, err := Encode(map[string]any{}, "secret", "HS256"); err == nil {
		t.Error("expected error for empty claims")
	}
}

func TestEncodeUnsupportedAlgorithm(t *testing.T) {
	if _, err := Encode(map[string]any{"a": 1}, "secret", "RS256"); err == nil {
		t.Error("expected error for unsupported algorithm")
	}
}
