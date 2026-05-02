package form

import fayna "github.com/erniealice/fayna-golang"

// Data is the template-facing data shape for the job drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.JobLabels verbatim — templates read
// .Labels.Columns.Name, .Labels.Form.NamePlaceholder, etc. directly.
type Data struct {
	FormAction   string
	IsEdit       bool
	ID           string
	Name         string
	ClientID     string
	LocationID   string
	Labels       fayna.JobLabels
	CommonLabels any
}
