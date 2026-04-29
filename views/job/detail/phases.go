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

// PhaseRow is the per-phase view-model rendered on the Phases tab. Used by
// the `job-phases-card` template to expose the Mark Complete CTA + status
// badge alongside each phase. 2026-04-29 milestone-billing plan §4.
type PhaseRow struct {
	ID            string
	Name          string
	Order         string
	Status        string // pending | active | completed (lowercase shorthand)
	StatusVariant string // badge variant (warning|success|info|default)
	StatusLabel   string // translated badge text
	IsCompleted   bool   // disables the Mark Complete CTA
	MarkURL       string // POST URL with id=… &status=PHASE_STATUS_COMPLETED
}

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
	pageData.PhasesList = buildPhasesList(phases, l, deps.Routes)
}

// buildPhasesList builds the per-phase view-model (separate from the
// denormalized phases+tasks table) so the Phases tab can render a "Mark
// Complete" CTA and status badge for each phase.
func buildPhasesList(
	phases []*jobphasepb.JobPhase,
	l fayna.JobLabels,
	routes fayna.JobRoutes,
) []PhaseRow {
	rows := make([]PhaseRow, 0, len(phases))
	for _, p := range phases {
		status := phaseStatusString(p.GetStatus())
		row := PhaseRow{
			ID:            p.GetId(),
			Name:          p.GetName(),
			Order:         fmt.Sprintf("%d", p.GetPhaseOrder()),
			Status:        status,
			StatusVariant: phaseStatusVariant(status),
			StatusLabel:   phaseStatusLabel(status, l),
			IsCompleted:   p.GetStatus() == jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED,
		}
		if routes.PhaseSetStatusURL != "" {
			row.MarkURL = fmt.Sprintf("%s?id=%s&status=PHASE_STATUS_COMPLETED",
				routes.PhaseSetStatusURL, row.ID)
		}
		rows = append(rows, row)
	}
	return rows
}

// phaseStatusLabel returns the translated badge text for a phase status.
func phaseStatusLabel(status string, l fayna.JobLabels) string {
	switch status {
	case "pending":
		if l.Detail.PhaseStatusPending != "" {
			return l.Detail.PhaseStatusPending
		}
		return "Pending"
	case "active":
		if l.Detail.PhaseStatusActive != "" {
			return l.Detail.PhaseStatusActive
		}
		return "Active"
	case "completed":
		if l.Detail.PhaseStatusCompleted != "" {
			return l.Detail.PhaseStatusCompleted
		}
		return "Completed"
	default:
		return status
	}
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
