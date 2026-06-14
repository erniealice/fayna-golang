package list

import (
	"context"
	"fmt"
	"log"

	workrequest "github.com/erniealice/fayna-golang/domain/operation/work_request"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	wrpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/work_request"
)

// workRequestSortableSQLCols is the sort allowlist (SEC-5: NO raw column from
// query string). Every sortable column is enumerated.
var workRequestSortableSQLCols = map[string]string{
	"request_number": "request_number",
	"title":          "title",
	"status":         "status",
	"origin":         "origin",
	"priority":       "priority",
	"sla_due_at":     "sla_due_at",
	"created_at":     "date_created",
}

// ListViewDeps holds view dependencies for the inbox list.
type ListViewDeps struct {
	Routes             workrequest.Routes
	Labels             workrequest.Labels
	CommonLabels       pyeza.CommonLabels
	TableLabels        types.TableLabels
	ListWorkRequests   func(ctx context.Context, req *wrpb.ListWorkRequestsRequest) (*wrpb.ListWorkRequestsResponse, error)
	GetListPageData    func(ctx context.Context, req *wrpb.GetWorkRequestListPageDataRequest) (*wrpb.GetWorkRequestListPageDataResponse, error)
}

// PageData holds the data for the work request inbox list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
	Labels          workrequest.Labels
	StatusTabs      []StatusTab
	ActiveStatus    string
}

// StatusTab represents a status filter tab in the inbox.
type StatusTab struct {
	Key      string
	Label    string
	Href     string
	HxGet    string
	Active   bool
	TestID   string
}

// NewView creates the work request inbox list view.
func NewView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "list") {
			return view.Forbidden("work_request:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "open"
		}

		resp, err := deps.ListWorkRequests(ctx, &wrpb.ListWorkRequestsRequest{})
		if err != nil {
			log.Printf("Failed to list work requests: %v", err)
			return view.Error(fmt.Errorf("failed to load work requests: %w", err))
		}

		l := deps.Labels
		columns := workRequestColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "work-request-list-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sla_due_at",
			DefaultSortDirection: "asc",
			RefreshURL:           route.ResolveURL(deps.Routes.TableURL, "status", status),
			Labels:               deps.TableLabels,

			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Actions.LogRequest,
				ActionURL:       deps.Routes.AddURL + "?origin=2",
				Icon:            "icon-plus",
				Disabled:        !perms.Can("work_request", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
			BulkActions: &types.BulkActionsConfig{
				Enabled:       true,
				SelectAllLabel: "Select all",
				SelectedLabel:  "selected",
				CancelLabel:    "Cancel",
				Actions: []types.BulkAction{
					{
						Key:      "assign",
						Label:    l.Actions.BulkAssign,
						Icon:     "icon-user-plus",
						Variant:  "primary",
						Endpoint: deps.Routes.BulkAssignURL,
					},
				},
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
				HeaderIcon:     "icon-inbox",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "work-request-list-content",
			Table:           tableConfig,
			Labels:          l,
			StatusTabs:      statusTabs,
			ActiveStatus:    status,
		}

		return view.OK("work-request-list", pageData)
	})
}

// NewTableView creates the table-only partial view for HTMX table swaps.
func NewTableView(deps *ListViewDeps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)
		if !perms.Can("work_request", "list") {
			return view.Forbidden("work_request:list")
		}

		status := viewCtx.Request.PathValue("status")
		if status == "" {
			status = "open"
		}

		resp, err := deps.ListWorkRequests(ctx, &wrpb.ListWorkRequestsRequest{})
		if err != nil {
			log.Printf("Failed to list work requests: %v", err)
			return view.Error(fmt.Errorf("failed to load work requests: %w", err))
		}

		l := deps.Labels
		columns := workRequestColumns(l)
		rows := buildTableRows(resp.GetData(), status, l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "work-request-list-table",
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "sla_due_at",
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

func workRequestColumns(l workrequest.Labels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "request_number", Label: l.Columns.RequestNumber},
		{Key: "origin", Label: l.Columns.Origin, WidthClass: "col-3xl"},
		{Key: "client", Label: l.Columns.Client},
		{Key: "raised", Label: l.Columns.Raised, WidthClass: "col-3xl"},
		{Key: "sla", Label: l.Columns.SLA, WidthClass: "col-3xl"},
		{Key: "assigned", Label: l.Columns.Assigned},
		{Key: "status", Label: l.Columns.Status, WidthClass: "col-3xl"},
	}
}

