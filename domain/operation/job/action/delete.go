package action

import (
	"context"
	"log"

	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// NewDeleteAction creates the job delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}

		_, err := deps.DeleteJob(ctx, &jobpb.DeleteJobRequest{
			Data: &jobpb.Job{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("jobs-table")
	})
}

// NewBulkDeleteAction creates the job bulk delete action (POST only).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError(deps.Labels.Errors.NoIDs)
		}

		for _, id := range ids {
			_, err := deps.DeleteJob(ctx, &jobpb.DeleteJobRequest{
				Data: &jobpb.Job{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("jobs-table")
	})
}
