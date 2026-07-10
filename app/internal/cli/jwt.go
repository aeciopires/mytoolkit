package cli

import (
	"encoding/json"
	"net/http"
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

func newJWTCommand() *cobra.Command {
	var decodeFlag, encodeFlag bool
	var token, secret, algorithm, claimsPath, keyPath, outPath string

	cmd := &cobra.Command{
		Use:   "jwt",
		Short: "Decode or encode a JWT for inspection and testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := readKeyFlag(keyPath)
			if err != nil {
				return err
			}
			if decodeFlag {
				result, err := jwttool.Decode(token, secret, key)
				if err != nil {
					return err
				}
				out, err := json.Marshal(result)
				if err != nil {
					return err
				}
				return textio.Write(outPath, append(out, '\n'))
			}
			if encodeFlag {
				raw, err := textio.Read(claimsPath)
				if err != nil {
					return err
				}
				claims, err := jwttool.ParseClaimsJSON(raw)
				if err != nil {
					return err
				}
				out, err := jwttool.Encode(claims, secret, key, algorithm)
				if err != nil {
					return err
				}
				return textio.Write(outPath, []byte(out+"\n"))
			}
			return apperr.New(400, "MISSING_MODE", "one of --decode or --encode is required")
		},
	}
	cmd.Flags().BoolVar(&decodeFlag, "decode", false, "decode a token")
	cmd.Flags().BoolVar(&encodeFlag, "encode", false, "encode claims into a signed token")
	cmd.Flags().StringVar(&token, "token", "", "token to decode (--decode mode)")
	cmd.Flags().StringVar(&secret, "secret", "", "HMAC shared secret (HS256/HS384/HS512 only): verification for --decode, signing for --encode")
	cmd.Flags().StringVar(&algorithm, "algorithm", jwttool.DefaultAlgorithm, "signing algorithm for --encode: "+strings.Join(jwttool.SupportedAlgorithms, ", "))
	cmd.Flags().StringVar(&claimsPath, "claims", "-", "claims JSON file, or - for stdin (--encode mode)")
	cmd.Flags().StringVar(&keyPath, "key", "", "path to a PEM key file, or - for stdin (RS*/PS*/ES*/EdDSA only): private key for --encode, public key for --decode verification")
	cmd.Flags().StringVar(&outPath, "out", "-", "output file, or - for stdout")
	return cmd
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

func jwtHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req struct {
		Input   string `json:"input"`
		Options struct {
			Mode      string `json:"mode"`
			Secret    string `json:"secret"`
			Key       string `json:"key"`
			Algorithm string `json:"algorithm"`
		} `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}

	switch req.Options.Mode {
	case "encode":
		claims, err := jwttool.ParseClaimsJSON([]byte(req.Input))
		if err != nil {
			response.WriteError(w, err)
			return
		}
		out, err := jwttool.Encode(claims, req.Options.Secret, req.Options.Key, req.Options.Algorithm)
		if err != nil {
			response.WriteError(w, err)
			return
		}
		response.WriteSuccess(w, "jwt", map[string]string{"output": out}, time.Since(start))
	default:
		result, err := jwttool.Decode(req.Input, req.Options.Secret, req.Options.Key)
		if err != nil {
			response.WriteError(w, err)
			return
		}
		response.WriteSuccess(w, "jwt", result, time.Since(start))
	}
}
