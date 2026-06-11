package action

import (
	"context"
	"log"

	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// NewSubmitAction creates the submit-for-approval action (POST).
func NewSubmitAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.SubmitForApproval(ctx, &jobactivitypb.SubmitForApprovalRequest{
			ActivityId: id,
		})
		if err != nil {
			log.Printf("Failed to submit job activity %s for approval: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activities-table")
	})
}
