package form

import (
	"context"
	"testing"

	scorescalepb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale"
	scorescalebandpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/score_scale_band"
	ttcrdpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria_rating_description"
)

func TestBuildRatingModeOptionsDefaultsLegacyStateToStandard(t *testing.T) {
	opts := BuildRatingModeOptions(RatingModeUnspecified, RatingModeLabels{
		Standard:               "Standard label",
		NumericWithDescription: "Numeric description label",
	})
	if len(opts) != 2 {
		t.Fatalf("option count = %d, want 2", len(opts))
	}
	if opts[0].Value != RatingModeStandard || opts[0].Label != "Standard label" || !opts[0].Selected {
		t.Fatalf("standard option = %+v, want selected standard", opts[0])
	}
	if opts[1].Value != RatingModeNumericWithDescription || opts[1].Label != "Numeric description label" || opts[1].Selected {
		t.Fatalf("numeric option = %+v, want unselected numeric mode", opts[1])
	}
}

func TestBuildRatingDescriptionRowsJoinsActiveBandsAndBindingText(t *testing.T) {
	min, max := 0.0, 4.0
	match := "5"
	ctx := context.Background()
	rows := BuildRatingDescriptionRows(
		ctx,
		func(context.Context, *scorescalebandpb.ListScoreScaleBandsRequest) (*scorescalebandpb.ListScoreScaleBandsResponse, error) {
			return &scorescalebandpb.ListScoreScaleBandsResponse{Data: []*scorescalebandpb.ScoreScaleBand{
				{Id: "band-b", ScoreScaleId: "scale-b", Active: true, OutputLabel: "Developing", InputMin: &min, InputMax: &max},
				{Id: "band-a", ScoreScaleId: "scale-a", Active: true, OutputLabel: "Excellent", InputMatch: &match},
				{Id: "inactive", ScoreScaleId: "scale-a", Active: false, OutputLabel: "Hidden"},
			}}, nil
		},
		func(context.Context, *ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaRequest) (*ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaResponse, error) {
			return &ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaResponse{TemplateTaskCriteriaRatingDescriptions: []*ttcrdpb.TemplateTaskCriteriaRatingDescription{
				{Id: "description-a", ScoreScaleBandId: "band-a", Description: "Outstanding performance", Active: true},
			}}, nil
		},
		"criteria-1",
		map[string]string{"scale-a": "Five point", "scale-b": "Range"},
	)

	if len(rows) != 2 {
		t.Fatalf("row count = %d, want 2 active bands", len(rows))
	}
	if rows[0].BandID != "band-a" || rows[0].ScaleName != "Five point" || rows[0].InputLabel != "= 5" || rows[0].Description != "Outstanding performance" || rows[0].ID != "description-a" {
		t.Fatalf("exact-match row = %+v", rows[0])
	}
	if rows[1].BandID != "band-b" || rows[1].InputLabel != "0 ≤ score < 4" || rows[1].Description != "" {
		t.Fatalf("range row = %+v", rows[1])
	}
}

func TestBuildScoreScaleOptionsFiltersInactiveAndMarksSelection(t *testing.T) {
	opts := BuildScoreScaleOptions(context.Background(), func(context.Context, *scorescalepb.ListScoreScalesRequest) (*scorescalepb.ListScoreScalesResponse, error) {
		return &scorescalepb.ListScoreScalesResponse{Data: []*scorescalepb.ScoreScale{
			{Id: "scale-z", Name: "Zed", Active: true},
			{Id: "scale-a", Name: "Alpha", Active: true, InputUnit: "%"},
			{Id: "scale-hidden", Name: "Hidden", Active: false},
		}}, nil
	}, "scale-a")

	if len(opts) != 2 || opts[0].Label != "Alpha (%)" || !opts[0].Selected || opts[1].Label != "Zed" || opts[1].Selected {
		t.Fatalf("scale options = %+v", opts)
	}
}
