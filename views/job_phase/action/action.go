// Package action contains HTTP/HTMX handlers for the job_phase view module.
// Dependency-bearing helpers that need full Deps live here; pure-function
// builders live in the sibling form/ package.
package action

import (
	"context"

	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	fayna "github.com/erniealice/fayna-golang"
)

// Deps holds dependencies shared across all job_phase action handlers.
type Deps struct {
	Routes fayna.JobPhaseRoutes
	Labels fayna.JobPhaseLabels

	// Job phase CRUD
	CreateJobPhase func(ctx context.Context, req *jobphasepb.CreateJobPhaseRequest) (*jobphasepb.CreateJobPhaseResponse, error)
	ReadJobPhase   func(ctx context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)
	UpdateJobPhase func(ctx context.Context, req *jobphasepb.UpdateJobPhaseRequest) (*jobphasepb.UpdateJobPhaseResponse, error)
	DeleteJobPhase func(ctx context.Context, req *jobphasepb.DeleteJobPhaseRequest) (*jobphasepb.DeleteJobPhaseResponse, error)
	ListJobPhases  func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)

	// ResourceSearchURL for the resource picker in the Add/Edit drawer.
	ResourceSearchURL string
}

// phaseStatusToEnum converts a status string (proto enum name or shorthand)
// to the protobuf PhaseStatus enum.
// Canonical conversion — also used by set_status.go and bulk_set_status.
func phaseStatusToEnum(status string) jobphasepb.PhaseStatus {
	switch status {
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

// phaseStatusString converts a PhaseStatus enum to a lowercase display string.
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
