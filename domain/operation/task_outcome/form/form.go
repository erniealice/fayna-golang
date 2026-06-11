package form

import (
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the task outcome recording form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// Labels is typed any to avoid an import cycle between the form sub-package
// and the parent task_outcome package. Templates read .Labels.* via
// Go template reflection — no cast required.
type Data struct {
	FormAction      string
	WorkspaceID     string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit          bool
	ID              string
	TaskID          string
	CriteriaID      string
	CriteriaName    string
	CriteriaType    string
	CriteriaOptions []pyeza.SelectOption
	NumericValue    float64
	TextValue       string
	Notes           string
	PassFailValue   bool
	SelectedOption  string
	ScoreMin        float64
	ScoreMax        float64
	Labels          any
	CommonLabels    any
}
