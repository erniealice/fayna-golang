package action

import (
	"context"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// NewReverseAction creates the reverse-activity action (POST).
func NewReverseAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "manage") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.ReverseActivity(ctx, &jobactivitypb.ReverseJobActivityRequest{
			ActivityId: id,
			ReversedBy: viewCtx.Request.FormValue("reversed_by"),
			Reason:     viewCtx.Request.FormValue("reason"),
		})
		if err != nil {
			log.Printf("Failed to reverse job activity %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("activities-table")
	})
}
