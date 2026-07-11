package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
	"github.com/aeciopires/mytoolkit/internal/textio"
	"github.com/aeciopires/mytoolkit/internal/tools/textcount"
)

func init() {
	rootCmd.AddCommand(newTextCountCommand())
	registerToolHandler("text-count", textCountHandler)
}

func newTextCountCommand() *cobra.Command {
	var inPath string
	var outPath string
	cmd := &cobra.Command{
		Use:   "text-count",
		Short: "Count characters, words, and lines in text",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := textio.Read(inPath)
			if err != nil {
				return err
			}
			counts, err := textcount.Count(input, textcount.Options{})
			if err != nil {
				return err
			}
			out := fmt.Sprintf("characters: %d\ncharacters_no_spaces: %d\nwords: %d\nlines: %d\n",
				counts.Characters, counts.CharactersNoSpaces, counts.Words, counts.Lines)
			return textio.Write(outPath, []byte(out))
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "-", "input file, or - for stdin")
	cmd.Flags().StringVar(&outPath, "out", "-", "output file, or - for stdout")
	return cmd
}

// textCountHandler godoc
// @Summary Count characters, words, and lines
// @Description Counts characters (Unicode-aware), characters excluding whitespace, words, and lines in the input text. Never errors — empty/whitespace input is valid and returns all-zero counts.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string} true "Text to count"
// @Success 200 {object} object{success=bool,data=textcount.Counts,meta=ToolMeta}
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/text-count [post]
func textCountHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}
	counts, err := textcount.Count([]byte(req.Input), textcount.Options{})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteSuccess(w, "text-count", counts, time.Since(start))
}
