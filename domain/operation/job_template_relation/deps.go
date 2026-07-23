package job_template_relation

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
//
// Create/Read/Update/Delete/List are wired against the full
// JobTemplateRelationDomainService proto contract; as of this build espyna's
// job_template_relation use-case package only ships ListByParent (the rest
// are landing in a parallel wave — see W4 report). Every closure here is
// nil-safe at the call site (action handlers early-return "not available";
// pickers/tables render empty) so this module compiles and degrades
// gracefully independent of espyna's rollout order.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// JobTemplateRelation CRUD
	CreateJobTemplateRelation func(ctx context.Context, req *relationpb.CreateJobTemplateRelationRequest) (*relationpb.CreateJobTemplateRelationResponse, error)
	ReadJobTemplateRelation   func(ctx context.Context, req *relationpb.ReadJobTemplateRelationRequest) (*relationpb.ReadJobTemplateRelationResponse, error)
	UpdateJobTemplateRelation func(ctx context.Context, req *relationpb.UpdateJobTemplateRelationRequest) (*relationpb.UpdateJobTemplateRelationResponse, error)
	DeleteJobTemplateRelation func(ctx context.Context, req *relationpb.DeleteJobTemplateRelationRequest) (*relationpb.DeleteJobTemplateRelationResponse, error)
	ListJobTemplateRelations  func(ctx context.Context, req *relationpb.ListJobTemplateRelationsRequest) (*relationpb.ListJobTemplateRelationsResponse, error)
	// ListByParent backs the job_template detail Spawn Graph tab's roster
	// table (spawn_graph.go) — the one closure that IS live in espyna today.
	ListByParent func(ctx context.Context, req *relationpb.ListJobTemplateRelationsByParentRequest) (*relationpb.ListJobTemplateRelationsByParentResponse, error)

	// ListJobTemplates populates the Parent/Child Template pickers. Optional
	// — nil-safe (empty picker).
	ListJobTemplates func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	// ReadJobTemplate resolves the parent template's display name when the
	// drawer opens in ContextTemplate. Optional — nil-safe (falls back to
	// the raw id).
	ReadJobTemplate func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)
}
