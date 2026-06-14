package operation

import (
	"context"

	itempkg "github.com/erniealice/fayna-golang/domain/operation/evaluation_template_item"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// EvaluationTemplateItemModuleDeps holds dependencies for the rubric-item
// drawer module. No standalone list/detail — it surfaces via the
// evaluation_template detail Items tab. Use-case closures are injected by the
// block.
type EvaluationTemplateItemModuleDeps struct {
	Routes       itempkg.Routes
	Labels       itempkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateEvaluationTemplateItem func(ctx context.Context, req *itempb.CreateEvaluationTemplateItemRequest) (*itempb.CreateEvaluationTemplateItemResponse, error)
	ReadEvaluationTemplateItem   func(ctx context.Context, req *itempb.ReadEvaluationTemplateItemRequest) (*itempb.ReadEvaluationTemplateItemResponse, error)
	UpdateEvaluationTemplateItem func(ctx context.Context, req *itempb.UpdateEvaluationTemplateItemRequest) (*itempb.UpdateEvaluationTemplateItemResponse, error)
	DeleteEvaluationTemplateItem func(ctx context.Context, req *itempb.DeleteEvaluationTemplateItemRequest) (*itempb.DeleteEvaluationTemplateItemResponse, error)
	ListEvaluationTemplateItems  func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)

	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	NewID func() string
}

// EvaluationTemplateItemModule holds the constructed rubric-item drawer views.
type EvaluationTemplateItemModule struct {
	routes itempkg.Routes
	Add    view.View
	Edit   view.View
	Remove view.View
}

// NewEvaluationTemplateItemModule constructs the rubric-item drawer module.
func NewEvaluationTemplateItemModule(deps *EvaluationTemplateItemModuleDeps) *EvaluationTemplateItemModule {
	entityDeps := &itempkg.ModuleDeps{
		Routes:                       deps.Routes,
		Labels:                       deps.Labels,
		CommonLabels:                 deps.CommonLabels,
		TableLabels:                  deps.TableLabels,
		CreateEvaluationTemplateItem: deps.CreateEvaluationTemplateItem,
		ReadEvaluationTemplateItem:   deps.ReadEvaluationTemplateItem,
		UpdateEvaluationTemplateItem: deps.UpdateEvaluationTemplateItem,
		DeleteEvaluationTemplateItem: deps.DeleteEvaluationTemplateItem,
		ListEvaluationTemplateItems:  deps.ListEvaluationTemplateItems,
		ListOutcomeCriterias:         deps.ListOutcomeCriterias,
		NewID:                        deps.NewID,
	}

	return &EvaluationTemplateItemModule{
		routes: deps.Routes,
		Add:    itempkg.NewAddAction(entityDeps),
		Edit:   itempkg.NewEditAction(entityDeps),
		Remove: itempkg.NewRemoveAction(entityDeps),
	}
}

// RegisterRoutes registers the rubric-item drawer routes.
func (m *EvaluationTemplateItemModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
	r.POST(m.routes.RemoveURL, m.Remove)
}
