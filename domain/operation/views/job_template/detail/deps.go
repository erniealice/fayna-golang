package detail

import (
	"context"

	operation "github.com/erniealice/fayna-golang/domain/operation"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// DetailViewDeps holds view dependencies.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes operation.JobTemplateRoutes
	// PhaseRoutes and TaskRoutes supply Edit/Delete URLs for per-row CTAs
	// and the Add CTA on the Phases / Tasks tabs.
	PhaseRoutes operation.JobTemplatePhaseRoutes
	TaskRoutes  operation.JobTemplateTaskRoutes

	ReadJobTemplate         func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)

	// Tab data stubs — wired in P6.template-children when view modules land.
	// Nil is safe: loaders no-op and render empty-state panels.
	//
	// ListTasksByPhase returns all JobTemplateTask rows for a given phase. The
	// tasks tab loader calls this once per phase to build the denormalised table.
	ListTasksByPhase func(ctx context.Context, req *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)
	// ListCriteriaByTask returns TemplateTaskCriteria pinnings for a given task.
	// The standards tab loader calls this once per task after collecting all tasks.
	ListCriteriaByTask func(ctx context.Context, req *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error)

	Labels       operation.JobTemplateLabels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
}
