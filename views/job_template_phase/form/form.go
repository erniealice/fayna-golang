// Package form holds the template-facing data shape for the job_template_phase drawer form.
// No Deps, no repo imports, no context.Context — pure types only.
package form

import (
	fayna "github.com/erniealice/fayna-golang"
	pyeza "github.com/erniealice/pyeza-golang"
)

// Data is the template-facing form data for the job-template-phase drawer (Add + Edit).
type Data struct {
	// FormAction is the POST URL for the drawer form.
	FormAction string
	WorkspaceID string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard

	// IsEdit distinguishes Add (false) vs Edit (true) in the template.
	IsEdit bool

	// ID is the job_template_phase PK — present on Edit, empty on Add.
	ID string

	// JobTemplateID is the context-locked FK to the parent job template.
	// Always present; hidden input emitted above sheet-body.
	JobTemplateID string

	// Phase fields
	Name       string
	PhaseOrder int32

	// Optional timing
	EstimatedDurationMinutes int32

	// Optional FK fields
	ResourceID   string
	ResourceName string

	// PredecessorTemplatePhaseID is the FK to another phase within the same template.
	// Only shown on Edit (guarded by .IsEdit in the template).
	PredecessorTemplatePhaseID string

	// ResourceSearchURL for the resource auto-complete (action mode).
	ResourceSearchURL string

	// Labels for the template.
	Labels fayna.JobTemplatePhaseLabels

	// CommonLabels for the sheet-form-footer.
	CommonLabels *pyeza.CommonLabels
}
