package detail

import (
	"context"
	"fmt"
	"log"

	operation "github.com/erniealice/fayna-golang/domain/operation"
	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// loadTasksTab loads tasks for this phase's Tasks tab.
//
// Calls ListJobTasksByPhase RPC (or in-memory filter when unavailable) and
// renders each row as a deep-link into the job_task module detail page.
// The "+ Add Task" CTA opens the job_task add drawer prefilled with this
// phase's ID via the ?job_phase_id= query parameter.
func loadTasksTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, phaseID string) {
	if deps.ListJobTasksByPhase == nil {
		// No use case wired yet — render empty state without error.
		return
	}

	resp, err := deps.ListJobTasksByPhase(ctx, &jobtaskpb.ListJobTasksByPhaseRequest{
		JobPhaseId: phaseID,
	})
	if err != nil {
		log.Printf("Failed to list tasks for phase %s: %v", phaseID, err)
		return
	}

	taskRoutes := operation.DefaultJobTaskRoutes()
	addTaskURL := taskRoutes.AddURL + "?job_phase_id=" + phaseID

	columns := []types.TableColumn{
		{Key: "name", Label: "Name"},
		{Key: "step_order", Label: "#", WidthClass: "col-sm"},
		{Key: "status", Label: "Status", WidthClass: "col-3xl"},
		{Key: "assigned_to", Label: "Assigned To", WidthClass: "col-4xl"},
		{Key: "percent_complete", Label: "% Done", WidthClass: "col-3xl"},
	}

	rows := []types.TableRow{}
	for _, t := range resp.GetJobTasks() {
		id := t.GetId()
		name := t.GetName()
		taskStatus := taskStatusString(t.GetStatus())
		assignedTo := ""
		if t.AssignedTo != nil {
			assignedTo = *t.AssignedTo
		}
		percentComplete := float64(0)
		if t.PercentComplete != nil {
			percentComplete = *t.PercentComplete
		}

		detailURL := route.ResolveURL(taskRoutes.DetailURL, "id", id)

		rows = append(rows, types.TableRow{
			ID:   id,
			Href: detailURL,
			Cells: []types.TableCell{
				{Type: "text", Value: name},
				{Type: "text", Value: fmt.Sprintf("%d", t.GetStepOrder())},
				{Type: "badge", Value: taskStatus, Variant: taskStatusVariant(taskStatus)},
				{Type: "text", Value: assignedTo},
				{Type: "text", Value: fmt.Sprintf("%.0f%%", percentComplete)},
			},
			DataAttrs: map[string]string{
				"name":   name,
				"status": taskStatus,
			},
			Actions: []types.TableAction{
				{Type: "view", Label: "View Task", Action: "view", Href: detailURL},
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	pageData.TasksTable = &types.TableConfig{
		ID:          "job-phase-tasks-table",
		Columns:     columns,
		Rows:        rows,
		ShowSearch:  false,
		ShowActions: true,
		ShowSort:    false,
		Labels:      deps.TableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No tasks",
			Message: "No tasks recorded for this phase.",
		},
		PrimaryAction: &types.PrimaryAction{
			Label:     "+ Add Task",
			ActionURL: addTaskURL,
			Icon:      "icon-plus",
		},
	}
}

// taskStatusString converts a TaskStatus enum to a lowercase display string
// for use within the job_phase detail tasks tab (mirrors job_task/list/page.go).
func taskStatusString(s jobtaskpb.TaskStatus) string {
	switch s {
	case jobtaskpb.TaskStatus_TASK_STATUS_PENDING:
		return "pending"
	case jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return "in_progress"
	case jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED:
		return "completed"
	case jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED:
		return "skipped"
	case jobtaskpb.TaskStatus_TASK_STATUS_HOLD:
		return "hold"
	case jobtaskpb.TaskStatus_TASK_STATUS_REWORK:
		return "rework"
	default:
		return "pending"
	}
}

// taskStatusVariant returns the badge variant for a task status string.
func taskStatusVariant(status string) string {
	switch status {
	case "pending":
		return "warning"
	case "in_progress":
		return "info"
	case "completed":
		return "success"
	case "skipped":
		return "default"
	case "hold":
		return "warning"
	case "rework":
		return "danger"
	default:
		return "default"
	}
}
