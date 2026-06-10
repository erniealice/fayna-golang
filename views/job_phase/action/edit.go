package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobphaseform "github.com/erniealice/fayna-golang/views/job_phase/form"

	"github.com/erniealice/pyeza-golang/view"
)

// NewEditAction creates the job_phase edit action (GET = drawer form pre-filled, POST = update).
func NewEditAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_phase", "update") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		id := viewCtx.Request.PathValue("id")

		if viewCtx.Request.Method == http.MethodGet {
			if deps.ReadJobPhase == nil {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			resp, err := deps.ReadJobPhase(ctx, &jobphasepb.ReadJobPhaseRequest{
				Data: &jobphasepb.JobPhase{Id: id},
			})
			if err != nil {
				log.Printf("Failed to read job phase %s: %v", id, err)
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			data := resp.GetData()
			if len(data) == 0 {
				return view.HTMXError(deps.Labels.Errors.NotFound)
			}
			p := data[0]
			statusStr := phaseStatusString(p.GetStatus()) // lowercase shorthand
			// Convert to proto enum string for the select Value
			statusEnum := "PHASE_STATUS_PENDING"
			switch statusStr {
			case "active":
				statusEnum = "PHASE_STATUS_ACTIVE"
			case "completed":
				statusEnum = "PHASE_STATUS_COMPLETED"
			}

			resourceID := ""
			if p.ResourceId != nil {
				resourceID = *p.ResourceId
			}
			templatePhaseID := ""
			if p.TemplatePhaseId != nil {
				templatePhaseID = *p.TemplatePhaseId
			}
			predecessorPhaseID := ""
			if p.PredecessorPhaseId != nil {
				predecessorPhaseID = *p.PredecessorPhaseId
			}
			setupMinutes := int32(0)
			if p.SetupMinutes != nil {
				setupMinutes = *p.SetupMinutes
			}
			runMinutesPerUnit := float64(0)
			if p.RunMinutesPerUnit != nil {
				runMinutesPerUnit = *p.RunMinutesPerUnit
			}

			return view.OK("job-phase-drawer-form", &jobphaseform.Data{
				FormAction:         deps.Routes.EditURL,
				IsEdit:             true,
				ID:                 p.GetId(),
				JobID:              p.GetJobId(),
				Name:               p.GetName(),
				PhaseOrder:         p.GetPhaseOrder(),
				Status:             statusEnum,
				StatusOptions:      jobphaseform.BuildPhaseStatusOptions(statusEnum),
				TemplatePhaseID:    templatePhaseID,
				ResourceID:         resourceID,
				PredecessorPhaseID: predecessorPhaseID,
				PlannedStart:       p.GetPlannedStartString(),
				PlannedEnd:         p.GetPlannedEndString(),
				ActualStart:        p.GetActualStartString(),
				ActualEnd:          p.GetActualEndString(),
				SetupMinutes:       setupMinutes,
				RunMinutesPerUnit:  runMinutesPerUnit,
				ResourceSearchURL:  deps.ResourceSearchURL,
				Labels:             deps.Labels,
			})
		}

		// POST — update phase
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		phaseOrder, _ := strconv.ParseInt(r.FormValue("phase_order"), 10, 32)
		setupMinutes, _ := strconv.ParseInt(r.FormValue("setup_minutes"), 10, 32)
		runMinutesPerUnit, _ := strconv.ParseFloat(r.FormValue("run_minutes_per_unit"), 64)

		phase := &jobphasepb.JobPhase{
			Id:         id,
			JobId:      r.FormValue("job_id"),
			Name:       r.FormValue("name"),
			PhaseOrder: int32(phaseOrder),
			Status:     phaseStatusToEnum(r.FormValue("status")),
		}
		if v := r.FormValue("template_phase_id"); v != "" {
			phase.TemplatePhaseId = &v
		}
		if v := r.FormValue("resource_id"); v != "" {
			phase.ResourceId = &v
		}
		if v := r.FormValue("predecessor_phase_id"); v != "" {
			phase.PredecessorPhaseId = &v
		}
		if v := r.FormValue("actual_start"); v != "" {
			phase.ActualStartString = &v
		}
		if v := r.FormValue("actual_end"); v != "" {
			phase.ActualEndString = &v
		}
		if setupMinutes > 0 {
			i32 := int32(setupMinutes)
			phase.SetupMinutes = &i32
		}
		if runMinutesPerUnit > 0 {
			phase.RunMinutesPerUnit = &runMinutesPerUnit
		}

		_, err := deps.UpdateJobPhase(ctx, &jobphasepb.UpdateJobPhaseRequest{Data: phase})
		if err != nil {
			log.Printf("Failed to update job phase %s: %v", id, err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("job-phases-table")
	})
}
