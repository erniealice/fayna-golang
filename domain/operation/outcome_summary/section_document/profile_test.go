package section_document

import (
	"testing"

	bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"
)

func TestRenderProfile_ManifestKeyMapping(t *testing.T) {
	want := bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1
	profile, ok := LookupProfile(want)
	if !ok || profile.Enum != want || profile.Key != SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key || profile.JobTemplateSlots != 11 || !profile.RequiresExactCategory {
		t.Fatalf("profile = %+v, %v", profile, ok)
	}
	byKey, ok := LookupProfileKey(profile.Key)
	if !ok || byKey != profile {
		t.Fatalf("inverse profile = %+v, %v", byKey, ok)
	}
	for _, unsupported := range []bindingpb.RenderProfile{
		bindingpb.RenderProfile_RENDER_PROFILE_UNSPECIFIED,
		bindingpb.RenderProfile(99),
	} {
		if got, ok := LookupProfile(unsupported); ok || got != (Profile{}) {
			t.Fatalf("unsupported %v resolved to %+v", unsupported, got)
		}
	}
	if _, ok := LookupProfileKey("academic"); ok {
		t.Fatal("vertical category vocabulary must not resolve as a render profile")
	}
}
