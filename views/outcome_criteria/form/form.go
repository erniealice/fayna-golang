package form

import (
	fayna "github.com/erniealice/fayna-golang"
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the outcome criteria drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.OutcomeCriteriaLabels verbatim — templates
// read .Labels.Errors.*, .Labels.Form.*, .Labels.Columns.* directly.
type Data struct {
	FormAction   string
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
	Labels       fayna.OutcomeCriteriaLabels
	CommonLabels any
}

// DefaultTypeOptions returns the selectable criteria type options.
// The labels parameter is accepted for future i18n but currently unused —
// options carry English display strings directly.
func DefaultTypeOptions(_ fayna.OutcomeCriteriaLabels) []pyeza.SelectOption {
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
func DefaultScopeOptions(_ fayna.OutcomeCriteriaLabels) []pyeza.SelectOption {
	return []pyeza.SelectOption{
		{Value: "TASK", Label: "Task"},
		{Value: "PHASE", Label: "Phase"},
		{Value: "JOB", Label: "Job"},
	}
}
