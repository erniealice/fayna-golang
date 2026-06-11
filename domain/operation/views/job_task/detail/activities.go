package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// loadActivitiesTab loads job activities for this task's Activities tab.
//
// Filters all activities to those whose job_task_id matches this task's id.
func loadActivitiesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, t *jobtaskpb.JobTask) {
	if deps.ListJobActivities == nil {
		return
	}

	resp, err := deps.ListJobActivities(ctx, &jobactivitypb.ListJobActivitiesRequest{})
	if err != nil {
		log.Printf("Failed to list activities for task %s: %v", t.GetId(), err)
		return
	}

	taskID := t.GetId()
	var activities []*jobactivitypb.JobActivity
	for _, a := range resp.GetData() {
		if a.GetJobTaskId() == taskID {
			activities = append(activities, a)
		}
	}

	columns := []types.TableColumn{
		{Key: "date", Label: "Date", WidthClass: "col-3xl"},
		{Key: "entry_type", Label: "Type", WidthClass: "col-3xl"},
		{Key: "description", Label: "Description"},
		{Key: "quantity", Label: "Qty", WidthClass: "col-2xl"},
		{Key: "approval_status", Label: "Approval", WidthClass: "col-3xl"},
	}

	rows := []types.TableRow{}
	for _, a := range activities {
		entryType := a.GetEntryType().String()
		approvalStatus := a.GetApprovalStatus().String()
		rows = append(rows, types.TableRow{
			ID: a.GetId(),
			Cells: []types.TableCell{
				types.DateTimeCell(a.GetEntryDateString(), types.DateReadable),
				{Type: "text", Value: entryType},
				{Type: "text", Value: a.GetDescription()},
				{Type: "text", Value: activityQuantityDisplay(a)},
				{Type: "badge", Value: approvalStatus, Variant: "default"},
			},
			DataAttrs: map[string]string{
				"entry_type":      entryType,
				"approval_status": approvalStatus,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	pageData.ActivitiesTable = &types.TableConfig{
		ID:          "job-task-activities-table",
		Columns:     columns,
		Rows:        rows,
		ShowSearch:  false,
		ShowActions: false,
		ShowSort:    false,
		Labels:      deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No activities",
			Message: "No activity entries recorded for this task.",
		},
	}
}

// activityQuantityDisplay returns a human-readable quantity string for a
// JobActivity row.
func activityQuantityDisplay(a *jobactivitypb.JobActivity) string {
	if q := a.GetQuantity(); q != 0 {
		return fmt.Sprintf("%.2f", q)
	}
	return ""
}
