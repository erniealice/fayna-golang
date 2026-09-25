package operation

import (
	"context"

	entrypkg "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry"
	entryform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry/form"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"
	"github.com/erniealice/pyeza-golang/view"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// RatingDescriptionSetEntryModuleDeps holds dependencies for the entry
// drawer module. No standalone list/detail — it surfaces via the parent
// rating_description_set detail's Descriptors matrix.
type RatingDescriptionSetEntryModuleDeps struct {
	Routes       entrypkg.Routes
	Labels       entrypkg.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	CreateRatingDescriptionSetEntry func(ctx context.Context, req *entrypb.CreateRatingDescriptionSetEntryRequest) (*entrypb.CreateRatingDescriptionSetEntryResponse, error)
	ReadRatingDescriptionSetEntry   func(ctx context.Context, req *entrypb.ReadRatingDescriptionSetEntryRequest) (*entrypb.ReadRatingDescriptionSetEntryResponse, error)
	UpdateRatingDescriptionSetEntry func(ctx context.Context, req *entrypb.UpdateRatingDescriptionSetEntryRequest) (*entrypb.UpdateRatingDescriptionSetEntryResponse, error)

	ReadRatingDescriptionSet func(ctx context.Context, req *setpb.ReadRatingDescriptionSetRequest) (*setpb.ReadRatingDescriptionSetResponse, error)

	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
	ListScoreScaleBands  func(ctx context.Context, req *scalebandpb.ListScoreScaleBandsRequest) (*scalebandpb.ListScoreScaleBandsResponse, error)

	// GetFormPageData — Drawer page-data picker, authorized under
	// rating_description_set:update (codex-review-impl4.out.md "Update-only
	// drawer"), also resolving the Edit entry and Level-picker scale
	// server-side.
	GetFormPageData func(ctx context.Context, ratingDescriptionSetID, entryID string) (*entryform.FormPageData, error)
}

// RatingDescriptionSetEntryModule holds the constructed entry drawer views.
type RatingDescriptionSetEntryModule struct {
	routes entrypkg.Routes
	Add    view.View
	Edit   view.View
}

// NewRatingDescriptionSetEntryModule constructs the entry drawer module.
func NewRatingDescriptionSetEntryModule(deps *RatingDescriptionSetEntryModuleDeps) *RatingDescriptionSetEntryModule {
	entityDeps := &entrypkg.ModuleDeps{
		Routes:                          deps.Routes,
		Labels:                          deps.Labels,
		CommonLabels:                    deps.CommonLabels,
		TableLabels:                     deps.TableLabels,
		CreateRatingDescriptionSetEntry: deps.CreateRatingDescriptionSetEntry,
		ReadRatingDescriptionSetEntry:   deps.ReadRatingDescriptionSetEntry,
		UpdateRatingDescriptionSetEntry: deps.UpdateRatingDescriptionSetEntry,
		ReadRatingDescriptionSet:        deps.ReadRatingDescriptionSet,
		ListOutcomeCriterias:            deps.ListOutcomeCriterias,
		ListScoreScaleBands:             deps.ListScoreScaleBands,
		GetFormPageData:                 deps.GetFormPageData,
	}

	return &RatingDescriptionSetEntryModule{
		routes: deps.Routes,
		Add:    entrypkg.NewAddAction(entityDeps),
		Edit:   entrypkg.NewEditAction(entityDeps),
	}
}

// RegisterRoutes registers the entry drawer routes.
func (m *RatingDescriptionSetEntryModule) RegisterRoutes(r view.RouteRegistrar) {
	r.GET(m.routes.AddURL, m.Add)
	r.POST(m.routes.AddURL, m.Add)
	r.GET(m.routes.EditURL, m.Edit)
	r.POST(m.routes.EditURL, m.Edit)
}
