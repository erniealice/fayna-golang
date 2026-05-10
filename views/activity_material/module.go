// Package activity_material provides the ActivityMaterial view module (Shape 2 sibling).
//
// ActivityMaterial is the charge-detail record for ENTRY_TYPE_MATERIAL job activities.
// PK = activity_id (not a separate id) — it is the FK to JobActivity (1:1 relationship).
//
// Access patterns:
//   - Primary: JobActivity detail page's charge tab (entry_type=MATERIAL) — drawer form.
//   - Deep link: /app/activity-material/{id} — direct URL from the charge tab Edit CTA.
//   - Debug/power-user: /app/activity-material/list — NOT in the sidebar.
//
// This module mirrors views/activity_labor/ (Shape 2 sibling template).
package activity_material

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"
	activitymaterialaction "github.com/erniealice/fayna-golang/views/activity_material/action"
	activitymaterialdetail "github.com/erniealice/fayna-golang/views/activity_material/detail"
	activitymateriallist "github.com/erniealice/fayna-golang/views/activity_material/list"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// ModuleDeps holds all dependencies for the activity material module.
type ModuleDeps struct {
	Routes       fayna.ActivityMaterialRoutes
	Labels       fayna.ActivityMaterialLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Use case functions wired by block.go via reflection.
	// All nil-safe — handlers return clear gap errors when absent.
	// TODO: add ActivityMaterial to espyna OperationUseCases and wire these
	//       in block/wiring.go wireActivityMaterialDeps().
	CreateActivityMaterial func(ctx context.Context, req *activitymaterialpb.CreateActivityMaterialRequest) (*activitymaterialpb.CreateActivityMaterialResponse, error)
	ReadActivityMaterial   func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	UpdateActivityMaterial func(ctx context.Context, req *activitymaterialpb.UpdateActivityMaterialRequest) (*activitymaterialpb.UpdateActivityMaterialResponse, error)
	DeleteActivityMaterial func(ctx context.Context, req *activitymaterialpb.DeleteActivityMaterialRequest) (*activitymaterialpb.DeleteActivityMaterialResponse, error)
	ListActivityMaterials  func(ctx context.Context, req *activitymaterialpb.ListActivityMaterialsRequest) (*activitymaterialpb.ListActivityMaterialsResponse, error)
}

// Module holds all constructed activity material views.
type Module struct {
	routes fayna.ActivityMaterialRoutes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewModule creates the activity material module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	actionDeps := &activitymaterialaction.Deps{
		Routes:                 deps.Routes,
		Labels:                 deps.Labels,
		CreateActivityMaterial: deps.CreateActivityMaterial,
		ReadActivityMaterial:   deps.ReadActivityMaterial,
		UpdateActivityMaterial: deps.UpdateActivityMaterial,
		DeleteActivityMaterial: deps.DeleteActivityMaterial,
	}

	listDeps := &activitymateriallist.ListViewDeps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CommonLabels:          deps.CommonLabels,
		TableLabels:           deps.TableLabels,
		ListActivityMaterials: deps.ListActivityMaterials,
	}

	detailDeps := &activitymaterialdetail.DetailViewDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		ReadActivityMaterial: deps.ReadActivityMaterial,
	}

	return &Module{
		routes: deps.Routes,
		List:   activitymateriallist.NewView(listDeps),
		Detail: activitymaterialdetail.NewView(detailDeps),
		Add:    activitymaterialaction.NewAddAction(actionDeps),
		Edit:   activitymaterialaction.NewEditAction(actionDeps),
		Delete: activitymaterialaction.NewDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all activity material routes.
// Note: /app/activity-material/list is registered but NOT added to the sidebar.
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
	// Note: no BulkDelete for v1 — ActivityMaterial is a 1:1 leaf of JobActivity.
}
