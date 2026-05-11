package action

import (
	"context"
	"log"
	"net/http"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/view"
)

// NewSetStatusAction creates the job_task status update action (POST only).
//
// Reads `id` and `status` from the query string (form fallback). Accepts the
// proto enum name (TASK_STATUS_PENDING / TASK_STATUS_IN_PROGRESS /
// TASK_STATUS_COMPLETED / TASK_STATUS_SKIPPED / TASK_STATUS_HOLD /
// TASK_STATUS_REWORK) and lowercase shorthands.
//
// Returns HX-Redirect to the task detail page so the status badge refreshes.
func NewSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.URL.Query().Get("id")
		targetStatus := viewCtx.Request.URL.Query().Get("status")
		if id == "" || targetStatus == "" {
			_ = viewCtx.Request.ParseForm()
			if id == "" {
				id = viewCtx.Request.FormValue("id")
			}
			if targetStatus == "" {
				targetStatus = viewCtx.Request.FormValue("status")
			}
		}
		if id == "" {
			return fayna.HTMXError(deps.Labels.Errors.IDRequired)
		}
		if targetStatus == "" {
			return fayna.HTMXError("Status is required")
		}

		statusEnum := taskStatusToEnum(targetStatus)
		if statusEnum == jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED {
			return fayna.HTMXError("Invalid task status")
		}

		// Read existing task first — UpdateJobTask requires Name and JobPhaseId in
		// the request payload (espyna validation). Fetching also verifies existence.
		jobPhaseID := ""
		taskName := ""
		if deps.ReadJobTask != nil {
			readResp, err := deps.ReadJobTask(ctx, &jobtaskpb.ReadJobTaskRequest{
				Data: &jobtaskpb.JobTask{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job task %s: %v", id, err)
				return fayna.HTMXError(err.Error())
			}
			data := readResp.GetData()
			if len(data) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			jobPhaseID = data[0].GetJobPhaseId()
			taskName = data[0].GetName()
		}

		if deps.UpdateJobTask == nil {
			return fayna.HTMXError("Task update not available")
		}

		_, err := deps.UpdateJobTask(ctx, &jobtaskpb.UpdateJobTaskRequest{
			Data: &jobtaskpb.JobTask{
				Id:         id,
				JobPhaseId: jobPhaseID,
				Name:       taskName,
				Status:     statusEnum,
			},
		})
		if err != nil {
			log.Printf("Failed to update task %s status to %s: %v", id, targetStatus, err)
			return fayna.HTMXError(err.Error())
		}

		// Redirect to the task detail page so the status badge refreshes.
		// Fall back to a 204 + trigger when the detail URL is unavailable.
		if deps.Routes.DetailURL != "" {
			return view.ViewResult{
				StatusCode: http.StatusNoContent,
				Headers: map[string]string{
					"HX-Redirect": deps.Routes.DetailURL + "?id=" + id,
				},
			}
		}
		return view.ViewResult{
			StatusCode: http.StatusNoContent,
			Headers:    map[string]string{"HX-Trigger": `{"jobTaskStatusChanged":true}`},
		}
	})
}

// NewBulkSetStatusAction creates the job_task bulk set-status action (POST only).
// Accepts multiple `id` fields + a single `target_status` field.
func NewBulkSetStatusAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}
		ids := viewCtx.Request.Form["id"]
		targetStatus := viewCtx.Request.FormValue("target_status")

		if len(ids) == 0 {
			return fayna.HTMXError("No IDs provided")
		}
		if targetStatus == "" {
			return fayna.HTMXError("Target status is required")
		}

		statusEnum := taskStatusToEnum(targetStatus)

		for _, id := range ids {
			// Read first to get Name + JobPhaseId (required by espyna UpdateJobTask).
			jobPhaseID := ""
			taskName := ""
			if deps.ReadJobTask != nil {
				readResp, err := deps.ReadJobTask(ctx, &jobtaskpb.ReadJobTaskRequest{
					Data: &jobtaskpb.JobTask{Id: id},
				})
				if err != nil {
					log.Printf("Bulk set-status: failed to read task %s: %v", id, err)
					continue
				}
				if data := readResp.GetData(); len(data) > 0 {
					jobPhaseID = data[0].GetJobPhaseId()
					taskName = data[0].GetName()
				}
			}

			if deps.UpdateJobTask == nil {
				continue
			}
			_, err := deps.UpdateJobTask(ctx, &jobtaskpb.UpdateJobTaskRequest{
				Data: &jobtaskpb.JobTask{
					Id:         id,
					JobPhaseId: jobPhaseID,
					Name:       taskName,
					Status:     statusEnum,
				},
			})
			if err != nil {
				log.Printf("Bulk set-status: failed to update task %s: %v", id, err)
			}
		}

		return fayna.HTMXSuccess("job-tasks-table")
	})
}
