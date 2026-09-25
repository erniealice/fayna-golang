package operation

import (
	"context"

	rdsetpkg "github.com/erniealice/fayna-golang/domain/operation/rating_description_set"
	rdsetdetail "github.com/erniealice/fayna-golang/domain/operation/rating_description_set/detail"
	rdsetform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set/form"
	rdsetlist "github.com/erniealice/fayna-golang/domain/operation/rating_description_set/list"
	rdsetlistdata "github.com/erniealice/fayna-golang/domain/operation/rating_description_set/listdata"
	rdentrypkg "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry"
	rdentryform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry/form"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	linkpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// RatingDescriptionSetModuleDeps holds all dependencies for the rating
// description set module (MVP scope: list + detail matrix + add + publish +
// deprecate; Edit/Delete/CreateVersion/tabs are deferred — TODO, see the
// plan sequence log).
type RatingDescriptionSetModuleDeps struct {
	Routes      rdsetpkg.Routes
	EntryRoutes rdentrypkg.Routes // resolved via compose.RoutesOf in the block Unit

	Labels       rdsetpkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateRatingDescriptionSet    func(ctx context.Context, req *setpb.CreateRatingDescriptionSetRequest) (*setpb.CreateRatingDescriptionSetResponse, error)
	ReadRatingDescriptionSet      func(ctx context.Context, req *setpb.ReadRatingDescriptionSetRequest) (*setpb.ReadRatingDescriptionSetResponse, error)
	ListRatingDescriptionSets     func(ctx context.Context, req *setpb.ListRatingDescriptionSetsRequest) (*setpb.ListRatingDescriptionSetsResponse, error)
	PublishRatingDescriptionSet   func(ctx context.Context, req *setpb.PublishRatingDescriptionSetRequest) (*setpb.PublishRatingDescriptionSetResponse, error)
	DeprecateRatingDescriptionSet func(ctx context.Context, req *setpb.DeprecateRatingDescriptionSetRequest) (*setpb.DeprecateRatingDescriptionSetResponse, error)

	ListRatingDescriptionSetEntries      func(ctx context.Context, req *entrypb.ListRatingDescriptionSetEntriesRequest) (*entrypb.ListRatingDescriptionSetEntriesResponse, error)
	ListRatingDescriptionSetProductPlans func(ctx context.Context, req *linkpb.ListRatingDescriptionSetProductPlansRequest) (*linkpb.ListRatingDescriptionSetProductPlansResponse, error)

	ListScoreScales      func(ctx context.Context, req *scalepb.ListScoreScalesRequest) (*scalepb.ListScoreScalesResponse, error)
	ListScoreScaleBands  func(ctx context.Context, req *scalebandpb.ListScoreScaleBandsRequest) (*scalebandpb.ListScoreScaleBandsResponse, error)
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// GetFormPageData/GetEntryFormPageData — page-data pickers authorized
	// under rating_description_set:create/:read instead of separate
	// score_scale:list / outcome_criteria:list / score_scale_band:list
	// grants (codex-review-impl2.out.md finding #6).
	GetFormPageData      func(ctx context.Context) (*rdsetform.FormPageData, error)
	GetEntryFormPageData func(ctx context.Context) (*rdentryform.FormPageData, error)
	// GetListSummaryPageData — the LIST page's SOLE data source, authorized
	// ONCE under rating_description_set:list (codex-review-impl3.out.md
	// finding #1).
	GetListSummaryPageData func(ctx context.Context) (*rdsetlistdata.PageData, error)
}

// RatingDescriptionSetModule holds all constructed rating description set views.
type RatingDescriptionSetModule struct {
	routes    rdsetpkg.Routes
	List      view.View
	Detail    view.View
	Add       view.View
	Publish   view.View
	Deprecate view.View
}

// NewRatingDescriptionSetModule creates a new rating description set module
// with all views wired.
func NewRatingDescriptionSetModule(deps *RatingDescriptionSetModuleDeps) *RatingDescriptionSetModule {
	listDeps := &rdsetlist.ListViewDeps{
		Routes:                 deps.Routes,
		GetListSummaryPageData: deps.GetListSummaryPageData,
		Labels:                 deps.Labels,
		CommonLabels:           deps.CommonLabels,
		TableLabels:            deps.TableLabels,
	}

	detailDeps := &rdsetdetail.DetailViewDeps{
		Routes:                          deps.Routes,
		EntryRoutes:                     deps.EntryRoutes,
		Labels:                          deps.Labels,
		CommonLabels:                    deps.CommonLabels,
		TableLabels:                     deps.TableLabels,
		ReadRatingDescriptionSet:        deps.ReadRatingDescriptionSet,
		ListRatingDescriptionSetEntries: deps.ListRatingDescriptionSetEntries,
		ListScoreScaleBands:             deps.ListScoreScaleBands,
		ListOutcomeCriterias:            deps.ListOutcomeCriterias,
		GetEntryFormPageData:            deps.GetEntryFormPageData,
	}

	entityDeps := &rdsetpkg.ModuleDeps{
		Routes:                               deps.Routes,
		Labels:                               deps.Labels,
		CommonLabels:                         deps.CommonLabels,
		TableLabels:                          deps.TableLabels,
		CreateRatingDescriptionSet:           deps.CreateRatingDescriptionSet,
		ReadRatingDescriptionSet:             deps.ReadRatingDescriptionSet,
		ListRatingDescriptionSets:            deps.ListRatingDescriptionSets,
		PublishRatingDescriptionSet:          deps.PublishRatingDescriptionSet,
		DeprecateRatingDescriptionSet:        deps.DeprecateRatingDescriptionSet,
		ListRatingDescriptionSetEntries:      deps.ListRatingDescriptionSetEntries,
		ListRatingDescriptionSetProductPlans: deps.ListRatingDescriptionSetProductPlans,
		ListScoreScales:                      deps.ListScoreScales,
		ListScoreScaleBands:                  deps.ListScoreScaleBands,
		ListOutcomeCriterias:                 deps.ListOutcomeCriterias,
		GetFormPageData:                      deps.GetFormPageData,
	}

	return &RatingDescriptionSetModule{
		routes:    deps.Routes,
		List:      rdsetlist.NewView(listDeps),
		Detail:    rdsetdetail.NewView(detailDeps),
		Add:       rdsetpkg.NewAddAction(entityDeps),
		Publish:   rdsetpkg.NewPublishAction(entityDeps),
		Deprecate: rdsetpkg.NewDeprecateAction(entityDeps),
	}
}

// RegisterRoutes registers all rating description set routes.
func (m *RatingDescriptionSetModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.ListURL, m.List)
	r.GET(m.routes.DetailURL, m.Detail)

	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.POST(m.routes.PublishURL, m.Publish)
	r.POST(m.routes.DeprecateURL, m.Deprecate)
}
