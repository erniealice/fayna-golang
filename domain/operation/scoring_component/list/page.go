package list

import (
	"context"
	"fmt"
	"log"

	scoring_component "github.com/erniealice/fayna-golang/domain/operation/scoring_component"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	scoringpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                scoring_component.Routes
	ListScoringComponents func(ctx context.Context, req *scoringpb.ListScoringComponentsRequest) (*scoringpb.ListScoringComponentsResponse, error)
	Labels                scoring_component.Labels
	CommonLabels          pyeza.CommonLabels
	TableLabels           types.TableLabels
}

// PageData holds the data for the scoring component list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the scoring component list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component", "list") {
			return view.Forbidden("scoring_component:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListScoringComponents(ctx, &scoringpb.ListScoringComponentsRequest{})
		if err != nil {
			log.Printf("Failed to list scoring components: %v", err)
			return view.Error(fmt.Errorf("failed to load scoring components: %w", err))
		}

		l := deps.Labels
		columns := componentColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "scoring-components-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sequence_order",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddComponent,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("scoring_component", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          statusPageTitle(l, status),
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    statusPageTitle(l, status),
				HeaderSubtitle: statusPageCaption(l, status),
				HeaderIcon:     "icon-bar-chart-2",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "scoring-component-list-content",
			Table:           tableConfig,
		}

		return view.OK("scoring-component-list", pageData)
	})
}

func componentColumns(l scoring_component.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "code", Label: l.Columns.Code},
		{Key: "label", Label: l.Columns.Label},
		{Key: "weight", Label: l.Columns.Weight, WidthClass: "col-lg"},
		{Key: "sequence_order", Label: l.Columns.SequenceOrder, WidthClass: "col-lg"},
		{Key: "active", Label: l.Columns.Active, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*scoringpb.ScoringComponent,
	status string,
	l scoring_component.Labels,
	routes scoring_component.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, c := range items {
		isActive := c.GetActive()
		rowStatus := "inactive"
		if isActive {
			rowStatus = "active"
		}
		if rowStatus != status {
			continue
		}

		id := c.GetId()
		code := c.GetCode()
		label := c.GetLabel()
		weight := fmt.Sprintf("%.2f", c.GetWeight())
		seqOrder := fmt.Sprintf("%d", c.GetSequenceOrder())
		activeStr := "Yes"
		activeVariant := "success"
		if !isActive {
			activeStr = "No"
			activeVariant = "default"
		}
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: code},
				{Type: "text", Value: label},
				{Type: "text", Value: weight},
				{Type: "text", Value: seqOrder},
				{Type: "badge", Value: activeStr, Variant: activeVariant},
			},
			DataAttrs: map[string]string{
				"code":           code,
				"label":          label,
				"weight":         weight,
				"sequence_order": seqOrder,
				"active":         rowStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("scoring_component", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: label, Disabled: !perms.Can("scoring_component", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func statusPageTitle(l scoring_component.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l scoring_component.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l scoring_component.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l scoring_component.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
