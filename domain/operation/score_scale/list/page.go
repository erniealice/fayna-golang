package list

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/score_scale"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
)

// ListViewDeps holds view dependencies for the score scale list view.
type ListViewDeps struct {
	Routes          score_scale.Routes
	ListScoreScales func(ctx context.Context, req *scalepb.ListScoreScalesRequest) (*scalepb.ListScoreScalesResponse, error)
	Labels          score_scale.Labels
	CommonLabels    pyeza.CommonLabels
	TableLabels     types.TableLabels
}

// PageData holds the data for the score scale list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the score scale list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("score_scale", "list") {
			return view.Forbidden("score_scale:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListScoreScales(ctx, &scalepb.ListScoreScalesRequest{})
		if err != nil {
			log.Printf("Failed to list score scales: %v", err)
			return view.Error(fmt.Errorf("failed to load score scales: %w", err))
		}

		l := deps.Labels
		columns := scaleColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "score-scale-table",
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
				Label:           l.Buttons.AddScoreScale,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("score_scale", "create"),
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
			ContentTemplate: "score-scale-list-content",
			Table:           tableConfig,
		}

		return view.OK("score-scale-list", pageData)
	})
}

func scaleColumns(l score_scale.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "scale_kind", Label: l.Columns.ScaleKind, WidthClass: "col-4xl"},
		{Key: "version_status", Label: l.Columns.VersionStatus, WidthClass: "col-3xl"},
		{Key: "version", Label: l.Columns.Version, WidthClass: "col-lg"},
		{Key: "input_unit", Label: l.Columns.InputUnit, WidthClass: "col-3xl"},
		{Key: "output_unit", Label: l.Columns.OutputUnit, WidthClass: "col-3xl"},
	}
}

func buildTableRows(
	items []*scalepb.ScoreScale,
	status string,
	l score_scale.Labels,
	routes score_scale.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	rows := []types.TableRow{}
	for _, s := range items {
		vs := versionStatusString(s.GetVersionStatus())
		if vs != status {
			continue
		}

		id := s.GetId()
		name := s.GetName()
		kindStr := scaleKindString(s.GetScaleKind())
		version := fmt.Sprintf("v%d", s.GetVersion())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "badge", Value: kindStr, Variant: scaleKindVariant(s.GetScaleKind())},
				{Type: "badge", Value: vs, Variant: versionStatusVariant(s.GetVersionStatus())},
				{Type: "text", Value: version},
				{Type: "text", Value: s.GetInputUnit()},
				{Type: "text", Value: s.GetOutputUnit()},
			},
			DataAttrs: map[string]string{
				"name":           name,
				"scale_kind":     kindStr,
				"version_status": vs,
				"version":        version,
				"input_unit":     s.GetInputUnit(),
				"output_unit":    s.GetOutputUnit(),
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("score_scale", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: name, Disabled: !perms.Can("score_scale", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
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

func scaleKindString(k enums.ScaleKind) string {
	switch k {
	case enums.ScaleKind_SCALE_KIND_RANGE_MAP:
		return "Range Map"
	case enums.ScaleKind_SCALE_KIND_EXACT_MAP:
		return "Exact Map"
	default:
		return "Unspecified"
	}
}

func scaleKindVariant(k enums.ScaleKind) string {
	switch k {
	case enums.ScaleKind_SCALE_KIND_RANGE_MAP:
		return "info"
	case enums.ScaleKind_SCALE_KIND_EXACT_MAP:
		return "default"
	default:
		return "default"
	}
}

func statusPageTitle(l score_scale.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "inactive":
		return l.Page.HeadingInactive
	default:
		return l.Page.Heading
	}
}

func statusPageCaption(l score_scale.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.CaptionActive
	case "inactive":
		return l.Page.CaptionInactive
	default:
		return l.Page.Caption
	}
}

func emptyTitle(l score_scale.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveTitle
	case "inactive":
		return l.Empty.InactiveTitle
	default:
		return l.Empty.Title
	}
}

func emptyMessage(l score_scale.Labels, status string) string {
	switch status {
	case "active":
		return l.Empty.ActiveMessage
	case "inactive":
		return l.Empty.InactiveMessage
	default:
		return l.Empty.Message
	}
}
