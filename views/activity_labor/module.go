// Package activity_labor provides the ActivityLabor view module (Shape 2 sibling).
//
// ActivityLabor is the charge-detail record for ENTRY_TYPE_LABOR job activities.
// PK = activity_id (not a separate id) — it is the FK to JobActivity (1:1 relationship).
//
// Access patterns:
//   - Primary: JobActivity detail page's charge tab (entry_type=LABOR) — drawer form.
//   - Deep link: /app/activity-labor/{id} — direct URL from the charge tab Edit CTA.
//   - Debug/power-user: /app/activity-labor/list — NOT in the sidebar.
//
// This module is the canonical Shape 2 sibling template for activity_material and
// activity_expense (wave 3). Keep it clean and copy-paste-ready.
package activity_labor

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"
	activitylaboraction "github.com/erniealice/fayna-golang/views/activity_labor/action"
	activitylabordetail "github.com/erniealice/fayna-golang/views/activity_labor/detail"
	activitylaborlist "github.com/erniealice/fayna-golang/views/activity_labor/list"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// ModuleDeps holds all dependencies for the activity labor module.
type ModuleDeps struct {
	Routes       fayna.ActivityLaborRoutes
	Labels       fayna.ActivityLaborLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Use case functions wired by block.go via reflection.
	// All nil-safe — handlers return clear gap errors when absent.
	// TODO: add ActivityLabor to espyna OperationUseCases and wire these
	//       in block/wiring.go wireActivityLaborDeps().
	CreateActivityLabor func(ctx context.Context, req *activitylaborpb.CreateActivityLaborRequest) (*activitylaborpb.CreateActivityLaborResponse, error)
	ReadActivityLabor   func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	UpdateActivityLabor func(ctx context.Context, req *activitylaborpb.UpdateActivityLaborRequest) (*activitylaborpb.UpdateActivityLaborResponse, error)
	DeleteActivityLabor func(ctx context.Context, req *activitylaborpb.DeleteActivityLaborRequest) (*activitylaborpb.DeleteActivityLaborResponse, error)
	ListActivityLabors  func(ctx context.Context, req *activitylaborpb.ListActivityLaborsRequest) (*activitylaborpb.ListActivityLaborsResponse, error)
}

// Module holds all constructed activity labor views.
type Module struct {
	routes fayna.ActivityLaborRoutes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewModule creates the activity labor module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	actionDeps := &activitylaboraction.Deps{
		Routes:              deps.Routes,
		Labels:              deps.Labels,
		CreateActivityLabor: deps.CreateActivityLabor,
		ReadActivityLabor:   deps.ReadActivityLabor,
		UpdateActivityLabor: deps.UpdateActivityLabor,
		DeleteActivityLabor: deps.DeleteActivityLabor,
	}

	listDeps := &activitylaborlist.ListViewDeps{
		Routes:             deps.Routes,
		Labels:             deps.Labels,
		CommonLabels:       deps.CommonLabels,
		TableLabels:        deps.TableLabels,
		ListActivityLabors: deps.ListActivityLabors,
	}

	detailDeps := &activitylabordetail.DetailViewDeps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CommonLabels:      deps.CommonLabels,
		ReadActivityLabor: deps.ReadActivityLabor,
	}

	return &Module{
		routes: deps.Routes,
		List:   activitylaborlist.NewView(listDeps),
		Detail: activitylabordetail.NewView(detailDeps),
		Add:    activitylaboraction.NewAddAction(actionDeps),
		Edit:   activitylaboraction.NewEditAction(actionDeps),
		Delete: activitylaboraction.NewDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all activity labor routes.
// Note: /app/activity-labor/list is registered but NOT added to the sidebar.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	// List (power-user/debug — no sidebar entry)
	r.GET(m.routes.ListURL, m.List)

	// Detail (deep link from charge tab)
	r.GET(m.routes.DetailURL, m.Detail)

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	// Note: no BulkDelete for v1 — ActivityLabor is a 1:1 leaf of JobActivity.
}
