package action

import (
	"context"
	"log"
	"net/http"
	"strconv"

	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatephaseform "github.com/erniealice/fayna-golang/domain/operation/job_template_phase/form"

	"github.com/erniealice/pyeza-golang/view"
)

// NewAddAction creates the job_template_phase add action (GET = drawer form, POST = create).
// The job_template_id FK is taken from the query string on GET (?job_template_id=) and from
// the hidden form input on POST. This keeps the phase in context of its parent template.
func NewAddAction(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("job_template_phase", "create") {
			return view.HTMXError(deps.Labels.Errors.PermissionDenied)
		}

		if viewCtx.Request.Method == http.MethodGet {
			jobTemplateID := viewCtx.Request.URL.Query().Get("job_template_id")
			return view.OK("job-template-phase-drawer-form", &jobtemplatephaseform.Data{
				FormAction:           deps.Routes.AddURL,
				IsEdit:               false,
				JobTemplateID:        jobTemplateID,
				ResourceSearchURL:    deps.ResourceSearchURL,
				Labels:               deps.Labels,
				ScoringSchemeOptions: jobtemplatephaseform.BuildScoringSchemeOptions(ctx, deps.ListScoringSchemes, ""),
			})
		}

		// POST — create template phase
		if err := viewCtx.Request.ParseForm(); err != nil {
			return view.HTMXError("Invalid form data")
		}
		r := viewCtx.Request

		phaseOrder, _ := strconv.ParseInt(r.FormValue("phase_order"), 10, 32)
		estDuration, _ := strconv.ParseInt(r.FormValue("estimated_duration_minutes"), 10, 32)

		phase := &jobtemplatephasepb.JobTemplatePhase{
			JobTemplateId: r.FormValue("job_template_id"),
			Name:          r.FormValue("name"),
			PhaseOrder:    int32(phaseOrder),
		}
		if v := r.FormValue("code"); v != "" {
			phase.Code = &v
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
		if v := r.FormValue("scoring_scheme_id"); v != "" {
			phase.ScoringSchemeId = &v
		}

		if deps.CreateJobTemplatePhase == nil {
			return view.HTMXError("Create not available")
		}

		_, err := deps.CreateJobTemplatePhase(ctx, &jobtemplatephasepb.CreateJobTemplatePhaseRequest{Data: phase})
		if err != nil {
			log.Printf("Failed to create job template phase: %v", err)
			return view.HTMXError(err.Error())
		}

		return view.HTMXSuccess("job-template-phases-table")
	})
}
