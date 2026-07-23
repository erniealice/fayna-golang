package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	jobtemplateTaskform "github.com/erniealice/fayna-golang/domain/operation/job_template_task/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"
)

// NewEditAction creates the job_template_task edit action (GET = drawer form pre-filled, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_task", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadJobTemplateTask == nil {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			resp, err := deps.ReadJobTemplateTask(ctx, &jobtemplateTaskpb.ReadJobTemplateTaskRequest{
				Data: &jobtemplateTaskpb.JobTemplateTask{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job template task %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := resp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			t := data[0]

			resourceID := ""
			if t.ResourceId != nil {
				resourceID = *t.ResourceId
			}
			estDuration := int32(0)
			if t.EstimatedDurationMinutes != nil {
				estDuration = *t.EstimatedDurationMinutes
			}

			return view.OK("job-template-task-drawer-form", &jobtemplateTaskform.Data{
				FormAction:               route.ResolveURL(deps.Routes.EditURL, "id", id),
				IsEdit:                   true,
				ID:                       t.GetId(),
				JobTemplatePhaseID:       t.GetJobTemplatePhaseId(),
				Name:                     t.GetName(),
				Code:                     t.GetCode(),
				StepOrder:                t.GetStepOrder(),
				EstimatedDurationMinutes: estDuration,
				ResourceID:               resourceID,
				ResourceSearchURL:        deps.ResourceSearchURL,
				Labels:                   deps.Labels,
			})
		}

		// POST — update template task
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		stepOrder, _ := strconv.ParseInt(r.FormValue("step_order"), 10, 32)
		estDuration, _ := strconv.ParseInt(r.FormValue("estimated_duration_minutes"), 10, 32)

		task := &jobtemplateTaskpb.JobTemplateTask{
			Id:                 id,
			JobTemplatePhaseId: r.FormValue("job_template_phase_id"),
			Name:               r.FormValue("name"),
			StepOrder:          int32(stepOrder),
		}
		if v := r.FormValue("code"); v != "" {
			task.Code = &v
		}
		if v := r.FormValue("resource_id"); v != "" {
			task.ResourceId = &v
		}
		if estDuration > 0 {
			i32 := int32(estDuration)
			task.EstimatedDurationMinutes = &i32
		}

		if deps.UpdateJobTemplateTask == nil {
			return view.HTMXError("Update not available")
		}

		_, err := deps.UpdateJobTemplateTask(ctx, &jobtemplateTaskpb.UpdateJobTemplateTaskRequest{Data: task})
		if err != nil {
			log.Printf("Failed to update job template task %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("jt-tasks-table")
	})
}
