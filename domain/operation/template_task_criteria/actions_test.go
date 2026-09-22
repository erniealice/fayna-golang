package template_task_criteria

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	ttcrdpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/template_task_criteria_rating_description"
	ttcform "github.com/erniealice/fayna-golang/domain/operation/template_task_criteria/form"
)

func TestRatingConfigurationFromRequestKeepsScaleOnlyForNumericMode(t *testing.T) {
	standardRequest := httptest.NewRequest("POST", "/", strings.NewReader(url.Values{
		"rating_mode":     {ttcform.RatingModeStandard},
		"rating_scale_id": {"stale-scale"},
	}.Encode()))
	standardRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := standardRequest.ParseForm(); err != nil {
		t.Fatal(err)
	}
	mode, scale := ratingConfigurationFromRequest(standardRequest)
	if mode == nil || *mode != enums.RatingMode_RATING_MODE_STANDARD || scale != nil {
		t.Fatalf("standard configuration = mode %v, scale %v; want standard and nil scale", mode, scale)
	}

	numericRequest := httptest.NewRequest("POST", "/", strings.NewReader(url.Values{
		"rating_mode":     {ttcform.RatingModeNumericWithDescription},
		"rating_scale_id": {"scale-1"},
	}.Encode()))
	numericRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := numericRequest.ParseForm(); err != nil {
		t.Fatal(err)
	}
	mode, scale = ratingConfigurationFromRequest(numericRequest)
	if mode == nil || *mode != enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION || scale == nil || *scale != "scale-1" {
		t.Fatalf("numeric configuration = mode %v, scale %v; want numeric and scale-1", mode, scale)
	}
}

func TestSaveRatingDescriptionsUpsertsSelectedScaleAndRemovesStaleRows(t *testing.T) {
	var updated, created, deleted []string
	deps := &ModuleDeps{
		ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteria: func(context.Context, *ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaRequest) (*ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaResponse, error) {
			return &ttcrdpb.ListTemplateTaskCriteriaRatingDescriptionsByTemplateTaskCriteriaResponse{TemplateTaskCriteriaRatingDescriptions: []*ttcrdpb.TemplateTaskCriteriaRatingDescription{
				{Id: "description-1", ScoreScaleBandId: "band-1", Description: "old", Active: true},
				{Id: "description-stale", ScoreScaleBandId: "band-stale", Description: "remove", Active: true},
			}}, nil
		},
		CreateTemplateTaskCriteriaRatingDescription: func(_ context.Context, req *ttcrdpb.CreateTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.CreateTemplateTaskCriteriaRatingDescriptionResponse, error) {
			created = append(created, req.GetData().GetScoreScaleBandId()+":"+req.GetData().GetDescription())
			return &ttcrdpb.CreateTemplateTaskCriteriaRatingDescriptionResponse{Success: true}, nil
		},
		UpdateTemplateTaskCriteriaRatingDescription: func(_ context.Context, req *ttcrdpb.UpdateTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.UpdateTemplateTaskCriteriaRatingDescriptionResponse, error) {
			updated = append(updated, req.GetData().GetId()+":"+req.GetData().GetDescription())
			return &ttcrdpb.UpdateTemplateTaskCriteriaRatingDescriptionResponse{Success: true}, nil
		},
		DeleteTemplateTaskCriteriaRatingDescription: func(_ context.Context, req *ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionRequest) (*ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionResponse, error) {
			deleted = append(deleted, req.GetData().GetId())
			return &ttcrdpb.DeleteTemplateTaskCriteriaRatingDescriptionResponse{Success: true}, nil
		},
	}

	r := httptest.NewRequest("POST", "/", strings.NewReader(url.Values{
		"rating_description_band_id":  {"band-1", "band-2"},
		"rating_description_scale_id": {"scale-1", "scale-1"},
		"rating_description_text":     {"revised", "new wording"},
	}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	mode := enums.RatingMode_RATING_MODE_NUMERIC_WITH_DESCRIPTION
	scale := "scale-1"
	if err := saveRatingDescriptions(context.Background(), deps, "ttc-1", &mode, &scale, r); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	if len(updated) != 1 || updated[0] != "description-1:revised" {
		t.Fatalf("updated = %v, want description-1:revised", updated)
	}
	if len(created) != 1 || created[0] != "band-2:new wording" {
		t.Fatalf("created = %v, want band-2:new wording", created)
	}
	if len(deleted) != 1 || deleted[0] != "description-stale" {
		t.Fatalf("deleted = %v, want stale row only", deleted)
	}
}
