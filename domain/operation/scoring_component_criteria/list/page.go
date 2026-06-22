package list

import (
	"context"
	"fmt"
	"log"

	scc "github.com/erniealice/fayna-golang/domain/operation/scoring_component_criteria"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	sccpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_component_criteria"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                        scc.Routes
	ListScoringComponentCriterias func(ctx context.Context, req *sccpb.ListScoringComponentCriteriasRequest) (*sccpb.ListScoringComponentCriteriasResponse, error)
	Labels                        scc.Labels
	CommonLabels                  pyeza.CommonLabels
	TableLabels                   types.TableLabels
}

// PageData holds the data for the scoring component criteria list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the scoring component criteria list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_component_criteria", "list") {
			return view.Forbidden("scoring_component_criteria:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListScoringComponentCriterias(ctx, &sccpb.ListScoringComponentCriteriasRequest{})
		if err != nil {
			log.Printf("Failed to list scoring component criterias: %v", err)
			return view.Error(fmt.Errorf("failed to load scoring component criteria: %w", err))
		}

		l := deps.Labels
		columns := listColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "scoring-component-criteria-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "scoring_scheme_id",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddLink,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("scoring_component_criteria", "create"),
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
				HeaderIcon:     "icon-link",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "scoring-component-criteria-list-content",
			Table:           tableConfig,
		}

		return view.OK("scoring-component-criteria-list", pageData)
	})
}

func listColumns(l scc.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "scoring_scheme_id", Label: l.Columns.ScoringSchemeID},
		{Key: "scoring_component_id", Label: l.Columns.ScoringComponentID},
		{Key: "outcome_criteria_id", Label: l.Columns.OutcomeCriteriaID},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*sccpb.ScoringComponentCriteria,
	status string,
	l scc.Labels,
	routes scc.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, c := range items {
		activeStatus := "inactive"
		if c.GetActive() {
			activeStatus = "active"
		}
		if activeStatus != status {
			continue
		}

		id := c.GetId()
		schemeID := c.GetScoringSchemeId()
		componentID := c.GetScoringComponentId()
		criteriaID := c.GetOutcomeCriteriaId()
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		statusVariant := "warning"
		if c.GetActive() {
			statusVariant = "success"
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: schemeID},
				{Type: "text", Value: componentID},
				{Type: "text", Value: criteriaID},
				{Type: "badge", Value: activeStatus, Variant: statusVariant},
			},
			DataAttrs: map[string]string{
				"scoring_scheme_id":    schemeID,
				"scoring_component_id": componentID,
				"outcome_criteria_id":  criteriaID,
				"status":               activeStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("scoring_component_criteria", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("scoring_component_criteria", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func statusPageTitle(l scc.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l scc.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l scc.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l scc.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
