package list

import (
	"context"
	"fmt"
	"log"
	"math"

	espynahttp "github.com/erniealice/espyna-golang/contrib/http"
	fayna "github.com/erniealice/fayna-golang"
	lynguaV1 "github.com/erniealice/lyngua/golang/v1"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	fulfillmentpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/fulfillment"
)

// StatusTab represents a status filter tab for the fulfillment list page.
type StatusTab struct {
	Key    string
	Label  string
	Href   string
	Active bool
}

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
	Routes                     fayna.FulfillmentRoutes
	Labels                     fayna.FulfillmentLabels
	CommonLabels               pyeza.CommonLabels
	TableLabels                types.TableLabels
	GetFulfillmentListPageData func(ctx context.Context, req *fulfillmentpb.GetFulfillmentListPageDataRequest) (*fulfillmentpb.GetFulfillmentListPageDataResponse, error)
}

// PageData holds the data for the fulfillment list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	StatusTabs      []StatusTab
	ActiveStatus    string
}

var fulfillmentAllowedSortCols = []string{
	"date_created", "date_modified", "status", "fulfillment_method",
}

var fulfillmentSearchFields = []string{"reference_number", "fulfillment_method"}

// NewView creates the fulfillment list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "PENDING"
		}

		p, err := espynahttp.ParseTableParams(viewCtx.Request, fulfillmentAllowedSortCols)
		if err != nil {
			return view.Error(err)
		}

		tableConfig, err := buildTableConfig(ctx, deps, status, p)
		if err != nil {
			return view.Error(err)
		}

		l := deps.Labels
		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Title,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "fulfillment",
				HeaderTitle:    l.Title,
				HeaderSubtitle: l.AppLabel,
				HeaderIcon:     "icon-package",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "fulfillment-list-content",
			Table:           tableConfig,
			StatusTabs:      buildStatusTabs(l, deps.Routes, status),
			ActiveStatus:    status,
		}

		// KB help content
		if viewCtx.Translations != nil {
			if provider, ok := viewCtx.Translations.(*lynguaV1.TranslationProvider); ok {
				if kb, _ := provider.LoadKBIfExists(viewCtx.Lang, viewCtx.BusinessType, "fulfillment"); kb != nil {
					pageData.HasHelp = true
					pageData.HelpContent = kb.Body
				}
			}
		}

		return view.OK("fulfillment-list", pageData)
	})
}

// NewTableView creates a view that returns only the table-card HTML.
// Used as the refresh target after CRUD operations so that only the table
// is swapped (not the entire page content).
func NewTableView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "PENDING"
		}

		p, err := espynahttp.ParseTableParams(viewCtx.Request, fulfillmentAllowedSortCols)
		if err != nil {
			return view.Error(err)
		}

		tableConfig, err := buildTableConfig(ctx, deps, status, p)
		if err != nil {
			return view.Error(err)
		}

		return view.OK("table-card", tableConfig)
	})
}

