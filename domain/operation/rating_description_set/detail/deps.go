package detail

import (
	"context"

	"github.com/erniealice/fayna-golang/domain/operation/rating_description_set"
	rdentry "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry"
	rdentryform "github.com/erniealice/fayna-golang/domain/operation/rating_description_set_entry/form"

	pyeza "github.com/erniealice/pyeza-golang"
	"github.com/erniealice/pyeza-golang/types"

	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
	setpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	entrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	scalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
)

// DetailViewDeps holds view dependencies for the rating description set
// detail view. MVP scope renders one view (the Descriptors matrix, read-only
// text + entry-drawer links) — no separate tabs yet (TODO: assignments,
// versions, info tabs per interfaces.md §8).
type DetailViewDeps struct {
	Routes      rating_description_set.Routes
	EntryRoutes rdentry.Routes // for the matrix's Add/Edit entry links

	Labels       rating_description_set.Labels
	CommonLabels pyeza.CommonLabels
	TableLabels  types.TableLabels

	ReadRatingDescriptionSet       func(ctx context.Context, req *setpb.ReadRatingDescriptionSetRequest) (*setpb.ReadRatingDescriptionSetResponse, error)
	ListRatingDescriptionSetEntries func(ctx context.Context, req *entrypb.ListRatingDescriptionSetEntriesRequest) (*entrypb.ListRatingDescriptionSetEntriesResponse, error)
	ListScoreScaleBands             func(ctx context.Context, req *scalebandpb.ListScoreScaleBandsRequest) (*scalebandpb.ListScoreScaleBandsResponse, error)
	ListOutcomeCriterias            func(ctx context.Context, req *criteriapb.ListOutcomeCriteriasRequest) (*criteriapb.ListOutcomeCriteriasResponse, error)

	// GetEntryFormPageData is the matrix's band-row + criterion-name source,
	// sourced from a page-data use case authorized under
	// rating_description_set:read (finding #6 — live-confirmed: "the
	// rating_description_set detail matrix's band ROWS never render") —
	// preferred over ListScoreScaleBands/ListOutcomeCriterias above. Reuses
	// rating_description_set_entry's page data (same criteria/bands, shared
	// by both the entry drawer and this matrix).
	GetEntryFormPageData func(ctx context.Context) (*rdentryform.FormPageData, error)
}
