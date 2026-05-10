package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	fayna "github.com/erniealice/fayna-golang"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobphaseform "github.com/erniealice/fayna-golang/views/job_phase/form"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/view"
)

// NewAddAction creates the job_phase add action (GET = drawer form, POST = create).
// The job_id FK is taken from the query string on GET (?job_id=) and from the
// hidden form input on POST. This keeps the phase in the context of its parent job.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_phase", "create") {
			return fayna.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			jobID := viewCtx.Request.URL.Query().Get("job_id")
			const defaultStatus = "PHASE_STATUS_PENDING"
			return view.OK("job-phase-drawer-form", &jobphaseform.Data{
				FormAction:        deps.Routes.AddURL,
				IsEdit:            false,
				JobID:             jobID,
				Status:            defaultStatus,
				StatusOptions:     jobphaseform.BuildPhaseStatusOptions(defaultStatus),
				ResourceSearchURL: deps.ResourceSearchURL,
				Labels:            deps.Labels,
			})
		}

		// POST — create phase
		if err := viewCtx.Request.ParseForm(); err != nil {
			return fayna.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		phaseOrder, _ := strconv.ParseInt(r.FormValue("phase_order"), 10, 32)
		setupMinutes, _ := strconv.ParseInt(r.FormValue("setup_minutes"), 10, 32)
		runMinutesPerUnit, _ := strconv.ParseFloat(r.FormValue("run_minutes_per_unit"), 64)

		phase := &jobphasepb.JobPhase{
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
		if setupMinutes > 0 {
			i32 := int32(setupMinutes)
			phase.SetupMinutes = &i32
		}
		if runMinutesPerUnit > 0 {
			phase.RunMinutesPerUnit = &runMinutesPerUnit
		}

		resp, err := deps.CreateJobPhase(ctx, &jobphasepb.CreateJobPhaseRequest{Data: phase})
		if err != nil {
			log.Printf("Failed to create job phase: %v", err)
			return fayna.HTMXError(err.Error())
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

		return fayna.HTMXSuccess("job-phases-table")
	})
}
