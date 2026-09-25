package action

import (
	"testing"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

func TestRatingDescriptions_Describe(t *testing.T) {
	lo, mid, hi := 0.0, 3.0, 6.0
	match := "8"
	descs := []*matrixpb.RatingDescription{
		{ScaleKind: enums.ScaleKind_SCALE_KIND_RANGE_MAP, InputMin: &lo, InputMax: &mid, Description: "low"},
		{ScaleKind: enums.ScaleKind_SCALE_KIND_RANGE_MAP, InputMin: &mid, InputMax: &hi, Description: "mid"},
		{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: &match, Description: "top"},
		{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: &match, Description: "   "}, // blank wording is dropped
		{ScaleKind: enums.ScaleKind_SCALE_KIND_UNSPECIFIED, Description: "ignored"},
	}
	r := ratingDescriptionsFromList(descs)
	for v, want := range map[float64]string{0: "low", 2.9: "low", 3: "mid", 5.99: "mid", 6: "", 8: "top", 8.0000000001: "top", -1: ""} {
		if got := r.describe(v); got != want {
			t.Errorf("describe(%v) = %q, want %q", v, got, want)
		}
	}
	if len(r) != 3 {
		t.Errorf("expected 3 usable matchers, got %d", len(r))
	}
}

func TestRatingDescriptions_EmptyIsInert(t *testing.T) {
	if r := ratingDescriptionsFromList(nil); r.describe(4) != "" {
		t.Fatal("nil descriptions must be inert")
	}
	if r := ratingDescriptionsFromList([]*matrixpb.RatingDescription{nil}); r.describe(4) != "" {
		t.Fatal("a nil entry must be skipped, not panic")
	}
}

// ratingResolutionFrom is the status → write-path mapping (schema-proposal
// §4/§5, interfaces.md §1 RatingDescriptionResolutionStatus). RESOLVED and
// NO_LINK both save (Q21); everything else rejects the cell (Q22).
func TestRatingResolutionFrom_StatusMapping(t *testing.T) {
	match := "4"
	resolved := &matrixpb.CellRatingResolution{
		Status: enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED,
		Descriptions: []*matrixpb.RatingDescription{
			{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: &match, Description: "four"},
		},
	}
	if got := ratingResolutionFrom(resolved); got.reject || got.ratings.describe(4) != "four" {
		t.Fatalf("RESOLVED must carry usable ratings, got %+v", got)
	}

	resolvedNoMatch := &matrixpb.CellRatingResolution{Status: enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_RESOLVED}
	if got := ratingResolutionFrom(resolvedNoMatch); got.reject || got.ratings.describe(4) != "" {
		t.Fatalf("RESOLVED with no entries must be usable but describe \"\" (no entry), got %+v", got)
	}

	noLink := &matrixpb.CellRatingResolution{Status: enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_NO_LINK}
	if got := ratingResolutionFrom(noLink); got.reject || got.ratings.describe(4) != "" {
		t.Fatalf("NO_LINK must behave like no entry (Q21), got %+v", got)
	}

	for _, tc := range []struct {
		name   string
		status enums.RatingDescriptionResolutionStatus
		reason string
	}{
		{"unresolved identity", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_UNRESOLVED_IDENTITY, "rating_description_unresolved_identity"},
		{"ambiguous", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_AMBIGUOUS, "rating_description_ambiguous"},
		{"invalid config", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_INVALID_CONFIG, "rating_description_invalid_config"},
		{"placeholder unresolved", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_PLACEHOLDER_UNRESOLVED, "rating_description_placeholder_unresolved"},
		{"unspecified/unknown status", enums.RatingDescriptionResolutionStatus_RATING_DESCRIPTION_RESOLUTION_STATUS_UNSPECIFIED, "rating_description_resolution_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ratingResolutionFrom(&matrixpb.CellRatingResolution{Status: tc.status, Reason: "diagnostic"})
			if !got.reject || got.reason != tc.reason {
				t.Errorf("want reject=true reason=%q, got reject=%v reason=%q", tc.reason, got.reject, got.reason)
			}
		})
	}
}

func strp(s string) *string { return &s }
