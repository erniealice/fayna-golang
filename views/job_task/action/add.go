package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtaskform "github.com/erniealice/fayna-golang/views/job_task/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"
)

// NewAddAction creates the job_task add action (GET = drawer form, POST = create).
// The job_phase_id FK is taken from the query string on GET (?job_phase_id=) and
// from the hidden form input on POST. This keeps the task in the context of its
// parent phase.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_task", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			jobPhaseID := viewCtx.Request.URL.Query().Get("job_phase_id")
			const defaultStatus = "TASK_STATUS_PENDING"
			return view.OK("job-task-drawer-form", &jobtaskform.Data{
				FormAction:            deps.Routes.AddURL,
				IsEdit:                false,
				JobPhaseID:            jobPhaseID,
				Status:                defaultStatus,
				StatusOptions:         jobtaskform.BuildTaskStatusOptions(defaultStatus),
				StaffSearchURL:        deps.StaffSearchURL,
				ResourceSearchURL:     deps.ResourceSearchURL,
				TemplateTaskSearchURL: deps.TemplateTaskSearchURL,
				Labels:                deps.Labels,
			})
		}

		// POST — create task
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		stepOrder, _ := strconv.ParseInt(r.FormValue("step_order"), 10, 32)
		plannedQty, _ := strconv.ParseFloat(r.FormValue("planned_quantity"), 64)
		completedQty, _ := strconv.ParseFloat(r.FormValue("completed_quantity"), 64)
		percentComplete, _ := strconv.ParseFloat(r.FormValue("percent_complete"), 64)

		task := &jobtaskpb.JobTask{
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

		resp, err := deps.CreateJobTask(ctx, &jobtaskpb.CreateJobTaskRequest{Data: task})
		if err != nil {
			log.Printf("Failed to create job task: %v", err)
			return view.HTMXError(err.Error())
		}

		newID := ""
		if respData := resp.GetData(); len(respData) > 0 {
			newID = respData[0].GetId()
		}
		if newID != "" {
			return view.ViewResult{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"HX-Trigger":  `{"formSuccess":true}`,
					"HX-Redirect": route.ResolveURL(deps.Routes.DetailURL, "id", newID),
				},
			}
		}

		return view.HTMXSuccess("job-tasks-table")
	})
}
