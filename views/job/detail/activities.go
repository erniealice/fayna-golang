package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
)

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
