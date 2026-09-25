package list

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/rating_description_set"
	"github.com/erniealice/fayna-golang/domain/operation/rating_description_set/listdata"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
)

// ListViewDeps holds view dependencies for the rating description set list view.
type ListViewDeps struct {
	Routes rating_description_set.Routes
	// GetListSummaryPageData is the SOLE data source for this page — one
	// espyna page-data use case authorized ONCE under
	// rating_description_set:list, returning every set row already
	// enriched with its scale name and entry/link counts
	// (codex-review-impl3.out.md finding #1 — "the management lists still
	// fail under the prescribed permissions": the previous ordinary
	// ListScoreScales / ListRatingDescriptionSetEntries /
	// ListRatingDescriptionSetProductPlans calls each require a SEPARATE
	// permission — score_scale:list (held by NO role in the clone's
	// catalog) and rating_description_set_product_plan:list (not granted
	// to Section Template Manager, per copya.md's locked grant matrix: set
	// list/read only) — that made this page fail for every principal, or
	// for Section Template Manager specifically). This view calls nothing
	// else.
	GetListSummaryPageData func(ctx context.Context) (*listdata.PageData, error)
	Labels                 rating_description_set.Labels
	CommonLabels           pyeza.CommonLabels
	TableLabels            types.TableLabels
}

// PageData holds the data for the rating description set list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the rating description set list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set", "list") {
			return view.Forbidden("rating_description_set:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		if deps.GetListSummaryPageData == nil {
			log.Printf("rating_description_set list page: GetListSummaryPageData not wired")
			return view.Error(fmt.Errorf("rating description set list page data is not available"))
		}
		data, err := deps.GetListSummaryPageData(ctx)
		if err != nil {
			log.Printf("Failed to load rating description set list page data: %v", err)
			return view.Error(fmt.Errorf("failed to load rating description sets: %w", err))
		}

		var summaryRows []listdata.Row
		if data != nil {
			summaryRows = data.Rows
		}

		l := deps.Labels
		columns := setColumns(l)
		rows := buildTableRows(summaryRows, status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "rating-description-set-table",
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
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.Add,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("rating_description_set", "create"),
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
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-file-text",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "rating-description-set-list-content",
			Table:           tableConfig,
		}

		return view.OK("rating-description-set-list", pageData)
	})
}

func setColumns(l rating_description_set.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "name", Label: l.Columns.Name},
		{Key: "version", Label: l.Columns.Version, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
		{Key: "scale", Label: l.Columns.Scale, WidthClass: "col-4xl"},
		{Key: "entries", Label: l.Columns.Entries, WidthClass: "col-lg"},
		{Key: "links", Label: l.Columns.Links, WidthClass: "col-lg"},
	}
}

func buildTableRows(
	rows []listdata.Row,
	status string,
	l rating_description_set.Labels,
	routes rating_description_set.Routes,
	perms *types.UserPermissions,
) []types.TableRow {
	out := []types.TableRow{}
	for _, r := range rows {
		s := r.Set
		if s == nil {
			continue
		}
		st := statusString(s.GetVersionStatus())
		if status == "active" && st == "deprecated" {
			continue
		}
		if status == "deprecated" && st != "deprecated" {
			continue
		}

		id := s.GetId()
		name := s.GetName()
		version := fmt.Sprintf("v%d", s.GetVersion())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		scaleName := r.ScoreScaleName
		if scaleName == "" {
			scaleName = s.GetScoreScaleId()
		}

		out = append(out, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: version},
				{Type: "badge", Value: statusLabel(l, s.GetVersionStatus()), Variant: statusVariant(s.GetVersionStatus())},
				{Type: "text", Value: scaleName},
				{Type: "text", Value: fmt.Sprintf("%d", r.EntryCount)},
				{Type: "text", Value: fmt.Sprintf("%d", r.LinkCount)},
			},
			DataAttrs: map[string]string{
				"name":    name,
				"version": version,
				"status":  st,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: "View", Action: "view", Href: detailURL},
			},
		})
	}
	return out
}

func statusString(s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return "draft"
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return "published"
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return "deprecated"
	default:
		return "draft"
	}
}

func statusLabel(l rating_description_set.Labels, s enums.VersionStatus) string {
	switch s {
	case enums.VersionStatus_VERSION_STATUS_DRAFT:
		return l.Status.Draft
	case enums.VersionStatus_VERSION_STATUS_PUBLISHED:
		return l.Status.Published
	case enums.VersionStatus_VERSION_STATUS_DEPRECATED:
		return l.Status.Deprecated
	default:
		return l.Status.Draft
	}
}

func statusVariant(s enums.VersionStatus) string {
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

func statusPageTitle(l rating_description_set.Labels, status string) string {
	switch status {
	case "active":
		return l.Page.HeadingActive
	case "deprecated":
		return l.Page.HeadingDeprecated
	default:
		return l.Page.HeadingAll
	}
}
