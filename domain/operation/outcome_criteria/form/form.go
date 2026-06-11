package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the outcome criteria drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent outcome_criteria package. Templates read .Labels.* via
// Go template reflection — no cast required.
type Data struct {
	FormAction   string
	WorkspaceID  string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit       bool
	ID           string
	Name         string
	Type         string
	Scope        string
	Description  string
	Required     bool
	Weight       float64
	TypeOptions  []pyeza.SelectOption
	ScopeOptions []pyeza.SelectOption
	Labels       any
	CommonLabels any
}

// DefaultTypeOptions returns the selectable criteria type options.
// The labels parameter is accepted for future i18n but currently unused —
// options carry English display strings directly.
func DefaultTypeOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "NUMERIC_RANGE", Label: "Numeric Range"},
		{Value: "NUMERIC_SCORE", Label: "Numeric Score"},
		{Value: "PASS_FAIL", Label: "Pass / Fail"},
		{Value: "CATEGORICAL", Label: "Categorical"},
		{Value: "TEXT", Label: "Text"},
		{Value: "MULTI_CHECK", Label: "Multi-check"},
	}
}

// DefaultScopeOptions returns the selectable criteria scope options.
// The labels parameter is accepted for future i18n but currently unused.
func DefaultScopeOptions() []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "TASK", Label: "Task"},
		{Value: "PHASE", Label: "Phase"},
		{Value: "JOB", Label: "Job"},
	}
}
