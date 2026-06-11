package detail

import (
	"context"

	job_task "github.com/erniealice/fayna-golang/domain/operation/job_task"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobactivitypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_activity"
	jobtaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_task"
)

// DetailViewDeps holds view dependencies for the job_task detail views.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes       job_task.Routes
	Labels       job_task.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Task read (required)
	ReadJobTask func(ctx context.Context, req *jobtaskpb.ReadJobTaskRequest) (*jobtaskpb.ReadJobTaskResponse, error)

	// Activities tab — list all activities for the task.
	// JobActivity rows carry job_task_id; we filter to those matching this task.
	ListJobActivities func(ctx context.Context, req *jobactivitypb.ListJobActivitiesRequest) (*jobactivitypb.ListJobActivitiesResponse, error)
}
