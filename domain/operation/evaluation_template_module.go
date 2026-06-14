package operation

import (
	"context"

	evaluationtemplatepkg "github.com/erniealice/fayna-golang/domain/operation/evaluation_template"
	evaluationtemplatedetail "github.com/erniealice/fayna-golang/domain/operation/evaluation_template/detail"
	evaluationtemplatelist "github.com/erniealice/fayna-golang/domain/operation/evaluation_template/list"
	evaluationtemplateitempkg "github.com/erniealice/fayna-golang/domain/operation/evaluation_template_item"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
	templatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template"
)

// EvaluationTemplateModuleDeps holds all dependencies for the evaluation
// template module. The use-case closures are injected by the block (which
// wires the espyna use cases); this module never calls espyna directly.
type EvaluationTemplateModuleDeps struct {
	Routes       evaluationtemplatepkg.Routes
	Labels       evaluationtemplatepkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// EvaluationTemplate CRUD + lifecycle (status flip).
	CreateEvaluationTemplate func(ctx context.Context, req *templatepb.CreateEvaluationTemplateRequest) (*templatepb.CreateEvaluationTemplateResponse, error)
	ReadEvaluationTemplate   func(ctx context.Context, req *templatepb.ReadEvaluationTemplateRequest) (*templatepb.ReadEvaluationTemplateResponse, error)
	UpdateEvaluationTemplate func(ctx context.Context, req *templatepb.UpdateEvaluationTemplateRequest) (*templatepb.UpdateEvaluationTemplateResponse, error)
	DeleteEvaluationTemplate func(ctx context.Context, req *templatepb.DeleteEvaluationTemplateRequest) (*templatepb.DeleteEvaluationTemplateResponse, error)
	ListEvaluationTemplates  func(ctx context.Context, req *templatepb.ListEvaluationTemplatesRequest) (*templatepb.ListEvaluationTemplatesResponse, error)

	// EvaluationTemplateItem read — Items column count + Items tab rubric list.
	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)

	// OutcomeCriteria read — criterion type lookup for the rubric builder +
	// activation guard.
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// Item-drawer routes (from the evaluation_template_item module) so the
	// detail Items tab can mount Add Question / edit / remove endpoints.
	ItemRoutes evaluationtemplateitempkg.Routes

	NewID func() string
}

// EvaluationTemplateModule holds all constructed evaluation template views.
type EvaluationTemplateModule struct {
	routes        evaluationtemplatepkg.Routes
	List          view.View
	Table         view.View
	Detail        view.View
	TabAction     view.View
	Add           view.View
	Edit          view.View
	Activate      view.View
	Deprecate     view.View
	Clone         view.View
	BulkDeprecate view.View
}

// NewEvaluationTemplateModule creates the module with all views wired.
func NewEvaluationTemplateModule(deps *EvaluationTemplateModuleDeps) *EvaluationTemplateModule {
	listDeps := &evaluationtemplatelist.ListViewDeps{
		Routes:                      deps.Routes,
		ListEvaluationTemplates:     deps.ListEvaluationTemplates,
		ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
		Labels:                      deps.Labels,
		CommonLabels:                deps.CommonLabels,
		TableLabels:                 deps.TableLabels,
	}

	detailDeps := &evaluationtemplatedetail.DetailViewDeps{
		Routes:                      deps.Routes,
		Labels:                      deps.Labels,
		CommonLabels:                deps.CommonLabels,
		TableLabels:                 deps.TableLabels,
		ReadEvaluationTemplate:      deps.ReadEvaluationTemplate,
		ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
		ListOutcomeCriterias:        deps.ListOutcomeCriterias,
		ItemAddURL:                  deps.ItemRoutes.AddURL,
		ItemEditURL:                 deps.ItemRoutes.EditURL,
		ItemRemoveURL:               deps.ItemRoutes.RemoveURL,
	}

	entityDeps := &evaluationtemplatepkg.ModuleDeps{
		Routes:                      deps.Routes,
		Labels:                      deps.Labels,
		CommonLabels:                deps.CommonLabels,
		TableLabels:                 deps.TableLabels,
		CreateEvaluationTemplate:    deps.CreateEvaluationTemplate,
		ReadEvaluationTemplate:      deps.ReadEvaluationTemplate,
		UpdateEvaluationTemplate:    deps.UpdateEvaluationTemplate,
		DeleteEvaluationTemplate:    deps.DeleteEvaluationTemplate,
		ListEvaluationTemplates:     deps.ListEvaluationTemplates,
		ListEvaluationTemplateItems: deps.ListEvaluationTemplateItems,
		ListOutcomeCriterias:        deps.ListOutcomeCriterias,
		NewID:                       deps.NewID,
	}

	return &EvaluationTemplateModule{
		routes:        deps.Routes,
		List:          evaluationtemplatelist.NewView(listDeps),
		Table:         evaluationtemplatelist.NewTableView(listDeps),
		Detail:        evaluationtemplatedetail.NewView(detailDeps),
		TabAction:     evaluationtemplatedetail.NewTabAction(detailDeps),
		Add:           evaluationtemplatepkg.NewAddAction(entityDeps),
		Edit:          evaluationtemplatepkg.NewEditAction(entityDeps),
		Activate:      evaluationtemplatepkg.NewActivateAction(entityDeps),
		Deprecate:     evaluationtemplatepkg.NewDeprecateAction(entityDeps),
		Clone:         evaluationtemplatepkg.NewCloneAction(entityDeps),
		BulkDeprecate: evaluationtemplatepkg.NewBulkDeprecateAction(entityDeps),
	}
}

// RegisterRoutes registers all evaluation template routes.
func (m *EvaluationTemplateModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.TableURL, m.Table)
	r.GET(m.routes.DetailURL, m.Detail)
	r.GET(m.routes.TabActionURL, m.TabAction)

	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)

	r.POST(m.routes.ActivateURL, m.Activate)
	r.POST(m.routes.DeprecateURL, m.Deprecate)
	r.POST(m.routes.CloneURL, m.Clone)
	r.POST(m.routes.BulkDeprecateURL, m.BulkDeprecate)
}
