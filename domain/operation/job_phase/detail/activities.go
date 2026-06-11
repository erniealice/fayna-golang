package detail

import (
	"context"
	"fmt"
	"log"

	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// loadActivitiesTab loads job activities for this phase's Activities tab.
//
// v1 strategy: list all activities for the phase's parent job, then filter
// to those whose job_task_id is non-empty — a proxy for "task-attributed"
// activities belonging to this job. We cannot filter by phase_id directly
// because JobActivity only carries job_task_id (not job_phase_id), and the
// task→phase join is not available without a separate ListJobTasksByPhase call.
//
// TODO(WaveN): replace with a dedicated ListJobActivitiesByPhase RPC once
// espyna exposes it, or extend JobActivity to carry job_phase_id for a fast
// report join (referenced in the milestone-billing plan).
func loadActivitiesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, p *jobphasepb.JobPhase) {
	if deps.ListJobActivities == nil {
		return
	}

	resp, err := deps.ListJobActivities(ctx, &jobactivitypb.ListJobActivitiesRequest{})
	if err != nil {
		log.Printf("Failed to list activities for phase %s: %v", p.GetId(), err)
		return
	}

	jobID := p.GetJobId()
	var activities []*jobactivitypb.JobActivity
	for _, a := range resp.GetData() {
		if a.GetJobId() == jobID && a.GetJobTaskId() != "" {
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
		ID:          "job-phase-activities-table",
		Columns:     columns,
		Rows:        rows,
		ShowSearch:  false,
		ShowActions: false,
		ShowSort:    false,
		Labels:      deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No activities",
			Message: "No activity entries recorded for tasks in this phase.",
		},
	}
}

// activityQuantityDisplay returns a human-readable quantity string for a
// JobActivity row. Uses Quantity for most entry types; falls back to the
// total cost display for expense entries.
func activityQuantityDisplay(a *jobactivitypb.JobActivity) string {
	if q := a.GetQuantity(); q != 0 {
		return fmt.Sprintf("%.2f", q)
	}
	return ""
}
