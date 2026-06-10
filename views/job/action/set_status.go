package action

import (
	"context"
	"log"

	"github.com/erniealice/pyeza-golang/view"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
)

// NewSetStatusAction creates the job status update action (POST only).
func NewSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		targetStatus := viewCtx.Request.URL.Query().Get("status")

		if id == "" {
			_ = viewCtx.Request.ParseForm()
			id = viewCtx.Request.FormValue("id")
			targetStatus = viewCtx.Request.FormValue("status")
		}
		if id == "" {
			return view.HTMXError(deps.Labels.Errors.IDRequired)
		}
		if targetStatus == "" {
			return view.HTMXError(deps.Labels.Errors.StatusRequired)
		}

		statusEnum := jobStatusToEnum(targetStatus)

		_, err := deps.UpdateJob(ctx, &jobpb.UpdateJobRequest{
			Data: &jobpb.Job{Id: id, Status: statusEnum},
		})
		if err != nil {
			log.Printf("Failed to update job status %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("jobs-table")
	})
}

// NewBulkSetStatusAction creates the job bulk status update action (POST only).
func NewBulkSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		_ = viewCtx.Request.ParseMultipartForm(32 << 20)

		ids := viewCtx.Request.Form["id"]
		targetStatus := viewCtx.Request.FormValue("target_status")

		if len(ids) == 0 {
			return view.HTMXError("No IDs provided")
		}
		if targetStatus == "" {
			return view.HTMXError("Target status is required")
		}

		statusEnum := jobStatusToEnum(targetStatus)

		for _, id := range ids {
			if _, err := deps.UpdateJob(ctx, &jobpb.UpdateJobRequest{
				Data: &jobpb.Job{Id: id, Status: statusEnum},
			}); err != nil {
				log.Printf("Failed to update job status %s: %v", id, err)
			}
		}

		return view.HTMXSuccess("jobs-table")
	})
}
