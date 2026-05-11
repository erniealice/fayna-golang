package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	fayna "github.com/erniealice/fayna-golang"
	jobtaskform "github.com/erniealice/fayna-golang/views/job_task/form"

	"github.com/erniealice/pyeza-golang/view"
)

// NewEditAction creates the job_task edit action (GET = drawer form pre-filled, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadJobTask == nil {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			resp, err := deps.ReadJobTask(ctx, &jobtaskpb.ReadJobTaskRequest{
				Data: &jobtaskpb.JobTask{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job task %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := resp.GetData()
			if len(data) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			t := data[0]
			statusStr := taskStatusString(t.GetStatus())
			// Convert to proto enum string for the select Value
			statusEnum := "TASK_STATUS_PENDING"
			switch statusStr {
			case "in_progress":
				statusEnum = "TASK_STATUS_IN_PROGRESS"
			case "completed":
				statusEnum = "TASK_STATUS_COMPLETED"
			case "skipped":
				statusEnum = "TASK_STATUS_SKIPPED"
			case "hold":
				statusEnum = "TASK_STATUS_HOLD"
			case "rework":
				statusEnum = "TASK_STATUS_REWORK"
			}

			assignedTo := ""
			if t.AssignedTo != nil {
				assignedTo = *t.AssignedTo
			}
			resourceID := ""
			if t.ResourceId != nil {
				resourceID = *t.ResourceId
			}
			templateTaskID := ""
			if t.TemplateTaskId != nil {
				templateTaskID = *t.TemplateTaskId
			}
			plannedQty := float64(0)
			if t.PlannedQuantity != nil {
				plannedQty = *t.PlannedQuantity
			}
			completedQty := float64(0)
			if t.CompletedQuantity != nil {
				completedQty = *t.CompletedQuantity
			}
			percentComplete := float64(0)
			if t.PercentComplete != nil {
				percentComplete = *t.PercentComplete
			}
			allowParallel := false
			if t.AllowParallel != nil {
				allowParallel = *t.AllowParallel
			}

			return view.OK("job-task-drawer-form", &jobtaskform.Data{
				FormAction:            deps.Routes.EditURL,
				IsEdit:                true,
				ID:                    t.GetId(),
				JobPhaseID:            t.GetJobPhaseId(),
				Name:                  t.GetName(),
				StepOrder:             t.GetStepOrder(),
				Status:                statusEnum,
				StatusOptions:         jobtaskform.BuildTaskStatusOptions(statusEnum),
				IsAdHoc:               t.GetIsAdHoc(),
				AssignedTo:            assignedTo,
				ResourceID:            resourceID,
				TemplateTaskID:        templateTaskID,
				PlannedQuantity:       plannedQty,
				CompletedQuantity:     completedQty,
				PercentComplete:       percentComplete,
				AllowParallel:         allowParallel,
				ActualStart:           t.GetActualStartString(),
				ActualEnd:             t.GetActualEndString(),
				StaffSearchURL:        deps.StaffSearchURL,
				ResourceSearchURL:     deps.ResourceSearchURL,
				TemplateTaskSearchURL: deps.TemplateTaskSearchURL,
				Labels:                deps.Labels,
			})
		}

		// POST — update task
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		stepOrder, _ := strconv.ParseInt(r.FormValue("step_order"), 10, 32)
		plannedQty, _ := strconv.ParseFloat(r.FormValue("planned_quantity"), 64)
		completedQty, _ := strconv.ParseFloat(r.FormValue("completed_quantity"), 64)
		percentComplete, _ := strconv.ParseFloat(r.FormValue("percent_complete"), 64)

		task := &jobtaskpb.JobTask{
			Id:         id,
			JobPhaseId: r.FormValue("job_phase_id"),
			Name:       r.FormValue("name"),
			StepOrder:  int32(stepOrder),
			Status:     taskStatusToEnum(r.FormValue("status")),
			IsAdHoc:    r.FormValue("is_ad_hoc") == "true" || r.FormValue("is_ad_hoc") == "1",
		}
		if v := r.FormValue("assigned_to"); v != "" {
			task.AssignedTo = &v
		}
		if v := r.FormValue("resource_id"); v != "" {
			task.ResourceId = &v
		}
		if v := r.FormValue("template_task_id"); v != "" {
			task.TemplateTaskId = &v
		}
		if v := r.FormValue("actual_start"); v != "" {
			task.ActualStartString = &v
		}
		if v := r.FormValue("actual_end"); v != "" {
			task.ActualEndString = &v
		}
		if plannedQty > 0 {
			task.PlannedQuantity = &plannedQty
		}
		if completedQty > 0 {
			task.CompletedQuantity = &completedQty
		}
		if percentComplete > 0 {
			task.PercentComplete = &percentComplete
		}
		allowParallel := r.FormValue("allow_parallel") == "true" || r.FormValue("allow_parallel") == "1"
		task.AllowParallel = &allowParallel

		_, err := deps.UpdateJobTask(ctx, &jobtaskpb.UpdateJobTaskRequest{Data: task})
		if err != nil {
			log.Printf("Failed to update job task %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("job-tasks-table")
	})
}
