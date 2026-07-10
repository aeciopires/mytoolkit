package jwttool

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	claims := map[string]any{"sub": "123"}
	token, err := Encode(claims, "mysecret", "", "HS256")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Decode(token, "mysecret", "")
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
	token, err := Encode(map[string]any{"sub": "123"}, "mysecret", "", "HS256")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(token, "wrongsecret", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || *result.Valid {
		t.Error("expected Valid=false with wrong secret")
	}
}

func TestDecodeWithoutSecret(t *testing.T) {
	token, _ := Encode(map[string]any{"sub": "123"}, "mysecret", "", "HS256")
	result, err := Decode(token, "", "")
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
	if _, err := Decode("not-a-jwt", "", ""); err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestEncodeEmptyClaims(t *testing.T) {
	if _, err := Encode(map[string]any{}, "secret", "", "HS256"); err == nil {
		t.Error("expected error for empty claims")
	}
}

func TestEncodeUnsupportedAlgorithm(t *testing.T) {
	if _, err := Encode(map[string]any{"a": 1}, "secret", "", "RS9000"); err == nil {
		t.Error("expected error for unsupported algorithm")
	}
}

func TestEncodeDefaultAlgorithmIsHS256(t *testing.T) {
	token, err := Encode(map[string]any{"a": 1}, "secret", "", "")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(token, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Header["alg"] != DefaultAlgorithm {
		t.Errorf("expected default algorithm %q, got %v", DefaultAlgorithm, result.Header["alg"])
	}
}

// --- Asymmetric algorithms (RSA, RSA-PSS, ECDSA, EdDSA) ---

func genRSAKeyPEM(t *testing.T) (privPEM, pubPEM string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
}

func genECKeyPEM(t *testing.T) (privPEM, pubPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
}

func genEdKeyPEM(t *testing.T) (privPEM, pubPEM string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
}

func TestEncodeDecodeRoundTripRSA(t *testing.T) {
	priv, pub := genRSAKeyPEM(t)
	for _, alg := range []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512"} {
		t.Run(alg, func(t *testing.T) {
			token, err := Encode(map[string]any{"sub": "rsa"}, "", priv, alg)
			if err != nil {
				t.Fatal(err)
			}
			result, err := Decode(token, "", pub)
			if err != nil {
				t.Fatal(err)
			}
			if result.Valid == nil || !*result.Valid {
				t.Errorf("expected Valid=true for %s with correct public key", alg)
			}
			if result.Claims["sub"] != "rsa" {
				t.Errorf("claims not decoded correctly: %+v", result.Claims)
			}
		})
	}
}

func TestEncodeDecodeRoundTripECDSA(t *testing.T) {
	priv, pub := genECKeyPEM(t)
	// ES256 needs a P-256 key (what genECKeyPEM produces); this exercises
	// the ECDSA code path end-to-end, not every curve/algorithm pairing.
	token, err := Encode(map[string]any{"sub": "ec"}, "", priv, "ES256")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(token, "", pub)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || !*result.Valid {
		t.Error("expected Valid=true for ES256 with correct public key")
	}
}

func TestEncodeDecodeRoundTripEdDSA(t *testing.T) {
	priv, pub := genEdKeyPEM(t)
	token, err := Encode(map[string]any{"sub": "ed"}, "", priv, "EdDSA")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(token, "", pub)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || !*result.Valid {
		t.Error("expected Valid=true for EdDSA with correct public key")
	}
}

func TestDecodeRSAWrongPublicKeyFails(t *testing.T) {
	priv, _ := genRSAKeyPEM(t)
	_, otherPub := genRSAKeyPEM(t)
	token, err := Encode(map[string]any{"sub": "rsa"}, "", priv, "RS256")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(token, "", otherPub)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || *result.Valid {
		t.Error("expected Valid=false when verifying against a different key pair's public key")
	}
}

func TestEncodeRSAInvalidKeyPEM(t *testing.T) {
	if _, err := Encode(map[string]any{"a": 1}, "", "not a pem key", "RS256"); err == nil {
		t.Error("expected error for invalid RSA private key PEM")
	}
}

func TestDecodeSecretAgainstAsymmetricTokenFailsCleanly(t *testing.T) {
	priv, _ := genRSAKeyPEM(t)
	token, err := Encode(map[string]any{"sub": "rsa"}, "", priv, "RS256")
	if err != nil {
		t.Fatal(err)
	}
	// Pasting an HMAC secret into "secret" for an RS256 token should fail
	// verification, not panic or silently succeed.
	result, err := Decode(token, "some-secret", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid == nil || *result.Valid {
		t.Error("expected Valid=false when verifying an RS256 token with an HMAC secret instead of a public key")
	}
}
