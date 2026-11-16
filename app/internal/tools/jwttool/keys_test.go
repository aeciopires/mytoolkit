package jwttool

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	ae, ok := errors.AsType[*apperr.Error](err)
	if !ok {
		t.Fatalf("got error %v (%T), want *apperr.Error with code %s", err, err, code)
	}
	if ae.Code != code {
		t.Fatalf("got code %s (%s), want %s", ae.Code, ae.Message, code)
	}
}

func mustValid(t *testing.T, r DecodeResult) {
	t.Helper()
	if r.Valid == nil || !*r.Valid {
		t.Fatalf("expected Valid=true, got %v (verify_error=%q)", r.Valid, r.VerifyError)
	}
}

func mustInvalid(t *testing.T, r DecodeResult) {
	t.Helper()
	if r.Valid == nil || *r.Valid {
		t.Fatalf("expected Valid=false, got %v", r.Valid)
	}
	if r.VerifyError == "" {
		t.Error("expected a verify_error explaining the failure")
	}
}

// Every asymmetric algorithm: generate → sign with each private form (PEM
// and private JWK) → verify with each public form (PEM, JWK, JWKS).
func TestGenerateKeySignVerifyAllForms(t *testing.T) {
	for _, alg := range SupportedAlgorithms {
		if strings.HasPrefix(alg, "HS") {
			continue
		}
		t.Run(alg, func(t *testing.T) {
			kp, err := GenerateKey(KeyOptions{Algorithm: alg})
			if err != nil {
				t.Fatal(err)
			}
			if kp.Algorithm != alg || kp.KeyID == "" || kp.JWKS == nil || len(kp.JWKS.Keys) != 1 {
				t.Fatalf("unexpected key pair: %+v", kp)
			}
			pub := kp.JWKS.Keys[0]
			if pub.D != "" || pub.P != "" {
				t.Fatal("JWKS must not contain private members")
			}
			if pub.Kid != pub.Thumbprint() {
				t.Errorf("default kid %q should be the RFC 7638 thumbprint %q", pub.Kid, pub.Thumbprint())
			}

			privJWK, _ := json.Marshal(kp.PrivateJWK)
			pubJWK, _ := json.Marshal(kp.PublicJWK)
			jwks, _ := json.Marshal(kp.JWKS)

			for name, signKey := range map[string]string{"pem": kp.PrivateKey, "jwk": string(privJWK)} {
				token, err := Encode(map[string]any{"sub": alg}, EncodeOptions{Algorithm: alg, Key: signKey, KeyID: kp.KeyID})
				if err != nil {
					t.Fatalf("sign with %s: %v", name, err)
				}
				for vname, verifyKey := range map[string]string{
					"pem": kp.PublicKey, "jwk": string(pubJWK), "jwks": string(jwks), "private-pem": kp.PrivateKey,
				} {
					r, err := Decode(token, "", verifyKey)
					if err != nil {
						t.Fatalf("sign=%s verify=%s: %v", name, vname, err)
					}
					mustValid(t, r)
				}
			}
		})
	}
}

func TestGenerateKeyHMAC(t *testing.T) {
	kp, err := GenerateKey(KeyOptions{Algorithm: "HS512"})
	if err != nil {
		t.Fatal(err)
	}
	if len(kp.Secret) < 64 || kp.JWKS != nil || kp.PublicKey != "" || kp.PrivateJWK.Kty != "oct" {
		t.Fatalf("unexpected HMAC key pair: %+v", kp)
	}
	token, err := Encode(map[string]any{"sub": "x"}, EncodeOptions{Algorithm: "HS512", Secret: kp.Secret})
	if err != nil {
		t.Fatal(err)
	}
	r, _ := Decode(token, kp.Secret, "")
	mustValid(t, r)
}

func TestGenerateKeyRSABits(t *testing.T) {
	if _, err := GenerateKey(KeyOptions{Algorithm: "RS256", Bits: 1024}); err == nil {
		t.Fatal("expected 1024-bit RSA to be rejected")
	} else {
		assertCode(t, err, "INVALID_OPTION")
	}
	if _, err := GenerateKey(KeyOptions{Algorithm: "none"}); err == nil {
		t.Fatal(`expected "none" to be rejected`)
	} else {
		assertCode(t, err, "UNSUPPORTED_ALGORITHM")
	}
}

func TestJWKSelectsByKid(t *testing.T) {
	a, _ := GenerateKey(KeyOptions{Algorithm: "ES256", KeyID: "a"})
	b, _ := GenerateKey(KeyOptions{Algorithm: "ES256", KeyID: "b"})
	set, err := BuildJWKS(a.PublicKey, JWKSOptions{KeyID: "a"})
	if err != nil {
		t.Fatal(err)
	}
	setB, _ := BuildJWKS(b.PublicKey, JWKSOptions{KeyID: "b"})
	set.Keys = append(set.Keys, setB.Keys...)
	jwks, _ := json.Marshal(set)

	token, _ := Encode(map[string]any{"sub": "x"}, EncodeOptions{Algorithm: "ES256", Key: b.PrivateKey, KeyID: "b"})
	r, _ := Decode(token, "", string(jwks))
	mustValid(t, r)

	token, _ = Encode(map[string]any{"sub": "x"}, EncodeOptions{Algorithm: "ES256", Key: b.PrivateKey, KeyID: "zzz"})
	r, _ = Decode(token, "", string(jwks))
	mustInvalid(t, r)
	if !strings.Contains(r.VerifyError, `"zzz"`) {
		t.Errorf("verify_error should name the missing kid, got %q", r.VerifyError)
	}
}

func TestBuildJWKSMultiplePEMBlocks(t *testing.T) {
	rsaKP, _ := GenerateKey(KeyOptions{Algorithm: "PS256"})
	edKP, _ := GenerateKey(KeyOptions{Algorithm: "EdDSA"})
	set, err := BuildJWKS(rsaKP.PrivateKey+edKP.PublicKey, JWKSOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Keys) != 2 || set.Keys[0].Alg != "RS256" || set.Keys[1].Alg != "EdDSA" {
		t.Fatalf("unexpected set: %+v", set)
	}
	if set.Keys[0].D != "" {
		t.Error("private PEM input must be reduced to its public half")
	}
	if _, err := BuildJWKS(rsaKP.PublicKey+edKP.PublicKey, JWKSOptions{KeyID: "x"}); err == nil {
		t.Error("expected kid with multiple keys to be rejected")
	}
	if _, err := BuildJWKS(rsaKP.PublicKey, JWKSOptions{Algorithm: "ES256"}); err == nil {
		t.Error("expected an RSA key labeled ES256 to be rejected")
	}
	if _, err := BuildJWKS("garbage", JWKSOptions{}); err == nil {
		t.Error("expected garbage input to be rejected")
	}
}

// RFC 7638 §3.1's worked example.
func TestThumbprintRFC7638(t *testing.T) {
	j := JWK{
		Kty: "RSA",
		E:   "AQAB",
		N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
	}
	if got, want := j.Thumbprint(), "NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"; got != want {
		t.Fatalf("thumbprint: got %s, want %s", got, want)
	}
}

func TestEncodeECDSACurveMismatchIsInvalidKey(t *testing.T) {
	kp, _ := GenerateKey(KeyOptions{Algorithm: "ES256"})
	_, err := Encode(map[string]any{"a": 1}, EncodeOptions{Algorithm: "ES384", Key: kp.PrivateKey})
	assertCode(t, err, "INVALID_KEY")
}
