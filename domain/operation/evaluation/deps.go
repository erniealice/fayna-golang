package evaluation

import (
	"context"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	evalpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation"
	resppb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_response"
	itempb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/evaluation_template_item"
)

// ModuleDeps holds the typed use-case closures + label/route config that the
// list/detail/action/portal/form builders need. The Integrator (block catalog)
// injects every closure; the view layer NEVER calls espyna directly.
//
// Defined here (not in the module assembler) to avoid a self-import cycle:
// the package-level action builders (actions.go) and sub-packages (list/,
// detail/, portal/, form/) all read from this single struct.
//
// All IDOR / CR-5 servicing gates are enforced INSIDE these closures
// (use-case/adapter QUERY PREDICATE), per Q-EVAL-IDOR-1 — the view does not
// supply client_id and cannot re-introduce it (the proto request types carry
// no client_id field). See descriptor.go / pages.md §A.3, §C, §H.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// Evaluation CRUD + page-data (List/Item) — servicing-gated for staff (CR-5)
	CreateEvaluation func(ctx context.Context, req *evalpb.CreateEvaluationRequest) (*evalpb.CreateEvaluationResponse, error)
	ReadEvaluation   func(ctx context.Context, req *evalpb.ReadEvaluationRequest) (*evalpb.ReadEvaluationResponse, error)
	UpdateEvaluation func(ctx context.Context, req *evalpb.UpdateEvaluationRequest) (*evalpb.UpdateEvaluationResponse, error)
	DeleteEvaluation func(ctx context.Context, req *evalpb.DeleteEvaluationRequest) (*evalpb.DeleteEvaluationResponse, error)
	ListEvaluations  func(ctx context.Context, req *evalpb.ListEvaluationsRequest) (*evalpb.ListEvaluationsResponse, error)

	// Staff list / detail page-data (servicing-gated projection — §C, §D)
	GetListPageData func(ctx context.Context, req *evalpb.GetEvaluationListPageDataRequest) (*evalpb.GetEvaluationListPageDataResponse, error)
	GetItemPageData func(ctx context.Context, req *evalpb.GetEvaluationItemPageDataRequest) (*evalpb.GetEvaluationItemPageDataResponse, error)

	// Named lifecycle use cases (SUBMITTED→SIGNED_OFF / →ARCHIVED).
	// SignOff: SUBMITTED→SIGNED_OFF, stamps signed_off_* arcs (CR-4), is_owner-gated.
	// Archive/Submit are modelled as UpdateEvaluation status transitions at the
	// espyna layer; the closure shape stays Read/Update-based so the Integrator
	// can wire either the dedicated UC or the shaped update.
	SignOffEvaluation func(ctx context.Context, req *evalpb.UpdateEvaluationRequest) (*evalpb.UpdateEvaluationResponse, error)
	ArchiveEvaluation func(ctx context.Context, req *evalpb.UpdateEvaluationRequest) (*evalpb.UpdateEvaluationResponse, error)

	// Scores tab — evaluation_response SNAPSHOT rows (filtered evaluation_id, §D.1)
	ListEvaluationResponses func(ctx context.Context, req *resppb.ListEvaluationResponsesRequest) (*resppb.ListEvaluationResponsesResponse, error)
	CreateEvaluationResponse func(ctx context.Context, req *resppb.CreateEvaluationResponseRequest) (*resppb.CreateEvaluationResponseResponse, error)

	// Polymorphic dimension slot (§A.2) — loads the active template's ordered
	// items so the drawer can render one form-group per item (branch on the
	// linked OutcomeCriteria.criteria_type).
	ListEvaluationTemplateItems func(ctx context.Context, req *itempb.ListEvaluationTemplateItemsRequest) (*itempb.ListEvaluationTemplateItemsResponse, error)

	NewID func() string
}
