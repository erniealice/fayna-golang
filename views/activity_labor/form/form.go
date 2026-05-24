// Package form holds the template-facing data shape for the activity labor drawer form.
// No Deps, no repo imports, no context.Context parameters (drawer-form-subpackage-convention.md).
package form

import fayna "github.com/erniealice/fayna-golang"

// Data is the template-facing data shape for the activity labor drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// ActivityID is the PK of ActivityLabor and the FK to JobActivity (1:1).
// It is context-locked on Edit — emitted as a hidden input above sheet-body.
// On Add it is sourced from the parent JobActivity's ?activity_id query param.
//
// TimeStart / TimeEnd are stored as int64 Unix timestamps in the proto.
// The proto also carries optional *_string companions (human-readable).
// The form uses datetime-local inputs whose values are formatted as
// "2006-01-02T15:04" (HTML datetime-local format); the handler converts
// to Unix int64 on submit.
type Data struct {
	FormAction string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit     bool

	// ActivityID is both the PK of ActivityLabor and the FK to JobActivity.
	// Context-locked on Edit; sourced from ?activity_id on Add.
	ActivityID string

	// StaffID is the FK to staff. StaffName is the display label for the
	// pre-selected staff auto-complete option.
	StaffID   string
	StaffName string

	// Hours — decimal hours worked (e.g. 1.5 = 90 minutes).
	Hours float64

	// RateType — proto enum string: "RATE_TYPE_STANDARD", "RATE_TYPE_OVERTIME",
	// "RATE_TYPE_PREMIUM", or "" (unspecified).
	RateType string

	// RateTypeOptions — pre-built select options for the rate type picker.
	// Built by form.BuildRateTypeOptions in options.go.
	RateTypeOptions []Option

	// TimeStart / TimeEnd — formatted as "2006-01-02T15:04" for datetime-local inputs.
	// Empty string = no value set.
	TimeStart string
	TimeEnd   string

	// StaffSearchURL — action-mode auto-complete endpoint (returns [{value,label}] JSON).
	// Empty = staff search not available; the auto-complete falls back to filter mode.
	StaffSearchURL string

	Labels       fayna.ActivityLaborLabels
	CommonLabels any
}

// Option is a single select option used in RateTypeOptions and other selects.
type Option struct {
	Value    string
	Label    string
	Selected bool
}
