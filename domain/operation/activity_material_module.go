package operation

import (
	"context"

	activitymaterialpkg "github.com/erniealice/fayna-golang/domain/operation/activity_material"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"

	activitymaterialaction "github.com/erniealice/fayna-golang/domain/operation/activity_material/action"
	activitymaterialdetail "github.com/erniealice/fayna-golang/domain/operation/activity_material/detail"
	activitymateriallist "github.com/erniealice/fayna-golang/domain/operation/activity_material/list"
)

// ActivityMaterialModuleDeps holds all dependencies for the activity material module.
type ActivityMaterialModuleDeps struct {
	Routes       activitymaterialpkg.Routes
	Labels       activitymaterialpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateActivityMaterial func(ctx context.Context, req *activitymaterialpb.CreateActivityMaterialRequest) (*activitymaterialpb.CreateActivityMaterialResponse, error)
	ReadActivityMaterial   func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	UpdateActivityMaterial func(ctx context.Context, req *activitymaterialpb.UpdateActivityMaterialRequest) (*activitymaterialpb.UpdateActivityMaterialResponse, error)
	DeleteActivityMaterial func(ctx context.Context, req *activitymaterialpb.DeleteActivityMaterialRequest) (*activitymaterialpb.DeleteActivityMaterialResponse, error)
	ListActivityMaterials  func(ctx context.Context, req *activitymaterialpb.ListActivityMaterialsRequest) (*activitymaterialpb.ListActivityMaterialsResponse, error)
}

// ActivityMaterialModule holds all constructed activity material views.
type ActivityMaterialModule struct {
	routes activitymaterialpkg.Routes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewActivityMaterialModule creates the activity material module with all views wired.
func NewActivityMaterialModule(deps *ActivityMaterialModuleDeps) *ActivityMaterialModule {
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

	return &ActivityMaterialModule{
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
func (m *ActivityMaterialModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
}
