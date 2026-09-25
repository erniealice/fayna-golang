package operation

import (
	"context"

	linkpkg "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan"
	linkform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan/form"
	linklist "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan/list"
	linklistdata "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_product_plan/listdata"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	linkpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
	productplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/product/product_plan"
	priceschedulepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/subscription/price_schedule"
)

// RatingDescriptionSetProductPlanModuleDeps holds all dependencies for the
// AY setup ("Descriptor Assignments") module (MVP scope: list of active
// links per academic year + Relink + Unlink; bulk assign / copy-previous-AY
// are deferred — TODO, see the plan sequence log).
type RatingDescriptionSetProductPlanModuleDeps struct {
	Routes       linkpkg.Routes
	Labels       linkpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	GetRatingDescriptionSetProductPlanListPageData func(ctx context.Context, req *linkpb.GetRatingDescriptionSetProductPlanListPageDataRequest) (*linkpb.GetRatingDescriptionSetProductPlanListPageDataResponse, error)
	RelinkRatingDescriptionSetProductPlan          func(ctx context.Context, req *linkpb.RelinkRatingDescriptionSetProductPlanRequest) (*linkpb.RelinkRatingDescriptionSetProductPlanResponse, error)
	UnlinkRatingDescriptionSetProductPlan          func(ctx context.Context, req *linkpb.UnlinkRatingDescriptionSetProductPlanRequest) (*linkpb.UnlinkRatingDescriptionSetProductPlanResponse, error)

	ListRatingDescriptionSets func(ctx context.Context, req *setpb.ListRatingDescriptionSetsRequest) (*setpb.ListRatingDescriptionSetsResponse, error)
	ListProductPlans          func(ctx context.Context, req *productplanpb.ListProductPlansRequest) (*productplanpb.ListProductPlansResponse, error)
	ListPriceSchedules        func(ctx context.Context, req *priceschedulepb.ListPriceSchedulesRequest) (*priceschedulepb.ListPriceSchedulesResponse, error)

	// GetFormPageData — Relink drawer's offering/AY/PUBLISHED-set picker,
	// authorized under rating_description_set_product_plan:read instead of
	// separate product_plan:list / price_schedule:list /
	// rating_description_set:list grants (codex-review-impl2.out.md finding
	// #6); its offering options are workspace-scoped through the offering's
	// PARENT product AND plan (codex-review-impl3.out.md finding #3).
	GetFormPageData func(ctx context.Context) (*linkform.FormPageData, error)
	// GetListSummaryPageData — the LIST page's SOLE data source, authorized
	// ONCE under rating_description_set_product_plan:list
	// (codex-review-impl3.out.md findings #1 and #3). NOT used by the
	// Relink drawer (GetFormPageData above).
	GetListSummaryPageData func(ctx context.Context, priceScheduleID string) (*linklistdata.PageData, error)
}

// RatingDescriptionSetProductPlanModule holds all constructed AY setup views.
type RatingDescriptionSetProductPlanModule struct {
	routes linkpkg.Routes
	List   view.View
	Relink view.View
	Unlink view.View
}

// NewRatingDescriptionSetProductPlanModule creates a new AY setup module with
// all views wired.
func NewRatingDescriptionSetProductPlanModule(deps *RatingDescriptionSetProductPlanModuleDeps) *RatingDescriptionSetProductPlanModule {
	listDeps := &linklist.ListViewDeps{
		Routes:                 deps.Routes,
		GetListSummaryPageData: deps.GetListSummaryPageData,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		TableLabels:            deps.TableLabels,
	}

	entityDeps := &linkpkg.ModuleDeps{
		Routes:       deps.Routes,
		Labels:       deps.Labels,
		CommonLabels: deps.CommonLabels,
		TableLabels:  deps.TableLabels,
		GetRatingDescriptionSetProductPlanListPageData: deps.GetRatingDescriptionSetProductPlanListPageData,
		RelinkRatingDescriptionSetProductPlan:          deps.RelinkRatingDescriptionSetProductPlan,
		UnlinkRatingDescriptionSetProductPlan:          deps.UnlinkRatingDescriptionSetProductPlan,
		ListRatingDescriptionSets:                      deps.ListRatingDescriptionSets,
		ListProductPlans:                               deps.ListProductPlans,
		ListPriceSchedules:                             deps.ListPriceSchedules,
		GetFormPageData:                                deps.GetFormPageData,
	}

	return &RatingDescriptionSetProductPlanModule{
		routes: deps.Routes,
		List:   linklist.NewView(listDeps),
		Relink: linkpkg.NewRelinkAction(entityDeps),
		Unlink: linkpkg.NewUnlinkAction(entityDeps),
	}
}

// RegisterRoutes registers all AY setup routes.
func (m *RatingDescriptionSetProductPlanModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)

	r.GET(m.routes.RelinkURL, m.Relink)
	r.POST(m.routes.RelinkURL, m.Relink)
	r.POST(m.routes.UnlinkURL, m.Unlink)
}
