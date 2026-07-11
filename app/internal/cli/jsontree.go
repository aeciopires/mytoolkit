package cli

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
	"github.com/aeciopires/mytoolkit/internal/textio"
	"github.com/aeciopires/mytoolkit/internal/tools/jsontree"
)

func init() {
	rootCmd.AddCommand(newJSONTreeCommand())
	registerToolHandler("json-tree", jsonTreeHandler)
}

func newJSONTreeCommand() *cobra.Command {
	var inPath, outPath string
	cmd := &cobra.Command{
		Use:   "json-tree",
		Short: "Parse JSON into a navigable tree structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := textio.Read(inPath)
			if err != nil {
				return err
			}
			node, err := jsontree.Parse(input, jsontree.Options{})
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(node, "", "  ")
			if err != nil {
				return err
			}
			return textio.Write(outPath, append(out, '\n'))
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "-", "input file, or - for stdin")
	cmd.Flags().StringVar(&outPath, "out", "-", "output file, or - for stdout")
	return cmd
}

// jsonTreeHandler godoc
// @Summary Parse JSON into a navigable tree
// @Description Parses raw JSON text into a key-order-preserving tree structure. Error messages include the exact 1-indexed line/column of any syntax problem. Trailing content after a complete JSON value is rejected, not silently ignored.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string} true "Raw JSON to parse"
// @Success 200 {object} object{success=bool,data=object{tree=jsontree.Node},meta=ToolMeta}
// @Failure 400 {object} ToolErrorResponse "e.g. INVALID_JSON with a position suffix like (at line 4, column 1)"
// @Router /api/v1/tools/json-tree [post]
func jsonTreeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}
	node, err := jsontree.Parse([]byte(req.Input), jsontree.Options{})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteSuccess(w, "json-tree", map[string]any{"tree": node}, time.Since(start))
}
