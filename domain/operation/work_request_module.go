package operation

import (
	"context"

	wrpkg "github.com/erniealice/fayna-golang/domain/operation/work_request"
	wrdetail "github.com/erniealice/fayna-golang/domain/operation/work_request/detail"
	wrlist "github.com/erniealice/fayna-golang/domain/operation/work_request/list"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	wrpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request"
)

// WorkRequestModuleDeps holds all dependencies for the work_request module.
type WorkRequestModuleDeps struct {
	Routes       wrpkg.Routes
	Labels       wrpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Work request CRUD
	CreateWorkRequest func(ctx context.Context, req *wrpb.CreateWorkRequestRequest) (*wrpb.CreateWorkRequestResponse, error)
	ReadWorkRequest   func(ctx context.Context, req *wrpb.ReadWorkRequestRequest) (*wrpb.ReadWorkRequestResponse, error)
	UpdateWorkRequest func(ctx context.Context, req *wrpb.UpdateWorkRequestRequest) (*wrpb.UpdateWorkRequestResponse, error)
	DeleteWorkRequest func(ctx context.Context, req *wrpb.DeleteWorkRequestRequest) (*wrpb.DeleteWorkRequestResponse, error)
	ListWorkRequests  func(ctx context.Context, req *wrpb.ListWorkRequestsRequest) (*wrpb.ListWorkRequestsResponse, error)

	// Page data
	GetListPageData func(ctx context.Context, req *wrpb.GetWorkRequestListPageDataRequest) (*wrpb.GetWorkRequestListPageDataResponse, error)
	GetItemPageData func(ctx context.Context, req *wrpb.GetWorkRequestItemPageDataRequest) (*wrpb.GetWorkRequestItemPageDataResponse, error)
}

// WorkRequestModule holds all constructed work_request views.
type WorkRequestModule struct {
	routes     wrpkg.Routes
	List       view.View
	Table      view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	SetStatus  view.View
	Assign     view.View
	BulkAssign view.View
	Resolve    view.View
}

// NewWorkRequestModule creates a new work_request module with all views wired.
func NewWorkRequestModule(deps *WorkRequestModuleDeps) *WorkRequestModule {
	listDeps := &wrlist.ListViewDeps{
		Routes:           deps.Routes,
		Labels:           deps.Labels,
		CommonLabels:     deps.CommonLabels,
		TableLabels:      deps.TableLabels,
		ListWorkRequests: deps.ListWorkRequests,
		GetListPageData:  deps.GetListPageData,
	}

	detailDeps := &wrdetail.DetailViewDeps{
		Routes:          deps.Routes,
		Labels:          deps.Labels,
		CommonLabels:    deps.CommonLabels,
		TableLabels:     deps.TableLabels,
		ReadWorkRequest: deps.ReadWorkRequest,
		GetItemPageData: deps.GetItemPageData,
	}

	actionDeps := &wrpkg.ActionDeps{
		Routes:            deps.Routes,
		Labels:            deps.Labels,
		CreateWorkRequest: deps.CreateWorkRequest,
		ReadWorkRequest:   deps.ReadWorkRequest,
		UpdateWorkRequest: deps.UpdateWorkRequest,
		DeleteWorkRequest: deps.DeleteWorkRequest,
		ListWorkRequests:  deps.ListWorkRequests,
	}

	return &WorkRequestModule{
		routes:     deps.Routes,
		List:       wrlist.NewView(listDeps),
		Table:      wrlist.NewTableView(listDeps),
		Detail:     wrdetail.NewView(detailDeps),
		TabAction:  wrdetail.NewTabAction(detailDeps),
		Add:        wrpkg.NewAddAction(actionDeps),
		Edit:       wrpkg.NewEditAction(actionDeps),
		SetStatus:  wrpkg.NewSetStatusAction(actionDeps),
		Assign:     wrpkg.NewAssignAction(actionDeps),
		BulkAssign: wrpkg.NewBulkAssignAction(actionDeps),
		Resolve:    wrpkg.NewResolveAction(actionDeps),
	}
}

// RegisterRoutes registers all work_request routes.
func (m *WorkRequestModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)
	r.GET(m.routes.TableURL, m.Table)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.GET(m.routes.SetStatusURL, m.SetStatus)
	r.POST(m.routes.SetStatusURL, m.SetStatus)
	r.GET(m.routes.AssignURL, m.Assign)
	r.POST(m.routes.AssignURL, m.Assign)
	r.GET(m.routes.BulkAssignURL, m.BulkAssign)
	r.POST(m.routes.BulkAssignURL, m.BulkAssign)
	r.POST(m.routes.ResolveURL, m.Resolve)
}
