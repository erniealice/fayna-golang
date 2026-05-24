package form

import fayna "github.com/erniealice/fayna-golang"

// Data is the template-facing data shape for the job template drawer form.
// Used by both Add (FormAction = AddURL, IsEdit = false) and
// Edit (FormAction = EditURL, IsEdit = true) handlers.
//
// No mapper: the Labels field is fayna.JobTemplateLabels verbatim — templates
// read .Labels.Columns.Name, .Labels.Form.NamePlaceholder, etc. directly.
type Data struct {
	FormAction   string
	WorkspaceID   string // injected by C1: populated by ViewAdapter.injectWorkspaceID for action_workspace_guard
	IsEdit       bool
	ID           string
	Name         string
	Description  string
	Active       bool
	Labels       fayna.JobTemplateLabels
	CommonLabels any
}