func buildTableRows(items []*wrpb.WorkRequest, status string, l workrequest.Labels, routes workrequest.Routes, perms *types.UserPermissions) []types.TableRow {
	rows := []types.TableRow{}
	for _, wr := range items {
		wrStatus := workRequestStatusString(wr.GetStatus())
		if !matchesStatusTab(wrStatus, status, wr.GetActive()) {
			continue
		}

		id := wr.GetId()
		title := wr.GetTitle()
		reqNum := wr.GetRequestNumber()
		displayName := title
		if reqNum != "" {
			displayName = reqNum + " - " + title
		}

		originLabel := originString(wr.GetOrigin(), l)
		detailURL := route.ResolveURL(routes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: displayName},
				{Type: "badge", Value: originLabel, Variant: originVariant(wr.GetOrigin())},
				{Type: "text", Value: wr.GetClientId()},
				types.DateTimeCell(wr.GetDateCreatedString(), types.DateReadable),
				{Type: "badge", Value: slaLabel(wr, l), Variant: slaVariant(wr)},
				{Type: "text", Value: stringOrDash(wr.GetAssignedToWorkspaceUserId())},
				{Type: "badge", Value: wrStatus, Variant: statusVariant(wrStatus)},
			},
			DataAttrs: map[string]string{
				"request-number": reqNum,
				"title":          title,
				"status":         wrStatus,
				"origin":         originLabel,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: l.Actions.Open, Action: "view", Href: detailURL},
				{Type: "edit", Label: l.Actions.Assign, Action: "assign", URL: route.ResolveURL(routes.AssignURL, "id", id), DrawerTitle: l.Actions.Assign, Disabled: !perms.Can("work_request", "assign"), DisabledTooltip: l.Errors.PermissionDenied},
			},
		})
	}
	return rows
}

func matchesStatusTab(wrStatus string, tabStatus string, active bool) bool {
	switch tabStatus {
	case "open":
		return active // status NOT IN (5,6,7)
	case "all":
		return true
	default:
		return wrStatus == tabStatus
	}
}

func buildStatusTabs(l workrequest.Labels, active string, routes workrequest.Routes) []StatusTab {
	tabs := []struct {
		key   string
		label string
	}{
		{"open", "Open"},
		{"in-review", l.Status.InReview},
		{"returned-for-info", l.Status.ReturnedForInfo},
		{"on-hold", l.Status.OnHold},
		{"escalated", l.Status.Escalated},
		{"pending-override", l.Status.PendingOverride},
		{"approved", l.Status.Approved},
		{"declined", l.Status.Declined},
		{"completed", l.Status.Completed},
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
			TestID: "work-request-tab-" + t.key,
		})
	}
	return result
}

func workRequestStatusString(s wrpb.WorkRequestStatus) string {
	switch s {
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_NEW:
		return "new"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_SUBMITTED:
		return "submitted"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_IN_REVIEW:
		return "in-review"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_APPROVED:
		return "approved"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_DECLINED:
		return "declined"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_COMPLETED:
		return "completed"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_CANCELLED:
		return "cancelled"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_RETURNED_FOR_INFO:
		return "returned-for-info"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ON_HOLD:
		return "on-hold"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_ESCALATED:
		return "escalated"
	case wrpb.WorkRequestStatus_WORK_REQUEST_STATUS_PENDING_OVERRIDE:
		return "pending-override"
	default:
		return "new"
	}
}

func statusVariant(status string) string {
	switch status {
	case "new", "submitted":
		return "default"
	case "in-review":
		return "info"
	case "approved":
		return "success"
	case "declined":
		return "danger"
	case "completed":
		return "success"
	case "cancelled":
		return "default"
	case "returned-for-info":
		return "warning"
	case "on-hold":
		return "secondary"
	case "escalated":
		return "danger"
	case "pending-override":
		return "warning"
	default:
		return "default"
	}
}

func originString(o wrpb.WorkRequestOrigin, l workrequest.Labels) string {
	switch o {
	case wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_CLIENT_ORIGINATED:
		return l.Origin.ClientOriginated
	case wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_CLIENT_RELATED_INTERNAL:
		return l.Origin.ClientRelatedInternal
	case wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_INTERNAL:
		return l.Origin.Internal
	default:
		return ""
	}
}

func originVariant(o wrpb.WorkRequestOrigin) string {
	switch o {
	case wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_CLIENT_ORIGINATED:
		return "info"
	case wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_CLIENT_RELATED_INTERNAL:
		return "secondary"
	case wrpb.WorkRequestOrigin_WORK_REQUEST_ORIGIN_INTERNAL:
		return "default"
	default:
		return "default"
	}
}

func slaLabel(wr *wrpb.WorkRequest, l workrequest.Labels) string {
	if wr.GetSlaDueAt() == 0 {
		return l.SLA.NoSLA
	}
	if wr.GetSlaBreachedAt() != 0 {
		return l.SLA.Breached
	}
	return l.SLA.OnTrack
}

func slaVariant(wr *wrpb.WorkRequest) string {
	if wr.GetSlaDueAt() == 0 {
		return "default"
	}
	if wr.GetSlaBreachedAt() != 0 {
		return "danger"
	}
	return "success"
}

func stringOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
