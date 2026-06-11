package form

import (
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// BuildTaskStatusOptions returns a []map[string]any for the status <select>.
// The current value is pre-selected so that the Edit drawer shows the existing state.
func BuildTaskStatusOptions(current string) []map[string]any {
	rows := []struct {
		value string
		label string
	}{
		{"TASK_STATUS_PENDING", "Pending"},
		{"TASK_STATUS_IN_PROGRESS", "In Progress"},
		{"TASK_STATUS_COMPLETED", "Completed"},
		{"TASK_STATUS_SKIPPED", "Skipped"},
		{"TASK_STATUS_HOLD", "Hold"},
		{"TASK_STATUS_REWORK", "Rework"},
	}

	opts := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		opts = append(opts, map[string]any{
			"Value":    r.value,
			"Label":    r.label,
			"Selected": current == r.value,
		})
	}
	return opts
}

// TaskStatusFromString converts a form status string (proto enum name or
// shorthand) to the protobuf TaskStatus enum.
// Mirrors the logic in action/action.go taskStatusToEnum so templates and
// handlers stay in sync.
func TaskStatusFromString(s string) jobtaskpb.TaskStatus {
	switch s {
	case "TASK_STATUS_PENDING", "pending":
		return jobtaskpb.TaskStatus_TASK_STATUS_PENDING
	case "TASK_STATUS_IN_PROGRESS", "in_progress":
		return jobtaskpb.TaskStatus_TASK_STATUS_IN_PROGRESS
	case "TASK_STATUS_COMPLETED", "completed":
		return jobtaskpb.TaskStatus_TASK_STATUS_COMPLETED
	case "TASK_STATUS_SKIPPED", "skipped":
		return jobtaskpb.TaskStatus_TASK_STATUS_SKIPPED
	case "TASK_STATUS_HOLD", "hold":
		return jobtaskpb.TaskStatus_TASK_STATUS_HOLD
	case "TASK_STATUS_REWORK", "rework":
		return jobtaskpb.TaskStatus_TASK_STATUS_REWORK
	default:
		return jobtaskpb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}
