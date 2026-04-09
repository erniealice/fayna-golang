package list

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes               fayna.OutcomeCriteriaRoutes
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
	Labels               fayna.OutcomeCriteriaLabels
	CommonLabels         pyeza.CommonLabels
	TableLabels          types.TableLabels
}

// PageData holds the data for the outcome criteria list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the outcome criteria list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListOutcomeCriterias(ctx, &criteriapb.ListOutcomeCriteriasRequest{})
		if err != nil {
			log.Printf("Failed to list outcome criterias: %v", err)
			return view.Error(fmt.Errorf("failed to load criteria: %w", err))
		}

		l := deps.Labels
		columns := criteriaColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "outcome-criteria-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "name",
			DefaultSortDirection: "asc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   emptyTitle(l, status),
				Message: emptyMessage(l, status),
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddCriterion,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("outcome_criteria", "create"),
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
				HeaderIcon:     "icon-check-circle",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "outcome-criteria-list-content",
			Table:           tableConfig,
		}

		return view.OK("outcome-criteria-list", pageData)
	})
}

func criteriaColumns(l fayna.OutcomeCriteriaLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name, Sortable: true},
		{Key: "type", Label: l.Columns.Type, Sortable: true, WidthClass: "col-4xl"},
		{Key: "scope", Label: l.Columns.Scope, Sortable: true, WidthClass: "col-3xl"},
		{Key: "version", Label: l.Columns.Version, Sortable: true, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, Sortable: true, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*criteriapb.OutcomeCriteria,
	status string,
	l fayna.OutcomeCriteriaLabels,
	routes fayna.OutcomeCriteriaRoutes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, c := range items {
		versionStatus := versionStatusString(c.GetVersionStatus())
		if versionStatus != status {
			continue
		}

		id := c.GetId()
		name := c.GetName()
		criteriaType := criteriaTypeString(c.GetCriteriaType())
		scope := criteriaScopeString(c.GetScope())
		version := fmt.Sprintf("v%d", c.GetVersion())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "badge", Value: criteriaType, Variant: criteriaTypeVariant(c.GetCriteriaType())},
				{Type: "text", Value: scope},
				{Type: "text", Value: version},
				{Type: "badge", Value: versionStatus, Variant: versionStatusVariant(c.GetVersionStatus())},
			},
			DataAttrs: map[string]string{
				"name":    name,
				"type":    criteriaType,
				"scope":   scope,
				"version": version,
				"status":  versionStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("outcome_criteria", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: !perms.Can("outcome_criteria", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func versionStatusString(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "draft"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "active"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "inactive"
	default:
		return "draft"
	}
}

func versionStatusVariant(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "default"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "success"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "warning"
	default:
		return "default"
	}
}

func criteriaTypeString(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE:
		return "Numeric Range"
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "Numeric Score"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "Pass/Fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "Categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "Text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "Multi-Check"
	default:
		return "Unspecified"
	}
}

func criteriaTypeVariant(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "success"
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE, enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "info"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "warning"
	default:
		return "default"
	}
}

func criteriaScopeString(s enums.CriteriaScope) string {
	switch s {
	case enums.CriteriaScope_CRITERIA_SCOPE_SYSTEM:
		return "System"
	case enums.CriteriaScope_CRITERIA_SCOPE_INDUSTRY:
		return "Industry"
	case enums.CriteriaScope_CRITERIA_SCOPE_WORKSPACE:
		return "Workspace"
	default:
		return "Unspecified"
	}
}

func statusPageTitle(l fayna.OutcomeCriteriaLabels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l fayna.OutcomeCriteriaLabels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l fayna.OutcomeCriteriaLabels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l fayna.OutcomeCriteriaLabels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
