package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

// ActivityRow is the per-activity view-model rendered on the Activities tab.
// Drives the data-testid="job-activity-row" mini-list (above the legacy
// table-card) so phase5 specs 09 (INCLUDED) and 11 (BILLABLE T&M overage)
// can target individual rows by activity-id and assert the billable_status
// badge value.
//
// 2026-04-29 milestone-billing plan §5/§6.
type ActivityRow struct {
	ID                   string
	EntryType            string
	EntryTypeLabel       string
	BillableStatus       string // shorthand: included | billable | non_billable | write_off | unspecified
	BillableStatusLabel  string
	BillableStatusVariant string // badge variant
	Description          string
	Quantity             string
	UnitCost             string
	TotalCost            string
	BillRate             string
	BillAmount           string
	Currency             string
	EditURL              string
}

// loadActivitiesTab populates the PageData with the activities table data.
func loadActivitiesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, jobID string) {
	if deps.ListJobActivities == nil {
		return
	}

	resp, err := deps.ListJobActivities(ctx, &jobactivitypb.ListJobActivitiesRequest{})
	if err != nil {
		log.Printf("Failed to list job activities for job %s: %v", jobID, err)
		return
	}

	// Filter activities by job ID
	var activities []*jobactivitypb.JobActivity
	for _, a := range resp.GetData() {
		if a.GetJobId() == jobID {
			activities = append(activities, a)
		}
	}

	l := deps.Labels
	pageData.ActivitiesTable = buildActivitiesTable(activities, l, deps.TableLabels)

	// 2026-04-29 milestone-billing plan §5/§6 — operator-facing add/edit CTAs
	// are wired via JobActivityRoutes. Empty deps = CTA suppressed (back-compat
	// for callers that haven't wired the cross-module deps yet).
	pageData.JobActivityLabels = deps.JobActivityLabels
	if deps.JobActivityRoutes.AddURL != "" {
		pageData.AddActivityURL = fmt.Sprintf("%s?job_id=%s", deps.JobActivityRoutes.AddURL, jobID)
	}
	if deps.JobActivityRoutes.EditURL != "" {
		pageData.EditActivityURL = deps.JobActivityRoutes.EditURL
	}
	pageData.ActivitiesList = buildActivitiesList(activities, deps.JobActivityRoutes, deps.JobActivityLabels)
}

// buildActivitiesList builds the per-activity view-model used by the
// data-testid-tagged mini-list above the legacy table-card. Each row exposes
// a stable activity_id and badge selectors so phase5 specs 09/11 can assert
// against billable_status.
func buildActivitiesList(
	activities []*jobactivitypb.JobActivity,
	routes fayna.JobActivityRoutes,
	l fayna.JobActivityLabels,
) []ActivityRow {
	rows := make([]ActivityRow, 0, len(activities))
	for _, a := range activities {
		entryType := activityEntryTypeString(a.GetEntryType())
		billable := activityBillableStatusString(a.GetBillableStatus())

		desc := a.GetDescription()

		row := ActivityRow{
			ID:                    a.GetId(),
			EntryType:             entryType,
			EntryTypeLabel:        activityEntryTypeLabel(entryType, l),
			BillableStatus:        billable,
			BillableStatusLabel:   activityBillableStatusLabel(billable, l),
			BillableStatusVariant: activityBillableStatusVariant(billable),
			Description:           desc,
			Quantity:              fmt.Sprintf("%.2f", a.GetQuantity()),
			UnitCost:              fmt.Sprintf("%.2f", float64(a.GetUnitCost())/100),
			TotalCost:             fmt.Sprintf("%.2f", float64(a.GetTotalCost())/100),
			Currency:              a.GetCurrency(),
		}
		if a.BillRate != nil {
			row.BillRate = fmt.Sprintf("%.2f", float64(*a.BillRate)/100)
		}
		if a.BillAmount != nil {
			row.BillAmount = fmt.Sprintf("%.2f", float64(*a.BillAmount)/100)
		}
		if routes.EditURL != "" {
			row.EditURL = route.ResolveURL(routes.EditURL, "id", row.ID)
		}
		rows = append(rows, row)
	}
	return rows
}

