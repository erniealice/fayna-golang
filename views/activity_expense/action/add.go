package action

import (
	"context"
	"log"
	"net/http"

	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// NewAddAction creates the activity expense create action (GET = form, POST = create).
// The activity_id is sourced from the ?activity_id query param on GET and
// round-tripped as a hidden input on POST. ActivityExpense.activity_id is the PK
// (1:1 with JobActivity) — there is no separate ID generation.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_expense", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			activityID := viewCtx.Request.URL.Query().Get("activity_id")
			formData := buildEmptyFormData(activityID, deps.Routes, deps.Labels)
			formData.FormAction = addFormAction(deps.Routes)
			formData.CommonLabels = nil // injected by ViewAdapter
			return view.OK("activity-expense-drawer-form", formData)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		activityID := r.FormValue("activity_id")
		if activityID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if deps.CreateActivityExpense == nil {
			// TODO: wire CreateActivityExpense from espyna OperationUseCases.ActivityExpense
			// when the use case is added. For now return a clear gap error.
			return view.HTMXError("CreateActivityExpense use case not wired — add ActivityExpense to espyna OperationUseCases")
		}

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

		_, err := deps.CreateActivityExpense(ctx, &activityexpensepb.CreateActivityExpenseRequest{
			Data: record,
		})
		if err != nil {
			log.Printf("Failed to create activity expense for activity %s: %v", activityID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activity-expense-charge-section")
	})
}
