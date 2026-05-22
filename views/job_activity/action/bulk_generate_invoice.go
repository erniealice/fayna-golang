package action

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"
)

// NewBulkGenerateInvoiceAction creates the bulk generate invoice action (POST).
// It receives a list of selected activity IDs via multipart form-data and calls
// GenerateInvoiceFromActivities, then redirects to the new revenue detail page.
func NewBulkGenerateInvoiceAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P1: codex caught this handler ungated.
		// The mutation produces invoices, so the most specific gate is
		// invoice:create. Also require job_activity:post since the use case
		// posts the activities to flip them from DRAFT → POSTED before
		// invoicing.
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("invoice", "create") || !perms.Can("job_activity", "post") {
			return fayna.HTMXError("You do not have permission to generate invoices from activities")
		}

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

		revenueDetailBase := deps.RevenueDetailURLTemplate
		if revenueDetailBase == "" {
			revenueDetailBase = "/app/revenue/detail/"
		} else {
			revenueDetailBase = strings.Split(revenueDetailBase, "{id}")[0]
		}
		redirectURL := revenueDetailBase + revenueID + "?tab=items"
		return view.ViewResult{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"HX-Redirect": redirectURL,
				"HX-Trigger":  `{"formSuccess":true}`,
			},
		}
	})
}

