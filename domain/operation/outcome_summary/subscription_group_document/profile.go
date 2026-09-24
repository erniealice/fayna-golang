// Package subscription_group_document owns generic render-profile contracts for
// subscription-group outcome documents. It contains no education vocabulary;
// apps map their trusted job-category codes to these canonical profiles.
package subscription_group_document

import "strings"

import bindingpb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/subscription_group_document_template"

const SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key = "subscription_group_outcome_matrix_single_period_11_v1"
const SubscriptionGroupClientPhaseOutcomeReportV1Key = "subscription_group_client_phase_outcome_report_v1"

// GroupMatrixRenderProfile is the one group-matrix data source: one
// authorized subscription group × one job category × one period. Its stored
// "…_SINGLE_PERIOD_11_V1" name is kept only for compatibility with existing
// bindings; the column count is chosen by each uploaded template, never here.
const GroupMatrixRenderProfile = bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1

// MaxColumns bounds a group matrix; it matches the DOCX engine's table
// column-loop limit (Word's own 63-column table limit).
const MaxColumns = 63

// CategoryBindingScope describes which job-category binding identity a
// template profile accepts. This is part of the profile contract rather than
// an app-specific switch, so settings and render paths can validate scope in
// the same way as more profiles are added.
type CategoryBindingScope uint8

const (
	CategoryBindingScopeUnspecified CategoryBindingScope = iota
	CategoryBindingScopeExactCategory
	CategoryBindingScopeAllCategories
)

// Profile describes the stable data-to-DOCX contract attached to a binding.
type Profile struct {
	Enum                 bindingpb.RenderProfile
	Key                  string
	CategoryBindingScope CategoryBindingScope
}

var subscriptionGroupOutcomeMatrixSinglePeriod11V1 = Profile{
	Enum:                 bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1,
	Key:                  SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key,
	CategoryBindingScope: CategoryBindingScopeExactCategory,
}

var subscriptionGroupClientPhaseOutcomeReportV1 = Profile{
	Enum:                 bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1,
	Key:                  SubscriptionGroupClientPhaseOutcomeReportV1Key,
	CategoryBindingScope: CategoryBindingScopeAllCategories,
}

// AcceptsCategoryBinding reports whether categoryID matches this profile's
// declared binding scope. An empty category ID is the all-categories scope.
func (p Profile) AcceptsCategoryBinding(categoryID string) bool {
	hasCategory := strings.TrimSpace(categoryID) != ""
	switch p.CategoryBindingScope {
	case CategoryBindingScopeExactCategory:
		return hasCategory
	case CategoryBindingScopeAllCategories:
		return !hasCategory
	default:
		return false
	}
}

// LookupProfile returns only registered, executable render profiles.
func LookupProfile(value bindingpb.RenderProfile) (Profile, bool) {
	switch value {
	case bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_OUTCOME_MATRIX_SINGLE_PERIOD_11_V1:
		return subscriptionGroupOutcomeMatrixSinglePeriod11V1, true
	case bindingpb.RenderProfile_RENDER_PROFILE_SUBSCRIPTION_GROUP_CLIENT_PHASE_OUTCOME_REPORT_V1:
		return subscriptionGroupClientPhaseOutcomeReportV1, true
	default:
		return Profile{}, false
	}
}

// LookupProfileKey is the inverse enum-to-manifest-key registry.
func LookupProfileKey(key string) (Profile, bool) {
	switch key {
	case SubscriptionGroupOutcomeMatrixSinglePeriod11V1Key:
		return subscriptionGroupOutcomeMatrixSinglePeriod11V1, true
	case SubscriptionGroupClientPhaseOutcomeReportV1Key:
		return subscriptionGroupClientPhaseOutcomeReportV1, true
	default:
		return Profile{}, false
	}
}
