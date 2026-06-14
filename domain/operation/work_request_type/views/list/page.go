package list

import (
	"context"
	"fmt"
	"log"

	workrequesttype "github.com/erniealice/fayna-golang/domain/operation/work_request_type"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	wrtpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request_type"
)

// workRequestTypeSortableSQLCols is the sort allowlist (SEC-5: NO raw column from
// query string). Every sortable column is enumerated.
var workRequestTypeSortableSQLCols = map[string]string{
	"code":              "code",
	"name":              "label_key",
	"category":          "category",
	"default_sla_hours": "default_sla_hours",
	"status":            "status",
	"sort_order":        "sort_order",
}

// ListViewDeps holds view dependencies for the catalog list.
type ListViewDeps struct {
	Routes               workrequesttype.Routes
	Labels               workrequesttype.Labels
	CommonLabels         pyeza.CommonLabels
	TableLabels          types.TableLabels
	ListWorkRequestTypes func(ctx context.Context, req *wrtpb.ListWorkRequestTypesRequest) (*wrtpb.ListWorkRequestTypesResponse, error)
}

// PageData holds the data for the work request type catalog list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          workrequesttype.Labels
	StatusTabs      []StatusTab
	ActiveStatus    string
}

// StatusTab represents a status filter tab in the catalog.
type StatusTab struct {
	Key    string
	Label  string
	Href   string
	HxGet  string
	Active bool
	TestID string
}

// NewView creates the work request type catalog list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request_type", "list") {
			return view.Forbidden("work_request_type:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListWorkRequestTypes(ctx, &wrtpb.ListWorkRequestTypesRequest{})
		if err != nil {
			log.Printf("Failed to list work request types: %v", err)
			return view.Error(fmt.Errorf("failed to load request types: %w", err))
		}

		l := deps.Labels
		columns := workRequestTypeColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "work-request-type-list-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sort_order",
			DefaultSortDirection: "asc",
			RefreshURL:           route.ResolveURL(deps.Routes.TableURL, "status", status),
			Labels:               deps.TableLabels,

			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Actions.Add,
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("work_request_type", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		statusTabs := buildStatusTabs(l, status, deps.Routes)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      deps.Routes.ActiveNav,
				ActiveSubNav:   deps.Routes.ActiveSubNav,
				HeaderTitle:    l.Page.Heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-list",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "work-request-type-list-content",
			Table:           tableConfig,
			Labels:          l,
			StatusTabs:      statusTabs,
			ActiveStatus:    status,
		}

		return view.OK("work-request-type-list", pageData)
	})
}

// NewTableView creates the table-only partial view for HTMX table swaps.
func NewTableView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request_type", "list") {
			return view.Forbidden("work_request_type:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "active"
		}

		resp, err := deps.ListWorkRequestTypes(ctx, &wrtpb.ListWorkRequestTypesRequest{})
		if err != nil {
			log.Printf("Failed to list work request types: %v", err)
			return view.Error(fmt.Errorf("failed to load request types: %w", err))
		}

		l := deps.Labels
		columns := workRequestTypeColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "work-request-type-list-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sort_order",
			DefaultSortDirection: "asc",
			RefreshURL:           route.ResolveURL(deps.Routes.TableURL, "status", status),
			Labels:               deps.TableLabels,

			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
		}
		types.ApplyTableSettings(tableConfig)

		return view.OK("table-card", tableConfig)
	})
}

