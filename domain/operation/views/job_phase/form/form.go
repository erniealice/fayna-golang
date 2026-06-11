// Package form holds the template-facing data shape for the job_phase drawer form.
// No Deps, no repo imports, no context.Context — pure types only.
package form

import (
	operation "github.com/erniealice/fayna-golang/domain/operation"
	pyeza "github.com/erniealice/pyeza-golang"
)

// Data is the template-facing form data for the job-phase drawer (Add + Edit).
type Data struct {
	// FormAction is the POST URL for the drawer form.
	FormAction string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard

	// IsEdit distinguishes Add (false) vs Edit (true) in the template.
	IsEdit bool

	// ID is the job_phase PK — present on Edit, empty on Add.
	ID string

	// JobID is the context-locked FK to the parent job.
	// Always present; hidden input emitted above sheet-body.
	JobID string

	// Phase fields
	Name       string
	PhaseOrder int32
	Status     string

	// StatusOptions for the status select.
	StatusOptions []map[string]any

	// Optional FKs
	TemplatePhaseID    string
	ResourceID         string
	ResourceName       string
	PredecessorPhaseID string

	// Timing
	PlannedStart string
	PlannedEnd   string
	ActualStart  string
	ActualEnd    string

	// Resource & timing scalars
	SetupMinutes      int32
	RunMinutesPerUnit float64

	// ResourceSearchURL for the resource auto-complete (action mode).
	ResourceSearchURL string

	// Labels for the template.
	Labels operation.JobPhaseLabels

	// CommonLabels for the sheet-form-footer.
	CommonLabels *pyeza.CommonLabels
}
