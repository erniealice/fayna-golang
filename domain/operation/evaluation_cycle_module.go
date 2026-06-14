package operation

import (
	"context"

	cyclepkg "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle"
	cycledetail "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle/detail"
	cyclelist "github.com/erniealice/fayna-golang/domain/operation/evaluation_cycle/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	cyclepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle"
	memberpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_cycle_member"
	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
)

// EvaluationCycleModuleDeps holds all dependencies for the evaluation_cycle module.
//
// The use-case closures are INJECTED by the Integrator / block — the views never
// call espyna directly. OpenEvaluationCycle / CloseEvaluationCycle map to the
// espyna evaluation_cycle/orchestration.go OpenUseCase / CloseUseCase. The
// X-of-Y banner prefers GetCycleProgress (espyna read-UC, view-delta F-GATE.2);
// when nil it computes inline from ListEvaluationCycleMembers + ListEvaluations.
type EvaluationCycleModuleDeps struct {
	Routes       cyclepkg.Routes
	Labels       cyclepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateEvaluationCycle func(ctx context.Context, req *cyclepb.CreateEvaluationCycleRequest) (*cyclepb.CreateEvaluationCycleResponse, error)
	ReadEvaluationCycle   func(ctx context.Context, req *cyclepb.ReadEvaluationCycleRequest) (*cyclepb.ReadEvaluationCycleResponse, error)
	ListEvaluationCycles  func(ctx context.Context, req *cyclepb.ListEvaluationCyclesRequest) (*cyclepb.ListEvaluationCyclesResponse, error)
	OpenEvaluationCycle   func(ctx context.Context, req *cyclepb.UpdateEvaluationCycleRequest) (*cyclepb.UpdateEvaluationCycleResponse, error)
	CloseEvaluationCycle  func(ctx context.Context, req *cyclepb.UpdateEvaluationCycleRequest) (*cyclepb.UpdateEvaluationCycleResponse, error)

	ListEvaluationCycleMembers func(ctx context.Context, req *memberpb.ListEvaluationCycleMembersRequest) (*memberpb.ListEvaluationCycleMembersResponse, error)

	GetCycleProgress func(ctx context.Context, cycleID string) (*cyclepkg.CycleProgress, error)
	ListEvaluations  func(ctx context.Context, req *evalpb.ListEvaluationsRequest) (*evalpb.ListEvaluationsResponse, error)

	NewID func() string
}

// EvaluationCycleModule holds all constructed evaluation_cycle views.
type EvaluationCycleModule struct {
	routes     cyclepkg.Routes
	List       view.View
	Table      view.View
	Detail     view.View
	MembersTab view.View
	Add        view.View
	Open       view.View
	Close      view.View
}

// NewEvaluationCycleModule creates a new evaluation_cycle module with all views wired.
func NewEvaluationCycleModule(deps *EvaluationCycleModuleDeps) *EvaluationCycleModule {
	listDeps := &cyclelist.ListViewDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		ListEvaluationCycles: deps.ListEvaluationCycles,
		GetCycleProgress:     deps.GetCycleProgress,
	}

	detailDeps := &cycledetail.DetailViewDeps{
		Routes:                     deps.Routes,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
		ReadEvaluationCycle:        deps.ReadEvaluationCycle,
		ListEvaluationCycleMembers: deps.ListEvaluationCycleMembers,
		GetCycleProgress:           deps.GetCycleProgress,
		ListEvaluations:            deps.ListEvaluations,
	}

	actionDeps := &cyclepkg.ModuleDeps{
		Routes:                     deps.Routes,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
		CreateEvaluationCycle:      deps.CreateEvaluationCycle,
		ReadEvaluationCycle:        deps.ReadEvaluationCycle,
		ListEvaluationCycles:       deps.ListEvaluationCycles,
		OpenEvaluationCycle:        deps.OpenEvaluationCycle,
		CloseEvaluationCycle:       deps.CloseEvaluationCycle,
		ListEvaluationCycleMembers: deps.ListEvaluationCycleMembers,
		GetCycleProgress:           deps.GetCycleProgress,
		ListEvaluations:            deps.ListEvaluations,
		NewID:                      deps.NewID,
	}

	return &EvaluationCycleModule{
		routes:     deps.Routes,
		List:       cyclelist.NewView(listDeps),
		Table:      cyclelist.NewTableView(listDeps),
		Detail:     cycledetail.NewView(detailDeps),
		MembersTab: cycledetail.NewMembersTabAction(detailDeps),
		Add:        cyclepkg.NewAddAction(actionDeps),
		Open:       cyclepkg.NewOpenAction(actionDeps),
		Close:      cyclepkg.NewCloseAction(actionDeps),
	}
}

// RegisterRoutes registers all evaluation_cycle routes.
func (m *EvaluationCycleModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TableURL, m.Table)
	r.GET(m.routes.MembersTabURL, m.MembersTab)

	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.POST(m.routes.OpenURL, m.Open)
	r.POST(m.routes.CloseURL, m.Close)
}
