// Package activity_expense provides the ActivityExpense view module (Shape 2 sibling).
//
// ActivityExpense is the charge-detail record for ENTRY_TYPE_EXPENSE job activities.
// PK = activity_id (not a separate id) — it is the FK to JobActivity (1:1 relationship).
//
// Access patterns:
//   - Primary: JobActivity detail page's charge tab (entry_type=EXPENSE) — drawer form.
//   - Deep link: /app/activity-expense/{id} — direct URL from the charge tab Edit CTA.
//   - Debug/power-user: /app/activity-expense/list — NOT in the sidebar.
//
// This module mirrors views/activity_labor/ (Shape 2 sibling template).
package activity_expense

import (
	"context"

	operation "github.com/erniealice/fayna-golang/domain/operation"
	activityexpenseaction "github.com/erniealice/fayna-golang/domain/operation/views/activity_expense/action"
	activityexpensedetail "github.com/erniealice/fayna-golang/domain/operation/views/activity_expense/detail"
	activityexpenselist "github.com/erniealice/fayna-golang/domain/operation/views/activity_expense/list"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// ModuleDeps holds all dependencies for the activity expense module.
type ModuleDeps struct {
	Routes       operation.ActivityExpenseRoutes
	Labels       operation.ActivityExpenseLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Use case functions wired by block.go via reflection.
	// All nil-safe — handlers return clear gap errors when absent.
	// TODO: add ActivityExpense to espyna OperationUseCases and wire these
	//       in block/wiring.go wireActivityExpenseDeps().
	CreateActivityExpense func(ctx context.Context, req *activityexpensepb.CreateActivityExpenseRequest) (*activityexpensepb.CreateActivityExpenseResponse, error)
	ReadActivityExpense   func(ctx context.Context, req *activityexpensepb.ReadActivityExpenseRequest) (*activityexpensepb.ReadActivityExpenseResponse, error)
	UpdateActivityExpense func(ctx context.Context, req *activityexpensepb.UpdateActivityExpenseRequest) (*activityexpensepb.UpdateActivityExpenseResponse, error)
	DeleteActivityExpense func(ctx context.Context, req *activityexpensepb.DeleteActivityExpenseRequest) (*activityexpensepb.DeleteActivityExpenseResponse, error)
	ListActivityExpenses  func(ctx context.Context, req *activityexpensepb.ListActivityExpensesRequest) (*activityexpensepb.ListActivityExpensesResponse, error)
}

// Module holds all constructed activity expense views.
type Module struct {
	routes operation.ActivityExpenseRoutes
	List   view.View
	Detail view.View
	Add    view.View
	Edit   view.View
	Delete view.View
}

// NewModule creates the activity expense module with all views wired.
func NewModule(deps *ModuleDeps) *Module {
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

	return &Module{
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
	// Note: no BulkDelete for v1 — ActivityExpense is a 1:1 leaf of JobActivity.
}
