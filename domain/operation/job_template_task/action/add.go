package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	jobtemplateTaskform "github.com/erniealice/fayna-golang/domain/operation/job_template_task/form"

	"github.com/erniealice/pyeza-golang/view"
)

// NewAddAction creates the job_template_task add action (GET = drawer form, POST = create).
// The job_template_phase_id FK is taken from the query string on GET (?job_template_phase_id=)
// and from the hidden form input on POST. This keeps the task in context of its parent phase.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_task", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			phaseID := viewCtx.Request.URL.Query().Get("job_template_phase_id")
			return view.OK("job-template-task-drawer-form", &jobtemplateTaskform.Data{
				FormAction:         deps.Routes.AddURL,
				IsEdit:             false,
				JobTemplatePhaseID: phaseID,
				ResourceSearchURL:  deps.ResourceSearchURL,
				Labels:             deps.Labels,
			})
		}

		// POST — create template task
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		stepOrder, _ := strconv.ParseInt(r.FormValue("step_order"), 10, 32)
		estDuration, _ := strconv.ParseInt(r.FormValue("estimated_duration_minutes"), 10, 32)

		task := &jobtemplateTaskpb.JobTemplateTask{
			JobTemplatePhaseId: r.FormValue("job_template_phase_id"),
			Name:               r.FormValue("name"),
			StepOrder:          int32(stepOrder),
		}
		if v := r.FormValue("resource_id"); v != "" {
			task.ResourceId = &v
		}
		if estDuration > 0 {
			i32 := int32(estDuration)
			task.EstimatedDurationMinutes = &i32
		}

		if deps.CreateJobTemplateTask == nil {
			return view.HTMXError("Create not available")
		}

		_, err := deps.CreateJobTemplateTask(ctx, &jobtemplateTaskpb.CreateJobTemplateTaskRequest{Data: task})
		if err != nil {
			log.Printf("Failed to create job template task: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("jt-tasks-table")
	})
}
