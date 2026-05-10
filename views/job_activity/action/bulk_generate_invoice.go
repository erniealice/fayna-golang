package action

import (
	"context"
	"fmt"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"
)

// NewBulkGenerateInvoiceAction creates the bulk generate invoice action (POST).
// It receives a list of selected activity IDs via multipart form-data and calls
// GenerateInvoiceFromActivities, then redirects to the new revenue detail page.
func NewBulkGenerateInvoiceAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		r := viewCtx.Request
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			// Fall back to regular form parse (non-multipart submissions)
			if err2 := r.ParseForm(); err2 != nil {
				return fayna.HTMXError("Invalid form data")
			}
		}

		ids := r.Form["id"]
		if len(ids) == 0 {
			return fayna.HTMXError("No activities selected")
		}

		if deps.GenerateInvoiceFromActivities == nil {
			return fayna.HTMXError("Invoice generation not available")
		}

		revenueID, err := deps.GenerateInvoiceFromActivities(ctx, ids, "", "", "PHP", "")
		if err != nil {
			log.Printf("Failed to generate invoice from activities: %v", err)
			return fayna.HTMXError(fmt.Sprintf("Failed to generate invoice: %v", err))
		}

		redirectURL := fmt.Sprintf("/app/revenue/detail/%s?tab=items", revenueID)
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Redirect": redirectURL,
				"HX-Trigger":  `{"formSuccess":true}`,
			},
		}
	})
}
