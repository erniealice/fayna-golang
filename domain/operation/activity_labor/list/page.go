// Package list provides a basic list view for ActivityLabor records.
// This page is registered at /app/activity-labor/list but has NO sidebar entry.
// It is a power-user / debugging surface — the primary operator interface is the
// JobActivity detail page's charge tab.
package list

import (
	"context"
	"fmt"
	"log"

	activity_labor "github.com/erniealice/fayna-golang/domain/operation/activity_labor"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitylaborpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_labor"
)

// ListViewDeps holds the dependencies needed by the activity labor list page.
type ListViewDeps struct {
	Routes       activity_labor.Routes
	Labels       activity_labor.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ListActivityLabors is the paginated list use case.
	// Nil-safe: renders an empty state with a gap notice when not wired.
	ListActivityLabors func(ctx context.Context, req *activitylaborpb.ListActivityLaborsRequest) (*activitylaborpb.ListActivityLaborsResponse, error)
}

// PageData holds the data for the activity labor list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          activity_labor.Labels
}

// NewView creates the activity labor list view.
// Not wired in the sidebar — accessed directly at /app/activity-labor/list.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a + P2b: add GetUserPermissions
		// (page previously had none) and reject direct-URL access without
		// activity_labor:list (catalog row added in Phase 1).
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_labor", "list") {
			return view.Forbidden("activity_labor:list")
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
				HeaderIcon:     "icon-clock",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "activity-labor-list-content",
			Table:           table,
			Labels:          l,
		}

		return view.OK("activity-labor-list", pageData)
	})
}

func buildTable(ctx context.Context, deps *ListViewDeps) *types.TableConfig {
	l := deps.Labels

	columns := []types.TableColumn{
		{Key: "activity_id", Label: "Activity ID"},
		{Key: "staff_id", Label: l.Columns.Staff},
		{Key: "hours", Label: l.Columns.Hours},
		{Key: "rate_type", Label: l.Columns.RateType},
		{Key: "time_start", Label: l.Columns.TimeStart},
		{Key: "time_end", Label: l.Columns.TimeEnd},
	}

	var rows []types.TableRow

	if deps.ListActivityLabors == nil {
		// Use case not wired — return empty state with gap notice.
		return &types.TableConfig{
			ID:      "activity-labor-list-table",
			Columns: columns,
			Rows:    rows,
			EmptyState: types.TableEmptyState{
				Title:   "ListActivityLabors not wired",
				Message: "Add ActivityLabor to espyna OperationUseCases to populate this list.",
			},
		}
	}

	resp, err := deps.ListActivityLabors(ctx, &activitylaborpb.ListActivityLaborsRequest{})
	if err != nil {
		log.Printf("Failed to list activity labors: %v", err)
		return &types.TableConfig{
			ID:      "activity-labor-list-table",
			Columns: columns,
			Rows:    rows,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
		}
	}

	for _, labor := range resp.GetData() {
		row := types.TableRow{
			ID: labor.GetActivityId(),
			Cells: []types.TableCell{
				{Type: "text", Value: labor.GetActivityId()},
				{Type: "text", Value: labor.GetStaffId()},
				{Type: "text", Value: fmt.Sprintf("%.2f", labor.GetHours())},
				{Type: "text", Value: labor.GetRateType().String()},
				{Type: "text", Value: labor.GetTimeStartString()},
				{Type: "text", Value: labor.GetTimeEndString()},
			},
		}
		rows = append(rows, row)
	}

	return &types.TableConfig{
		ID:      "activity-labor-list-table",
		Columns: columns,
		Rows:    rows,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
}
