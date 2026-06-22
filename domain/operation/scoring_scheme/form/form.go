package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the scoring scheme drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent scoring_scheme package.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit      bool
	ID          string
	Name        string
	// Enum fields: round-tripped as proto enum string name (e.g. "SCORING_METHOD_WEIGHTED_AVERAGE")
	CompositeMethod     string
	RoundingMode        string
	WeightsMustSumToOne bool
	// Optional FK — empty string means unset
	ScoreScaleId string

	CompositeMethodOptions []pyeza.SelectOption
	RoundingModeOptions    []pyeza.SelectOption
	Labels                 any
	CommonLabels           any
}

// DefaultCompositeMethodOptions returns selectable ScoringMethod enum options.
// Values are proto enum string names for round-trip fidelity.
func DefaultCompositeMethodOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "SCORING_METHOD_EQUAL_WEIGHT", Label: "Equal Weight"},
		{Value: "SCORING_METHOD_WEIGHTED_AVERAGE", Label: "Weighted Average"},
		{Value: "SCORING_METHOD_MINIMUM_DETERMINATION", Label: "Minimum Determination"},
		{Value: "SCORING_METHOD_PERCENTAGE_PASS", Label: "Percentage Pass"},
		{Value: "SCORING_METHOD_SUM", Label: "Sum"},
	}
}

// DefaultRoundingModeOptions returns selectable RoundingMode enum options.
// Values are proto enum string names for round-trip fidelity.
func DefaultRoundingModeOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "ROUNDING_MODE_HALF_UP", Label: "Half Up"},
		{Value: "ROUNDING_MODE_HALF_DOWN", Label: "Half Down"},
		{Value: "ROUNDING_MODE_HALF_EVEN", Label: "Half Even (Banker's)"},
	}
}
