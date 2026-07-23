// Package form holds the template-facing data shape for the job_template_phase drawer form.
// No Deps, no repo imports, no context.Context — pure types only.
package form

import (
	"github.com/erniealice/fayna-golang/domain/operation/job_template_phase"
	"github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing form data for the job-template-phase drawer (Add + Edit).
type Data struct {
	// FormAction is the POST URL for the drawer form.
	FormAction  string
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

	// Code is the optional stable machine key for this phase (path segment in
	// document placeholders). Empty = unset. Operator-assigned in the drawer.
	Code string

	// Optional timing
	EstimatedDurationMinutes int32

	// Optional FK fields
	ResourceID   string
	ResourceName string

	// PredecessorTemplatePhaseID is the FK to another phase within the same template.
	// Only shown on Edit (guarded by .IsEdit in the template).
	PredecessorTemplatePhaseID string

	// ScoringSchemeID is the optional FK to the grading/scoring scheme this
	// phase's outcomes roll up under (education tier: "Grading Scheme" via
	// lyngua). ScoringSchemeOptions is pre-built by BuildScoringSchemeOptions
	// in options.go.
	ScoringSchemeID      string
	ScoringSchemeOptions []types.SelectOption

	// ResourceSearchURL for the resource auto-complete (action mode).
	ResourceSearchURL string

	// Labels for the template.
	Labels job_template_phase.Labels

	// CommonLabels for the sheet-form-footer. Typed `any` (NOT a concrete
	// pointer) to match the job_template form precedent: the app-side render
	// pipeline (pyeza render.(*Pipeline).InjectPageData) injects the app's
	// CommonLabels VALUE via reflection, and a mismatched concrete field type
	// panics reflect.Value.Set on every drawer GET (found live 2026-07-21,
	// AY-2627 Phase-3 canary grounding).
	CommonLabels any
}
