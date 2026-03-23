package detail

import (
	"context"
	"fmt"
	"log"

	fayna "github.com/erniealice/fayna-golang"

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
	pageData.PhasesTable = buildPhasesTable(phases, tasksByPhase, l, deps.TableLabels)
}

// buildPhasesTable builds the phases table config.
func buildPhasesTable(
	phases []*jobphasepb.JobPhase,
	tasksByPhase map[string][]*jobtaskpb.JobTask,
	l fayna.JobLabels,
	tableLabels types.TableLabels,
) *types.TableConfig {
	columns := []types.TableColumn{
		{Key: "order", Label: l.Detail.PhaseOrder, Sortable: false, Width: "60px"},
		{Key: "name", Label: l.Detail.PhaseName, Sortable: true},
		{Key: "status", Label: l.Detail.PhaseStatus, Sortable: true, Width: "120px"},
		{Key: "tasks", Label: l.Detail.TaskName, Sortable: false},
	}

	rows := []types.TableRow{}
	for _, phase := range phases {
		id := phase.GetId()
		name := phase.GetName()
		order := fmt.Sprintf("%d", phase.GetPhaseOrder())
		status := phaseStatusString(phase.GetStatus())

		// Build task summary
		tasks := tasksByPhase[id]
		taskSummary := ""
		if len(tasks) > 0 {
			taskSummary = fmt.Sprintf("%d task(s)", len(tasks))
		}

		rows = append(rows, types.TableRow{
			ID: id,
			Cells: []types.TableCell{
				{Type: "text", Value: order},
				{Type: "text", Value: name},
				{Type: "badge", Value: status, Variant: phaseStatusVariant(status)},
				{Type: "text", Value: taskSummary},
			},
			DataAttrs: map[string]string{
				"order":  order,
				"name":   name,
				"status": status,
			},
		})
	}

	types.ApplyColumnStyles(columns, rows)

	return &types.TableConfig{
		ID:         "phases-table",
		Columns:    columns,
		Rows:       rows,
		ShowSearch: false,
		Labels:     tableLabels,
		EmptyState: types.TableEmptyState{
			Title:   "No phases",
			Message: "This job has no phases defined yet.",
		},
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
