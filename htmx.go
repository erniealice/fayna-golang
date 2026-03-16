package fayna

import (
	"fmt"
	"net/http"

	"github.com/erniealice/pyeza-golang/view"
)

// HTMXSuccess returns a header-only response that signals the sheet to close
// and the table to refresh. The ViewAdapter handles header-only responses
// (no template, just headers + status code).
func HTMXSuccess(tableID string) view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"HX-Trigger": fmt.Sprintf(`{"formSuccess":true,"refreshTable":"%s"}`, tableID),
		},
	}
}

// HTMXError returns a header-only response that signals a form error.
// The sheet.js handleResponse reads HX-Error-Message on non-2xx responses.
func HTMXError(message string) view.ViewResult {
	return view.ViewResult{
		StatusCode: http.StatusUnprocessableEntity,
		Headers: map[string]string{
			"HX-Error-Message": message,
		},
	}
}
