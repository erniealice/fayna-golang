package form

import (
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
)

// BuildPhaseStatusOptions returns a []map[string]any for the status <select>.
// The current value is pre-selected so that the Edit drawer shows the existing state.
func BuildPhaseStatusOptions(current string) []map[string]any {
	rows := []struct {
		value string
		enum  jobphasepb.PhaseStatus
		label string
	}{
		{"PHASE_STATUS_PENDING", jobphasepb.PhaseStatus_PHASE_STATUS_PENDING, "Pending"},
		{"PHASE_STATUS_ACTIVE", jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE, "Active"},
		{"PHASE_STATUS_COMPLETED", jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED, "Completed"},
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

// PhaseStatusFromString converts a form status string (proto enum name or
// shorthand) to the protobuf PhaseStatus enum.
// Mirrors the logic in action/action.go phaseStatusToEnum so templates and
// handlers stay in sync.
func PhaseStatusFromString(s string) jobphasepb.PhaseStatus {
	switch s {
	case "PHASE_STATUS_PENDING", "pending":
		return jobphasepb.PhaseStatus_PHASE_STATUS_PENDING
	case "PHASE_STATUS_ACTIVE", "active":
		return jobphasepb.PhaseStatus_PHASE_STATUS_ACTIVE
	case "PHASE_STATUS_COMPLETED", "completed":
		return jobphasepb.PhaseStatus_PHASE_STATUS_COMPLETED
	default:
		return jobphasepb.PhaseStatus_PHASE_STATUS_UNSPECIFIED
	}
}
