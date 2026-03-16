package task_outcome

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	outcomepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/task_outcome"

	taskoutcomedetail "github.com/erniealice/fayna-golang/views/task_outcome/detail"
	taskoutcomelist "github.com/erniealice/fayna-golang/views/task_outcome/list"
)

// ModuleDeps holds all dependencies for the task outcome module.
type ModuleDeps struct {
	Routes       fayna.TaskOutcomeRoutes
	Labels       fayna.TaskOutcomeLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Task outcome CRUD
	CreateTaskOutcome func(ctx context.Context, req *outcomepb.CreateTaskOutcomeRequest) (*outcomepb.CreateTaskOutcomeResponse, error)
	ReadTaskOutcome   func(ctx context.Context, req *outcomepb.ReadTaskOutcomeRequest) (*outcomepb.ReadTaskOutcomeResponse, error)
	UpdateTaskOutcome func(ctx context.Context, req *outcomepb.UpdateTaskOutcomeRequest) (*outcomepb.UpdateTaskOutcomeResponse, error)
	DeleteTaskOutcome func(ctx context.Context, req *outcomepb.DeleteTaskOutcomeRequest) (*outcomepb.DeleteTaskOutcomeResponse, error)
	ListTaskOutcomes  func(ctx context.Context, req *outcomepb.ListTaskOutcomesRequest) (*outcomepb.ListTaskOutcomesResponse, error)

	// Outcome criteria read (for linking criteria details)
	ReadOutcomeCriteria func(ctx context.Context, req *criteriapb.ReadOutcomeCriteriaRequest) (*criteriapb.ReadOutcomeCriteriaResponse, error)
}

// Module holds all constructed task outcome views.
type Module struct {
	routes fayna.TaskOutcomeRoutes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewModule creates a new task outcome module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	detailDeps := &taskoutcomedetail.Deps{
		Routes:          deps.Routes,
		Labels:          deps.Labels,
		CommonLabels:    deps.CommonLabels,
		ReadTaskOutcome: deps.ReadTaskOutcome,
	}

	return &Module{
		routes: deps.Routes,
		List: taskoutcomelist.NewView(&taskoutcomelist.Deps{
			Routes:           deps.Routes,
			ListTaskOutcomes: deps.ListTaskOutcomes,
			Labels:           deps.Labels,
			CommonLabels:     deps.CommonLabels,
			TableLabels:      deps.TableLabels,
		}),
		Detail: taskoutcomedetail.NewView(detailDeps),
	}
}

// RegisterRoutes registers all task outcome routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
}
