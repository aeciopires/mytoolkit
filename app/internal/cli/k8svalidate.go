package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/response"
	"github.com/aeciopires/mytoolkit/internal/textio"
	"github.com/aeciopires/mytoolkit/internal/tools/k8svalidate"
)

func init() {
	rootCmd.AddCommand(newK8sValidateCommand())
	registerToolHandler("k8s-validate", k8sValidateHandler)
}

func newK8sValidateCommand() *cobra.Command {
	var inPath, outPath string
	cmd := &cobra.Command{
		Use:   "k8s-validate",
		Short: "Validate that a YAML document is well-formed and Kubernetes-API-shaped",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := textio.Read(inPath)
			if err != nil {
				return err
			}
			result, err := k8svalidate.Validate(input, k8svalidate.Options{})
			if err != nil {
				return err
			}
			if err := textio.Write(outPath, []byte(formatReport(result))); err != nil {
				return err
			}
			if !result.Valid {
				invalid := 0
				for _, d := range result.Documents {
					if !d.Valid {
						invalid++
					}
				}
				return fmt.Errorf("%d of %d document(s) failed Kubernetes validation", invalid, len(result.Documents))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "-", "input file, or - for stdin")
	cmd.Flags().StringVar(&outPath, "out", "-", "output file, or - for stdout")
	return cmd
}

func formatReport(result k8svalidate.Result) string {
	var b strings.Builder
	valid := 0
	for _, d := range result.Documents {
		var detail []string
		if d.APIVersion != "" {
			detail = append(detail, "apiVersion="+d.APIVersion)
		}
		if d.Kind != "" {
			detail = append(detail, "kind="+d.Kind)
		}
		if d.Name != "" {
			detail = append(detail, "name="+d.Name)
		}
		suffix := ""
		if len(detail) > 0 {
			suffix = " (" + strings.Join(detail, ", ") + ")"
		}
		if d.Valid {
			valid++
			fmt.Fprintf(&b, "Document %d: VALID%s\n", d.Index, suffix)
		} else {
			fmt.Fprintf(&b, "Document %d: INVALID: %s%s\n", d.Index, d.Error, suffix)
		}
	}
	fmt.Fprintf(&b, "\n%d/%d document(s) valid.\n", valid, len(result.Documents))
	return b.String()
}

// k8sValidateHandler godoc
// @Summary Validate Kubernetes-shaped YAML
// @Description Validates a YAML document (or a "---"-separated multi-document stream) against the two fields every Kubernetes API object requires: non-empty apiVersion and kind, plus an object-shaped metadata if present. Does not validate against any specific resource's full schema. HTTP 200 with "data.valid":false means the manifests parsed but aren't valid; only a hard YAML syntax error returns a non-2xx status.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string} true "YAML document or multi-document stream"
// @Success 200 {object} object{success=bool,data=k8svalidate.Result,meta=ToolMeta}
// @Failure 400 {object} ToolErrorResponse "e.g. INVALID_YAML (syntax error), NO_DOCUMENTS"
// @Router /api/v1/tools/k8s-validate [post]
func k8sValidateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, apperr.New(http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request body"))
		return
	}
	result, err := k8svalidate.Validate([]byte(req.Input), k8svalidate.Options{})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteSuccess(w, "k8s-validate", result, time.Since(start))
}
