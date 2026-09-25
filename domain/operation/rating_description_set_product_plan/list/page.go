package list

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan"
	linkdata "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan/listdata"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"
)

// ListViewDeps holds view dependencies for the AY setup ("Descriptor
// Assignments") list view.
type ListViewDeps struct {
	Routes rating_description_set_product_plan.Routes
	// GetListSummaryPageData is the SOLE data source for this page — one
	// espyna page-data use case authorized ONCE under
	// rating_description_set_product_plan:list, returning every AY option,
	// the selected/default AY, and that AY's links already enriched with
	// the offering name (scoped through the offering's PARENT product AND
	// plan workspace, not a generic cross-workspace list — finding #3) and
	// the target set's name/version at ANY status
	// (codex-review-impl3.out.md finding #1). This view calls nothing
	// else.
	GetListSummaryPageData func(ctx context.Context, priceScheduleID string) (*linkdata.PageData, error)
	Labels                 rating_description_set_product_plan.Labels
	CommonLabels           pyeza.CommonLabels
	TableLabels            types.TableLabels
}

// ScheduleOption is one academic-year selector entry.
type ScheduleOption struct {
	ID       string
	Name     string
	Selected bool
}

// PageData holds the data for the AY setup list page.
type PageData struct {
	types.PageData
	ContentTemplate    string
	Table              *types.TableConfig
	ScheduleOptions    []ScheduleOption
	SelectedScheduleID string
	ListBaseURL        string
	// Labels — templates/list.html's schedule-selector branch reads
	// .Labels.Filters.PriceSchedule directly off PageData (not through
	// Table) — mirrors rating_description_set/list's PageData.
	Labels rating_description_set_product_plan.Labels
}

// NewView creates the AY setup list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("rating_description_set_product_plan", "list") {
			return view.Forbidden("rating_description_set_product_plan:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "linked"
		}

		l := deps.Labels
		listBase := "/rating-description-set-links/list/" + status
		if deps.Routes.ListURL != "" {
			listBase = deps.Routes.ListURL
		}

		requestedScheduleID := viewCtx.Request.URL.Query().Get("price_schedule_id")

		if deps.GetListSummaryPageData == nil {
			log.Printf("rating_description_set_product_plan list page: GetListSummaryPageData not wired")
			return view.Error(fmt.Errorf("descriptor assignment page data is not available"))
		}
		data, err := deps.GetListSummaryPageData(ctx, requestedScheduleID)
		if err != nil {
			log.Printf("Failed to load descriptor assignments page data: %v", err)
			return view.Error(fmt.Errorf("failed to load descriptor assignments: %w", err))
		}

		scheduleID := data.SelectedScheduleID
		if scheduleID == "" {
			scheduleID = requestedScheduleID
		}

		scheduleOptions := make([]ScheduleOption, 0, len(data.Schedules))
		for _, s := range data.Schedules {
			scheduleOptions = append(scheduleOptions, ScheduleOption{ID: s.ID, Name: s.Name, Selected: s.ID == scheduleID})
		}

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion: viewCtx.CacheVersion,
				Title:        l.Page.Heading,
				CurrentPath:  viewCtx.CurrentPath,
				ActiveNav:    deps.Routes.ActiveNav,
				ActiveSubNav: deps.Routes.ActiveSubNav,
				HeaderTitle:  l.Page.Heading,
				HeaderIcon:   "icon-link",
				CommonLabels: deps.CommonLabels,
			},
			ContentTemplate:    "rating-description-set-link-list-content",
			ScheduleOptions:    scheduleOptions,
			SelectedScheduleID: scheduleID,
			ListBaseURL:        listBase,
			Labels:             l,
		}

		if scheduleID == "" {
			pageData.Table = emptyTable(l, true)
			return view.OK("rating-description-set-link-list", pageData)
		}

		pageData.Table = buildTable(data.Links, l, deps.Routes, scheduleID, perms, deps.TableLabels)

		return view.OK("rating-description-set-link-list", pageData)
	})
}

func emptyTable(l rating_description_set_product_plan.Labels, noSchedule bool) *types.TableConfig {
	title, msg := l.Empty.Title, l.Empty.Message
	if noSchedule {
		title, msg = l.Empty.NoScheduleTitle, l.Empty.NoScheduleMsg
	}
	return &types.TableConfig{
		ID:      "rating-description-set-link-table",
		Columns: []types.TableColumn{{Key: "offering", Label: l.Columns.Offering}},
		Rows:    []types.TableRow{},
		EmptyState: types.TableEmptyState{
			Title:   title,
			Message: msg,
		},
	}
}

func buildTable(
	rows []linkdata.LinkRow,
	l rating_description_set_product_plan.Labels,
	routes rating_description_set_product_plan.Routes,
	scheduleID string,
	perms *types.UserPermissions,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "offering", Label: l.Columns.Offering},
		{Key: "set", Label: l.Columns.Set},
		{Key: "version", Label: l.Columns.Version, WidthClass: "col-lg"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}

	tableRows := make([]types.TableRow, 0, len(rows))
	for _, row := range rows {
		link := row.Link
		if link == nil || !link.GetActive() {
			continue
		}
		offeringName := row.ProductPlanName
		if offeringName == "" {
			offeringName = link.GetProductPlanId()
		}

		relinkURL := fmt.Sprintf("%s?product_plan_id=%s&price_schedule_id=%s&expected_current_link_id=%s",
			routes.RelinkURL, link.GetProductPlanId(), scheduleID, link.GetId())
		unlinkURL := fmt.Sprintf("%s?link_id=%s", routes.UnlinkURL, link.GetId())

		tableRows = append(tableRows, types.TableRow{
			ID: link.GetId(),
			Cells: []types.TableCell{
				{Type: "text", Value: offeringName},
				{Type: "text", Value: row.Set.Name},
				{Type: "text", Value: fmt.Sprintf("v%d", row.Set.Version)},
				{Type: "badge", Value: l.Status.Linked, Variant: "success"},
			},
			DataAttrs: map[string]string{
				"testid": "rdl-row-" + link.GetProductPlanId(),
			},
			Actions: []types.TableAction{
				{Type: "edit", Label: l.Actions.Relink, Action: "edit", URL: relinkURL, DrawerTitle: l.Actions.Relink, Disabled: !perms.Can("rating_description_set_product_plan", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Actions.Unlink, Action: "delete", URL: unlinkURL, ItemName: offeringName, Disabled: !perms.Can("rating_description_set_product_plan", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}

	types.ApplyColumnStyles(columns, tableRows)

	tableConfig := &types.TableConfig{
		ID:          "rating-description-set-link-table",
		Columns:     columns,
		Rows:        tableRows,
		ShowSearch:  true,
		ShowActions: true,
		ShowEntries: true,
		Labels:      tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
		PrimaryAction: &types.PrimaryAction{
			Label:           l.Actions.Link,
			ActionURL:       fmt.Sprintf("%s?price_schedule_id=%s", routes.RelinkURL, scheduleID),
			Icon:            "icon-plus",
			Disabled:        !perms.Can("rating_description_set_product_plan", "create"),
			DisabledTooltip: l.Errors.PermissionDenied,
		},
	}
	types.ApplyTableSettings(tableConfig)
	return tableConfig
}