func workRequestTypeColumns(l workrequesttype.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "code", Label: l.Columns.Code},
		{Key: "name", Label: l.Columns.Name},
		{Key: "category", Label: l.Columns.Category, WidthClass: "col-3xl"},
		{Key: "default_sla_hours", Label: l.Columns.DefaultSLAHours, WidthClass: "col-3xl"},
		{Key: "sort_order", Label: l.Columns.SortOrder, WidthClass: "col-3xl"},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(items []*wrtpb.WorkRequestType, status string, l workrequesttype.Labels, routes workrequesttype.Routes, perms *types.UserPermissions) []types.TableRow {
	rows := []types.TableRow{}
	for _, wrt := range items {
		wrtStatus := workRequestTypeStatusString(wrt.GetStatus())
		if !matchesStatusTab(wrtStatus, status, wrt.GetActive()) {
			continue
		}

		id := wrt.GetId()
		code := wrt.GetCode()
		labelKey := wrt.GetLabelKey()
		categoryLabel := categoryString(wrt.GetCategory(), l)
		slaHours := fmt.Sprintf("%d", wrt.GetDefaultSlaHours())
		sortOrder := fmt.Sprintf("%d", wrt.GetSortOrder())

		// Build row actions
		actions := []types.TableAction{
			{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("work_request_type", "update"), DisabledTooltip: l.Errors.PermissionDenied},
		}

		// Archive/Unarchive action
		if wrt.GetActive() {
			actions = append(actions, types.TableAction{
				Type: "archive", Label: l.Actions.Archive, Action: "archive", URL: route.ResolveURL(routes.EditURL, "id", id),
				Disabled: !perms.Can("work_request_type", "update"), DisabledTooltip: l.Errors.PermissionDenied,
			})
		} else {
			actions = append(actions, types.TableAction{
				Type: "unarchive", Label: l.Actions.Unarchive, Action: "unarchive", URL: route.ResolveURL(routes.EditURL, "id", id),
				Disabled: !perms.Can("work_request_type", "update"), DisabledTooltip: l.Errors.PermissionDenied,
			})
		}

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: code},
				{Type: "text", Value: labelKey},
				{Type: "badge", Value: categoryLabel, Variant: categoryVariant(wrt.GetCategory())},
				{Type: "number", Value: slaHours},
				{Type: "number", Value: sortOrder},
				{Type: "badge", Value: wrtStatus, Variant: statusVariant(wrtStatus)},
			},
			DataAttrs: map[string]string{
				"code":     code,
				"category": categoryLabel,
				"status":   wrtStatus,
			},
			Actions: actions,
		})
	}
	return rows
}

func matchesStatusTab(wrtStatus string, tabStatus string, active bool) bool {
	switch tabStatus {
	case "active":
		return active
	case "archived":
		return !active
	case "all":
		return true
	default:
		return wrtStatus == tabStatus
	}
}

func buildStatusTabs(l workrequesttype.Labels, active string, routes workrequesttype.Routes) []StatusTab {
	tabs := []struct {
		key   string
		label string
	}{
		{"active", l.Status.Active},
		{"archived", l.Status.Archived},
		{"all", "All"},
	}

	result := make([]StatusTab, 0, len(tabs))
	for _, t := range tabs {
		listURL := route.ResolveURL(routes.ListURL, "status", t.key)
		tableURL := route.ResolveURL(routes.TableURL, "status", t.key)
		result = append(result, StatusTab{
			Key:    t.key,
			Label:  t.label,
			Href:   listURL,
			HxGet:  tableURL,
			Active: t.key == active,
			TestID: "work-request-type-tab-" + t.key,
		})
	}
	return result
}

func workRequestTypeStatusString(s wrtpb.WorkRequestTypeStatus) string {
	switch s {
	case wrtpb.WorkRequestTypeStatus_WORK_REQUEST_TYPE_STATUS_ACTIVE:
		return "active"
	case wrtpb.WorkRequestTypeStatus_WORK_REQUEST_TYPE_STATUS_ARCHIVED:
		return "archived"
	default:
		return "active"
	}
}

func statusVariant(status string) string {
	switch status {
	case "active":
		return "success"
	case "archived":
		return "default"
	default:
		return "default"
	}
}

func categoryString(c wrtpb.WorkRequestTypeCategory, l workrequesttype.Labels) string {
	switch c {
	case wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_PERSON_SCOPED:
		return l.Category.PersonScoped
	case wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_ACCOUNT_SCOPED:
		return l.Category.AccountScoped
	default:
		return ""
	}
}

func categoryVariant(c wrtpb.WorkRequestTypeCategory) string {
	switch c {
	case wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_PERSON_SCOPED:
		return "info"
	case wrtpb.WorkRequestTypeCategory_WORK_REQUEST_TYPE_CATEGORY_ACCOUNT_SCOPED:
		return "secondary"
	default:
		return "default"
	}
}
