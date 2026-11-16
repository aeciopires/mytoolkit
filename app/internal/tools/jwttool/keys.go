package jwttool

import (
	"cmp"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json/v2"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

// DefaultRSABits is the RSA modulus size GenerateKey uses when
// KeyOptions.Bits is zero. 2048 is the minimum RFC 7518 §3.3 allows for
// RS*/PS* and what every mainstream JWT library generates by default.
const DefaultRSABits = 2048

// SupportedRSABits lists the RSA modulus sizes GenerateKey accepts.
var SupportedRSABits = []int{2048, 3072, 4096}

// JWK is a JSON Web Key (RFC 7517) for the key types this tool supports:
// "RSA" (RS*/PS*), "EC" (ES*), "OKP" (EdDSA/Ed25519), and "oct" (HS*).
// Private members (d, p, q, dp, dq, qi, k) are only populated on a private
// JWK; a public JWK — and therefore every entry of a JWKS — never carries
// them.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use,omitempty"`
	Alg string `json:"alg,omitempty"`
	Kid string `json:"kid,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	D   string `json:"d,omitempty"`
	P   string `json:"p,omitempty"`
	Q   string `json:"q,omitempty"`
	DP  string `json:"dp,omitempty"`
	DQ  string `json:"dq,omitempty"`
	QI  string `json:"qi,omitempty"`
	K   string `json:"k,omitempty"`
}

// JWKSet is a JSON Web Key Set (RFC 7517 §5), the document an issuer
// publishes (typically at /.well-known/jwks.json) so verifiers can find
// the public key matching a token's "kid" header.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// KeyOptions configures GenerateKey.
type KeyOptions struct {
	// Algorithm is one of SupportedAlgorithms; empty means DefaultAlgorithm.
	Algorithm string
	// Bits is the RSA modulus size for RS*/PS* (one of SupportedRSABits);
	// zero means DefaultRSABits. Ignored for every other algorithm.
	Bits int
	// KeyID is the "kid" to stamp on the generated JWKs; empty means the
	// RFC 7638 SHA-256 thumbprint of the public key.
	KeyID string
}

// KeyPair is GenerateKey's result. For HS* only Secret and PrivateJWK (an
// "oct" JWK) are set — a symmetric key has no public half and must never
// be published in a JWKS. For every other algorithm, PrivateKey/PublicKey
// are PEM (PKCS#8 / PKIX), and JWKS holds the public JWK alone, ready to
// publish.
type KeyPair struct {
	Algorithm  string  `json:"algorithm"`
	KeyID      string  `json:"kid"`
	Secret     string  `json:"secret,omitempty"`
	PrivateKey string  `json:"private_key,omitempty"`
	PublicKey  string  `json:"public_key,omitempty"`
	PrivateJWK *JWK    `json:"private_jwk,omitempty"`
	PublicJWK  *JWK    `json:"public_jwk,omitempty"`
	JWKS       *JWKSet `json:"jwks,omitempty"`
}

// GenerateKey creates fresh key material for opts.Algorithm entirely
// offline (crypto/rand only, no network): a random secret for HS*, an RSA
// key for RS*/PS*, an ECDSA key on the curve RFC 7518 mandates for ES*
// (P-256/P-384/P-521), or an Ed25519 key for EdDSA.
func GenerateKey(opts KeyOptions) (KeyPair, error) {
	alg := cmp.Or(opts.Algorithm, DefaultAlgorithm)
	method, err := lookupMethod(alg)
	if err != nil {
		return KeyPair{}, err
	}

	var priv crypto.Signer
	switch m := method.(type) {
	case *jwt.SigningMethodHMAC:
		return generateSecret(alg, m, opts.KeyID)
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		bits := opts.Bits
		if bits == 0 {
			bits = DefaultRSABits
		}
		if !slices.Contains(SupportedRSABits, bits) {
			return KeyPair{}, apperr.Newf(400, "INVALID_OPTION", "bits must be one of %v, got %d", SupportedRSABits, bits)
		}
		priv, err = rsa.GenerateKey(rand.Reader, bits)
	case *jwt.SigningMethodECDSA:
		priv, err = ecdsa.GenerateKey(curveForBits(m.CurveBits), rand.Reader)
	case *jwt.SigningMethodEd25519:
		_, priv, err = ed25519.GenerateKey(rand.Reader)
	}
	if err != nil {
		return KeyPair{}, apperr.Newf(500, "KEYGEN_FAILED", "key generation failed: %s", err)
	}

	pubJWK, err := publicJWK(priv.Public(), alg, opts.KeyID)
	if err != nil {
		return KeyPair{}, err
	}
	privJWK, err := privateJWK(priv, alg, pubJWK.Kid)
	if err != nil {
		return KeyPair{}, err
	}
	privPEM, pubPEM, err := encodePEMPair(priv)
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{
		Algorithm:  alg,
		KeyID:      pubJWK.Kid,
		PrivateKey: privPEM,
		PublicKey:  pubPEM,
		PrivateJWK: &privJWK,
		PublicJWK:  &pubJWK,
		JWKS:       &JWKSet{Keys: []JWK{pubJWK}},
	}, nil
}

// generateSecret creates an HS* secret with as many random bytes as the
// hash's output size (RFC 7518 §3.2's minimum key size), rendered as
// base64url text. Encode/Decode use a secret's UTF-8 bytes as the HMAC key,
// so the "oct" JWK's "k" is the base64url of that text, keeping the two
// representations interchangeable.
func generateSecret(alg string, m *jwt.SigningMethodHMAC, kid string) (KeyPair, error) {
	raw := make([]byte, m.Hash.Size())
	if _, err := rand.Read(raw); err != nil {
		return KeyPair{}, apperr.Newf(500, "KEYGEN_FAILED", "key generation failed: %s", err)
	}
	secret := b64(raw)
	k := b64([]byte(secret))
	if kid == "" {
		kid = thumbprint(map[string]string{"k": k, "kty": "oct"})
	}
	return KeyPair{
		Algorithm:  alg,
		KeyID:      kid,
		Secret:     secret,
		PrivateJWK: &JWK{Kty: "oct", Use: "sig", Alg: alg, Kid: kid, K: k},
	}, nil
}

// JWKSOptions configures BuildJWKS.
type JWKSOptions struct {
	// Algorithm sets every key's "alg"; empty infers it from each key's
	// type (RSA → RS256, EC → ES256/384/512 by curve, Ed25519 → EdDSA).
	Algorithm string
	// KeyID sets the "kid"; only valid when the input holds a single key.
	// Empty means each key's RFC 7638 thumbprint.
	KeyID string
}

// BuildJWKS converts one or more keys — concatenated PEM blocks (public
// keys, private keys, or certificates), a single JWK, or an existing JWKS —
// into a JWKS holding only public JWKs. Private input is reduced to its
// public half, so the result is always safe to publish. Multiple PEM
// blocks produce a multi-key set, e.g. to publish the current and the
// next key during a rotation.
func BuildJWKS(input string, opts JWKSOptions) (JWKSet, error) {
	keys, err := parseKeys(input)
	if err != nil {
		return JWKSet{}, apperr.Newf(400, "INVALID_KEY", "%s", err)
	}
	if len(keys) == 0 {
		return JWKSet{}, apperr.New(400, "EMPTY_INPUT", "input must contain at least one PEM key, JWK, or JWKS")
	}
	if opts.KeyID != "" && len(keys) > 1 {
		return JWKSet{}, apperr.New(400, "INVALID_OPTION", "kid can only be set when the input holds a single key")
	}
	if opts.Algorithm != "" {
		if _, err := lookupMethod(opts.Algorithm); err != nil {
			return JWKSet{}, err
		}
	}

	set := JWKSet{Keys: make([]JWK, 0, len(keys))}
	for _, k := range keys {
		pub, err := publicKeyOf(k.key)
		if err != nil {
			return JWKSet{}, apperr.Newf(400, "INVALID_KEY", "%s", err)
		}
		alg := cmp.Or(opts.Algorithm, k.alg)
		if alg == "" {
			alg, err = inferAlgorithm(pub)
			if err != nil {
				return JWKSet{}, apperr.Newf(400, "INVALID_KEY", "%s", err)
			}
		}
		if err := checkKeyFitsAlgorithm(pub, alg); err != nil {
			return JWKSet{}, apperr.Newf(400, "INVALID_KEY", "%s", err)
		}
		j, err := publicJWK(pub, alg, cmp.Or(opts.KeyID, k.kid))
		if err != nil {
			return JWKSet{}, err
		}
		set.Keys = append(set.Keys, j)
	}
	return set, nil
}

// parsedKey is one key read by parseKeys, with the kid/alg it carried (a
// JWK may declare both; a PEM block never does).
type parsedKey struct {
	key any
	kid string
	alg string
}

// parseKeys reads every key in input: a JWKS ({"keys":[...]}), a single
// JWK, or one or more concatenated PEM blocks.
func parseKeys(input string) ([]parsedKey, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	if strings.HasPrefix(input, "{") {
		return parseJWKInput(input)
	}

	var keys []parsedKey
	rest := []byte(input)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		k, err := parsePEMBlock(block)
		if err != nil {
			return nil, err
		}
		keys = append(keys, parsedKey{key: k})
	}
	if len(keys) == 0 {
		return nil, errors.New("key must be PEM-encoded, a JWK, or a JWKS")
	}
	return keys, nil
}

func parseJWKInput(input string) ([]parsedKey, error) {
	var probe struct {
		Keys []JWK `json:"keys"`
		JWK  `json:",inline"`
	}
	if err := json.Unmarshal([]byte(input), &probe); err != nil {
		return nil, fmt.Errorf("invalid JWK/JWKS JSON: %w", err)
	}
	jwks := probe.Keys
	if jwks == nil {
		jwks = []JWK{probe.JWK}
	}
	keys := make([]parsedKey, 0, len(jwks))
	for _, j := range jwks {
		k, err := j.Key()
		if err != nil {
			return nil, err
		}
		keys = append(keys, parsedKey{key: k, kid: j.Kid, alg: j.Alg})
	}
	return keys, nil
}

func parsePEMBlock(block *pem.Block) (any, error) {
	var (
		k   any
		err error
	)
	switch block.Type {
	case "PUBLIC KEY":
		k, err = x509.ParsePKIXPublicKey(block.Bytes)
	case "RSA PUBLIC KEY":
		k, err = x509.ParsePKCS1PublicKey(block.Bytes)
	case "PRIVATE KEY":
		k, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		k, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		k, err = x509.ParseECPrivateKey(block.Bytes)
	case "CERTIFICATE":
		var cert *x509.Certificate
		if cert, err = x509.ParseCertificate(block.Bytes); err == nil {
			k = cert.PublicKey
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type %q", block.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", strings.ToLower(block.Type), err)
	}
	return k, nil
}

// publicKeyOf returns k's public half (k itself if it is already public),
// rejecting key types JWS can't use (e.g. X25519, DSA).
func publicKeyOf(k any) (crypto.PublicKey, error) {
	if s, ok := k.(crypto.Signer); ok {
		k = s.Public()
	}
	switch k.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey:
		return k, nil
	default:
		return nil, fmt.Errorf("unsupported key type %T", k)
	}
}

func inferAlgorithm(pub crypto.PublicKey) (string, error) {
	switch p := pub.(type) {
	case *rsa.PublicKey:
		return "RS256", nil
	case *ecdsa.PublicKey:
		switch p.Curve {
		case elliptic.P256():
			return "ES256", nil
		case elliptic.P384():
			return "ES384", nil
		case elliptic.P521():
			return "ES512", nil
		}
		return "", fmt.Errorf("unsupported EC curve %s", p.Curve.Params().Name)
	case ed25519.PublicKey:
		return "EdDSA", nil
	}
	return "", fmt.Errorf("unsupported key type %T", pub)
}

// checkKeyFitsAlgorithm rejects a key whose type (or EC curve) can't
// produce/verify alg's signatures, e.g. an RSA key labeled ES256 or a
// P-256 key labeled ES384 — mismatches golang-jwt only reports at
// sign/verify time with a generic error.
func checkKeyFitsAlgorithm(key any, alg string) error {
	method, err := lookupMethod(alg)
	if err != nil {
		return err
	}
	pub, err := publicKeyOf(key)
	if err != nil {
		return err
	}
	switch m := method.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		if _, ok := pub.(*rsa.PublicKey); ok {
			return nil
		}
	case *jwt.SigningMethodECDSA:
		if ec, ok := pub.(*ecdsa.PublicKey); ok {
			if ec.Curve.Params().BitSize != m.CurveBits {
				return fmt.Errorf("%s requires a %s key, got %s", alg, curveForBits(m.CurveBits).Params().Name, ec.Curve.Params().Name)
			}
			return nil
		}
	case *jwt.SigningMethodEd25519:
		if _, ok := pub.(ed25519.PublicKey); ok {
			return nil
		}
	}
	return fmt.Errorf("a %s key cannot be used with %s", keyTypeName(pub), alg)
}

func keyTypeName(pub crypto.PublicKey) string {
	switch pub.(type) {
	case *rsa.PublicKey:
		return "RSA"
	case *ecdsa.PublicKey:
		return "ECDSA"
	case ed25519.PublicKey:
		return "Ed25519"
	}
	return fmt.Sprintf("%T", pub)
}

func curveForBits(bits int) elliptic.Curve {
	switch bits {
	case 384:
		return elliptic.P384()
	case 521:
		return elliptic.P521()
	default:
		return elliptic.P256()
	}
}

func curveByName(name string) (elliptic.Curve, error) {
	switch name {
	case "P-256":
		return elliptic.P256(), nil
	case "P-384":
		return elliptic.P384(), nil
	case "P-521":
		return elliptic.P521(), nil
	}
	return nil, fmt.Errorf("unsupported EC curve %q", name)
}

// publicJWK renders pub as a public JWK with use=sig, the given alg, and
// kid (or the RFC 7638 thumbprint when kid is empty).
func publicJWK(pub crypto.PublicKey, alg, kid string) (JWK, error) {
	var j JWK
	switch p := pub.(type) {
	case *rsa.PublicKey:
		j = JWK{Kty: "RSA", N: b64(p.N.Bytes()), E: b64(big.NewInt(int64(p.E)).Bytes())}
	case *ecdsa.PublicKey:
		raw, err := p.Bytes() // uncompressed: 0x04 || X || Y, fixed width
		if err != nil {
			return JWK{}, apperr.Newf(400, "INVALID_KEY", "%s", err)
		}
		size := (len(raw) - 1) / 2
		j = JWK{Kty: "EC", Crv: p.Curve.Params().Name, X: b64(raw[1 : 1+size]), Y: b64(raw[1+size:])}
	case ed25519.PublicKey:
		j = JWK{Kty: "OKP", Crv: "Ed25519", X: b64(p)}
	default:
		return JWK{}, apperr.Newf(400, "INVALID_KEY", "unsupported key type %T", pub)
	}
	j.Use, j.Alg = "sig", alg
	j.Kid = cmp.Or(kid, j.Thumbprint())
	return j, nil
}

// privateJWK renders priv as a private JWK (public members plus the
// private ones RFC 7518 §6 defines for its key type).
func privateJWK(priv crypto.Signer, alg, kid string) (JWK, error) {
	j, err := publicJWK(priv.Public(), alg, kid)
	if err != nil {
		return JWK{}, err
	}
	switch p := priv.(type) {
	case *rsa.PrivateKey:
		p.Precompute()
		j.D = b64(p.D.Bytes())
		j.P, j.Q = b64(p.Primes[0].Bytes()), b64(p.Primes[1].Bytes())
		j.DP, j.DQ, j.QI = b64(p.Precomputed.Dp.Bytes()), b64(p.Precomputed.Dq.Bytes()), b64(p.Precomputed.Qinv.Bytes())
	case *ecdsa.PrivateKey:
		d, err := p.Bytes()
		if err != nil {
			return JWK{}, apperr.Newf(400, "INVALID_KEY", "%s", err)
		}
		j.D = b64(d)
	case ed25519.PrivateKey:
		j.D = b64(p.Seed())
	}
	return j, nil
}

// Key converts j to a Go crypto key: a private key (*rsa.PrivateKey,
// *ecdsa.PrivateKey, ed25519.PrivateKey) when its private member "d" is
// present, otherwise the public key. "oct" keys return their raw bytes.
func (j JWK) Key() (any, error) {
	switch j.Kty {
	case "RSA":
		return j.rsaKey()
	case "EC":
		return j.ecKey()
	case "OKP":
		return j.okpKey()
	case "oct":
		k, err := unb64("k", j.K)
		if err != nil {
			return nil, err
		}
		return k, nil
	case "":
		return nil, errors.New(`JWK is missing "kty"`)
	}
	return nil, fmt.Errorf("unsupported JWK kty %q", j.Kty)
}

func (j JWK) rsaKey() (any, error) {
	n, err := unb64Int("n", j.N)
	if err != nil {
		return nil, err
	}
	e, err := unb64Int("e", j.E)
	if err != nil {
		return nil, err
	}
	if !e.IsInt64() || e.Int64() > 1<<31-1 {
		return nil, errors.New(`JWK "e" is too large`)
	}
	pub := rsa.PublicKey{N: n, E: int(e.Int64())}
	if j.D == "" {
		return &pub, nil
	}
	d, err := unb64Int("d", j.D)
	if err != nil {
		return nil, err
	}
	p, err := unb64Int("p", j.P)
	if err != nil {
		return nil, err
	}
	q, err := unb64Int("q", j.Q)
	if err != nil {
		return nil, err
	}
	priv := &rsa.PrivateKey{PublicKey: pub, D: d, Primes: []*big.Int{p, q}}
	if err := priv.Validate(); err != nil {
		return nil, fmt.Errorf("invalid RSA JWK: %w", err)
	}
	priv.Precompute()
	return priv, nil
}

func (j JWK) ecKey() (any, error) {
	curve, err := curveByName(j.Crv)
	if err != nil {
		return nil, err
	}
	x, err := unb64("x", j.X)
	if err != nil {
		return nil, err
	}
	y, err := unb64("y", j.Y)
	if err != nil {
		return nil, err
	}
	if j.D != "" {
		d, err := unb64("d", j.D)
		if err != nil {
			return nil, err
		}
		priv, err := ecdsa.ParseRawPrivateKey(curve, d)
		if err != nil {
			return nil, fmt.Errorf("invalid EC JWK: %w", err)
		}
		return priv, nil
	}
	pub, err := ecdsa.ParseUncompressedPublicKey(curve, slices.Concat([]byte{4}, x, y))
	if err != nil {
		return nil, fmt.Errorf("invalid EC JWK: %w", err)
	}
	return pub, nil
}

func (j JWK) okpKey() (any, error) {
	if j.Crv != "Ed25519" {
		return nil, fmt.Errorf("unsupported OKP curve %q", j.Crv)
	}
	if j.D != "" {
		seed, err := unb64("d", j.D)
		if err != nil {
			return nil, err
		}
		if len(seed) != ed25519.SeedSize {
			return nil, errors.New(`invalid Ed25519 JWK "d" length`)
		}
		return ed25519.NewKeyFromSeed(seed), nil
	}
	x, err := unb64("x", j.X)
	if err != nil {
		return nil, err
	}
	if len(x) != ed25519.PublicKeySize {
		return nil, errors.New(`invalid Ed25519 JWK "x" length`)
	}
	return ed25519.PublicKey(x), nil
}

// Thumbprint is j's RFC 7638 JWK thumbprint: base64url(SHA-256) of the
// key's required public members, serialized with sorted keys and no
// whitespace. Used as the default "kid".
func (j JWK) Thumbprint() string {
	var members map[string]string
	switch j.Kty {
	case "RSA":
		members = map[string]string{"e": j.E, "kty": j.Kty, "n": j.N}
	case "EC":
		members = map[string]string{"crv": j.Crv, "kty": j.Kty, "x": j.X, "y": j.Y}
	case "OKP":
		members = map[string]string{"crv": j.Crv, "kty": j.Kty, "x": j.X}
	case "oct":
		members = map[string]string{"k": j.K, "kty": j.Kty}
	default:
		return ""
	}
	return thumbprint(members)
}

// thumbprint hashes members per RFC 7638 §3: json.Deterministic sorts map
// keys and compact output has no insignificant whitespace, which is exactly
// the required canonical form for these ASCII, string-only members.
func thumbprint(members map[string]string) string {
	canonical, _ := json.Marshal(members, json.Deterministic(true)) // map[string]string cannot fail to marshal
	sum := sha256.Sum256(canonical)
	return b64(sum[:])
}

func encodePEMPair(priv crypto.Signer) (privPEM, pubPEM string, err error) {
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", apperr.Newf(500, "KEYGEN_FAILED", "encode private key: %s", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return "", "", apperr.Newf(500, "KEYGEN_FAILED", "encode public key: %s", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})), nil
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func unb64(member, s string) ([]byte, error) {
	if s == "" {
		return nil, fmt.Errorf("JWK is missing %q", member)
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("JWK %q is not valid base64url: %w", member, err)
	}
	return b, nil
}

func unb64Int(member, s string) (*big.Int, error) {
	b, err := unb64(member, s)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(b), nil
}
