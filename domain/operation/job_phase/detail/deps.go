package detail

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/job_phase"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// DetailViewDeps holds view dependencies for the job_phase detail views.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes       job_phase.Routes
	Labels       job_phase.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Phase read (required)
	ReadJobPhase func(ctx context.Context, req *jobphasepb.ReadJobPhaseRequest) (*jobphasepb.ReadJobPhaseResponse, error)

	// Activities tab — list all activities for the phase's parent job, then
	// filter in-memory by job_task.job_phase_id == this phase.
	// TODO(WaveN): replace with dedicated ListJobActivitiesByPhase RPC once
	// espyna exposes it — the in-memory filter is an acceptable v1 trade-off.
	ListJobActivities func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)

	// Tasks tab — list all tasks for this phase via the job_task module.
	ListJobTasksByPhase func(ctx context.Context, req *jobtaskpb.ListJobTasksByPhaseRequest) (*jobtaskpb.ListJobTasksByPhaseResponse, error)
}
