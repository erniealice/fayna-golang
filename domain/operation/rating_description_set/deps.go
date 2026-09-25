package rating_description_set

import (
	"context"

	setform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set/form"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	linkpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
	scalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// ModuleDeps holds the typed closures that action builders and sub-packages
// need. Injected by the block (espyna use cases); this package never calls
// espyna or SQL directly.
type ModuleDeps struct {
	Routes       Routes
	Labels       Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	// RatingDescriptionSet CRUD + lifecycle (MVP: Create/Read/List + Publish/Deprecate;
	// Update/Delete/CreateVersion are deferred — TODO, see sequence log).
	CreateRatingDescriptionSet    func(ctx context.Context, req *setpb.CreateRatingDescriptionSetRequest) (*setpb.CreateRatingDescriptionSetResponse, error)
	ReadRatingDescriptionSet      func(ctx context.Context, req *setpb.ReadRatingDescriptionSetRequest) (*setpb.ReadRatingDescriptionSetResponse, error)
	ListRatingDescriptionSets     func(ctx context.Context, req *setpb.ListRatingDescriptionSetsRequest) (*setpb.ListRatingDescriptionSetsResponse, error)
	PublishRatingDescriptionSet   func(ctx context.Context, req *setpb.PublishRatingDescriptionSetRequest) (*setpb.PublishRatingDescriptionSetResponse, error)
	DeprecateRatingDescriptionSet func(ctx context.Context, req *setpb.DeprecateRatingDescriptionSetRequest) (*setpb.DeprecateRatingDescriptionSetResponse, error)

	// Cross-entity reads for the list (entries/links counts) and the
	// descriptors matrix tab. Bare lists, filtered client-side — the same
	// "fetch all, small dataset" pattern score_scale and template_task_criteria
	// already use in this package family.
	ListRatingDescriptionSetEntries      func(ctx context.Context, req *entrypb.ListRatingDescriptionSetEntriesRequest) (*entrypb.ListRatingDescriptionSetEntriesResponse, error)
	ListRatingDescriptionSetProductPlans func(ctx context.Context, req *linkpb.ListRatingDescriptionSetProductPlansRequest) (*linkpb.ListRatingDescriptionSetProductPlansResponse, error)

	// Pickers — optional/nil-safe (falls back to raw-id text inputs per the
	// established fayna convention, e.g. template_task_criteria's job_template
	// pickers).
	ListScoreScales      func(ctx context.Context, req *scalepb.ListScoreScalesRequest) (*scalepb.ListScoreScalesResponse, error)
	ListScoreScaleBands  func(ctx context.Context, req *scalebandpb.ListScoreScaleBandsRequest) (*scalebandpb.ListScoreScaleBandsResponse, error)
	ListOutcomeCriterias func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// GetFormPageData is the Add-set drawer's score_scale picker, sourced
	// from a page-data use case authorized under rating_description_set:create
	// (finding #6) — preferred over ListScoreScales above, which NewAddAction
	// no longer calls once this is wired.
	GetFormPageData func(ctx context.Context) (*setform.FormPageData, error)
}