// activityEntryTypeLabel resolves the translated entry-type display string
// with a fallback to the proto-shorthand if the lyngua key is missing.
func activityEntryTypeLabel(entryType string, l fayna.JobActivityLabels) string {
	switch entryType {
	case "labor":
		if l.Form.EntryTypeLabor != "" {
			return l.Form.EntryTypeLabor
		}
		return "Labor"
	case "material":
		if l.Form.EntryTypeMaterial != "" {
			return l.Form.EntryTypeMaterial
		}
		return "Material"
	case "expense":
		if l.Form.EntryTypeExpense != "" {
			return l.Form.EntryTypeExpense
		}
		return "Expense"
	default:
		return entryType
	}
}

// activityBillableStatusString converts a BillableStatus enum to a stable
// shorthand used by data-testid selectors: included | billable | non_billable
// | write_off | unspecified.
func activityBillableStatusString(s jobactivitypb.BillableStatus) string {
	switch s {
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_BILLABLE:
		return "billable"
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_NON_BILLABLE:
		return "non_billable"
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_INCLUDED:
		return "included"
	case jobactivitypb.BillableStatus_BILLABLE_STATUS_WRITE_OFF:
		return "write_off"
	default:
		return "unspecified"
	}
}

// activityBillableStatusLabel resolves the translated badge text for the
// billable_status, falling back to title-case shorthand when the lyngua key
// is missing.
func activityBillableStatusLabel(status string, l fayna.JobActivityLabels) string {
	switch status {
	case "included":
		if l.Form.BillableStatusIncluded != "" {
			return l.Form.BillableStatusIncluded
		}
		return "Included"
	case "billable":
		if l.Form.BillableStatusBillable != "" {
			return l.Form.BillableStatusBillable
		}
		return "Billable"
	case "non_billable":
		if l.Form.BillableStatusNonBillable != "" {
			return l.Form.BillableStatusNonBillable
		}
		return "Non-billable"
	case "write_off":
		return "Write-off"
	default:
		return "Unspecified"
	}
}

// activityBillableStatusVariant returns the badge variant for a billable
// status shorthand.
func activityBillableStatusVariant(status string) string {
	switch status {
	case "included":
		return "info"
	case "billable":
		return "success"
	case "non_billable":
		return "default"
	case "write_off":
		return "warning"
	default:
		return "default"
	}
}

// buildActivitiesTable builds the activities table config.
func buildActivitiesTable(
	activities []*jobactivitypb.JobActivity,
	l fayna.JobLabels,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "entry_type", Label: l.Detail.EntryType, Sortable: true, WidthClass: "col-xl"},
		{Key: "entry_date", Label: l.Detail.EntryDate, Sortable: true, WidthClass: "col-3xl"},
		{Key: "description", Label: l.Detail.Description, Sortable: false},
		{Key: "quantity", Label: l.Detail.Quantity, Sortable: false, WidthClass: "col-md"},
		{Key: "unit_cost", Label: l.Detail.UnitCost, Sortable: false, WidthClass: "col-2xl"},
		{Key: "total_cost", Label: l.Detail.TotalCost, Sortable: false, WidthClass: "col-2xl"},
	}

	rows := []types.TableRow{}
	for _, a := range activities {
		id := a.GetId()
		entryType := activityEntryTypeString(a.GetEntryType())
		entryDate := a.GetEntryDateString()
		description := a.GetDescription()
		quantity := fmt.Sprintf("%.2f", a.GetQuantity())
		currency := a.GetCurrency()

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "badge", Value: entryType, Variant: activityEntryTypeVariant(entryType)},
				types.DateTimeCell(entryDate, types.DateReadable),
				{Type: "text", Value: description},
				{Type: "text", Value: quantity},
				types.MoneyCell(float64(a.GetUnitCost()), currency, true),
				types.MoneyCell(float64(a.GetTotalCost()), currency, true),
			},
			DataAttrs: map[string]string{
				"entry_type": entryType,
				"entry_date": entryDate,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "activities-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: true,
		ShowSort:   true,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No activities",
			Message: "No activity entries have been recorded for this job yet.",
		},
	}
}

// activityEntryTypeString converts an EntryType enum to a display string.
func activityEntryTypeString(t jobactivitypb.EntryType) string {
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
		return "labor"
	}
}

// activityEntryTypeVariant returns the badge variant for an entry type string.
func activityEntryTypeVariant(entryType string) string {
	switch entryType {
	case "labor":
		return "info"
	case "material":
		return "warning"
	case "expense":
		return "danger"
	case "Equipment":
		return "secondary"
	case "Subcontract":
		return "secondary"
	default:
		return "default"
	}
}