// buildTableConfig fetches fulfillment data and builds the table configuration.
func buildTableConfig(ctx context.Context, deps *ListViewDeps, status string, p espynahttp.TableQueryParams) (*types.TableConfig, error) {
	perms := view.GetUserPermissions(ctx)

	listParams := espynahttp.ToListParams(p, fulfillmentSearchFields)
	resp, err := deps.GetFulfillmentListPageData(ctx, &fulfillmentpb.GetFulfillmentListPageDataRequest{
		Search:     listParams.Search,
		Filter:     listParams.Filters,
		Sort:       listParams.Sort,
		Pagination: listParams.Pagination,
	})
	if err != nil {
		log.Printf("Failed to list fulfillments: %v", err)
		return nil, fmt.Errorf("failed to load fulfillments: %w", err)
	}

	l := deps.Labels
	columns := fulfillmentColumns(l)
	rows := buildTableRows(resp.GetRows(), status, l, deps.Routes, perms)
	types.ApplyColumnStyles(columns, rows)

	refreshURL := route.ResolveURL(deps.Routes.ListURL, "status", status)

	// Build ServerPagination
	totalRows := int(resp.GetPagination().GetTotalItems())
	sp := &types.ServerPagination{
		Enabled:       true,
		Mode:          "offset",
		CurrentPage:   p.Page,
		PageSize:      p.PageSize,
		TotalRows:     totalRows,
		TotalPages:    int(math.Ceil(float64(totalRows) / float64(p.PageSize))),
		SearchQuery:   p.Search,
		SortColumn:    p.SortColumn,
		SortDirection: p.SortDir,
		FiltersJSON:   p.FiltersRaw,
		PaginationURL: refreshURL,
	}
	sp.BuildDisplay()

	tableConfig := &types.TableConfig{
		ID:                   "fulfillments-table",
		RefreshURL:           refreshURL,
		Columns:              columns,
		Rows:                 rows,
		ShowSearch:           true,
		ShowActions:          true,
		ShowSort:             true,
		ShowColumns:          true,
		ShowDensity:          true,
		ShowEntries:          true,
		DefaultSortColumn:    "date_created",
		DefaultSortDirection: "desc",
		Labels:               deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   l.Empty.Title,
			Message: l.Empty.Message,
		},
		PrimaryAction: &types.PrimaryAction{
			Label:           l.Buttons.AddFulfillment,
			ActionURL:       deps.Routes.AddURL,
			Icon:            "icon-plus",
			Disabled:        !perms.Can("fulfillment", "create"),
			DisabledTooltip: l.Errors.PermissionDenied,
		},
		ServerPagination: sp,
	}
	types.ApplyTableSettings(tableConfig)

	return tableConfig, nil
}

func buildStatusTabs(l fayna.FulfillmentLabels, routes fayna.FulfillmentRoutes, activeStatus string) []StatusTab {
	statuses := []struct {
		key   string
		label string
	}{
		{"PENDING", l.Status.Pending},
		{"READY", l.Status.Ready},
		{"IN_TRANSIT", l.Status.InTransit},
		{"DELIVERED", l.Status.Delivered},
		{"PARTIALLY_DELIVERED", l.Status.PartiallyDelivered},
		{"FAILED", l.Status.Failed},
		{"CANCELLED", l.Status.Cancelled},
	}

	tabs := make([]StatusTab, 0, len(statuses))
	for _, s := range statuses {
		tabs = append(tabs, StatusTab{
			Key:    s.key,
			Label:  s.label,
			Href:   route.ResolveURL(routes.ListURL, "status", s.key),
			Active: s.key == activeStatus,
		})
	}
	return tabs
}

func buildTableRows(rows []*fulfillmentpb.FulfillmentListRow, status string, l fayna.FulfillmentLabels, routes fayna.FulfillmentRoutes, perms *types.UserPermissions) []types.TableRow {
	result := []types.TableRow{}
	for _, row := range rows {
		f := row.GetFulfillment()
		if f == nil {
			continue
		}

		fStatus := f.GetStatus()
		if fStatus != status {
			continue
		}

		id := f.GetId()
		supplierName := row.GetSupplierName()
		method := f.GetFulfillmentMethod()
		itemCount := fmt.Sprintf("%d", row.GetItemCount())
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		result = append(result, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: method},
				{Type: "badge", Value: fStatus, Variant: fulfillmentStatusVariant(fStatus)},
				{Type: "text", Value: supplierName},
				{Type: "text", Value: itemCount},
			},
			DataAttrs: map[string]string{
				"fulfillment_method": method,
				"status":             fStatus,
				"supplier_name":      supplierName,
				"item_count":         itemCount,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Buttons.Edit, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Buttons.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Buttons.Edit, Disabled: !perms.Can("fulfillment", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				{Type: "delete", Label: l.Buttons.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("fulfillment", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return result
}

func fulfillmentStatusVariant(status string) string {
	switch status {
	case "PENDING":
		return "warning"
	case "READY":
		return "info"
	case "IN_TRANSIT":
		return "info"
	case "DELIVERED":
		return "success"
	case "PARTIALLY_DELIVERED":
		return "warning"
	case "FAILED":
		return "danger"
	case "CANCELLED":
		return "default"
	default:
		return "default"
	}
}
