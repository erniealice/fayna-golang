// Package section_document owns generic render-profile contracts for
// subscription-group outcome documents. It contains no education vocabulary;
// apps map their trusted job-category codes to these canonical profiles.
package section_document

import bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"

const SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key = "subscription_group_outcome_matrix_single_period_11_v1"

// Profile describes the stable data-to-DOCX contract attached to a binding.
type Profile struct {
	Enum                  bindingpb.RenderProfile
	Key                   string
	JobTemplateSlots      int
	RequiresExactCategory bool
}

var subscriptionGroupOutcomeMatrixSinglePeriod11V1 = Profile{
	Enum:                  bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1,
	Key:                   SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key,
	JobTemplateSlots:      11,
	RequiresExactCategory: true,
}

// LookupProfile returns only registered, executable render profiles.
func LookupProfile(value bindingpb.RenderProfile) (Profile, bool) {
	switch value {
	case bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1:
		return subscriptionGroupOutcomeMatrixSinglePeriod11V1, true
	default:
		return Profile{}, false
	}
}

// LookupProfileKey is the inverse enum-to-manifest-key registry.
func LookupProfileKey(key string) (Profile, bool) {
	switch key {
	case SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key:
		return subscriptionGroupOutcomeMatrixSinglePeriod11V1, true
	default:
		return Profile{}, false
	}
}
