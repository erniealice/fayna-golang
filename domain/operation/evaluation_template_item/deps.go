package evaluation_template_item

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// ModuleDeps holds the typed closures the rubric-item drawer actions need.
// Injected by the block (espyna use cases); this package never calls espyna.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// EvaluationTemplateItem CRUD.
	CreateEvaluationTemplateItem func(ctx context.Context, req *itempb.CreateEvaluationTemplateItemRequest) (*itempb.CreateEvaluationTemplateItemResponse, error)
	ReadEvaluationTemplateItem   func(ctx context.Context, req *itempb.ReadEvaluationTemplateItemRequest) (*itempb.ReadEvaluationTemplateItemResponse, error)
	UpdateEvaluationTemplateItem func(ctx context.Context, req *itempb.UpdateEvaluationTemplateItemRequest) (*itempb.UpdateEvaluationTemplateItemResponse, error)
	DeleteEvaluationTemplateItem func(ctx context.Context, req *itempb.DeleteEvaluationTemplateItemRequest) (*itempb.DeleteEvaluationTemplateItemResponse, error)
	ListEvaluationTemplateItems  func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)

	// OutcomeCriteria read — criterion autocomplete filtered
	// scope=CRITERIA_SCOPE_EVALUATION (surfaces the criterion's criteria_type
	// read-only in the drawer).
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	NewID func() string
}
