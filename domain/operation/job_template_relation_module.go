package operation

import (
	"context"

	jtrpkg "github.com/erniealice/fayna-golang/domain/operation/job_template_relation"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	jobtemplatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template"
	relationpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/job_template_relation"

	jtrdetail "github.com/erniealice/fayna-golang/domain/operation/job_template_relation/detail"
	jtrlist "github.com/erniealice/fayna-golang/domain/operation/job_template_relation/list"
)

// JobTemplateRelationModuleDeps holds all dependencies for the job template
// relation module (NEW, Q-TPL W4). See job_template_relation/deps.go for the
// espyna-rollout note on which closures are live today.
type JobTemplateRelationModuleDeps struct {
	Routes       jtrpkg.Routes
	Labels       jtrpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// JobTemplateRelation CRUD
	CreateJobTemplateRelation func(ctx context.Context, req *relationpb.CreateJobTemplateRelationRequest) (*relationpb.CreateJobTemplateRelationResponse, error)
	ReadJobTemplateRelation   func(ctx context.Context, req *relationpb.ReadJobTemplateRelationRequest) (*relationpb.ReadJobTemplateRelationResponse, error)
	UpdateJobTemplateRelation func(ctx context.Context, req *relationpb.UpdateJobTemplateRelationRequest) (*relationpb.UpdateJobTemplateRelationResponse, error)
	DeleteJobTemplateRelation func(ctx context.Context, req *relationpb.DeleteJobTemplateRelationRequest) (*relationpb.DeleteJobTemplateRelationResponse, error)
	ListJobTemplateRelations  func(ctx context.Context, req *relationpb.ListJobTemplateRelationsRequest) (*relationpb.ListJobTemplateRelationsResponse, error)
	ListByParent              func(ctx context.Context, req *relationpb.ListJobTemplateRelationsByParentRequest) (*relationpb.ListJobTemplateRelationsByParentResponse, error)

	ListJobTemplates func(ctx context.Context, req *jobtemplatepb.ListJobTemplatesRequest) (*jobtemplatepb.ListJobTemplatesResponse, error)
	ReadJobTemplate  func(ctx context.Context, req *jobtemplatepb.ReadJobTemplateRequest) (*jobtemplatepb.ReadJobTemplateResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// JobTemplateRelationModule holds all constructed job template relation views.
type JobTemplateRelationModule struct {
	routes     jtrpkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewJobTemplateRelationModule creates a new job template relation module with all views wired.
func NewJobTemplateRelationModule(deps *JobTemplateRelationModuleDeps) *JobTemplateRelationModule {
	detailDeps := &jtrdetail.DetailViewDeps{
		AuditOps:                deps.AuditOps,
		Routes:                  deps.Routes,
		Labels:                  deps.Labels,
		CommonLabels:            deps.CommonLabels,
		TableLabels:             deps.TableLabels,
		ReadJobTemplateRelation: deps.ReadJobTemplateRelation,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &jtrpkg.ModuleDeps{
		Routes:                    deps.Routes,
		Labels:                    deps.Labels,
		CommonLabels:              deps.CommonLabels,
		TableLabels:               deps.TableLabels,
		CreateJobTemplateRelation: deps.CreateJobTemplateRelation,
		ReadJobTemplateRelation:   deps.ReadJobTemplateRelation,
		UpdateJobTemplateRelation: deps.UpdateJobTemplateRelation,
		DeleteJobTemplateRelation: deps.DeleteJobTemplateRelation,
		ListJobTemplateRelations:  deps.ListJobTemplateRelations,
		ListByParent:              deps.ListByParent,
		ListJobTemplates:          deps.ListJobTemplates,
		ReadJobTemplate:           deps.ReadJobTemplate,
	}

	return &JobTemplateRelationModule{
		routes: deps.Routes,
		List: jtrlist.NewView(&jtrlist.ListViewDeps{
			Routes:                   deps.Routes,
			ListJobTemplateRelations: deps.ListJobTemplateRelations,
			Labels:                   deps.Labels,
			CommonLabels:             deps.CommonLabels,
			TableLabels:              deps.TableLabels,
		}),
		Detail:     jtrdetail.NewView(detailDeps),
		TabAction:  jtrdetail.NewTabAction(detailDeps),
		Add:        jtrpkg.NewAddAction(entityDeps),
		Edit:       jtrpkg.NewEditAction(entityDeps),
		Delete:     jtrpkg.NewDeleteAction(entityDeps),
		BulkDelete: jtrpkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all job template relation routes.
func (m *JobTemplateRelationModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)

	// CRUD actions (GET = drawer form, POST = process submission)
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.DeleteURL, m.Delete)
	r.POST(m.routes.BulkDeleteURL, m.BulkDelete)
}
