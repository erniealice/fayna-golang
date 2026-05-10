package action

import (
	"context"
	"log"

	fayna "github.com/erniealice/fayna-golang"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"

	"github.com/erniealice/pyeza-golang/view"
)

// NewDeleteAction creates the job_task delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}

		if deps.DeleteJobTask == nil {
			return fayna.HTMXError("Delete not available")
		}

		_, err := deps.DeleteJobTask(ctx, &jobtaskpb.DeleteJobTaskRequest{
			Data: &jobtaskpb.JobTask{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job task %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("job-tasks-table")
	})
}

// NewBulkDeleteAction creates the job_task bulk delete action (POST only).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}
		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}

		if deps.DeleteJobTask == nil {
			return fayna.HTMXError("Delete not available")
		}

		for _, id := range ids {
			_, err := deps.DeleteJobTask(ctx, &jobtaskpb.DeleteJobTaskRequest{
				Data: &jobtaskpb.JobTask{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job task %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("job-tasks-table")
	})
}
