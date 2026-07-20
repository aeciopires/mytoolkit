package mcp

import (
	"errors"
	"fmt"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

// toolErr formats an error returned by an internal/tools/<name> function
// for an MCP tool result. The SDK automatically converts any non-nil error
// returned by a tool handler into CallToolResult{IsError: true, ...} (see
// mcp.ToolHandlerFor), so handlers only need to return a well-formatted
// error, not build the result envelope themselves. *apperr.Error's own
// Error() method returns just the message (no code), so the code is
// prefixed back on here — it's the same information REST callers get from
// the JSON error envelope's "code" field.
func toolErr(err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		return fmt.Errorf("%s: %s", appErr.Code, appErr.Message)
	}
	return err
}
