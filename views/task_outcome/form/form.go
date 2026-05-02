package form

import (
	fayna "github.com/erniealice/fayna-golang"
	pyeza "github.com/erniealice/pyeza-golang/types"
)

// Data is the template-facing data shape for the task outcome recording form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.TaskOutcomeLabels verbatim — templates
// read .Labels.Errors.*, .Labels.Form.*, .Labels.Columns.* directly.
type Data struct {
	FormAction      string
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
	Labels          fayna.TaskOutcomeLabels
	CommonLabels    any
}
