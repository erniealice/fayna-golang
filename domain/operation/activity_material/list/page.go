// Package list provides a basic list view for ActivityMaterial records.
// This page is registered at /app/activity-material/list but has NO sidebar entry.
// It is a power-user / debugging surface — the primary operator interface is the
// JobActivity detail page's charge tab.
package list

import (
	"context"
	"log"

	activity_material "github.com/erniealice/fayna-golang/domain/operation/activity_material"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	activitymaterialpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/activity_material"
)

// ListViewDeps holds the dependencies needed by the activity material list page.
type ListViewDeps struct {
	Routes       activity_material.Routes
	Labels       activity_material.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// ListActivityMaterials is the paginated list use case.
	// Nil-safe: renders an empty state with a gap notice when not wired.
	ListActivityMaterials func(ctx context.Context, req *activitymaterialpb.ListActivityMaterialsRequest) (*activitymaterialpb.ListActivityMaterialsResponse, error)
}

// PageData holds the data for the activity material list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          activity_material.Labels
}

// NewView creates the activity material list view.
// Not wired in the sidebar — accessed directly at /app/activity-material/list.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		// 2026-05-14 permission-gates P2a + P2b.
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("activity_material", "list") {
			return view.Forbidden("activity_material:list")
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
				HeaderIcon:     "icon-package",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "activity-material-list-content",
			Table:           table,
			Labels:          l,
		}

		return view.OK("activity-material-list", pageData)
	})
}

func buildTable(ctx context.Context, deps *ListViewDeps) *types.TableConfig {
	l := deps.Labels

	columns := []types.TableColumn{
		{Key: "activity_id", Label: "Activity ID"},
		{Key: "product_id", Label: l.Columns.Product},
		{Key: "unit_of_measure", Label: l.Columns.UnitOfMeasure},
		{Key: "lot_number", Label: l.Columns.LotNumber},
		{Key: "location_id", Label: l.Columns.Location},
	}

	var rows []types.TableRow

	if deps.ListActivityMaterials == nil {
		// Use case not wired — return empty state with gap notice.
		return &types.TableConfig{
			ID:      "activity-material-list-table",
			Columns: columns,
			Rows:    rows,
			EmptyState: types.TableEmptyState{
				Title:   "ListActivityMaterials not wired",
				Message: "Add ActivityMaterial to espyna OperationUseCases to populate this list.",
			},
		}
	}

	resp, err := deps.ListActivityMaterials(ctx, &activitymaterialpb.ListActivityMaterialsRequest{})
	if err != nil {
		log.Printf("Failed to list activity materials: %v", err)
		return &types.TableConfig{
			ID:      "activity-material-list-table",
			Columns: columns,
			Rows:    rows,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
		}
	}

	for _, mat := range resp.GetData() {
		productName := mat.GetProductId()
		if p := mat.GetProduct(); p != nil && p.GetName() != "" {
			productName = p.GetName()
		}
		row := types.TableRow{
			ID: mat.GetActivityId(),
			Cells: []types.TableCell{
				{Type: "text", Value: mat.GetActivityId()},
				{Type: "text", Value: productName},
				{Type: "text", Value: mat.GetUnitOfMeasure()},
				{Type: "text", Value: mat.GetLotNumber()},
				{Type: "text", Value: mat.GetLocationId()},
			},
		}
		rows = append(rows, row)
	}

	return &types.TableConfig{
		ID:      "activity-material-list-table",
		Columns: columns,
		Rows:    rows,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
	}
}
