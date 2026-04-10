package job_activity

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"

	jobactivitydetail "github.com/erniealice/fayna-golang/views/job_activity/detail"
	jobactivitylist "github.com/erniealice/fayna-golang/views/job_activity/list"
)

// ModuleDeps holds all dependencies for the job activity module.
type ModuleDeps struct {
	Routes       fayna.JobActivityRoutes
	Labels       fayna.JobActivityLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Job activity use case functions
	GetJobActivityListPageData func(ctx context.Context, req *jobactivitypb.GetJobActivityListPageDataRequest) (*jobactivitypb.GetJobActivityListPageDataResponse, error)
	ReadJobActivity            func(ctx context.Context, req *jobactivitypb.ReadJobActivityRequest) (*jobactivitypb.ReadJobActivityResponse, error)
	CreateJobActivity          func(ctx context.Context, req *jobactivitypb.CreateJobActivityRequest) (*jobactivitypb.CreateJobActivityResponse, error)
	UpdateJobActivity          func(ctx context.Context, req *jobactivitypb.UpdateJobActivityRequest) (*jobactivitypb.UpdateJobActivityResponse, error)
	DeleteJobActivity          func(ctx context.Context, req *jobactivitypb.DeleteJobActivityRequest) (*jobactivitypb.DeleteJobActivityResponse, error)

	// Approval workflow
	SubmitForApproval func(ctx context.Context, req *jobactivitypb.SubmitForApprovalRequest) (*jobactivitypb.SubmitForApprovalResponse, error)
	ApproveActivity   func(ctx context.Context, req *jobactivitypb.ApproveJobActivityRequest) (*jobactivitypb.ApproveJobActivityResponse, error)
	RejectActivity    func(ctx context.Context, req *jobactivitypb.RejectJobActivityRequest) (*jobactivitypb.RejectJobActivityResponse, error)

	// Activity subtype read functions (for detail page)
	ReadActivityLabor    func(ctx context.Context, req *activitylaborpb.ReadActivityLaborRequest) (*activitylaborpb.ReadActivityLaborResponse, error)
	ReadActivityMaterial func(ctx context.Context, req *activitymaterialpb.ReadActivityMaterialRequest) (*activitymaterialpb.ReadActivityMaterialResponse, error)
	ReadActivityExpense  func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)

	NewID func() string
}

// Module holds all constructed job activity views.
type Module struct {
	routes  fayna.JobActivityRoutes
	List    view.View
	Detail  view.View
	Create  view.View
	Update  view.View
	Delete  view.View
	Submit  view.View
	Approve view.View
	Reject  view.View
}

// NewModule creates the job activity module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
	listView := jobactivitylist.NewView(&jobactivitylist.ListViewDeps{
		Routes:                     deps.Routes,
		GetJobActivityListPageData: deps.GetJobActivityListPageData,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
	})

	detailDeps := &jobactivitydetail.DetailViewDeps{
		Routes:               deps.Routes,
		ReadJobActivity:      deps.ReadJobActivity,
		ReadActivityLabor:    deps.ReadActivityLabor,
		ReadActivityMaterial: deps.ReadActivityMaterial,
		ReadActivityExpense:  deps.ReadActivityExpense,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
	}

	return &Module{
		routes:  deps.Routes,
		List:    listView,
		Detail:  jobactivitydetail.NewView(detailDeps),
		Create:  newCreateAction(deps),
		Update:  newUpdateAction(deps),
		Delete:  newDeleteAction(deps),
		Submit:  newSubmitAction(deps),
		Approve: newApproveAction(deps),
		Reject:  newRejectAction(deps),
	}
}

// RegisterRoutes registers all job activity routes.
func (m *Module) RegisterRoutes(r view.RouteRegistrar) {
	// List (full page + HTMX partial share the same view)
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.ListURL+"/content", m.List)

	// Detail
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.DetailURL+"/content", m.Detail)

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Create)
	r.POST(m.routes.AddURL, m.Create)
	r.GET(m.routes.EditURL, m.Update)
	r.POST(m.routes.EditURL, m.Update)
	r.POST(m.routes.DeleteURL, m.Delete)

	// Approval workflow
	r.POST(m.routes.SubmitURL, m.Submit)
	r.POST(m.routes.ApproveURL, m.Approve)
	r.POST(m.routes.RejectURL, m.Reject)
}
