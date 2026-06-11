// Package list provides a basic list view for ActivityExpense records.
// This page is registered at /app/activity-expense/list but has NO sidebar entry.
// It is a power-user / debugging surface — the primary operator interface is the
// JobActivity detail page's charge tab.
package list

import (
	"context"
	"log"

	activity_expense "github.com/erniealice/fayna-golang/domain/operation/activity_expense"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activityexpensepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_expense"
)

// ListViewDeps holds the dependencies needed by the activity expense list page.
type ListViewDeps struct {
	Routes       activity_expense.Routes
	Labels       activity_expense.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ListActivityExpenses is the paginated list use case.
	// Nil-safe: renders an empty state with a gap notice when not wired.
	ListActivityExpenses func(ctx context.Context, req *activityexpensepb.ListActivityExpensesRequest) (*activityexpensepb.ListActivityExpensesResponse, error)
}

// PageData holds the data for the activity expense list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          activity_expense.Labels
}

// NewView creates the activity expense list view.
// Not wired in the sidebar — accessed directly at /app/activity-expense/list.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a: reject direct-URL access without
		// activity_expense:list. (Catalog rows for activity_expense were added
		// in Phase 1.)
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_expense", "list") {
			return view.Forbidden("activity_expense:list")
		}
		_ = perms

		l := deps.Labels
		headerTitle := l.Page.Heading

		table := buildTable(ctx, deps)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          headerTitle,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				HeaderTitle:    headerTitle,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-receipt",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "activity-expense-list-content",
			Table:           table,
			Labels:          l,
		}

		return view.OK("activity-expense-list", pageData)
	})
}

func buildTable(ctx context.Context, deps *ListViewDeps) *types.TableConfig {
	l := deps.Labels

	columns := []types.TableColumn{
		{Key: "activity_id", Label: "Activity ID"},
		{Key: "expense_category_id", Label: l.Columns.ExpenseCategory},
		{Key: "vendor_ref", Label: l.Columns.VendorRef},
		{Key: "payment_method", Label: l.Columns.PaymentMethod},
		{Key: "markup_pct_override", Label: l.Columns.MarkupPct},
	}

	var rows []types.TableRow

	if deps.ListActivityExpenses == nil {
		// Use case not wired — return empty state with gap notice.
		return &types.TableConfig{
			ID:      "activity-expense-list-table",
			Columns: columns,
			Rows:    rows,
			EmptyState: types.TableEmptyState{
				Title:   "ListActivityExpenses not wired",
				Message: "Add ActivityExpense to espyna OperationUseCases to populate this list.",
			},
		}
	}

	resp, err := deps.ListActivityExpenses(ctx, &activityexpensepb.ListActivityExpensesRequest{})
	if err != nil {
		log.Printf("Failed to list activity expenses: %v", err)
		return &types.TableConfig{
			ID:      "activity-expense-list-table",
			Columns: columns,
			Rows:    rows,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
		}
	}

	for _, exp := range resp.GetData() {
		row := types.TableRow{
			ID: exp.GetActivityId(),
			Cells: []types.TableCell{
				{Type: "text", Value: exp.GetActivityId()},
				{Type: "text", Value: exp.GetExpenseCategoryId()},
				{Type: "text", Value: exp.GetVendorRef()},
				{Type: "text", Value: exp.GetPaymentMethod()},
				{Type: "text", Value: exp.GetExpenseCategory()},
			},
		}
		rows = append(rows, row)
	}

	return &types.TableConfig{
		ID:      "activity-expense-list-table",
		Columns: columns,
		Rows:    rows,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
}
