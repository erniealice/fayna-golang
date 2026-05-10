package action

import (
	"context"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
)

// NewDeleteAction creates the job template delete action (POST only).
func NewDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template", "delete") {
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

		_, err := deps.DeleteJobTemplate(ctx, &jobtemplatepb.DeleteJobTemplateRequest{
			Data: &jobtemplatepb.JobTemplate{Id: id},
		})
		if err != nil {
			log.Printf("Failed to delete job template %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("job-templates-table")
	})
}

// NewBulkDeleteAction creates the job template bulk delete action (POST only).
func NewBulkDeleteAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template", "delete") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}

		for _, id := range ids {
			_, err := deps.DeleteJobTemplate(ctx, &jobtemplatepb.DeleteJobTemplateRequest{
				Data: &jobtemplatepb.JobTemplate{Id: id},
			})
			if err != nil {
				log.Printf("Failed to delete job template %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("job-templates-table")
	})
}
