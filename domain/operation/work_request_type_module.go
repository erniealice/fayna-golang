package operation

import (
	"context"

	wrtpkg "github.com/erniealice/fayna-golang/domain/operation/work_request_type"
	wrtlist "github.com/erniealice/fayna-golang/domain/operation/work_request_type/views/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	wrtpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request_type"
)

// WorkRequestTypeModuleDeps holds all dependencies for the work_request_type module.
type WorkRequestTypeModuleDeps struct {
	Routes       wrtpkg.Routes
	Labels       wrtpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Work request type CRUD
	CreateWorkRequestType func(ctx context.Context, req *wrtpb.CreateWorkRequestTypeRequest) (*wrtpb.CreateWorkRequestTypeResponse, error)
	ReadWorkRequestType   func(ctx context.Context, req *wrtpb.ReadWorkRequestTypeRequest) (*wrtpb.ReadWorkRequestTypeResponse, error)
	UpdateWorkRequestType func(ctx context.Context, req *wrtpb.UpdateWorkRequestTypeRequest) (*wrtpb.UpdateWorkRequestTypeResponse, error)
	ListWorkRequestTypes  func(ctx context.Context, req *wrtpb.ListWorkRequestTypesRequest) (*wrtpb.ListWorkRequestTypesResponse, error)
}

// WorkRequestTypeModule holds all constructed work_request_type views.
type WorkRequestTypeModule struct {
	routes wrtpkg.Routes
	List   view.View
	Table  view.View
	Add    view.View
	Edit   view.View
}

// NewWorkRequestTypeModule creates a new work_request_type module with all views wired.
func NewWorkRequestTypeModule(deps *WorkRequestTypeModuleDeps) *WorkRequestTypeModule {
	listDeps := &wrtlist.ListViewDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		ListWorkRequestTypes: deps.ListWorkRequestTypes,
	}

	actionDeps := &wrtpkg.ActionDeps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CreateWorkRequestType: deps.CreateWorkRequestType,
		ReadWorkRequestType:   deps.ReadWorkRequestType,
		UpdateWorkRequestType: deps.UpdateWorkRequestType,
		ListWorkRequestTypes:  deps.ListWorkRequestTypes,
	}

	return &WorkRequestTypeModule{
		routes: deps.Routes,
		List:   wrtlist.NewView(listDeps),
		Table:  wrtlist.NewTableView(listDeps),
		Add:    wrtpkg.NewAddAction(actionDeps),
		Edit:   wrtpkg.NewEditAction(actionDeps),
	}
}

// RegisterRoutes registers all work_request_type routes.
func (m *WorkRequestTypeModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.TableURL, m.Table)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
}
