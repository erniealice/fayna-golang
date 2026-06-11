package operation

import (
	"context"

	activityexpensepkg "github.com/erniealice/fayna-golang/domain/operation/activity_expense"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"

	activityexpenseaction "github.com/erniealice/fayna-golang/domain/operation/activity_expense/action"
	activityexpensedetail "github.com/erniealice/fayna-golang/domain/operation/activity_expense/detail"
	activityexpenselist "github.com/erniealice/fayna-golang/domain/operation/activity_expense/list"
)

// ActivityExpenseModuleDeps holds all dependencies for the activity expense module.
type ActivityExpenseModuleDeps struct {
	Routes       activityexpensepkg.Routes
	Labels       activityexpensepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateActivityExpense func(ctx context.Context, req *activityexpensepb.CreateActivityExpenseRequest) (*activityexpensepb.CreateActivityExpenseResponse, error)
	ReadActivityExpense   func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
	UpdateActivityExpense func(ctx context.Context, req *activityexpensepb.UpdateActivityExpenseRequest) (*activityexpensepb.UpdateActivityExpenseResponse, error)
	DeleteActivityExpense func(ctx context.Context, req *activityexpensepb.DeleteActivityExpenseRequest) (*activityexpensepb.DeleteActivityExpenseResponse, error)
	ListActivityExpenses  func(ctx context.Context, req *activityexpensepb.ListActivityExpensesRequest) (*activityexpensepb.ListActivityExpensesResponse, error)
}

// ActivityExpenseModule holds all constructed activity expense views.
type ActivityExpenseModule struct {
	routes activityexpensepkg.Routes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewActivityExpenseModule creates the activity expense module with all views wired.
func NewActivityExpenseModule(deps *ActivityExpenseModuleDeps) *ActivityExpenseModule {
	actionDeps := &activityexpenseaction.Deps{
		Routes:                deps.Routes,
		Labels:                deps.Labels,
		CreateActivityExpense: deps.CreateActivityExpense,
		ReadActivityExpense:   deps.ReadActivityExpense,
		UpdateActivityExpense: deps.UpdateActivityExpense,
		DeleteActivityExpense: deps.DeleteActivityExpense,
	}

	listDeps := &activityexpenselist.ListViewDeps{
		Routes:               deps.Routes,
		Labels:               deps.Labels,
		CommonLabels:         deps.CommonLabels,
		TableLabels:          deps.TableLabels,
		ListActivityExpenses: deps.ListActivityExpenses,
	}

	detailDeps := &activityexpensedetail.DetailViewDeps{
		Routes:              deps.Routes,
		Labels:              deps.Labels,
		CommonLabels:        deps.CommonLabels,
		ReadActivityExpense: deps.ReadActivityExpense,
	}

	return &ActivityExpenseModule{
		routes: deps.Routes,
		List:   activityexpenselist.NewView(listDeps),
		Detail: activityexpensedetail.NewView(detailDeps),
		Add:    activityexpenseaction.NewAddAction(actionDeps),
		Edit:   activityexpenseaction.NewEditAction(actionDeps),
		Delete: activityexpenseaction.NewDeleteAction(actionDeps),
	}
}

// RegisterRoutes registers all activity expense routes.
// Note: /app/activity-expense/list is registered but NOT added to the sidebar.
func (m *ActivityExpenseModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
}
