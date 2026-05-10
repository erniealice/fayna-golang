package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatephaseform "github.com/erniealice/fayna-golang/views/job_template_phase/form"

	"github.com/erniealice/pyeza-golang/view"
)

// NewEditAction creates the job_template_phase edit action (GET = drawer form pre-filled, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_phase", "update") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadJobTemplatePhase == nil {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			resp, err := deps.ReadJobTemplatePhase(ctx, &jobtemplatephasepb.ReadJobTemplatePhaseRequest{
				Data: &jobtemplatephasepb.JobTemplatePhase{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job template phase %s: %v", id, err)
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := resp.GetData()
			if len(data) == 0 {
				return fayna.HTMXError(deps.Labels.Errors.NotFound)
			}
			p := data[0]

			resourceID := ""
			if p.ResourceId != nil {
				resourceID = *p.ResourceId
			}
			predecessorID := ""
			if p.PredecessorTemplatePhaseId != nil {
				predecessorID = *p.PredecessorTemplatePhaseId
			}
			estDuration := int32(0)
			if p.SetupMinutes != nil {
				estDuration = *p.SetupMinutes
			}

			return view.OK("job-template-phase-drawer-form", &jobtemplatephaseform.Data{
				FormAction:                 deps.Routes.EditURL,
				IsEdit:                     true,
				ID:                         p.GetId(),
				JobTemplateID:              p.GetJobTemplateId(),
				Name:                       p.GetName(),
				PhaseOrder:                 p.GetPhaseOrder(),
				EstimatedDurationMinutes:   estDuration,
				ResourceID:                 resourceID,
				PredecessorTemplatePhaseID: predecessorID,
				ResourceSearchURL:          deps.ResourceSearchURL,
				Labels:                     deps.Labels,
			})
		}

		// POST — update template phase
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		phaseOrder, _ := strconv.ParseInt(r.FormValue("phase_order"), 10, 32)
		estDuration, _ := strconv.ParseInt(r.FormValue("estimated_duration_minutes"), 10, 32)

		phase := &jobtemplatephasepb.JobTemplatePhase{
			Id:            id,
			JobTemplateId: r.FormValue("job_template_id"),
			Name:          r.FormValue("name"),
			PhaseOrder:    int32(phaseOrder),
		}
		if v := r.FormValue("resource_id"); v != "" {
			phase.ResourceId = &v
		}
		if estDuration > 0 {
			i32 := int32(estDuration)
			phase.SetupMinutes = &i32
		}
		if v := r.FormValue("predecessor_template_phase_id"); v != "" {
			phase.PredecessorTemplatePhaseId = &v
		}

		if deps.UpdateJobTemplatePhase == nil {
			return fayna.HTMXError("Update not available")
		}

		_, err := deps.UpdateJobTemplatePhase(ctx, &jobtemplatephasepb.UpdateJobTemplatePhaseRequest{Data: phase})
		if err != nil {
			log.Printf("Failed to update job template phase %s: %v", id, err)
			return fayna.HTMXError(err.Error())
		}

		return fayna.HTMXSuccess("job-template-phases-table")
	})
}
