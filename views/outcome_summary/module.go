package outcome_summary

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobsumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_outcome_summary"
	phasesumpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/phase_outcome_summary"

	jobsummary "github.com/erniealice/fayna-golang/views/outcome_summary/job_summary"
	summarylist "github.com/erniealice/fayna-golang/views/outcome_summary/list"
	phasesummary "github.com/erniealice/fayna-golang/views/outcome_summary/phase_summary"
)

// ModuleDeps holds all dependencies for the outcome summary module.
type ModuleDeps struct {
	Routes       fayna.OutcomeSummaryRoutes
	Labels       fayna.OutcomeSummaryLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Job outcome summary operations
	GetJobOutcomeSummaryByJob func(ctx context.Context, req *jobsumpb.GetJobOutcomeSummaryByJobRequest) (*jobsumpb.GetJobOutcomeSummaryByJobResponse, error)
	ListJobOutcomeSummarys    func(ctx context.Context, req *jobsumpb.ListJobOutcomeSummarysRequest) (*jobsumpb.ListJobOutcomeSummarysResponse, error)

	// Phase outcome summary operations
	GetPhaseOutcomeSummaryByJobPhase func(ctx context.Context, req *phasesumpb.GetPhaseOutcomeSummaryByJobPhaseRequest) (*phasesumpb.GetPhaseOutcomeSummaryByJobPhaseResponse, error)
	ListPhaseOutcomeSummarysByJob    func(ctx context.Context, req *phasesumpb.ListPhaseOutcomeSummarysByJobRequest) (*phasesumpb.ListPhaseOutcomeSummarysByJobResponse, error)
}

// Module holds all constructed outcome summary views.
type Module struct {
	routes       fayna.OutcomeSummaryRoutes
	List         view.View
	JobSummary   view.View
	PhaseSummary view.View
}

// NewModule creates a new outcome summary module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	return &Module{
		routes: deps.Routes,
		List: summarylist.NewView(&summarylist.ListViewDeps{
			Routes:                 deps.Routes,
			ListJobOutcomeSummarys: deps.ListJobOutcomeSummarys,
			Labels:                 deps.Labels,
			CommonLabels:           deps.CommonLabels,
			TableLabels:            deps.TableLabels,
		}),
		JobSummary: jobsummary.NewView(&jobsummary.Deps{
			Routes:                    deps.Routes,
			Labels:                    deps.Labels,
			CommonLabels:              deps.CommonLabels,
			GetJobOutcomeSummaryByJob: deps.GetJobOutcomeSummaryByJob,
		}),
		PhaseSummary: phasesummary.NewView(&phasesummary.Deps{
			Routes:                           deps.Routes,
			Labels:                           deps.Labels,
			CommonLabels:                     deps.CommonLabels,
			GetPhaseOutcomeSummaryByJobPhase: deps.GetPhaseOutcomeSummaryByJobPhase,
		}),
	}
}

// RegisterRoutes registers all outcome summary routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.JobSummaryURL, m.JobSummary)
	r.GET(m.routes.PhaseSummaryURL, m.PhaseSummary)
}
