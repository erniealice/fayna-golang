package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

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
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method != http.MethodPost {
			return fayna.HTMXError("Method not allowed")
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		r := viewCtx.Request
		// Support both "activity_id" and legacy "id" form fields.
		activityID := r.FormValue("activity_id")
		if activityID == "" {
			activityID = r.FormValue("id")
		}
		if activityID == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if deps.DeleteActivityExpense == nil {
			// TODO: wire DeleteActivityExpense from espyna OperationUseCases.ActivityExpense
			return fayna.HTMXError("DeleteActivityExpense use case not wired — add ActivityExpense to espyna OperationUseCases")
		}

		_, err := deps.DeleteActivityExpense(ctx, &activityexpensepb.DeleteActivityExpenseRequest{
			Data: &activityexpensepb.ActivityExpense{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to delete activity expense %s: %v", activityID, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activity-expense-charge-section")
	})
}
