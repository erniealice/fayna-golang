package action

import (
	"context"
	"log"
	"net/http"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// NewDeleteAction creates the activity labor delete action (POST only).
// ActivityLabor has no FKs from downstream tables (leaf entity) so delete is
// always safe — no reference checker needed for v1.
// The activity_id is read from the form field "activity_id" or from the
// "id" field (legacy path value shape) for consistency with other modules.
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_labor", "delete") {
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

		if deps.DeleteActivityLabor == nil {
			// TODO: wire DeleteActivityLabor from espyna OperationUseCases.ActivityLabor
			return fayna.HTMXError("DeleteActivityLabor use case not wired — add ActivityLabor to espyna OperationUseCases")
		}

		_, err := deps.DeleteActivityLabor(ctx, &activitylaborpb.DeleteActivityLaborRequest{
			Data: &activitylaborpb.ActivityLabor{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to delete activity labor %s: %v", activityID, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activity-labor-charge-section")
	})
}
