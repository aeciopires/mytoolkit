package cli

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
	"github.com/aeciopires/mytoolkit/internal/tools/qrcode"
)

func init() {
	rootCmd.AddCommand(newQRCodeCommand())
	registerToolHandler("qrcode", qrCodeHandler)
}

func newQRCodeCommand() *cobra.Command {
	var text, outPath string
	var size int
	cmd := &cobra.Command{
		Use:   "qrcode",
		Short: "Generate a QR code PNG image from text",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outPath == "" || outPath == "-" {
				return apperr.New(400, "OUT_REQUIRED", "--out <file> is required (binary PNG output cannot go to stdout)")
			}
			png, err := qrcode.Generate(text, qrcode.Options{Size: size})
			if err != nil {
				return err
			}
			return os.WriteFile(outPath, png, 0o644)
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "text or URL to encode")
	cmd.Flags().IntVar(&size, "size", 256, "image size in pixels (square)")
	cmd.Flags().StringVar(&outPath, "out", "", "output PNG file (required)")
	return cmd
}

// qrCodeHandler godoc
// @Summary Generate a QR code image
// @Description Generates a QR code PNG from text/URL/unicode content. The one deliberate exception to the app's shared JSON envelope: returns raw image/png bytes directly (so the response can be used as an <img src> or downloaded), not JSON.
// @Tags tools
// @Accept json
// @Produce png
// @Param request body object{input=string,options=object{size=int}} true "Text to encode and optional square size in pixels (default 256)"
// @Success 200 {file} binary "PNG image"
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/qrcode [post]
func qrCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Input   string `json:"input"`
		Options struct {
			Size int `json:"size"`
		} `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}
	png, err := qrcode.Generate(req.Input, qrcode.Options{Size: req.Options.Size})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `inline; filename="qrcode.png"`)
	_, _ = w.Write(png)
}
