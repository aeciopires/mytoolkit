package cli

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
	"github.com/aeciopires/mytoolkit/internal/textio"
	"github.com/aeciopires/mytoolkit/internal/tools/password"
)

func init() {
	rootCmd.AddCommand(newPasswordGenCommand())
	registerToolHandler("password-gen", passwordGenHandler)
}

func newPasswordGenCommand() *cobra.Command {
	var opts password.Options
	var outPath string
	cmd := &cobra.Command{
		Use:   "password-gen",
		Short: "Generate a strong, customizable random password",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := password.Generate(opts)
			if err != nil {
				return err
			}
			return textio.Write(outPath, []byte(out+"\n"))
		},
	}
	cmd.Flags().IntVar(&opts.Length, "length", 16, "password length")
	cmd.Flags().BoolVar(&opts.Lowercase, "lowercase", true, "include lowercase letters")
	cmd.Flags().BoolVar(&opts.Uppercase, "uppercase", true, "include uppercase letters")
	cmd.Flags().BoolVar(&opts.Numbers, "numbers", true, "include numbers")
	cmd.Flags().BoolVar(&opts.Symbols, "symbols", false, "include symbols")
	cmd.Flags().BoolVar(&opts.ExcludeConfusing, "exclude-confusing", false, "exclude i l L 1 o 0 O")
	cmd.Flags().BoolVar(&opts.ExcludeAmbiguous, "exclude-ambiguous", false, "exclude { } [ ] ( ) / \\ ' \" ` ~ , ; : . < >")
	cmd.Flags().StringVar(&outPath, "out", "-", "output file, or - for stdout")
	return cmd
}

// passwordGenHandler godoc
// @Summary Generate a random password
// @Description Generates a cryptographically random password (crypto/rand) from the requested character classes. Unlike other tools, this endpoint ignores the request body's "input" field entirely — only "options" is used.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{options=password.Options} true "Password options"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse "e.g. NO_CHARSET_SELECTED if every character class is disabled or empty after exclusions"
// @Router /api/v1/tools/password-gen [post]
func passwordGenHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req struct {
		Options password.Options `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}
	out, err := password.Generate(req.Options)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteSuccess(w, "password-gen", map[string]string{"output": out}, time.Since(start))
}
