package rating_description_set_entry

import (
	"context"

	entryform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry/form"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// ModuleDeps holds the typed closures the entry drawer actions need.
// Injected by the block (espyna use cases); this package never calls espyna.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// RatingDescriptionSetEntry — MVP: Create/Read/Update (Delete deferred, TODO).
	CreateRatingDescriptionSetEntry func(ctx context.Context, req *entrypb.CreateRatingDescriptionSetEntryRequest) (*entrypb.CreateRatingDescriptionSetEntryResponse, error)
	ReadRatingDescriptionSetEntry   func(ctx context.Context, req *entrypb.ReadRatingDescriptionSetEntryRequest) (*entrypb.ReadRatingDescriptionSetEntryResponse, error)
	UpdateRatingDescriptionSetEntry func(ctx context.Context, req *entrypb.UpdateRatingDescriptionSetEntryRequest) (*entrypb.UpdateRatingDescriptionSetEntryResponse, error)

	// Parent read — resolves the set's score_scale_id so the Level picker can
	// be scoped to the right band group.
	ReadRatingDescriptionSet func(ctx context.Context, req *setpb.ReadRatingDescriptionSetRequest) (*setpb.ReadRatingDescriptionSetResponse, error)

	// Pickers.
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)
	ListScoreScaleBands  func(ctx context.Context, req *scalebandpb.ListScoreScaleBandsRequest) (*scalebandpb.ListScoreScaleBandsResponse, error)

	// GetFormPageData is the entry drawer's criteria/band picker, sourced
	// from the Drawer page-data use case authorized under
	// rating_description_set:update (codex-review-impl4.out.md "Update-only
	// drawer") — preferred over ListOutcomeCriterias/ListScoreScaleBands
	// above. ratingDescriptionSetID scopes the Level picker (Add); entryID
	// pre-fills the entry (Edit) and, when ratingDescriptionSetID is empty,
	// resolves it from the loaded entry's own parent. Both parameters are
	// optional; the closure resolves whichever it can from what it is given.
	GetFormPageData func(ctx context.Context, ratingDescriptionSetID, entryID string) (*entryform.FormPageData, error)
}
