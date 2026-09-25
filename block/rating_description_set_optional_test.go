package block

import (
	"context"
	"testing"

	"github.com/erniealice/espyna-golang/consumer/compose"
	ratingdescriptionsetpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set"
	ratingdescriptionsetentrypb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_entry"
	ratingdescriptionsetproductplanpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/rating_description_set_product_plan"
)

// codex-review-impl2 #8: on a provider without the rating-description-set
// capability (mock_db / firestore) espyna leaves the use cases nil; the three
// units must skip mounting (no route registration, no panic, no error).
func TestRatingDescriptionSetUnits_SkipMountWhenUseCasesNil(t *testing.T) {
	uc := &UseCases{}
	units := map[string]compose.Unit{
		"rating_description_set":              RatingDescriptionSetUnit(uc, nil),
		"rating_description_set_entry":        RatingDescriptionSetEntryUnit(uc, nil),
		"rating_description_set_product_plan": RatingDescriptionSetProductPlanUnit(uc, nil),
	}
	for name, u := range units {
		// An empty MountContext has no route registrar: reaching RegisterRoutes
		// would panic, so a nil return proves the unit skipped mounting.
		if err := u.Mount(&compose.MountContext{}); err != nil {
			t.Fatalf("%s: skip must not error, got %v", name, err)
		}
	}
}

func TestRatingDescriptionSetsSupported(t *testing.T) {
	uc := &UseCases{}
	if ratingDescriptionSetsSupported(uc) || ratingDescriptionSetsSupported(nil) {
		t.Fatal("nil use cases must report unsupported")
	}
	uc.Operation.RatingDescriptionSet.ListRatingDescriptionSets = func(context.Context, *ratingdescriptionsetpb.ListRatingDescriptionSetsRequest) (*ratingdescriptionsetpb.ListRatingDescriptionSetsResponse, error) {
		return nil, nil
	}
	if ratingDescriptionSetsSupported(uc) {
		t.Fatal("partial wiring must report unsupported (all-or-nothing)")
	}
	uc.Operation.RatingDescriptionSetEntry.ListRatingDescriptionSetEntries = func(context.Context, *ratingdescriptionsetentrypb.ListRatingDescriptionSetEntriesRequest) (*ratingdescriptionsetentrypb.ListRatingDescriptionSetEntriesResponse, error) {
		return nil, nil
	}
	uc.Operation.RatingDescriptionSetProductPlan.ListRatingDescriptionSetProductPlans = func(context.Context, *ratingdescriptionsetproductplanpb.ListRatingDescriptionSetProductPlansRequest) (*ratingdescriptionsetproductplanpb.ListRatingDescriptionSetProductPlansResponse, error) {
		return nil, nil
	}
	if !ratingDescriptionSetsSupported(uc) {
		t.Fatal("fully wired use cases must report supported")
	}
}
