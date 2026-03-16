package list

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"
	"github.com/erniealice/fayna-golang/utils"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// Deps holds view dependencies.
type Deps struct {
	Routes                     fayna.JobActivityRoutes
	GetJobActivityListPageData func(ctx context.Context, req *jobactivitypb.GetJobActivityListPageDataRequest) (*jobactivitypb.GetJobActivityListPageDataResponse, error)
	Labels                     fayna.JobActivityLabels
	CommonLabels               pyeza.CommonLabels
	TableLabels                types.TableLabels
}

// PageData holds the data for the job activity list page.
type PageData struct {
	types.PageData
	ContentTemplate string
	Table           *types.TableConfig
}

// NewView creates the job activity list view.
func NewView(deps *Deps) view.View {
	return view.ViewFunc(func(ctx context.Context, viewCtx *view.ViewContext) view.ViewResult {
		perms := view.GetUserPermissions(ctx)

		resp, err := deps.GetJobActivityListPageData(ctx, &jobactivitypb.GetJobActivityListPageDataRequest{})
		if err != nil {
			log.Printf("Failed to load job activity list page data: %v", err)
			return view.Error(fmt.Errorf("failed to load activities: %w", err))
		}

		l := deps.Labels
		columns := activityColumns(l)
		rows := buildTableRows(resp.GetJobActivityList(), l, deps.Routes, perms)
		types.ApplyColumnStyles(columns, rows)

		tableConfig := &types.TableConfig{
			ID:                   "activities-table",
			RefreshURL:           deps.Routes.ListURL,
			Columns:              columns,
			Rows:                 rows,
			ShowSearch:           true,
			ShowActions:          true,
			ShowSort:             true,
			ShowColumns:          true,
			ShowDensity:          true,
			ShowEntries:          true,
			DefaultSortColumn:    "date",
			DefaultSortDirection: "desc",
			Labels:               deps.TableLabels,
			EmptyState: types.TableEmptyState{
				Title:   l.Empty.Title,
				Message: l.Empty.Message,
			},
			PrimaryAction: &types.PrimaryAction{
				Label:           l.Buttons.AddActivity,
				ActionURL:       deps.Routes.CreateURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_activity", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "operations",
				ActiveSubNav:   "activities",
				HeaderTitle:    l.Page.Heading,
				HeaderSubtitle: l.Page.Caption,
				HeaderIcon:     "icon-clock",
				CommonLabels:   deps.CommonLabels,
			},
			ContentTemplate: "job-activity-list-content",
			Table:           tableConfig,
		}

		return view.OK("job-activity-list", pageData)
	})
}

func activityColumns(l fayna.JobActivityLabels) []types.TableColumn {
	return []types.TableColumn{
		{Key: "date", Label: l.Columns.Date, Sortable: true, Width: "140px"},
		{Key: "job", Label: l.Columns.Job, Sortable: true},
		{Key: "entry_type", Label: l.Columns.EntryType, Sortable: true, Width: "120px"},
		{Key: "description", Label: l.Columns.Description, Sortable: true},
		{Key: "quantity", Label: l.Columns.Quantity, Sortable: true, Width: "100px", Align: "right"},
		{Key: "amount", Label: l.Columns.Amount, Sortable: true, Width: "140px", Align: "right"},
		{Key: "status", Label: l.Columns.Status, Sortable: true, Width: "120px"},
	}
}

func buildTableRows(activities []*jobactivitypb.JobActivity, l fayna.JobActivityLabels, routes fayna.JobActivityRoutes, perms *types.UserPermissions) []types.TableRow {
	rows := []types.TableRow{}
	for _, a := range activities {
		id := a.GetId()
		date := a.GetEntryDateString()
		entryType := entryTypeString(a.GetEntryType())
		description := a.GetDescription()
		currency := a.GetCurrency()
		quantity := fmt.Sprintf("%.2f", a.GetQuantity())
		amount := utils.FormatCentavoAmount(a.GetTotalCost(), currency)
		approvalStatus := approvalStatusString(a.GetApprovalStatus())

		jobName := ""
		if j := a.GetJob(); j != nil {
			jobName = j.GetName()
		}

		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		actions := []types.TableAction{
			{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
			{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.UpdateURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job_activity", "update"), DisabledTooltip: l.Errors.PermissionDenied},
			{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("job_activity", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
		}

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				types.DateTimeCell(date, types.DateReadable),
				{Type: "text", Value: jobName},
				{Type: "text", Value: entryType},
				{Type: "text", Value: description},
				{Type: "text", Value: quantity},
				{Type: "text", Value: amount},
				{Type: "badge", Value: approvalStatus, Variant: approvalStatusVariant(approvalStatus)},
			},
			DataAttrs: map[string]string{
				"date":        date,
				"job":         jobName,
				"entry_type":  entryType,
				"description": description,
				"quantity":    quantity,
				"amount":      amount,
				"status":      approvalStatus,
			},
			Actions: actions,
		})
	}
	return rows
}

func entryTypeString(t jobactivitypb.EntryType) string {
	switch t {
	case jobactivitypb.EntryType_ENTRY_TYPE_LABOR:
		return "labor"
	case jobactivitypb.EntryType_ENTRY_TYPE_MATERIAL:
		return "material"
	case jobactivitypb.EntryType_ENTRY_TYPE_EXPENSE:
		return "expense"
	default:
		return "unspecified"
	}
}

func approvalStatusString(s jobactivitypb.ActivityApprovalStatus) string {
	switch s {
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_DRAFT:
		return "draft"
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_SUBMITTED:
		return "submitted"
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_APPROVED:
		return "approved"
	case jobactivitypb.ActivityApprovalStatus_ACTIVITY_APPROVAL_STATUS_REJECTED:
		return "rejected"
	default:
		return "draft"
	}
}

func approvalStatusVariant(status string) string {
	switch status {
	case "draft":
		return "default"
	case "submitted":
		return "warning"
	case "approved":
		return "success"
	case "rejected":
		return "danger"
	default:
		return "default"
	}
}
