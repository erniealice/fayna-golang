package action

import (
	"context"
	"log"
	"net/http"

	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// NewDeleteAction creates the activity expense delete action (POST only).
// ActivityExpense has no FKs from downstream tables (leaf entity) so delete is
// always safe — no reference checker needed for v1.
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_expense", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method != http.MethodPost {
			return view.HTMXError("Method not allowed")
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		// Support both "activity_id" and legacy "id" form fields.
		activityID := r.FormValue("activity_id")
		if activityID == "" {
			activityID = r.FormValue("id")
		}
		if activityID == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if deps.DeleteActivityExpense == nil {
			// TODO: wire DeleteActivityExpense from espyna OperationUseCases.ActivityExpense
			return view.HTMXError("DeleteActivityExpense use case not wired — add ActivityExpense to espyna OperationUseCases")
		}

		_, err := deps.DeleteActivityExpense(ctx, &activityexpensepb.DeleteActivityExpenseRequest{
			Data: &activityexpensepb.ActivityExpense{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to delete activity expense %s: %v", activityID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activity-expense-charge-section")
	})
}
