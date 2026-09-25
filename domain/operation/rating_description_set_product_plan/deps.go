package rating_description_set_product_plan

import (
	"context"

	linkform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan/form"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
	productplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	linkpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
)

// ModuleDeps holds the typed closures that action builders and sub-packages
// need. Injected by the block (espyna use cases); this package never calls
// espyna or SQL directly.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// AY setup read — the "offerings x AY" grid per the frozen proto shape
	// returns ACTIVE LINK ROWS ONLY (no rows for never-linked offerings) —
	// see W3-ESPYNA.done "Known gap for the fayna views agent". MVP list
	// shows what's linked; a full offerings grid with Unlinked/Invalid rows
	// is deferred (TODO, see sequence log).
	GetRatingDescriptionSetProductPlanListPageData func(ctx context.Context, req *linkpb.GetRatingDescriptionSetProductPlanListPageDataRequest) (*linkpb.GetRatingDescriptionSetProductPlanListPageDataResponse, error)

	RelinkRatingDescriptionSetProductPlan func(ctx context.Context, req *linkpb.RelinkRatingDescriptionSetProductPlanRequest) (*linkpb.RelinkRatingDescriptionSetProductPlanResponse, error)
	UnlinkRatingDescriptionSetProductPlan func(ctx context.Context, req *linkpb.UnlinkRatingDescriptionSetProductPlanRequest) (*linkpb.UnlinkRatingDescriptionSetProductPlanResponse, error)

	// Set picker — PUBLISHED only (Relink target eligibility).
	ListRatingDescriptionSets func(ctx context.Context, req *setpb.ListRatingDescriptionSetsRequest) (*setpb.ListRatingDescriptionSetsResponse, error)

	// Offering name resolution + picker. Bare list (ground truth: 43 rows —
	// block/usecases.go ProductPlanUseCases doc comment), one call is
	// complete.
	ListProductPlans func(ctx context.Context, req *productplanpb.ListProductPlansRequest) (*productplanpb.ListProductPlansResponse, error)

	// Academic year selector. Bare list (ground truth: 2 rows on education1 —
	// block/usecases.go PriceScheduleUseCases doc comment).
	ListPriceSchedules func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)

	// GetFormPageData is the AY selector + Relink drawer's offering/
	// PUBLISHED-set picker, sourced from a page-data use case authorized
	// under rating_description_set_product_plan:read (finding #6) —
	// preferred over ListProductPlans/ListPriceSchedules/
	// ListRatingDescriptionSets above.
	GetFormPageData func(ctx context.Context) (*linkform.FormPageData, error)
}
