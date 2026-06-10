package action

import (
	"context"
	"log"
	"net/http"

	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// NewDeleteAction creates the activity material delete action (POST only).
// ActivityMaterial has no FKs from downstream tables (leaf entity) so delete is
// always safe — no reference checker needed for v1.
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_material", "delete") {
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

		if deps.DeleteActivityMaterial == nil {
			// TODO: wire DeleteActivityMaterial from espyna OperationUseCases.ActivityMaterial
			return view.HTMXError("DeleteActivityMaterial use case not wired — add ActivityMaterial to espyna OperationUseCases")
		}

		_, err := deps.DeleteActivityMaterial(ctx, &activitymaterialpb.DeleteActivityMaterialRequest{
			Data: &activitymaterialpb.ActivityMaterial{ActivityId: activityID},
		})
		if err != nil {
			log.Printf("Failed to delete activity material %s: %v", activityID, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activity-material-charge-section")
	})
}
