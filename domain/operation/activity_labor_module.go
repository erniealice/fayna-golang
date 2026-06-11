package operation

import (
	"context"

	activitylaborpkg "github.com/erniealice/fayna-golang/domain/operation/activity_labor"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"

	activitylaboraction "github.com/erniealice/fayna-golang/domain/operation/activity_labor/action"
	activitylabordetail "github.com/erniealice/fayna-golang/domain/operation/activity_labor/detail"
	activitylaborlist "github.com/erniealice/fayna-golang/domain/operation/activity_labor/list"
)

// ActivityLaborModuleDeps holds all dependencies for the activity labor module.
type ActivityLaborModuleDeps struct {
	Routes       activitylaborpkg.Routes
	Labels       activitylaborpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateActivityLabor func(ctx context.Context, req *activitylaborpb.CreateActivityLaborRequest) (*activitylaborpb.CreateActivityLaborResponse, error)
	ReadActivityLabor   func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	UpdateActivityLabor func(ctx context.Context, req *activitylaborpb.UpdateActivityLaborRequest) (*activitylaborpb.UpdateActivityLaborResponse, error)
	DeleteActivityLabor func(ctx context.Context, req *activitylaborpb.DeleteActivityLaborRequest) (*activitylaborpb.DeleteActivityLaborResponse, error)
	ListActivityLabors  func(ctx context.Context, req *activitylaborpb.ListActivityLaborsRequest) (*activitylaborpb.ListActivityLaborsResponse, error)
}

// ActivityLaborModule holds all constructed activity labor views.
type ActivityLaborModule struct {
	routes activitylaborpkg.Routes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewActivityLaborModule creates the activity labor module with all views wired.
func NewActivityLaborModule(deps *ActivityLaborModuleDeps) *ActivityLaborModule {
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

	return &ActivityLaborModule{
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
func (m *ActivityLaborModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
}
