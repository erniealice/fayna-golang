package action

import (
	"context"
	"log"

	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// NewBulkDeleteAction creates the job activity bulk delete action (POST).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteJobActivity(ctx, &jobactivitypb.DeleteJobActivityRequest{
				Data: &jobactivitypb.JobActivity{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job activity %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("activities-table")
	})
}

// NewDeleteAction creates the job activity delete action (POST).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_activity", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}

		id := viewCtx.Request.FormValue("id")
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteJobActivity(ctx, &jobactivitypb.DeleteJobActivityRequest{
			Data: &jobactivitypb.JobActivity{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job activity %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("activities-table")
	})
}
