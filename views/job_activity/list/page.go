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

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// ListViewDeps holds view dependencies.
type ListViewDeps struct {
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
func NewView(deps *ListViewDeps) view.View {
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
				ActionURL:       deps.Routes.AddURL,
				Icon:            "icon-plus",
				Disabled:        !perms.Can("job_activity", "create"),
				DisabledTooltip: l.Errors.PermissionDenied,
			},
		}
		tableConfig.BulkActions = &types.BulkActionsConfig{
			Enabled:        true,
			SelectAllLabel: "Select all",
			SelectedLabel:  "{count} selected",
			CancelLabel:    "Cancel",
			Actions: []types.BulkAction{
				{
					Key:            "generate-invoice",
					Label:          "Generate Invoice",
					Icon:           "icon-file-text",
					Variant:        "primary",
					Endpoint:       deps.Routes.BulkGenerateInvoiceURL,
					ConfirmTitle:   "Generate Invoice",
					ConfirmMessage: "Generate invoice from {{count}} selected activity(s)?",
				},
			},
		}
		types.ApplyTableSettings(tableConfig)

		pageData := &PageData{
			PageData: types.PageData{
				CacheVersion:   viewCtx.CacheVersion,
				Title:          l.Page.Heading,
				CurrentPath:    viewCtx.CurrentPath,
				ActiveNav:      "job",
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
		{Key: "date", Label: l.Columns.Date, Sortable: true, WidthClass: "col-3xl"},
		{Key: "job", Label: l.Columns.Job, Sortable: true},
		{Key: "entry_type", Label: l.Columns.EntryType, Sortable: true, WidthClass: "col-2xl"},
		{Key: "description", Label: l.Columns.Description, Sortable: true},
		{Key: "quantity", Label: l.Columns.Quantity, Sortable: true, WidthClass: "col-lg", Align: "right"},
		{Key: "amount", Label: l.Columns.Amount, Sortable: true, WidthClass: "col-3xl", Align: "right"},
		{Key: "status", Label: l.Columns.Status, Sortable: true, WidthClass: "col-2xl"},
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
		approvalStatus := approvalStatusString(a.GetApprovalStatus())
		postingStatus := postingStatusString(a.GetPostingStatus())

		jobName := ""
		if j := a.GetJob(); j != nil {
			jobName = j.GetName()
		}

		detailURL := route.ResolveURL(routes.DetailURL, "id", id)
		actions := []types.TableAction{
			{Type: "view", Label: l.Actions.View, Action: "view", Href: detailURL},
		}
		switch approvalStatus {
		case "draft":
			actions = append(actions,
				types.TableAction{Type: "edit", Label: l.Actions.Edit, Action: "edit", URL: route.ResolveURL(routes.EditURL, "id", id), DrawerTitle: l.Actions.Edit, Disabled: !perms.Can("job_activity", "update"), DisabledTooltip: l.Errors.PermissionDenied},
				types.TableAction{
					Type: "check", Label: l.Actions.Submit, Action: "deactivate",
					URL: routes.SubmitURL, ItemName: id,
					ConfirmTitle: l.Actions.Submit, ConfirmMessage: l.Actions.Submit + " " + id + "?",
					Disabled: !perms.Can("job_activity", "update"), DisabledTooltip: l.Errors.PermissionDenied,
				},
				types.TableAction{Type: "delete", Label: l.Actions.Delete, Action: "delete", URL: routes.DeleteURL, ItemName: id, Disabled: !perms.Can("job_activity", "delete"), DisabledTooltip: l.Errors.PermissionDenied},
			)
		case "submitted":
			actions = append(actions,
				types.TableAction{
					Type: "check", Label: l.Actions.Approve, Action: "activate",
					URL: routes.ApproveURL, ItemName: id,
					ConfirmTitle: l.Actions.Approve, ConfirmMessage: l.Actions.Approve + " " + id + "?",
					Disabled: !perms.Can("job_activity", "approve"), DisabledTooltip: l.Errors.PermissionDenied,
				},
				types.TableAction{
					Type: "delete", Label: l.Actions.Reject, Action: "cancel",
					URL: routes.RejectURL, ItemName: id,
					ConfirmTitle: l.Actions.Reject, ConfirmMessage: l.Actions.Reject + " " + id + "?",
					Disabled: !perms.Can("job_activity", "approve"), DisabledTooltip: l.Errors.PermissionDenied,
				},
			)
		case "approved":
			if postingStatus == "unposted" {
				actions = append(actions,
					types.TableAction{
						Type: "check", Label: l.Actions.Post, Action: "activate",
						URL: routes.PostURL, ItemName: id,
						ConfirmTitle: l.Actions.Post, ConfirmMessage: l.Actions.Post + " " + id + "?",
						Disabled: !perms.Can("job_activity", "post") && !perms.Can("job_activity", "manage"), DisabledTooltip: l.Errors.PermissionDenied,
					},
				)
			}
		case "rejected":
			// view only — no additional actions
		}
		if postingStatus == "posted" {
			actions = append(actions,
				types.TableAction{
					Type: "undo", Label: l.Actions.Reverse, Action: "cancel",
					URL: routes.ReverseURL, ItemName: id,
					ConfirmTitle: l.Actions.Reverse, ConfirmMessage: l.Actions.Reverse + " " + id + "?",
					Disabled: !perms.Can("job_activity", "manage"), DisabledTooltip: l.Errors.PermissionDenied,
				},
			)
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
				types.MoneyCell(float64(a.GetTotalCost()), currency, true),
				{Type: "badge", Value: approvalStatus, Variant: approvalStatusVariant(approvalStatus)},
			},
			DataAttrs: map[string]string{
				"date":        date,
				"job":         jobName,
				"entry_type":  entryType,
				"description": description,
				"quantity":    quantity,
				"amount":      fmt.Sprintf("%d", a.GetTotalCost()),
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
	case jobactivitypb.EntryType_ENTRY_TYPE_EQUIPMENT:
		return "Equipment"
	case jobactivitypb.EntryType_ENTRY_TYPE_SUBCONTRACT:
		return "Subcontract"
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

func postingStatusString(s jobactivitypb.ActivityPostingStatus) string {
	switch s {
	case jobactivitypb.ActivityPostingStatus_ACTIVITY_POSTING_STATUS_UNPOSTED:
		return "unposted"
	case jobactivitypb.ActivityPostingStatus_ACTIVITY_POSTING_STATUS_POSTED:
		return "posted"
	case jobactivitypb.ActivityPostingStatus_ACTIVITY_POSTING_STATUS_REVERSED:
		return "reversed"
	default:
		return "unposted"
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
