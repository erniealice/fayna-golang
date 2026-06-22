package operation

import (
	"context"

	ttcpkg "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria"

	"github.com/erniealice/hybra-golang/views/auditlog"
	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	ttcpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria"

	ttcdetail "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria/detail"
	ttclist "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria/list"
)

// TemplateTaskCriteriaModuleDeps holds all dependencies for the template task criteria module.
type TemplateTaskCriteriaModuleDeps struct {
	Routes       ttcpkg.Routes
	Labels       ttcpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// TemplateTaskCriteria CRUD
	CreateTemplateTaskCriteria func(ctx context.Context, req *ttcpb.CreateTemplateTaskCriteriaRequest) (*ttcpb.CreateTemplateTaskCriteriaResponse, error)
	ReadTemplateTaskCriteria   func(ctx context.Context, req *ttcpb.ReadTemplateTaskCriteriaRequest) (*ttcpb.ReadTemplateTaskCriteriaResponse, error)
	UpdateTemplateTaskCriteria func(ctx context.Context, req *ttcpb.UpdateTemplateTaskCriteriaRequest) (*ttcpb.UpdateTemplateTaskCriteriaResponse, error)
	DeleteTemplateTaskCriteria func(ctx context.Context, req *ttcpb.DeleteTemplateTaskCriteriaRequest) (*ttcpb.DeleteTemplateTaskCriteriaResponse, error)
	ListTemplateTaskCriterias  func(ctx context.Context, req *ttcpb.ListTemplateTaskCriteriasRequest) (*ttcpb.ListTemplateTaskCriteriasResponse, error)

	// Audit history (optional — nil = history tab hidden/empty)
	auditlog.AuditOps
}

// TemplateTaskCriteriaModule holds all constructed template task criteria views.
type TemplateTaskCriteriaModule struct {
	routes     ttcpkg.Routes
	List       view.View
	Detail     view.View
	TabAction  view.View
	Add        view.View
	Edit       view.View
	Delete     view.View
	BulkDelete view.View
}

// NewTemplateTaskCriteriaModule creates a new template task criteria module with all views wired.
func NewTemplateTaskCriteriaModule(deps *TemplateTaskCriteriaModuleDeps) *TemplateTaskCriteriaModule {
	detailDeps := &ttcdetail.DetailViewDeps{
		AuditOps:                 deps.AuditOps,
		Routes:                   deps.Routes,
		Labels:                   deps.Labels,
		CommonLabels:             deps.CommonLabels,
		TableLabels:              deps.TableLabels,
		ReadTemplateTaskCriteria: deps.ReadTemplateTaskCriteria,
	}

	// Build entity-package ModuleDeps for the exported action builders.
	entityDeps := &ttcpkg.ModuleDeps{
		Routes:                     deps.Routes,
		Labels:                     deps.Labels,
		CommonLabels:               deps.CommonLabels,
		TableLabels:                deps.TableLabels,
		CreateTemplateTaskCriteria: deps.CreateTemplateTaskCriteria,
		ReadTemplateTaskCriteria:   deps.ReadTemplateTaskCriteria,
		UpdateTemplateTaskCriteria: deps.UpdateTemplateTaskCriteria,
		DeleteTemplateTaskCriteria: deps.DeleteTemplateTaskCriteria,
		ListTemplateTaskCriterias:  deps.ListTemplateTaskCriterias,
	}

	return &TemplateTaskCriteriaModule{
		routes: deps.Routes,
		List: ttclist.NewView(&ttclist.ListViewDeps{
			Routes:                    deps.Routes,
			ListTemplateTaskCriterias: deps.ListTemplateTaskCriterias,
			Labels:                    deps.Labels,
			CommonLabels:              deps.CommonLabels,
			TableLabels:               deps.TableLabels,
		}),
		Detail:     ttcdetail.NewView(detailDeps),
		TabAction:  ttcdetail.NewTabAction(detailDeps),
		Add:        ttcpkg.NewAddAction(entityDeps),
		Edit:       ttcpkg.NewEditAction(entityDeps),
		Delete:     ttcpkg.NewDeleteAction(entityDeps),
		BulkDelete: ttcpkg.NewBulkDeleteAction(entityDeps),
	}
}

// RegisterRoutes registers all template task criteria routes.
func (m *TemplateTaskCriteriaModule) RegisterRoutes(r view.RouteRegistrar) {
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
