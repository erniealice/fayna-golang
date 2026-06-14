package evaluation_template

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	templatepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// ModuleDeps holds the typed closures that action builders and sub-packages need.
// Defined here (not in the module assembler) to avoid a self-import cycle:
// the module file cannot import its own package path.
//
// All use-case closures are INJECTED by the Integrator/block (which wires the
// espyna use cases). This package never calls espyna directly.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// EvaluationTemplate CRUD + lifecycle.
	// Activate/Deprecate are modeled as Update closures by the block (status
	// flip), and Clone as Create with copied_from_id; the block may instead
	// inject dedicated lifecycle closures. Keep the surface to the standard
	// CRUD verbs the espyna usecases/domain/operation/evaluation_template exposes.
	CreateEvaluationTemplate func(ctx context.Context, req *templatepb.CreateEvaluationTemplateRequest) (*templatepb.CreateEvaluationTemplateResponse, error)
	ReadEvaluationTemplate   func(ctx context.Context, req *templatepb.ReadEvaluationTemplateRequest) (*templatepb.ReadEvaluationTemplateResponse, error)
	UpdateEvaluationTemplate func(ctx context.Context, req *templatepb.UpdateEvaluationTemplateRequest) (*templatepb.UpdateEvaluationTemplateResponse, error)
	DeleteEvaluationTemplate func(ctx context.Context, req *templatepb.DeleteEvaluationTemplateRequest) (*templatepb.DeleteEvaluationTemplateResponse, error)
	ListEvaluationTemplates  func(ctx context.Context, req *templatepb.ListEvaluationTemplatesRequest) (*templatepb.ListEvaluationTemplatesResponse, error)

	// EvaluationTemplateItem read — for the Items tab COUNT (list column) and
	// the rubric-builder ordered list on the detail Items tab.
	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)

	// OutcomeCriteria read — for the rubric-item drawer criterion autocomplete
	// (ListByScope filtered scope=CRITERIA_SCOPE_EVALUATION); surfaces the
	// criterion's criteria_type read-only.
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	NewID func() string
}
