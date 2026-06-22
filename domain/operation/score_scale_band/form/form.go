package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the score scale band drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent score_scale_band package.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by ViewAdapter for action_workspace_guard
	IsEdit      bool
	ID          string

	ScoreScaleId  string
	SequenceOrder int32
	InputMin      *float64
	InputMax      *float64
	InputMatch    string
	OutputValue   *float64
	OutputLabel   string
	BandRole      string
	Determination string // proto enum string name e.g. "DETERMINATION_PASS"

	DeterminationOptions []pyeza.SelectOption

	Labels       any
	CommonLabels any
}

// DefaultDeterminationOptions returns the selectable Determination enum options.
// Values are proto enum string names for round-trip fidelity (Determination_value map lookup).
func DefaultDeterminationOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "DETERMINATION_UNSPECIFIED", Label: "Unspecified"},
		{Value: "DETERMINATION_PASS", Label: "Pass"},
		{Value: "DETERMINATION_FAIL", Label: "Fail"},
		{Value: "DETERMINATION_PASS_WITH_CONDITION", Label: "Pass with Condition"},
		{Value: "DETERMINATION_NOT_EVALUATED", Label: "Not Evaluated"},
		{Value: "DETERMINATION_NOT_APPLICABLE", Label: "Not Applicable"},
		{Value: "DETERMINATION_DEFERRED", Label: "Deferred"},
	}
}
