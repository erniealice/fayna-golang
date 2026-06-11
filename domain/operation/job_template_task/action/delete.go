package action

import (
	"context"
	"log"

	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"

	"github.com/erniealice/pyeza-golang/view"
)

// NewDeleteAction creates the job_template_task delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_task", "delete") {
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

		if deps.DeleteJobTemplateTask == nil {
			return view.HTMXError("Delete not available")
		}

		_, err := deps.DeleteJobTemplateTask(ctx, &jobtemplateTaskpb.DeleteJobTemplateTaskRequest{
			Data: &jobtemplateTaskpb.JobTemplateTask{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job template task %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("jt-tasks-table")
	})
}

// NewBulkDeleteAction creates the job_template_task bulk delete action (POST only).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_task", "delete") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}
		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}

		if deps.DeleteJobTemplateTask == nil {
			return view.HTMXError("Delete not available")
		}

		for _, id := range ids {
			_, err := deps.DeleteJobTemplateTask(ctx, &jobtemplateTaskpb.DeleteJobTemplateTaskRequest{
				Data: &jobtemplateTaskpb.JobTemplateTask{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job template task %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("jt-tasks-table")
	})
}
