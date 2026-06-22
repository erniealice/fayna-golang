package list

import (
	"context"
	"fmt"
	"log"

	scoring_scheme "github.com/erniealice/fayna-golang/domain/operation/scoring_scheme"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	schemepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/scoring_scheme"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes             scoring_scheme.Routes
	ListScoringSchemes func(ctx context.Context, req *schemepb.ListScoringSchemesRequest) (*schemepb.ListScoringSchemesResponse, error)
	Labels             scoring_scheme.Labels
	CommonLabels       pyeza.CommonLabels
	TableLabels        types.TableLabels
}

// PageData holds the data for the scoring scheme list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the scoring scheme list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("scoring_scheme", "list") {
			return view.Forbidden("scoring_scheme:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListScoringSchemes(ctx, &schemepb.ListScoringSchemesRequest{})
		if err != nil {
			log.Printf("Failed to list scoring schemes: %v", err)
			return view.Error(fmt.Errorf("failed to load scoring schemes: %w", err))
		}

		l := deps.Labels
		columns := schemeColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "scoring-scheme-table",
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
				Label:           l.Buttons.AddScheme,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("scoring_scheme", "create"),
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
				HeaderIcon:     "icon-sliders",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "scoring-scheme-list-content",
			Table:           tableConfig,
		}

		return view.OK("scoring-scheme-list", pageData)
	})
}

func schemeColumns(l scoring_scheme.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "composite_method", Label: l.Columns.CompositeMethod, WidthClass: "col-4xl"},
		{Key: "version", Label: l.Columns.Version, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*schemepb.ScoringScheme,
	status string,
	l scoring_scheme.Labels,
	routes scoring_scheme.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, s := range items {
		vStatus := versionStatusString(s.GetVersionStatus())
		if vStatus != status {
			continue
		}

		id := s.GetId()
		name := s.GetName()
		compositeMethod := compositeMethodString(s.GetCompositeMethod())
		version := fmt.Sprintf("v%d", s.GetVersion())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "badge", Value: compositeMethod, Variant: compositeMethodVariant(s.GetCompositeMethod())},
				{Type: "text", Value: version},
				{Type: "badge", Value: vStatus, Variant: versionStatusVariant(s.GetVersionStatus())},
			},
			DataAttrs: map[string]string{
				"name":             name,
				"composite_method": compositeMethod,
				"version":          version,
				"status":           vStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("scoring_scheme", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: !perms.Can("scoring_scheme", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
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

func compositeMethodString(m enums.ScoringMethod) string {
	switch m {
	case enums.ScoringMethod_SCORING_METHOD_EQUAL_WEIGHT:
		return "Equal Weight"
	case enums.ScoringMethod_SCORING_METHOD_WEIGHTED_AVERAGE:
		return "Weighted Average"
	case enums.ScoringMethod_SCORING_METHOD_MINIMUM_DETERMINATION:
		return "Minimum Determination"
	case enums.ScoringMethod_SCORING_METHOD_PERCENTAGE_PASS:
		return "Percentage Pass"
	case enums.ScoringMethod_SCORING_METHOD_SUM:
		return "Sum"
	default:
		return "Unspecified"
	}
}

func compositeMethodVariant(m enums.ScoringMethod) string {
	switch m {
	case enums.ScoringMethod_SCORING_METHOD_WEIGHTED_AVERAGE:
		return "info"
	case enums.ScoringMethod_SCORING_METHOD_EQUAL_WEIGHT:
		return "success"
	case enums.ScoringMethod_SCORING_METHOD_MINIMUM_DETERMINATION:
		return "warning"
	default:
		return "default"
	}
}

func statusPageTitle(l scoring_scheme.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l scoring_scheme.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l scoring_scheme.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l scoring_scheme.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
