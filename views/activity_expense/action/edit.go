package action

import (
	"context"
	"fmt"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// NewEditAction creates the activity expense update action (GET = form, POST = update).
// The {id} path value is the activity_id (PK of ActivityExpense = FK to JobActivity).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_expense", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		activityID := viewCtx.Request.PathValue("id")
		if activityID == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadActivityExpense == nil {
				// Stub: render an empty form pre-filled with just the activity_id.
				formData := buildEmptyFormData(activityID, deps.Routes, deps.Labels)
				formData.IsEdit = true
				formData.FormAction = editFormAction(deps.Routes, activityID)
				formData.CommonLabels = nil
				return view.OK("activity-expense-drawer-form", formData)
			}

			resp, err := deps.ReadActivityExpense(ctx, &activityexpensepb.ReadActivityExpenseRequest{
				Data: &activityexpensepb.ActivityExpense{ActivityId: activityID},
			})
			if err != nil {
				log.Printf("Failed to read activity expense %s: %v", activityID, err)
				return fayna.HTMXError(fmt.Sprintf("Failed to load expense record: %v", err))
			}
			data := resp.GetData()
			if len(data) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}

			formData := buildFormData(data[0], deps.Routes, deps.Labels)
			formData.FormAction = editFormAction(deps.Routes, activityID)
			formData.CommonLabels = nil
			return view.OK("activity-expense-drawer-form", formData)
		}

		// POST — process the update.
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		if deps.UpdateActivityExpense == nil {
			// TODO: wire UpdateActivityExpense from espyna OperationUseCases.ActivityExpense
			return fayna.HTMXError("UpdateActivityExpense use case not wired — add ActivityExpense to espyna OperationUseCases")
		}

		r := viewCtx.Request
		markupPct := parseFormFloat(r.FormValue("markup_pct_override"))
		expenseCategoryID := r.FormValue("expense_category_id")
		vendorRef := r.FormValue("vendor_ref")
		receiptURL := r.FormValue("receipt_url")
		paymentMethod := r.FormValue("payment_method")
		record := &activityexpensepb.ActivityExpense{
			ActivityId:        activityID,
			ExpenseCategoryId: &expenseCategoryID,
			VendorRef:         &vendorRef,
			ReceiptUrl:        &receiptURL,
			PaymentMethod:     &paymentMethod,
			MarkupPctOverride: &markupPct,
		}

		_, err := deps.UpdateActivityExpense(ctx, &activityexpensepb.UpdateActivityExpenseRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to update activity expense %s: %v", activityID, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activity-expense-charge-section")
	})
}
