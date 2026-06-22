package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the job outcome line drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent job_outcome_line package.
type Data struct {
	FormAction  string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit      bool
	ID          string

	// Core fields
	Label         string
	ReportingRole string // proto enum name string (e.g. "REPORTING_ROLE_PRIMARY")

	// Optional float fields
	WeightOrCredits float64
	OutputValue     float64

	// Optional string fields — *Set flags distinguish "empty" from "not set"
	OutputLabel         string
	OutputLabelSet      bool
	ScoreScaleBandId    string
	ScoreScaleBandIdSet bool

	ReportingRoleOptions []pyeza.SelectOption
	Labels               any
	CommonLabels         any
}

// DefaultReportingRoleOptions returns selectable options for the ReportingRole enum.
// Values are proto enum name strings for round-trip fidelity.
func DefaultReportingRoleOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "REPORTING_ROLE_UNSPECIFIED", Label: "Unspecified"},
		{Value: "REPORTING_ROLE_PRIMARY", Label: "Primary"},
		{Value: "REPORTING_ROLE_ALTERNATE", Label: "Alternate"},
		{Value: "REPORTING_ROLE_TRANSCRIPT", Label: "Transcript"},
		{Value: "REPORTING_ROLE_PERCENTILE", Label: "Percentile"},
	}
}
