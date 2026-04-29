package detail

import (
	"context"

	fayna "github.com/erniealice/fayna-golang"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job"
	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobphasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_phase"
	jobsettlementpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_settlement"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
	subscriptionpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/subscription"
)

// DetailViewDeps holds view dependencies for the job detail views.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes       fayna.JobRoutes
	Labels       fayna.JobLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

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
	// consuming app via fayna.WithSubscriptionDetailURL. Both nil/empty =
	// the breadcrumb is omitted.
	ReadSubscription        func(ctx context.Context, req *subscriptionpb.ReadSubscriptionRequest) (*subscriptionpb.ReadSubscriptionResponse, error)
	SubscriptionDetailURL string
}
