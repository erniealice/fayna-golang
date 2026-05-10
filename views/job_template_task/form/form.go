// Package form holds the template-facing data shape for the job_template_task drawer form.
// No Deps, no repo imports, no context.Context — pure types only.
package form

import (
	fayna "github.com/erniealice/fayna-golang"
	pyeza "github.com/erniealice/pyeza-golang"
)

// Data is the template-facing form data for the job-template-task drawer (Add + Edit).
type Data struct {
	// FormAction is the POST URL for the drawer form.
	FormAction string

	// IsEdit distinguishes Add (false) vs Edit (true) in the template.
	IsEdit bool

	// ID is the job_template_task PK — present on Edit, empty on Add.
	ID string

	// JobTemplatePhaseID is the context-locked FK to the parent phase.
	// Always present; hidden input emitted above sheet-body.
	JobTemplatePhaseID string

	// Task fields
	Name      string
	StepOrder int32

	// Optional timing
	EstimatedDurationMinutes int32

	// Optional FK fields
	ResourceID   string
	ResourceName string

	// ResourceSearchURL for the resource auto-complete (action mode).
	ResourceSearchURL string

	// Labels for the template.
	Labels fayna.JobTemplateTaskLabels

	// CommonLabels for the sheet-form-footer.
	CommonLabels *pyeza.CommonLabels
}
