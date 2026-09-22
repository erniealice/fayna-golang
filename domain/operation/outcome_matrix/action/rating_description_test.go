package action

import (
	"testing"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	matrixpb "github.com/erniealice/esqyma/pkg/schema/v1/service/operation/outcome_matrix"
)

func TestRatingDescriptions_Describe(t *testing.T) {
	lo, mid, hi := 0.0, 3.0, 6.0
	match := "8"
	col := &matrixpb.CriterionColumn{
		RatingMode: enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION,
		RatingDescriptions: []*matrixpb.RatingDescription{
			{ScaleKind: enums.ScaleKind_SCALE_KIND_RANGE_MAP, InputMin: &lo, InputMax: &mid, Description: "low"},
			{ScaleKind: enums.ScaleKind_SCALE_KIND_RANGE_MAP, InputMin: &mid, InputMax: &hi, Description: "mid"},
			{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: &match, Description: "top"},
			{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: &match, Description: "   "}, // blank wording is dropped
			{ScaleKind: enums.ScaleKind_SCALE_KIND_UNSPECIFIED, Description: "ignored"},
		},
	}
	r := ratingDescriptionsFromColumn(col)
	for v, want := range map[float64]string{0: "low", 2.9: "low", 3: "mid", 5.99: "mid", 6: "", 8: "top", 8.0000000001: "top", -1: ""} {
		if got := r.describe(v); got != want {
			t.Errorf("describe(%v) = %q, want %q", v, got, want)
		}
	}
	if len(r) != 3 {
		t.Errorf("expected 3 usable matchers, got %d", len(r))
	}
}

func TestRatingDescriptions_OffUnlessModeSet(t *testing.T) {
	col := &matrixpb.CriterionColumn{RatingDescriptions: []*matrixpb.RatingDescription{
		{ScaleKind: enums.ScaleKind_SCALE_KIND_EXACT_MAP, InputMatch: strp("4"), Description: "x"},
	}}
	if r := ratingDescriptionsFromColumn(col); r.enabled() || r.describe(4) != "" {
		t.Fatal("descriptions must be inert unless the binding is in numeric-with-description mode")
	}
	if r := ratingDescriptionsFromColumn(nil); r.enabled() {
		t.Fatal("nil column must be inert")
	}
}

func strp(s string) *string { return &s }
