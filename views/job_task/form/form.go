// Package form holds the template-facing data shape for the job_task drawer form.
// No Deps, no repo imports, no context.Context — pure types only.
package form

import (
	fayna "github.com/erniealice/fayna-golang"
	pyeza "github.com/erniealice/pyeza-golang"
)

// Data is the template-facing form data for the job-task drawer (Add + Edit).
type Data struct {
	// FormAction is the POST URL for the drawer form.
	FormAction string

	// IsEdit distinguishes Add (false) vs Edit (true) in the template.
	IsEdit bool

	// ID is the job_task PK — present on Edit, empty on Add.
	ID string

	// JobPhaseID is the context-locked FK to the parent job phase.
	// Always present; hidden input emitted above sheet-body.
	JobPhaseID string

	// Task fields
	Name      string
	StepOrder int32
	Status    string
	IsAdHoc   bool

	// StatusOptions for the status select.
	StatusOptions []map[string]any

	// Assignment & Resource FKs
	AssignedTo       string
	AssignedToName   string
	ResourceID       string
	ResourceName     string
	TemplateTaskID   string
	TemplateTaskName string

	// Schedule
	PlannedQuantity   float64
	CompletedQuantity float64
	PercentComplete   float64
	AllowParallel     bool

	// Actuals (only shown on Edit)
	ActualStart string
	ActualEnd   string

	// Search endpoints for auto-complete pickers.
	StaffSearchURL        string
	ResourceSearchURL     string
	TemplateTaskSearchURL string

	// Labels for the template.
	Labels fayna.JobTaskLabels

	// CommonLabels for the sheet-form-footer.
	CommonLabels *pyeza.CommonLabels
}
