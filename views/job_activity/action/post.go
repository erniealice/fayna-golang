package action

import (
	"context"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// NewPostAction creates the post-activity action (POST).
func NewPostAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "post") && !perms.Can("job_activity", "manage") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.PostActivity(ctx, &jobactivitypb.PostJobActivityRequest{
			ActivityId: id,
			PostedBy:   viewCtx.Request.FormValue("posted_by"),
		})
		if err != nil {
			log.Printf("Failed to post job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}
