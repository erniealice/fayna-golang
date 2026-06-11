package detail

import (
	"context"

	job "github.com/erniealice/fayna-golang/domain/operation/job"
	job_activity "github.com/erniealice/fayna-golang/domain/operation/job_activity"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobsettlementpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_settlement"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplatetaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	subscriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription"
)

// DetailViewDeps holds view dependencies for the job detail views.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes       job.Routes
	Labels       job.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// 2026-04-29 milestone-billing plan §5/§6 — Activities tab needs the
	// JobActivity add/edit URLs so the "+ Add Activity" CTA on Job detail
	// can launch the JobActivity drawer (with the BillableStatus selector
	// required by phase5 specs 09 and 11). Empty = CTA suppressed.
	JobActivityRoutes job_activity.Routes
	JobActivityLabels job_activity.Labels

	// Job read
	ReadJob func(ctx context.Context, req *jobpb.ReadJobRequest) (*jobpb.ReadJobResponse, error)

	// Sub-entity list operations (for tabs)
	ListJobPhases      func(ctx context.Context, req *jobphasepb.ListJobPhasesRequest) (*jobphasepb.ListJobPhasesResponse, error)
	ListJobTasks       func(ctx context.Context, req *jobtaskpb.ListJobTasksRequest) (*jobtaskpb.ListJobTasksResponse, error)
	UpdateJobTask      func(ctx context.Context, req *jobtaskpb.UpdateJobTaskRequest) (*jobtaskpb.UpdateJobTaskResponse, error)
	ListJobActivities  func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)
	ListJobSettlements func(ctx context.Context, req *jobsettlementpb.ListJobSettlementsRequest) (*jobsettlementpb.ListJobSettlementsResponse, error)

	// 2026-04-29 auto-spawn-jobs-from-subscription plan §5.4 — subscription
	// origin breadcrumb deps. ReadSubscription resolves the linked
	// Subscription's code; SubscriptionDetailURL is the cross-package URL
	// pattern (e.g. "/app/subscriptions/detail/{id}") supplied by the
	// consuming app via the fayna block option. Both nil/empty =
	// the breadcrumb is omitted.
	ReadSubscription      func(ctx context.Context, req *subscriptionpb.ReadSubscriptionRequest) (*subscriptionpb.ReadSubscriptionResponse, error)
	SubscriptionDetailURL string

	// Budget tab — reads the JobTemplate-derived phase/task hour plan.
	// All three deps are optional; nil = budget tab renders empty state.
	// Wave 3 will replace these with a JobInputPlan reader.
	ReadJobTemplate func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	// ListJobTemplatePhasesByTemplate lists phases by template_id using the
	// ListByJobTemplate RPC on the job_template_phase service.
	ListJobTemplatePhasesByTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
	// ListJobTemplateTasksByPhase lists tasks by phase_id using the
	// ListByPhase RPC on the job_template_task service.
	ListJobTemplateTasksByPhase func(ctx context.Context, req *jobtemplatetaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplatetaskpb.ListJobTemplateTasksByPhaseResponse, error)

	// Actuals tab — reads the JobActivity cost rollup via GetJobActivityRollup.
	// Optional; nil = actuals tab renders empty state.
	GetJobActivityRollup func(ctx context.Context, req *jobactivitypb.GetJobActivityRollupRequest) (*jobactivitypb.GetJobActivityRollupResponse, error)
}
