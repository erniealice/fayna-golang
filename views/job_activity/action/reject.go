package action

import (
	"context"
	"log"

	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// NewRejectAction creates the reject-activity action (POST).
func NewRejectAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "approve") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.RejectActivity(ctx, &jobactivitypb.RejectJobActivityRequest{
			ActivityId: id,
			RejectedBy: viewCtx.Request.FormValue("rejected_by"),
			Reason:     viewCtx.Request.FormValue("reason"),
		})
		if err != nil {
			log.Printf("Failed to reject job activity %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activities-table")
	})
}
