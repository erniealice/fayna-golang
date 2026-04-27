package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/pyeza-golang/route"
	"github.com/erniealice/pyeza-golang/types"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// loadPhasesTab populates the PageData with phases and tasks table data.
func loadPhasesTab(ctx context.Context, deps *DetailViewDeps, pageData *PageData, jobID string) {
	if deps.ListJobPhases == nil {
		return
	}

	phasesResp, err := deps.ListJobPhases(ctx, &jobphasepb.ListJobPhasesRequest{})
	if err != nil {
		log.Printf("Failed to list job phases for job %s: %v", jobID, err)
		return
	}

	var tasksResp *jobtaskpb.ListJobTasksResponse
	if deps.ListJobTasks != nil {
		tasksResp, err = deps.ListJobTasks(ctx, &jobtaskpb.ListJobTasksRequest{})
		if err != nil {
			log.Printf("Failed to list job tasks for job %s: %v", jobID, err)
		}
	}

	// Filter phases by job ID
	var phases []*jobphasepb.JobPhase
	for _, p := range phasesResp.GetData() {
		if p.GetJobId() == jobID {
			phases = append(phases, p)
		}
	}

	// Build a task map keyed by phase ID for quick lookup
	tasksByPhase := map[string][]*jobtaskpb.JobTask{}
	if tasksResp != nil {
		for _, t := range tasksResp.GetData() {
			if t.GetJobPhase() != nil && t.GetJobPhase().GetJobId() == jobID {
				phaseID := t.GetJobPhaseId()
				tasksByPhase[phaseID] = append(tasksByPhase[phaseID], t)
			}
		}
	}

	l := deps.Labels
	pageData.PhasesTable = buildPhasesTable(phases, tasksByPhase, jobID, l, deps.Routes, deps.TableLabels)
}

// buildPhasesTable builds a flat tasks table with phase name denormalized per row.
func buildPhasesTable(
	phases []*jobphasepb.JobPhase,
	tasksByPhase map[string][]*jobtaskpb.JobTask,
	jobID string,
	l fayna.JobLabels,
	routes fayna.JobRoutes,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "phase", Label: l.Detail.PhaseName, Sortable: true},
		{Key: "order", Label: "#", Sortable: false, WidthClass: "col-sm"},
		{Key: "name", Label: l.Detail.TaskName, Sortable: true},
		{Key: "status", Label: l.Detail.PhaseStatus, Sortable: true, WidthClass: "col-2xl"},
		{Key: "assigned_to", Label: "Assigned To", Sortable: false, WidthClass: "col-2xl"},
	}

	rows := []types.TableRow{}
	for _, phase := range phases {
		phaseID := phase.GetId()
		phaseName := phase.GetName()
		tasks := tasksByPhase[phaseID]
		for _, task := range tasks {
			taskID := task.GetId()
			stepOrder := fmt.Sprintf("%d", task.GetStepOrder())
			taskName := task.GetName()
			status := taskStatusString(task.GetStatus())

			assignee := "Unassigned"
			if a := task.GetAssignedTo(); a != "" {
				assignee = a
			}

			taskActions := []types.TableAction{}
			if routes.TaskAssignURL != "" {
				assignURL := route.ResolveURL(routes.TaskAssignURL, "id", jobID, "taskId", taskID)
				taskActions = append(taskActions, types.TableAction{
					Type:   "edit",
					Label:  "Assign",
					Action: "edit",
					URL:    assignURL,
				})
			}

			rows = append(rows, types.TableRow{
				ID: taskID,
				Cells: []types.TableCell{
					{Type: "text", Value: phaseName},
					{Type: "text", Value: stepOrder},
					{Type: "text", Value: taskName},
					{Type: "badge", Value: status, Variant: taskStatusVariant(status)},
					{Type: "text", Value: assignee},
				},
				DataAttrs: map[string]string{
					"phase":       phaseID,
					"phase_name":  phaseName,
					"step_order":  stepOrder,
					"name":        taskName,
					"status":      status,
					"assigned_to": assignee,
				},
				Actions: taskActions,
			})
		}
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "phases-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: false,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No tasks",
			Message: "This job has no tasks defined yet.",
		},
	}
}

// taskStatusString converts a TaskStatus enum to a display string.
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
		return "On Hold"
	case jobtaskpb.TaskStatus_TASK_STATUS_REWORK:
		return "Rework"
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
		return "success"
	case "completed":
		return "info"
	case "skipped":
		return "default"
	case "On Hold":
		return "secondary"
	case "Rework":
		return "danger"
	default:
		return "default"
	}
}

// phaseStatusString converts a PhaseStatus enum to a display string.
func phaseStatusString(s jobphasepb.PhaseStatus) string {
	switch s {
	case jobphasepb.PhaseStatus_PHASE_STATUS_PENDING:
		return "pending"
	case jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE:
		return "active"
	case jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED:
		return "completed"
	default:
		return "pending"
	}
}

// phaseStatusVariant returns the badge variant for a phase status string.
func phaseStatusVariant(status string) string {
	switch status {
	case "pending":
		return "warning"
	case "active":
		return "success"
	case "completed":
		return "info"
	default:
		return "default"
	}
}
