package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"

	enums "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/enums"
	criteriapb "github.com/erniealice/esqyma/pkg/schema/v1/domain/operation/outcome_criteria"
)

// Data is the template-facing data shape for the rubric-item drawer
// (evaluation_template_item-drawer-form). Used by Add (FormAction = AddURL) and
// Edit (FormAction = EditURL, IsEdit = true).
//
// Labels is typed any to avoid an import cycle with the parent package.
type Data struct {
	FormAction         string
	WorkspaceID        string // injected by ViewAdapter for the action_workspace_guard
	IsEdit             bool
	ID                 string
	EvaluationTemplateID string // hidden — parent FK; H1 workspace_id parity is copied server-side
	OutcomeCriteriaID  string
	SequenceOrder      int32
	WeightOverride     string // string form so "" round-trips as nil (optional float)
	QuestionLabel      string
	QuestionPrompt     string
	RequiredOverride   bool

	// Criteria autocomplete options (scope=CRITERIA_SCOPE_EVALUATION); the
	// option carries the criterion's criteria_type so the drawer can surface it
	// read-only.
	CriteriaOptions []pyeza.SelectOption
	// CriteriaTypeLabel — read-only display of the chosen criterion's response
	// type (set on edit).
	CriteriaTypeLabel string

	Labels       any
	CommonLabels any
}

// BuildCriteriaOptions filters the criteria list to scope=EVALUATION and maps
// to select options. Pure function (companions: form/options-style helper) —
// the caller fetches the list via ListOutcomeCriterias in the action layer.
func BuildCriteriaOptions(items []*criteriapb.OutcomeCriteria) []pyeza.SelectOption {
	opts := []pyeza.SelectOption{}
	for _, c := range items {
		if c.GetScope() != enums.CriteriaScope_CRITERIA_SCOPE_EVALUATION {
			continue
		}
		opts = append(opts, pyeza.SelectOption{
			Value: c.GetId(),
			Label: c.GetName(),
		})
	}
	return opts
}

// CriteriaTypeDisplay returns a human label for a criterion's response type.
func CriteriaTypeDisplay(t enums.CriteriaType) string {
	switch t {
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_RANGE:
		return "Numeric Range"
	case enums.CriteriaType_CRITERIA_TYPE_NUMERIC_SCORE:
		return "Numeric Score"
	case enums.CriteriaType_CRITERIA_TYPE_PASS_FAIL:
		return "Pass/Fail"
	case enums.CriteriaType_CRITERIA_TYPE_CATEGORICAL:
		return "Categorical"
	case enums.CriteriaType_CRITERIA_TYPE_TEXT:
		return "Text"
	case enums.CriteriaType_CRITERIA_TYPE_MULTI_CHECK:
		return "Multi-Check"
	default:
		return "Unspecified"
	}
}
