package detail

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/job_template"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_phase"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_relation"
	"github.com/erniealice/fayna-golang/domain/operation/job_template_task"
	"github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"

	"github.com/erniealice/hybra-golang/views/attachment"
	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	jobtemplatephasepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_phase"
	jobtemplaterelationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
	jobtemplateTaskpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_task"
	templatetaskcriteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"
)

// DetailViewDeps holds view dependencies.
type DetailViewDeps struct {
	attachment.AttachmentOps
	auditlog.AuditOps

	Routes job_template.Routes
	// PhaseRoutes and TaskRoutes supply Edit/Delete URLs for per-row CTAs
	// and the Add CTA on the Phases / Tasks tabs. CriteriaRoutes supplies
	// Add/Delete URLs for the Standards tab's "+ Add Standard" CTA and
	// per-row remove actions. RelationRoutes supplies Add/Delete URLs for
	// the Spawn Graph tab's "+ Add Relation" CTA and per-row remove actions.
	PhaseRoutes    job_template_phase.Routes
	TaskRoutes     job_template_task.Routes
	CriteriaRoutes template_task_criteria.Routes
	RelationRoutes job_template_relation.Routes

	ReadJobTemplate         func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
	ListPhasesByJobTemplate func(ctx context.Context, req *jobtemplatephasepb.ListByJobTemplateRequest) (*jobtemplatephasepb.ListByJobTemplateResponse, error)
	// ListRelationsByParent backs the Spawn Graph tab's roster table — the
	// one job_template_relation closure that IS live in espyna today (see
	// job_template_relation/deps.go). Nil is safe (empty-state panel).
	ListRelationsByParent func(ctx context.Context, req *jobtemplaterelationpb.ListJobTemplateRelationsByParentRequest) (*jobtemplaterelationpb.ListJobTemplateRelationsByParentResponse, error)

	// Tab data stubs — wired in P6.template-children when view modules land.
	// Nil is safe: loaders no-op and render empty-state panels.
	//
	// ListTasksByPhase returns all JobTemplateTask rows for a given phase. The
	// tasks tab loader calls this once per phase to build the denormalised table.
	ListTasksByPhase func(ctx context.Context, req *jobtemplateTaskpb.ListJobTemplateTasksByPhaseRequest) (*jobtemplateTaskpb.ListJobTemplateTasksByPhaseResponse, error)
	// ListCriteriaByTask returns TemplateTaskCriteria pinnings for a given task.
	// The standards tab loader calls this once per task after collecting all tasks.
	ListCriteriaByTask func(ctx context.Context, req *templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskRequest) (*templatetaskcriteriapb.ListTemplateTaskCriteriasByTemplateTaskResponse, error)

	Labels       job_template.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels
}
