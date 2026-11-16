package cli

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
	"github.com/aeciopires/mytoolkit/internal/textio"
	"github.com/aeciopires/mytoolkit/internal/tools/jwttool"
)

func init() {
	rootCmd.AddCommand(newJWTCommand())
	registerToolHandler("jwt", jwtHandler)
}

var algorithmsHelp = strings.Join(jwttool.SupportedAlgorithms, ", ")

func newJWTCommand() *cobra.Command {
	var decodeFlag, encodeFlag bool
	var token, secret, algorithm, claimsPath, keyPath, outPath string
	var spec claimFlags

	cmd := &cobra.Command{
		Use:   "jwt",
		Short: "Decode/verify, sign, and generate keys or a JWKS for JWTs — fully offline",
		Long: `Decode/verify (--decode) or sign (--encode) a JWT, or use the keygen/jwks
subcommands to create signing keys and a publishable JWKS. Everything runs
offline: no key, token or JWKS is ever fetched from or sent to the network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := readKeyFlag(keyPath)
			if err != nil {
				return err
			}
			switch {
			case decodeFlag && encodeFlag:
				return apperr.New(400, "INVALID_OPTION", "--decode and --encode are mutually exclusive")
			case decodeFlag:
				result, err := jwttool.Decode(token, secret, key)
				if err != nil {
					return err
				}
				return writeJSON(outPath, result, false)
			case encodeFlag:
				raw, err := readClaimsInput(cmd, claimsPath)
				if err != nil {
					return err
				}
				out, err := encodeToken(raw, spec.toSpec(), jwttool.EncodeOptions{
					Algorithm: algorithm, Secret: secret, Key: key, KeyID: spec.kid,
				})
				if err != nil {
					return err
				}
				return textio.Write(outPath, []byte(out+"\n"))
			}
			return apperr.New(400, "MISSING_MODE", "one of --decode or --encode (or a keygen/jwks subcommand) is required")
		},
	}
	f := cmd.Flags()
	f.BoolVar(&decodeFlag, "decode", false, "decode (and, with --secret/--key, verify) a token")
	f.BoolVar(&encodeFlag, "encode", false, "sign claims into a token")
	f.StringVar(&token, "token", "", "token to decode (--decode mode)")
	f.StringVar(&secret, "secret", "", "HMAC shared secret (HS256/HS384/HS512 only): verification for --decode, signing for --encode")
	f.StringVar(&algorithm, "algorithm", jwttool.DefaultAlgorithm, "signing algorithm for --encode: "+algorithmsHelp)
	f.StringVar(&claimsPath, "claims", "-", "base claims JSON file, or - for stdin (--encode mode); stdin is skipped when it is a terminal")
	f.StringVar(&keyPath, "key", "", "key file, or - for stdin (RS*/PS*/ES*/EdDSA only): private key (PEM or JWK) for --encode; public key, certificate, JWK or JWKS for --decode verification")
	f.StringVar(&outPath, "out", "-", "output file, or - for stdout")
	spec.register(cmd)

	cmd.AddCommand(newJWTKeygenCommand(), newJWTJWKSCommand())
	return cmd
}

// claimFlags are --encode's claim-builder flags (see jwttool.ClaimsSpec).
type claimFlags struct {
	custom, audience               []string
	issuer, subject, kid, exp, nbf string
	iat, jti                       bool
}

func (c *claimFlags) register(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringArrayVar(&c.custom, "claim", nil, "custom claim name=value (repeatable); repeating a name, or name[]=value, makes a list of strings")
	f.StringVar(&c.issuer, "iss", "", "issuer (iss) claim")
	f.StringVar(&c.subject, "sub", "", "subject (sub) claim")
	f.StringArrayVar(&c.audience, "aud", nil, "audience (aud) claim (repeatable; several values make a list)")
	f.StringVar(&c.exp, "exp", "", "expire after this duration (sets exp), e.g. 15m, 1h, 7d")
	f.StringVar(&c.nbf, "nbf", "", "not valid before now+duration (sets nbf), e.g. 0s, -30s")
	f.BoolVar(&c.iat, "iat", false, "set issued-at (iat) to now")
	f.BoolVar(&c.jti, "jti", false, "set a random JWT ID (jti)")
	f.StringVar(&c.kid, "kid", "", "key ID (kid) header; defaults to the JWK's kid when --key is a private JWK")
}

func (c *claimFlags) toSpec() claimsInput {
	return claimsInput{
		Custom: c.custom, Issuer: c.issuer, Subject: c.subject, Audience: c.audience,
		ExpiresIn: c.exp, NotBefore: c.nbf, IssuedAt: c.iat, JWTID: c.jti,
	}
}

// claimsInput is the transport-level (string durations) form of
// jwttool.ClaimsSpec, shared by the CLI flags and the REST options.
type claimsInput struct {
	Custom, Audience     []string
	Issuer, Subject      string
	ExpiresIn, NotBefore string
	IssuedAt, JWTID      bool
}

// encodeToken parses the base claims JSON, applies in, and signs.
func encodeToken(rawClaims []byte, in claimsInput, opts jwttool.EncodeOptions) (string, error) {
	base, err := jwttool.ParseClaimsJSON(rawClaims)
	if err != nil {
		return "", err
	}
	exp, err := jwttool.ParseDuration("expires_in", in.ExpiresIn)
	if err != nil {
		return "", err
	}
	nbf, err := jwttool.ParseDuration("not_before", in.NotBefore)
	if err != nil {
		return "", err
	}
	claims, err := jwttool.BuildClaims(base, jwttool.ClaimsSpec{
		Issuer: in.Issuer, Subject: in.Subject, Audience: in.Audience,
		ExpiresIn: exp, NotBefore: nbf, IssuedAt: in.IssuedAt, GenerateID: in.JWTID,
		Custom: in.Custom,
	}, time.Now())
	if err != nil {
		return "", err
	}
	return jwttool.Encode(claims, opts)
}

// readClaimsInput reads --claims. With the default "-" it skips an
// interactive terminal stdin, which would otherwise block forever waiting
// for input the user never meant to type (e.g. when every claim comes from
// --sub/--claim flags); with no claims at all Encode then reports
// EMPTY_CLAIMS instead of hanging.
func readClaimsInput(cmd *cobra.Command, path string) ([]byte, error) {
	if (path == "-" || path == "") && !cmd.Flags().Changed("claims") && stdinIsTerminal() {
		return nil, nil
	}
	return textio.Read(path)
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func newJWTKeygenCommand() *cobra.Command {
	var algorithm, kid, outPath, privOut, pubOut, jwksOut string
	var bits int
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate a signing key offline: HMAC secret, or RSA/EC/Ed25519 key pair with JWK/JWKS",
		RunE: func(cmd *cobra.Command, args []string) error {
			kp, err := jwttool.GenerateKey(jwttool.KeyOptions{Algorithm: algorithm, Bits: bits, KeyID: kid})
			if err != nil {
				return err
			}
			writeFiles := privOut != "" || pubOut != "" || jwksOut != ""
			if privOut != "" {
				priv := kp.PrivateKey
				if priv == "" { // HS*: the secret is the private material
					priv = kp.Secret + "\n"
				}
				if err := textio.WritePrivate(privOut, []byte(priv)); err != nil {
					return err
				}
			}
			if pubOut != "" {
				if kp.PublicKey == "" {
					return apperr.New(400, "INVALID_OPTION", "--public-out is not available for HMAC algorithms (a shared secret has no public key)")
				}
				if err := textio.Write(pubOut, []byte(kp.PublicKey)); err != nil {
					return err
				}
			}
			if jwksOut != "" {
				if kp.JWKS == nil {
					return apperr.New(400, "INVALID_OPTION", "--jwks-out is not available for HMAC algorithms (a shared secret must never be published)")
				}
				if err := writeJSON(jwksOut, kp.JWKS, true); err != nil {
					return err
				}
			}
			if writeFiles && !cmd.Flags().Changed("out") {
				return nil
			}
			return writeJSON(outPath, kp, true)
		},
	}
	f := cmd.Flags()
	f.StringVar(&algorithm, "algorithm", jwttool.DefaultAlgorithm, "algorithm the key is for: "+algorithmsHelp)
	f.IntVar(&bits, "bits", jwttool.DefaultRSABits, "RSA modulus size for RS*/PS*: 2048, 3072 or 4096")
	f.StringVar(&kid, "kid", "", "key ID (default: RFC 7638 JWK thumbprint)")
	f.StringVar(&outPath, "out", "-", "full JSON result (PEM, JWKs, JWKS), or - for stdout; skipped when a --*-out file flag is used unless set explicitly")
	f.StringVar(&privOut, "private-out", "", "write the private key PEM (or HMAC secret) to this file, mode 0600")
	f.StringVar(&pubOut, "public-out", "", "write the public key PEM to this file")
	f.StringVar(&jwksOut, "jwks-out", "", "write the public JWKS to this file")
	return cmd
}

func newJWTJWKSCommand() *cobra.Command {
	var algorithm, kid, keyPath, outPath string
	cmd := &cobra.Command{
		Use:   "jwks",
		Short: "Build a publishable JWKS (public keys only) from PEM keys/certificates or a JWK",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := textio.Read(keyPath)
			if err != nil {
				return err
			}
			set, err := jwttool.BuildJWKS(string(raw), jwttool.JWKSOptions{Algorithm: algorithm, KeyID: kid})
			if err != nil {
				return err
			}
			return writeJSON(outPath, set, true)
		},
	}
	f := cmd.Flags()
	f.StringVar(&keyPath, "key", "-", "PEM file with one or more keys/certificates (public or private), a JWK or a JWKS, or - for stdin")
	f.StringVar(&algorithm, "algorithm", "", "alg to stamp on every key (default: inferred from the key type)")
	f.StringVar(&kid, "kid", "", "key ID, single-key input only (default: RFC 7638 JWK thumbprint)")
	f.StringVar(&outPath, "out", "-", "output file, or - for stdout")
	return cmd
}

func writeJSON(path string, v any, indent bool) error {
	var (
		out []byte
		err error
	)
	if indent {
		out, err = json.MarshalIndent(v, "", "  ")
	} else {
		out, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	return textio.Write(path, append(out, '\n'))
}

// readKeyFlag reads --key's file (or stdin, for "-") if set; an unset
// --key (the default "") means "no key provided" rather than "read
// stdin", so it must be distinguished from textio.Read's own "-" meaning.
func readKeyFlag(keyPath string) (string, error) {
	if keyPath == "" {
		return "", nil
	}
	raw, err := textio.Read(keyPath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// flexList is a REST option that accepts either a JSON array of strings or
// a single string, which split turns into items — so the web UI (whose
// form controls always send strings) and API clients (which naturally send
// arrays) share one field.
type flexList struct {
	items []string
	split func(string) []string
}

func (l *flexList) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &l.items); err == nil {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	l.items = l.split(s)
	return nil
}

type jwtRequest struct {
	Input   string `json:"input"`
	Options struct {
		Mode      string   `json:"mode"`
		Secret    string   `json:"secret"`
		Key       string   `json:"key"`
		Algorithm string   `json:"algorithm"`
		KeyID     string   `json:"kid"`
		Bits      int      `json:"bits"`
		Issuer    string   `json:"issuer"`
		Subject   string   `json:"subject"`
		Audience  flexList `json:"audience"`
		ExpiresIn string   `json:"expires_in"`
		NotBefore string   `json:"not_before"`
		IssuedAt  bool     `json:"issued_at"`
		JWTID     bool     `json:"jwt_id"`
		Claims    flexList `json:"claims"`
	} `json:"options"`
}

// jwtHandler godoc
// @Summary Decode/verify, encode (sign), generate keys, or build a JWKS for JWTs
// @Description All modes run offline (no key, token or JWKS is fetched over the network).
// @Description mode=decode (default): parses "input" (a token) without requiring a key; supplying "secret" (HMAC) or "key" (PEM public key/certificate, JWK, or JWKS — matched by the token's kid) additionally verifies the signature and exp/nbf/iat, setting "valid" (and "verify_error" on failure).
// @Description mode=encode: signs "input" (optional JSON claims object) plus the registered-claim options (issuer, subject, audience, expires_in, not_before, issued_at, jwt_id) and "claims" (custom name=value pairs — array, or a newline-separated string; a repeated name or name[]=value yields a list of strings) using "algorithm" (default HS256) with "secret" (HMAC) or "key" (PEM or private JWK); "kid" sets the header.
// @Description mode=keygen: generates a secret (HS*) or key pair (RS*/PS* with "bits", ES*, EdDSA) as PEM, private/public JWK and JWKS.
// @Description mode=jwks: converts "input" (PEM keys/certificates, JWK or JWKS) into a JWKS of public keys.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{mode=string,secret=string,key=string,algorithm=string,kid=string,bits=int,issuer=string,subject=string,audience=[]string,expires_in=string,not_before=string,issued_at=bool,jwt_id=bool,claims=[]string}} true "Token (decode), claims JSON (encode), key material (jwks), or empty (keygen)"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse "e.g. INVALID_TOKEN, INVALID_CLAIMS, UNSUPPORTED_ALGORITHM, INVALID_KEY, MISSING_SECRET, EMPTY_CLAIMS, INVALID_OPTION"
// @Router /api/v1/tools/jwt [post]
func jwtHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	req := jwtRequest{}
	req.Options.Audience.split = jwttool.SplitList
	req.Options.Claims.split = jwttool.SplitLines
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}
	o := req.Options

	var (
		data any
		err  error
	)
	switch o.Mode {
	case "", "decode":
		data, err = jwttool.Decode(req.Input, o.Secret, o.Key)
	case "encode":
		var token string
		token, err = encodeToken([]byte(req.Input), claimsInput{
			Custom: o.Claims.items, Issuer: o.Issuer, Subject: o.Subject, Audience: o.Audience.items,
			ExpiresIn: o.ExpiresIn, NotBefore: o.NotBefore, IssuedAt: o.IssuedAt, JWTID: o.JWTID,
		}, jwttool.EncodeOptions{Algorithm: o.Algorithm, Secret: o.Secret, Key: o.Key, KeyID: o.KeyID})
		data = map[string]string{"output": token}
	case "keygen":
		data, err = jwttool.GenerateKey(jwttool.KeyOptions{Algorithm: o.Algorithm, Bits: o.Bits, KeyID: o.KeyID})
	case "jwks":
		data, err = jwttool.BuildJWKS(req.Input, jwttool.JWKSOptions{Algorithm: o.Algorithm, KeyID: o.KeyID})
	default:
		err = apperr.OneOf("mode", o.Mode, "decode", "encode", "keygen", "jwks")
	}
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteSuccess(w, "jwt", data, time.Since(start))
}
